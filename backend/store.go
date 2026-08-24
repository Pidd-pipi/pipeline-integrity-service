package main

import (
	"context"
	"errors"
	"sync"
)

var errCycleNotFound = errors.New("integrity cycle not found")

type CycleStore struct {
	mu    sync.RWMutex
	items map[string]IntegrityCycle
}

func newCycleStore() *CycleStore {
	return &CycleStore{items: map[string]IntegrityCycle{"cy-401": {ID: "cy-401", Segment: "North-12", Method: "ultrasonic", Risk: "elevated", Status: "planned", DueDate: "2026-09-02"}, "cy-402": {ID: "cy-402", Segment: "East-04", Method: "magnetic flux", Risk: "normal", Status: "active", DueDate: "2026-08-28"}}}
}
func (s *CycleStore) list(ctx context.Context) []IntegrityCycle {
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := make([]IntegrityCycle, 0, len(s.items))
	for _, v := range s.items {
		o = append(o, v)
	}
	return o
}
func (s *CycleStore) changeStatus(ctx context.Context, id, status string) (IntegrityCycle, error) {
	select {
	case <-ctx.Done():
		return IntegrityCycle{}, ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return IntegrityCycle{}, errCycleNotFound
	}
	v.Status = status
	s.items[id] = v
	return v, nil
}
