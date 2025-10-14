// +build !tinygo

package main

import (
    "fmt"

    "github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

// Logf for non-TinyGo builds — prints to stdout via fmt.Printf so unit tests can see logs if needed.
func Logf(level api.LogLevel, format string, args ...interface{}) {
    // Simple mapping of level to prefix
    prefix := ""
    switch level {
    case api.LogLevelError:
        prefix = "ERROR"
    case api.LogLevelInfo:
        prefix = "INFO"
    default:
        prefix = "DEBUG"
    }
    fmt.Printf("[%s] "+format+"\n", append([]interface{}{prefix}, args...)...)
}
