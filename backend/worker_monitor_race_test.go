package main

import (
	"context"
	"sync"
	"testing"
)

func TestCycleWorkerRefreshSnapshotNoRace(t *testing.T) {
	store := newCycleStore()
	worker := newCycleWorker(store)
	ctx := context.Background()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			for j := 0; j < 300; j++ {
				switch n % 3 {
				case 0:
					_ = worker.Refresh(ctx)
				case 1:
					_ = worker.Snapshot()
				case 2:
					_ = worker.Size()
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestOpsLatencyRecordSummaryNoRace(t *testing.T) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			for j := 0; j < 300; j++ {
				if n%2 == 0 {
					opsLatencyRecord("/api/cycles", int64(j))
				} else {
					_, _ = opsLatencySummary("/api/cycles")
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestOpsLatencyResetNoRace(t *testing.T) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				if n%2 == 0 {
					opsLatencyRecord("/api/cycles", int64(j))
				} else {
					opsLatencyReset()
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestOpsLatencySummaryBounded(t *testing.T) {
	opsLatencyReset()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 200; j++ {
				opsLatencyRecord("/api/cycles", int64(j))
			}
		}()
	}
	close(start)
	wg.Wait()
	count, _ := opsLatencySummary("/api/cycles")
	if count > 150 {
		t.Fatalf("latency samples must be capped, got %d", count)
	}
	opsLatencyReset()
}
