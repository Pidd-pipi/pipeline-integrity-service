package main

import (
	"errors"
	"fmt"
)

var (
	ErrOpsNotFound   = errors.New("operations record not found")
	ErrOpsConflict   = errors.New("operations revision conflict")
	ErrOpsInvalid    = errors.New("operations request is invalid")
	ErrOpsTransition = errors.New("operations status transition is not allowed")
	ErrOpsPolicy     = errors.New("operations policy rejected the request")
)

// OpsCode* are the stable, client-facing error codes mapped from the sentinel
// errors above. They are returned by opsCode and serialized over the wire so
// callers can branch on a stable string instead of parsing prose.
const (
	OpsCodeNotFound   = "not_found"
	OpsCodeConflict   = "conflict"
	OpsCodeInvalid    = "invalid"
	OpsCodeTransition = "transition"
	OpsCodePolicy     = "policy"
	OpsCodeInternal   = "internal"
)

// OpsError carries the operation context alongside a sentinel cause. It
// preserves the cause chain so errors.Is / errors.As keep working after
// wrapping — the previous implementation returned nil from Unwrap, which
// severed the chain and made every classified error read as "internal".
type OpsError struct {
	Code      string
	Operation string
	Cause     error
}

func (e *OpsError) Error() string {
	if e.Cause == nil {
		return e.Code + ": " + e.Operation
	}
	return fmt.Sprintf("%s: %s: %s", e.Code, e.Operation, e.Cause)
}

// Unwrap exposes Cause so errors.Is reaches the underlying sentinel error.
func (e *OpsError) Unwrap() error { return e.Cause }

// wrapOps annotates an error with the operation that produced it while keeping
// the sentinel cause reachable via errors.Is (note the %w verb, not %v).
func wrapOps(code, operation string, cause error) error {
	if cause == nil {
		return nil
	}
	return &OpsError{Code: code, Operation: operation, Cause: cause}
}

// opsCode returns the stable client-facing code for an error. It walks the
// chain (including OpsError.Cause) so wrapped sentinels classify correctly
// instead of collapsing to "internal".
func opsCode(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrOpsNotFound):
		return OpsCodeNotFound
	case errors.Is(err, ErrOpsConflict):
		return OpsCodeConflict
	case errors.Is(err, ErrOpsInvalid):
		return OpsCodeInvalid
	case errors.Is(err, ErrOpsTransition):
		return OpsCodeTransition
	case errors.Is(err, ErrOpsPolicy):
		return OpsCodePolicy
	}
	if code := opsCodeFromTyped(err); code != "" {
		return code
	}
	return OpsCodeInternal
}

// opsCodeFromTyped honors an explicit Code set on an *OpsError that is not
// backed by one of the known sentinels (e.g. a domain-specific code).
func opsCodeFromTyped(err error) string {
	var typed *OpsError
	if errors.As(err, &typed) && typed.Code != "" {
		return typed.Code
	}
	return ""
}

func opsIsNotFound(err error) bool   { return errors.Is(err, ErrOpsNotFound) }
func opsIsConflict(err error) bool   { return errors.Is(err, ErrOpsConflict) }
func opsIsInvalid(err error) bool    { return errors.Is(err, ErrOpsInvalid) }
func opsIsTransition(err error) bool { return errors.Is(err, ErrOpsTransition) }
func opsIsPolicy(err error) bool     { return errors.Is(err, ErrOpsPolicy) }
