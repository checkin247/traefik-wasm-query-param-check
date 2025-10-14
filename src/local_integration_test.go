//go:build integration
// +build integration

package main

import (
    "net/http"
    "os"
    "testing"
)

// Local integration test. Run with: go test -tags=integration -run TestLocalIntegration ./...
func TestLocalIntegration(t *testing.T) {
    base := os.Getenv("LOCAL_TEST_BASE")
    if base == "" {
        base = "http://localhost:80/"
    }
    client := &http.Client{}

    goodTokens := []string{"my-secret", "another-secret"}
    for _, tok := range goodTokens {
        resp, err := client.Get(base + "?Token=" + tok)
        if err != nil {
            t.Fatalf("request error: %v", err)
        }
        if resp.StatusCode != 200 {
            t.Fatalf("good token %q expected 200 got %d", tok, resp.StatusCode)
        }
    }
    // missing token
    resp, err := client.Get(base)
    if err != nil {
        t.Fatalf("request error: %v", err)
    }
    if resp.StatusCode != 401 {
        t.Fatalf("missing token expected 401 got %d", resp.StatusCode)
    }
}
