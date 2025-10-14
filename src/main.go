package main

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

type Config struct {
	ParamName     string   `json:"paramName"`
	AllowedValues []string `json:"allowedValues"`
	// DenyStatus accepts either a number or a quoted string ("401") in JSON.
	DenyStatus    HTTPStatus `json:"denyStatus,omitempty"` // default 401
	// DevMode enables extra logging and decision points for debugging in clusters.
	DevMode       DevFlag    `json:"devMode,omitempty"`
}

// HTTPStatus accepts either a JSON number (401) or JSON string ("401").
type HTTPStatus int

func (s *HTTPStatus) UnmarshalJSON(b []byte) error {
	// try number first
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*s = HTTPStatus(n)
		return nil
	}
	// try string
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	n, err := strconv.Atoi(str)
	if err != nil {
		return err
	}
	*s = HTTPStatus(n)
	return nil
}

// DevFlag accepts only a JSON boolean true to enable dev mode. Any other value
// (string, number, null, etc.) is treated as false and does not return an error.
type DevFlag bool

func (d *DevFlag) UnmarshalJSON(b []byte) error {
	var bv bool
	if err := json.Unmarshal(b, &bv); err == nil {
		*d = DevFlag(bv)
		return nil
	}
	// Some hosts (e.g. Traefik) may serialize booleans as strings ("true").
	// Accept the string value "true" (case-insensitive) as enabling devMode.
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if strings.EqualFold(s, "true") {
			*d = DevFlag(true)
			return nil
		}
	}

	// For any other non-boolean value, default to false and do not error.
	*d = DevFlag(false)
	return nil
}

type TokenMiddleware struct {
	param string
	allow map[string]struct{}
	deny  int
	dev   bool
}

// TinyGo/WASI requires a main entrypoint even if unused by http-wasm.
func main() {}

// runtime initialization (moved to TinyGo-only file)

func (m *TokenMiddleware) handleRequest(req api.Request, resp api.Response) (next bool, reqCtx uint32) {
	uri := req.GetURI() // e.g. "/path?Token=abc"
	ok, reason := tokenAllowed(uri, m.param, m.allow, m.dev)
	if ok {
		return true, 0
	}
	if m.dev {
		Logf(api.LogLevelInfo, "handleRequest: denying request uri=%s reason=%s", uri, reason)
	}
	return m.unauth(resp, reason)
}

func (m *TokenMiddleware) unauth(resp api.Response, reason string) (bool, uint32) {
	// Always set the status code. If dev mode is enabled, emit a body and
	// a content-type header to help debugging. In production (dev==false)
	// we avoid any headers or body so the response is a minimal 401.
	resp.SetStatusCode(uint32(m.deny))
	if m.dev {
		resp.Headers().Set("Content-Type", "text/plain; charset=utf-8")
		resp.Body().WriteString(unauthBody(m.dev, reason))
	}
	return false, 0
}

// unauthBody returns the response body when denying a request.
// When dev is enabled we return a friendly message; in non-dev mode return empty.
func unauthBody(dev bool, reason string) string {
	if !dev {
		return ""
	}
	switch reason {
	case "no-query":
		return "no query parameter is"
	case "param-missing":
		return "invalid query parameter"
	case "no-match":
		return "invalid token"
	default:
		return "invalid parameter value"
	}
}

// parseQuery extracts all values for key from "a=1&b=2&b=3"
func parseQuery(qs, key string) []string {
	out := make([]string, 0)
	if qs == "" {
		return out
	}
	k := key + "="
	parts := strings.Split(qs, "&")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, k) {
			out = append(out, urlDecode(part[len(k):]))
		} else if part == key { // key without "=", treat as empty value
			out = append(out, "")
		}
	}
	return out
}

// tokenAllowed determines whether the request URI contains an allowed token
// for the given param name. It is extracted to a helper so the decision logic
// can be unit tested without needing the http-wasm host types.
func tokenAllowed(uri, param string, allow map[string]struct{}, dev bool) (bool, string) {
	i := strings.IndexByte(uri, '?')
	if i < 0 {
		if dev {
			Logf(api.LogLevelDebug, "tokenAllowed: no query string for uri=%s", uri)
		}
		return false, "no-query"
	}
	qs := uri[i+1:]
	vals := parseQuery(qs, param)
	if len(vals) == 0 {
		if dev {
			Logf(api.LogLevelDebug, "tokenAllowed: no values for param=%s in uri=%s", param, uri)
		}
		return false, "param-missing"
	}
	for _, v := range vals {
		if _, ok := allow[v]; ok {
			if dev {
				Logf(api.LogLevelDebug, "tokenAllowed: allowed value=%s for param=%s", v, param)
			}
			return true, ""
		}
	}
	if dev {
		Logf(api.LogLevelDebug, "tokenAllowed: no matching allowed values for uri=%s", uri)
	}
	return false, "no-match"
}

func urlDecode(s string) string {
	if strings.IndexByte(s, '%') == -1 && strings.IndexByte(s, '+') == -1 {
		return s
	}
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '+':
			b = append(b, ' ')
		case '%':
			if i+2 < len(s) {
				h1, h2 := fromHex(s[i+1]), fromHex(s[i+2])
				if h1 >= 0 && h2 >= 0 {
					b = append(b, byte(h1<<4|h2))
					i += 2
					continue
				}
			}
			b = append(b, c)
		default:
			b = append(b, c)
		}
	}
	return string(b)
}

func fromHex(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}
