package main

import (
    "reflect"
    "testing"
)

func TestParseQuery(t *testing.T) {
    cases := []struct{
        name string
        qs   string
        key  string
        want []string
    }{
        {"multiple values", "a=1&b=2&b=3", "b", []string{"2","3"}},
        {"single value", "a=1&b=2", "a", []string{"1"}},
        {"key without value", "a=1&b", "b", []string{""}},
        {"empty qs", "", "a", []string{}},
        {"percent-decoding", "q=%20+%21", "q", []string{"  !"}},
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            got := parseQuery(tc.qs, tc.key)
            if !reflect.DeepEqual(got, tc.want) {
                t.Fatalf("parseQuery(%q,%q) = %#v; want %#v", tc.qs, tc.key, got, tc.want)
            }
        })
    }
}

func TestURLDecode(t *testing.T) {
    cases := []struct{
        in string
        want string
    }{
        {"plain", "plain"},
        {"plus+space", "plus space"},
        {"percent%21", "percent!"},
        {"mixed+%20%21", "mixed  !"},
    }
    for _, c := range cases {
        got := urlDecode(c.in)
        if got != c.want {
            t.Fatalf("urlDecode(%q) = %q; want %q", c.in, got, c.want)
        }
    }
}
