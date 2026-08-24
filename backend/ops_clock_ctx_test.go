package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestHTTPCanceledRequestStops(t *testing.T) {
	h := newRouter(newCycleStore())
	body := bytes.NewBufferString(`{"status":"active"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/cycles/cy-401/status", body)
	req = req.WithContext(canceledContext())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 499 {
		t.Fatalf("canceled status change should stop with 499, got %d", rec.Code)
	}
}

func TestHTTPCanceledListStops(t *testing.T) {
	h := newRouter(newCycleStore())
	req := httptest.NewRequest(http.MethodGet, "/api/cycles", nil)
	req = req.WithContext(canceledContext())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 499 {
		t.Fatalf("canceled list should stop with 499, got %d", rec.Code)
	}
}

func TestOpsDelayHonorsCancel(t *testing.T) {
	start := time.Now()
	err := opsDelay(canceledContext(), 300*time.Millisecond)
	if err != context.Canceled {
		t.Fatalf("opsDelay should return ctx.Err() when canceled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("opsDelay ignored cancel, waited %v", elapsed)
	}
}

func TestOpsContextKeepsParentDeadline(t *testing.T) {
	parent, cancelParent := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancelParent()
	ctx, cancel := opsContext(parent, 5*time.Second)
	defer cancel()
	time.Sleep(120 * time.Millisecond)
	if err := ctx.Err(); err == nil {
		t.Fatalf("opsContext dropped parent deadline: ctx still alive after parent expired")
	}
}
