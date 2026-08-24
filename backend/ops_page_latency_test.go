package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestPathLatencyRecordSummaryNoRace(t *testing.T) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			for j := 0; j < 300; j++ {
				if n%2 == 0 {
					opsPathLatencyRecord("/api/cycles", int64(j))
				} else {
					_, _ = opsPathLatencyNow("/api/cycles")
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestMiddlewareLatencyNoRace(t *testing.T) {
	h := opsEnterpriseMiddleware(requestIDMiddleware(recoveryMiddleware(http.NotFoundHandler())))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/cycles", nil))
				_, _ = opsPathLatencyNow("/api/cycles")
			}
		}()
	}
	close(start)
	wg.Wait()
}

func TestPageShowsLatency(t *testing.T) {
	inner := newRouter(newCycleStore())
	mux := http.NewServeMux()
	mux.Handle("/healthz", inner)
	mux.Handle("/api/", inner)
	mux.Handle("/", staticHandler())
	srv := newEnterpriseServer("127.0.0.1:0", mux)
	h := srv.Handler
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
				opsPathLatencyRecord("/", int64(j))
			}
		}()
	}
	close(start)
	wg.Wait()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "Latency samples: 1") && !strings.Contains(body, "Latency samples: 2") &&
		!strings.Contains(body, "Latency samples: 3") && !strings.Contains(body, "Latency samples: 4") &&
		!strings.Contains(body, "Latency samples: 5") && !strings.Contains(body, "Latency samples: 6") &&
		!strings.Contains(body, "Latency samples: 7") && !strings.Contains(body, "Latency samples: 8") {
		t.Fatalf("page must render a positive latency sample count, got body=%q", body)
	}
}

func TestPathLatencyBounded(t *testing.T) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				opsPathLatencyRecord("/", int64(j))
			}
		}()
	}
	close(start)
	wg.Wait()
	count, _ := opsPathLatencyNow("/")
	if count > 150 {
		t.Fatalf("latency samples must be capped, got %d", count)
	}
}
