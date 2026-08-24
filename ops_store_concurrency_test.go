package main

import (
	"context"
	"sync"
	"testing"
)

func seedOpsRecords001() []OpsRecord {
	return []OpsRecord{
		{ID: "op-101", Subject: "north segment inspection", Owner: "alice", Status: OpsStatusActive, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "north", "evidence": "e1"}, Revision: 1},
		{ID: "op-102", Subject: "east segment inspection", Owner: "bob", Status: OpsStatusQueued, Priority: OpsPriorityNormal, Labels: map[string]string{"site": "east", "evidence": "e2"}, Revision: 1},
		{ID: "op-103", Subject: "west segment inspection", Owner: "carol", Status: OpsStatusPaused, Priority: OpsPriorityCritical, Labels: map[string]string{"site": "west"}, Revision: 1},
		{ID: "op-104", Subject: "south segment inspection", Owner: "dave", Status: OpsStatusActive, Priority: OpsPriorityLow, Labels: map[string]string{"site": "south", "evidence": "e4"}, Revision: 1},
	}
}

func TestOpsStoreRaceFreeReads(t *testing.T) {
	store := newOpsStore(seedOpsRecords001())
	ctx := context.Background()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			for j := 0; j < 300; j++ {
				switch n % 4 {
				case 0:
					_, _ = store.Get(ctx, "op-101")
				case 1:
					_, _ = store.List(ctx)
				case 2:
					_ = store.Count()
				case 3:
					rec, err := store.Get(ctx, "op-102")
					if err == nil {
						_ = store.Update(ctx, rec, rec.Revision)
					}
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestOpsStoreLabelsIsolated(t *testing.T) {
	store := newOpsStore(seedOpsRecords001())
	ctx := context.Background()
	rec, err := store.Get(ctx, "op-101")
	if err != nil {
		t.Fatal(err)
	}
	rec.Labels["tampered"] = "yes"
	rec.Labels["site"] = "hacked"
	again, err := store.Get(ctx, "op-101")
	if err != nil {
		t.Fatal(err)
	}
	if got := again.Labels["tampered"]; got != "" {
		t.Fatalf("label tamper leaked into store: tampered=%q", got)
	}
	if got := again.Labels["site"]; got != "north" {
		t.Fatalf("store site label polluted: got %q want north", got)
	}
	items, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.ID == "op-101" {
			if item.Labels["site"] != "north" {
				t.Fatalf("list view polluted: site=%q", item.Labels["site"])
			}
		}
	}
}

func TestOpsStoreConcurrentWriteRead(t *testing.T) {
	store := newOpsStore(seedOpsRecords001())
	ctx := context.Background()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 150; j++ {
				rec, err := store.Get(ctx, "op-103")
				if err == nil {
					_ = store.Update(ctx, rec, rec.Revision)
				}
				_, _ = store.List(ctx)
				_ = store.Count()
			}
		}()
	}
	close(start)
	wg.Wait()
	final, err := store.Get(ctx, "op-103")
	if err != nil {
		t.Fatal(err)
	}
	if final.Revision < 1 {
		t.Fatalf("expected revisions applied, got %d", final.Revision)
	}
}

func TestOpsListLabelsIsolated(t *testing.T) {
	store := newOpsStore(seedOpsRecords001())
	ctx := context.Background()
	items, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := range items {
		if items[i].ID == "op-101" {
			items[i].Labels["via_list"] = "polluted"
		}
	}
	again, err := store.Get(ctx, "op-101")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := again.Labels["via_list"]; ok {
		t.Fatalf("mutating list result polluted the store")
	}
}
