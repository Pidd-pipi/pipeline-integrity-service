package main

import (
	"context"
	"sync"
)

// CycleWorker 后台周期复检 worker：定期拉取周期台账并维护只读快照。
type CycleWorker struct {
	store *CycleStore
	mu    sync.RWMutex
	items []IntegrityCycle
}

func newCycleWorker(store *CycleStore) *CycleWorker {
	return &CycleWorker{store: store, items: []IntegrityCycle{}}
}

// Refresh 拉取最新周期列表并更新快照。
func (w *CycleWorker) Refresh(ctx context.Context) error {
	items := w.store.list()
	w.items = items
	return nil
}

// Snapshot 返回最近一次刷新的周期快照副本。
func (w *CycleWorker) Snapshot() []IntegrityCycle {
	out := make([]IntegrityCycle, len(w.items))
	copy(out, w.items)
	return out
}

// Size 返回当前快照条数。
func (w *CycleWorker) Size() int {
	return len(w.items)
}
