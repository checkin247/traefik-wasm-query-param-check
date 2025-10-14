package main

import (
    "testing"
)

func TestTokenAllowed(t *testing.T) {
    allow := map[string]struct{}{"good":{}, "another":{}}
    cases := []struct{
        name string
        uri  string
        wantOk bool
        wantReason string
    }{
    {"good token", "/?Token=good", true, ""},
    {"another token", "/path?Token=another", true, ""},
        {"missing token", "/", false, "no-query"},
        {"bad token", "/?Token=bad", false, "no-match"},
        {"multiple values one good", "/?Token=bad&Token=good", true, ""},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got, reason := tokenAllowed(tc.uri, "Token", allow, false)
            if got != tc.wantOk {
                t.Fatalf("tokenAllowed(%q) = %v; want %v", tc.uri, got, tc.wantOk)
            }
            if tc.wantReason != "" && reason != tc.wantReason {
                t.Fatalf("tokenAllowed reason=%q want=%q", reason, tc.wantReason)
            }
        })
    }
}

