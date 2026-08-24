package main

import (
	"errors"
	"testing"
)

func TestWrappedErrorChainClassified(t *testing.T) {
	err := wrapOps("create", "store.put", ErrOpsConflict)
	if !errors.Is(err, ErrOpsConflict) {
		t.Fatalf("wrapped conflict error must be errors.Is-detectable")
	}
	if !opsIsConflict(err) {
		t.Fatalf("opsIsConflict must recognize the wrapped conflict error")
	}
	pol := wrapOps("create", "policy.check", ErrOpsPolicy)
	if !errors.Is(pol, ErrOpsPolicy) {
		t.Fatalf("wrapped policy error must be errors.Is-detectable")
	}
	if !opsIsPolicy(pol) {
		t.Fatalf("opsIsPolicy must recognize the wrapped policy error")
	}
}

func TestOpsErrorUnwrapsCause(t *testing.T) {
	err := wrapOps("get", "store.get", ErrOpsNotFound)
	var typed *OpsError
	if !errors.As(err, &typed) {
		t.Fatalf("expected *OpsError from wrapOps")
	}
	if !errors.Is(typed, ErrOpsNotFound) {
		t.Fatalf("OpsError must unwrap to its cause")
	}
	if !opsIsNotFound(err) {
		t.Fatalf("opsIsNotFound must recognize the wrapped not-found error")
	}
}

func TestTransitionErrorDetectable(t *testing.T) {
	m := newOpsStateMachine()
	err := m.Move(OpsStatusClosed, OpsStatusActive, "back")
	if err == nil {
		t.Fatalf("illegal transition must return an error")
	}
	if !errors.Is(err, ErrOpsTransition) {
		t.Fatalf("transition error must be errors.Is-detectable")
	}
	if !opsIsTransition(err) {
		t.Fatalf("opsIsTransition must recognize the transition error")
	}
	if got := opsCode(err); got != "transition" {
		t.Fatalf("opsCode(transition) = %q, want transition", got)
	}
}

func TestRawSentinelClassified(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{ErrOpsNotFound, "not_found"},
		{ErrOpsConflict, "conflict"},
		{ErrOpsInvalid, "invalid"},
		{ErrOpsTransition, "transition"},
		{ErrOpsPolicy, "policy"},
	}
	for _, c := range cases {
		if got := opsCode(c.err); got != c.want {
			t.Fatalf("opsCode(%v) = %q, want %q", c.err, got, c.want)
		}
	}
}
