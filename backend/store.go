package main

import (
	"errors"
	"sync"
)

var (
	errCycleNotFound           = errors.New("integrity cycle not found")
	errCycleInvalidTransition  = errors.New("integrity cycle status transition is not allowed")
)

type CycleStore struct {
	mu    sync.RWMutex
	items map[string]IntegrityCycle
}

func newCycleStore() *CycleStore {
	return &CycleStore{items: map[string]IntegrityCycle{"cy-401": {ID: "cy-401", Segment: "North-12", Method: "ultrasonic", Risk: "elevated", Status: "planned", DueDate: "2026-09-02"}, "cy-402": {ID: "cy-402", Segment: "East-04", Method: "magnetic flux", Risk: "normal", Status: "active", DueDate: "2026-08-28"}}}
}
func (s *CycleStore) list() []IntegrityCycle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o := make([]IntegrityCycle, 0, len(s.items))
	for _, v := range s.items {
		o = append(o, v)
	}
	return o
}
func (s *CycleStore) changeStatus(id, status string) (IntegrityCycle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.items[id]
	if !ok {
		return IntegrityCycle{}, errCycleNotFound
	}
	if err := validateCycleTransition(v.Status, status); err != nil {
		return IntegrityCycle{}, errCycleInvalidTransition
	}
	v.Status = status
	s.items[id] = v
	return v, nil
}
