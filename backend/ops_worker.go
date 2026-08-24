package main

import (
	"context"
	"sync"
)

// OpsRecheckWorker 后台复检 worker：刷新进行中作业快照。
type OpsRecheckWorker struct {
	store  *OpsStore
	mu     sync.Mutex
	active []OpsRecord
}

func newOpsRecheckWorker(store *OpsStore) *OpsRecheckWorker {
	return &OpsRecheckWorker{store: store, active: []OpsRecord{}}
}

// Refresh 拉取进行中作业并更新本地快照。
func (w *OpsRecheckWorker) Refresh(ctx context.Context) error {
	items, err := w.store.List(ctx)
	if err != nil {
		return err
	}
	snapshot := make([]OpsRecord, 0, len(items))
	for _, item := range items {
		if item.Status == OpsStatusActive {
			snapshot = append(snapshot, item)
		}
	}
	w.active = snapshot
	return nil
}

// ActiveSnapshot 返回最近一次刷新的进行中作业快照副本。
func (w *OpsRecheckWorker) ActiveSnapshot() []OpsRecord {
	out := make([]OpsRecord, len(w.active))
	copy(out, w.active)
	return out
}
