package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

func newRouter(store *CycleStore) http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", healthHandler("pipeline-integrity-service"))
	m.HandleFunc("/api/cycles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/cycles" {
			writeJSON(w, 405, map[string]string{"error": "method not allowed"})
			return
		}
		ctx, cancel := opsContext(r.Context(), 3*time.Second)
		defer cancel()
		if err := opsDelay(ctx, 10*time.Millisecond); err != nil {
			writeJSON(w, 499, map[string]string{"error": "request canceled"})
			return
		}
		writeJSON(w, 200, store.list(ctx))
	})
	m.HandleFunc("/api/cycles/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, 405, map[string]string{"error": "method not allowed"})
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/cycles/")
		if !strings.HasSuffix(path, "/status") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "integrity cycle status path not found"})
			return
		}
		id := strings.TrimSuffix(path, "/status")
		if id == "" || strings.Contains(id, "/") {
			writeJSON(w, 404, map[string]string{"error": "integrity cycle not found"})
			return
		}
		var c StatusChange
		if json.NewDecoder(r.Body).Decode(&c) != nil || validateCycleStatus(c.Status) != nil {
			writeJSON(w, 400, map[string]string{"error": "valid status is required"})
			return
		}
		ctx, cancel := opsContext(r.Context(), 3*time.Second)
		defer cancel()
		if err := opsDelay(ctx, 20*time.Millisecond); err != nil {
			writeJSON(w, 499, map[string]string{"error": "request canceled"})
			return
		}
		v, e := store.changeStatus(ctx, id, c.Status)
		if errors.Is(e, errCycleNotFound) {
			writeJSON(w, 404, map[string]string{"error": e.Error()})
			return
		}
		if e != nil {
			writeJSON(w, 499, map[string]string{"error": "request canceled"})
			return
		}
		writeJSON(w, 200, v)
	})
	return m
}
