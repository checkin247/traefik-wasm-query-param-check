package main

import (
    "encoding/json"
    "testing"
)

// helper to unmarshal and apply the default behavior used in init()
func parseAndDefault(t *testing.T, input string) (int, error) {
    t.Helper()
    var cfg Config
    if err := json.Unmarshal([]byte(input), &cfg); err != nil {
        return 0, err
    }
    deny := int(cfg.DenyStatus)
    if deny == 0 {
        deny = 401
    }
    return deny, nil
}

func TestConfigDenyStatus(t *testing.T) {
    tests := []struct{
        name string
        json string
        want int
        wantErr bool
    }{
        {"string status", `{"paramName":"Token","allowedValues":["a"],"denyStatus":"401"}`, 401, false},
        {"numeric status", `{"paramName":"Token","allowedValues":["a"],"denyStatus":403}`, 403, false},
        {"missing status -> default", `{"paramName":"Token","allowedValues":["a"]}`, 401, false},
        {"invalid status", `{"paramName":"Token","allowedValues":["a"],"denyStatus":"bad"}`, 0, true},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got, err := parseAndDefault(t, tc.json)
            if (err != nil) != tc.wantErr {
                t.Fatalf("unexpected error state: %v", err)
            }
            if !tc.wantErr && got != tc.want {
                t.Fatalf("got=%d want=%d", got, tc.want)
            }
        })
    }
}
