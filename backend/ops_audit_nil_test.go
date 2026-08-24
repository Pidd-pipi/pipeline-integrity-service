package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestAuditAnnotateNoPanic(t *testing.T) {
	audit := newOpsAudit()
	audit.Add("op-501", "created", "alice")
	audit.Annotate("op-501", "evidence", "e-1")
	events := audit.For("op-501")
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	if got := events[0].Details["evidence"]; got != "e-1" {
		t.Fatalf("annotated value = %q, want e-1", got)
	}
}

func TestAuditClearReleasesMemory(t *testing.T) {
	audit := newOpsAudit()
	for i := 0; i < 200; i++ {
		audit.Add("op-501", "created", "alice")
	}
	audit.Clear()
	if n := audit.Count(); n != 0 {
		t.Fatalf("clear must empty the log, got %d", n)
	}
	audit.mu.RLock()
	capLeft := cap(audit.events)
	audit.mu.RUnlock()
	if capLeft != 0 {
		t.Fatalf("clear must release the backing array, cap=%d", capLeft)
	}
}

func TestAuditTrimBounded(t *testing.T) {
	audit := newOpsAudit()
	for i := 0; i < 500; i++ {
		audit.Add("op-501", "created", "alice")
	}
	audit.Trim(10)
	if n := audit.Count(); n != 10 {
		t.Fatalf("trim must keep last 10, got %d", n)
	}
	audit.mu.RLock()
	capLeft := cap(audit.events)
	audit.mu.RUnlock()
	if capLeft >= 100 {
		t.Fatalf("trim must not retain the old big backing array, cap=%d", capLeft)
	}
}

func TestAuditForIsolated(t *testing.T) {
	audit := newOpsAudit()
	audit.Add("op-501", "created", "alice")
	audit.Annotate("op-501", "evidence", "e-1")
	events := audit.For("op-501")
	events[0].Details["tampered"] = "yes"
	again := audit.For("op-501")
	if got := again[0].Details["tampered"]; got != "" {
		t.Fatalf("mutating For result polluted the log: tampered=%q", got)
	}
	if got := again[0].Details["evidence"]; got != "e-1" {
		t.Fatalf("original detail lost: evidence=%q", got)
	}
}

func TestAuditSinceIsolated(t *testing.T) {
	audit := newOpsAudit()
	audit.Add("op-501", "created", "alice")
	audit.Annotate("op-501", "evidence", "e-1")
	events := audit.Since(time.Now().Add(-time.Hour))
	if len(events) != 1 {
		t.Fatalf("want 1 event from since, got %d", len(events))
	}
	events[0].Details["tampered"] = "yes"
	again := audit.Since(time.Now().Add(-time.Hour))
	if got := again[0].Details["tampered"]; got != "" {
		t.Fatalf("mutating Since result polluted the log: tampered=%q", got)
	}
}

func seedOpsRecords005() []OpsRecord {
	return []OpsRecord{
		{ID: "op-501", Subject: "north segment recheck", Owner: "alice", Status: OpsStatusActive, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "north"}},
		{ID: "op-502", Subject: "east segment recheck", Owner: "bob", Status: OpsStatusQueued, Priority: OpsPriorityNormal, Labels: map[string]string{"site": "east"}},
	}
}

func TestRecheckWorkerRefreshNoRace(t *testing.T) {
	store := newOpsStore(seedOpsRecords005())
	worker := newOpsRecheckWorker(store)
	ctx := context.Background()
	start := make(chan struct{})
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for j := 0; j < 300; j++ {
			_ = worker.Refresh(ctx)
		}
		close(done)
	}()
	go func() {
		defer wg.Done()
		<-start
		for {
			select {
			case <-done:
				return
			default:
				_ = worker.ActiveSnapshot()
			}
		}
	}()
	close(start)
	wg.Wait()
}
