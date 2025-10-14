package main

import (
    "testing"
)

func TestTokenAllowed(t *testing.T) {
    allow := map[string]struct{}{"good":{}, "another":{}}
    cases := []struct{
        name string
        uri  string
        want bool
    }{
        {"good token", "/?Token=good", true},
        {"another token", "/path?Token=another", true},
        {"missing token", "/", false},
        {"bad token", "/?Token=bad", false},
        {"multiple values one good", "/?Token=bad&Token=good", true},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got := tokenAllowed(tc.uri, "Token", allow)
            if got != tc.want {
                t.Fatalf("tokenAllowed(%q) = %v; want %v", tc.uri, got, tc.want)
            }
        })
    }
}

