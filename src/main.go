package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

type Config struct {
	ParamName     string   `json:"paramName"`
	AllowedValues []string `json:"allowedValues"`
	// DenyStatus accepts either a number or a quoted string ("401") in JSON.
	DenyStatus    HTTPStatus `json:"denyStatus,omitempty"` // default 401
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

type TokenMiddleware struct {
	param string
	allow map[string]struct{}
	deny  int
}

// TinyGo/WASI requires a main entrypoint even if unused by http-wasm.
func main() {}

// runtime initialization (moved to TinyGo-only file)

func (m *TokenMiddleware) handleRequest(req api.Request, resp api.Response) (next bool, reqCtx uint32) {
	uri := req.GetURI() // e.g. "/path?Token=abc"
	i := strings.IndexByte(uri, '?')
	if i < 0 {
		return m.unauth(resp, "missing query string")
	}
	qs := uri[i+1:]
	vals := parseQuery(qs, m.param)
	if len(vals) == 0 {
		return m.unauth(resp, fmt.Sprintf("missing %s", m.param))
	}
	for _, v := range vals {
		if _, ok := m.allow[v]; ok {
			return true, 0 // let Traefik continue
		}
	}
	return m.unauth(resp, "token not allowed")
}

func (m *TokenMiddleware) unauth(resp api.Response, msg string) (bool, uint32) {
	resp.SetStatusCode(uint32(m.deny))
	resp.Headers().Set("Content-Type", "text/plain; charset=utf-8")
	resp.Body().WriteString("Unauthorized: " + msg + "\n")
	return false, 0
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
func tokenAllowed(uri, param string, allow map[string]struct{}) bool {
	i := strings.IndexByte(uri, '?')
	if i < 0 {
		return false
	}
	qs := uri[i+1:]
	vals := parseQuery(qs, param)
	if len(vals) == 0 {
		return false
	}
	for _, v := range vals {
		if _, ok := allow[v]; ok {
			return true
		}
	}
	return false
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
