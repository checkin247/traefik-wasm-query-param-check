//go:build tinygo
// +build tinygo

package main

import (
    "fmt"

    "github.com/http-wasm/http-wasm-guest-tinygo/handler"
    "github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

func Logf(level api.LogLevel, format string, args ...interface{}) {
    // http-wasm handler API expects a string message
    msg := fmt.Sprintf(format, args...)
    if handler.Host != nil {
        handler.Host.Log(level, msg)
    }
}
