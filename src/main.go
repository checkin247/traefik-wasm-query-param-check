package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler"
	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

type Config struct {
	ParamName     string   `json:"paramName"`
	AllowedValues []string `json:"allowedValues"`
	DenyStatus    int      `json:"denyStatus,omitempty,string"` // default 401
}

type TokenMiddleware struct {
	param string
	allow map[string]struct{}
	deny  int
}

// TinyGo/WASI requires a main entrypoint even if unused by http-wasm.
func main() {}

func init() {
	var cfg Config
	if err := json.Unmarshal(handler.Host.GetConfig(), &cfg); err != nil {
		handler.Host.Log(api.LogLevelError, fmt.Sprintf("config unmarshal: %v", err))
		os.Exit(1)
	}
	if cfg.ParamName == "" || len(cfg.AllowedValues) == 0 {
		handler.Host.Log(api.LogLevelError, "invalid config: paramName and allowedValues are required")
		os.Exit(1)
	}
	deny := cfg.DenyStatus
	if deny == 0 {
		deny = 401
	}
	set := make(map[string]struct{}, len(cfg.AllowedValues))
	for _, v := range cfg.AllowedValues {
		set[v] = struct{}{}
	}
	mw := &TokenMiddleware{param: cfg.ParamName, allow: set, deny: deny}
	handler.HandleRequestFn = mw.handleRequest
}

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
	var out []string
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
