package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRolloutRetries429AndSendsRequiredFields(t *testing.T) {
	seen := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen++
		if r.Method != http.MethodPost || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fail()
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		for field, want := range map[string]any{
			"key": "care", "percentage": float64(10), "salt": "healthtech-rollout",
			"sticky_unit": "user", "version": float64(1),
		} {
			if body[field] != want {
				t.Fatalf("%s = %v, want %v", field, body[field], want)
			}
		}
		if _, ok := body["idempotency_key"]; ok {
			t.Fatal("unexpected idempotency_key")
		}
		if seen == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer server.Close()
	client := NewClient("test-key")
	client.base = server.URL
	if err := client.Rollout(context.Background(), "care", 10); err != nil {
		t.Fatal(err)
	}
	if seen != 2 {
		t.Fatalf("requests = %d", seen)
	}
}
