//go:build tinygo
// +build tinygo

package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/http-wasm/http-wasm-guest-tinygo/handler"
    "github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

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
    deny := int(cfg.DenyStatus)
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
