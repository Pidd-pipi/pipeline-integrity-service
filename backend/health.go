package main

import (
	"encoding/json"
	"net/http"
	"time"
)

var healthHistory []string
var healthHistoryLimit = 100

func healthSampleCount() int { return len(healthHistory) }

func healthHandler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, 405, map[string]string{"error": "method not allowed"})
			return
		}
		healthHistory = append(healthHistory, time.Now().UTC().Format(time.RFC3339Nano))
		writeJSON(w, 200, map[string]string{"status": "ok", "service": service, "time": time.Now().UTC().Format(time.RFC3339), "samples": formatOpsInt(healthSampleCount())})
	}
}
func writeJSON(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(v)
}
