package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCycleHTTP(t *testing.T) {
	h := newRouter(newCycleStore())
	cases := []struct {
		name, path, body string
		code             int
	}{{"collection", "/api/cycles", "", 200}, {"accept", "/api/cycles/cy-401/status", `{"status":"accepted"}`, 200}, {"invalid", "/api/cycles/cy-401/status", `{"status":"done"}`, 400}, {"missing", "/api/cycles/nope/status", `{"status":"active"}`, 404}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRecorder()
			method := http.MethodGet
			if tc.body != "" {
				method = http.MethodPost
			}
			h.ServeHTTP(r, httptest.NewRequest(method, tc.path, bytes.NewBufferString(tc.body)))
			if r.Code != tc.code {
				t.Fatalf("got %d", r.Code)
			}
		})
	}
}
