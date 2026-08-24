package main

import (
	"testing"
)

func TestCycleActiveCanTransitionToAccepted(t *testing.T) {
	found := false
	for _, next := range cycleTransitionTable["active"] {
		if next == "accepted" {
			found = true
		}
	}
	if !found {
		t.Fatalf("active must be able to transition to accepted, table=%v", cycleTransitionTable["active"])
	}
}

func TestCycleAcceptedIsTerminal(t *testing.T) {
	if len(cycleTransitionTable["accepted"]) != 0 {
		t.Fatalf("accepted is terminal, must have no outgoing transitions, got %v", cycleTransitionTable["accepted"])
	}
}

func TestValidateCycleTransitionRejectsIllegal(t *testing.T) {
	if err := validateCycleTransition("accepted", "active"); err == nil {
		t.Fatalf("accepted -> active must be rejected")
	}
	if err := validateCycleTransition("active", "planned"); err == nil {
		t.Fatalf("active -> planned must be rejected")
	}
}

func TestValidateCycleTransitionAllowsLegal(t *testing.T) {
	if err := validateCycleTransition("planned", "active"); err != nil {
		t.Fatalf("planned -> active must be allowed: %v", err)
	}
	if err := validateCycleTransition("planned", "accepted"); err != nil {
		t.Fatalf("planned -> accepted must be allowed: %v", err)
	}
	if err := validateCycleTransition("active", "accepted"); err != nil {
		t.Fatalf("active -> accepted must be allowed: %v", err)
	}
}

func TestCycleChangeStatusRejectsIllegal(t *testing.T) {
	store := newCycleStore()
	if _, err := store.changeStatus("cy-402", "accepted"); err != nil {
		t.Fatalf("active -> accepted should apply: %v", err)
	}
	if _, err := store.changeStatus("cy-402", "active"); err == nil {
		t.Fatalf("accepted -> active must be rejected")
	}
	if got := store.items["cy-402"].Status; got != "accepted" {
		t.Fatalf("cycle status must stay accepted, got %q", got)
	}
}

func TestCycleChangeStatusAllowsLegal(t *testing.T) {
	store := newCycleStore()
	if _, err := store.changeStatus("cy-401", "active"); err != nil {
		t.Fatalf("planned -> active should apply: %v", err)
	}
	if got := store.items["cy-401"].Status; got != "active" {
		t.Fatalf("cycle status should be active, got %q", got)
	}
}
