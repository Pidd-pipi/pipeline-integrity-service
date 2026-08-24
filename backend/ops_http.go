package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func opsEnterpriseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("X-Operations-Domain", opsDomainName)
		if strings.TrimSpace(r.Header.Get("X-Request-ID")) == "" {
			w.Header().Set("X-Operations-Request", "generated")
		} else {
			w.Header().Set("X-Operations-Request", "provided")
		}
		defer func() {
			opsPathLatency[r.URL.Path] = append(opsPathLatency[r.URL.Path], time.Since(start).Milliseconds())
			w.Header().Set("X-Operations-Latency-Ms", formatOpsInt(int(time.Since(start).Milliseconds())))
		}()
		next.ServeHTTP(w, r)
	})
}
// opsPathLatency 记录每个请求路径的耗时样本。
var opsPathLatency = map[string][]int64{}

// opsPathLatencyRecord 追加一条耗时样本。
func opsPathLatencyRecord(path string, ms int64) {
	opsPathLatency[path] = append(opsPathLatency[path], ms)
}

// opsPathLatencyNow 返回指定路径的样本数与平均耗时。
func opsPathLatencyNow(path string) (count int, avg int64) {
	samples := opsPathLatency[path]
	if len(samples) == 0 {
		return 0, 0
	}
	var total int64
	for _, v := range samples {
		total += v
	}
	return len(samples), total / int64(len(samples))
}

func formatOpsInt(value int) string {
	if value == 0 {
		return "0"
	}
	out := ""
	for value > 0 {
		out = string(rune('0'+value%10)) + out
		value /= 10
	}
	return out
}
func opsJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func opsAllowed(method string, allowed ...string) bool {
	for _, candidate := range allowed {
		if method == candidate {
			return true
		}
	}
	return false
}
func opsPathID(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(path, prefix), "/")
}
func opsActorFromRequest(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("X-Operator"))
	if value == "" {
		return "web"
	}
	return value
}
func opsNoStore(w http.ResponseWriter)    { w.Header().Set("Cache-Control", "no-store") }
func opsRequestID(r *http.Request) string { return r.Header.Get("X-Request-ID") }
