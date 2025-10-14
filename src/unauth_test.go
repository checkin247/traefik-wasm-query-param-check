package main

import (
    "testing"
)

func TestUnauthBody(t *testing.T) {
    if unauthBody(true, "other") != "invalid parameter value" {
        t.Fatalf("unauthBody(true) must be 'invalid parameter value'")
    }
    if unauthBody(false, "other") != "" {
        t.Fatalf("unauthBody(false) must be empty string")
    }
}

func TestUnauthBodyReasons(t *testing.T) {
    cases := []struct{
        reason string
        want string
    }{
        {"no-query", "no query parameter is"},
        {"param-missing", "invalid query parameter"},
        {"no-match", "invalid token"},
        {"other", "invalid parameter value"},
    }
    for _, c := range cases {
        got := unauthBody(true, c.reason)
        if got != c.want {
            t.Fatalf("unauthBody(true,%q) = %q; want %q", c.reason, got, c.want)
        }
    }
}
