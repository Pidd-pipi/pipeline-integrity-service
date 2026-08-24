package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
	"time"
)

func TestServeHTTPShutdownNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	srv := newEnterpriseServer("127.0.0.1:18765", http.NotFoundHandler())
	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	go func() {
		_ = serveHTTPWithSignals(srv, signals)
		close(done)
	}()
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:18765")
		if err == nil {
			_ = conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not become ready: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	close(signals)
	<-done
	time.Sleep(200 * time.Millisecond)
	if got := runtime.NumGoroutine(); got > before {
		t.Fatalf("goroutine leaked after graceful shutdown: before=%d after=%d", before, got)
	}
}

func TestStartAndServeNoBgLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	_ = startAndServe(Config{Port: "not-a-port"})
	time.Sleep(100 * time.Millisecond)
	// 首次 signal.Notify 会常驻一个全局 signal loop goroutine（+1），允许该常驻；
	// 自检/看门狗后台 goroutine 属于泄漏（env 会再多 +2）。
	if got := runtime.NumGoroutine(); got > before+1 {
		t.Fatalf("startAndServe leaked background goroutines: before=%d after=%d", before, got)
	}
}

func TestRecoveryMiddlewareReturns500(t *testing.T) {
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := recoveryMiddleware(panicHandler)
	rec := httptest.NewRecorder()
	func() {
		defer func() { _ = recover() }()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	}()
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("recovery should write 500 instead of re-panicking, got %d", rec.Code)
	}
}
