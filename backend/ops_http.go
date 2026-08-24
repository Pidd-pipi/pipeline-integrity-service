package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

func opsEnterpriseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header().Set("X-Operations-Domain", opsDomainName)
		if strings.TrimSpace(r.Header.Get("X-Request-ID")) == "" {
			w.Header().Set("X-Operations-Request", "generated")
		} else {
			w.Header().Set("X-Operations-Request", "provided")
		}
		defer func() { w.Header().Set("X-Operations-Latency-Ms", formatOpsInt(int(time.Since(start).Milliseconds()))) }()
		next.ServeHTTP(w, r)
	})
}
func formatOpsInt(value int) string {
	if value == 0 {
		return "0"
	}
	out := ""
	for value > 0 {
		out = string(rune('0'+value%10)) + out
		value /= 10
	}
	return out
}
func opsJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// opsHTTPStatus maps a classified error to its HTTP status. Sentinel errors
// carry enough meaning to pick a real status, so callers never have to fall
// back to 500 "internal error" for conditions the client can act on.
func opsHTTPStatus(code string) int {
	switch code {
	case OpsCodeNotFound:
		return http.StatusNotFound
	case OpsCodeConflict:
		return http.StatusConflict
	case OpsCodeInvalid, OpsCodePolicy:
		return http.StatusBadRequest
	case OpsCodeTransition:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// opsErrorResponse is the stable client-facing error body. The Code field lets
// clients branch on a machine-readable token; Message is human-readable.
type opsErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// opsWriteError classifies err and writes the matching HTTP status plus a
// stable JSON body. Unknown errors collapse to "internal" / 500 without
// leaking the underlying message, matching the prior client experience for
// genuinely unexpected failures.
func opsWriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code := opsCode(err)
	status := opsHTTPStatus(code)
	message := opsClientMessage(err, code)
	opsJSON(w, status, opsErrorResponse{Code: code, Message: message})
}

// opsClientMessage surfaces a safe, readable message for each classification.
// Sentinel-backed errors expose their own text; wrapped OpsErrors unwrap to
// that same sentinel. Anything unclassified is reported as internal so we
// never leak implementation detail.
func opsClientMessage(err error, code string) string {
	if code == OpsCodeInternal {
		return "internal error"
	}
	if msg := opsSentinelMessage(err); msg != "" {
		return msg
	}
	return err.Error()
}

func opsSentinelMessage(err error) string {
	switch {
	case errors.Is(err, ErrOpsNotFound):
		return ErrOpsNotFound.Error()
	case errors.Is(err, ErrOpsConflict):
		return ErrOpsConflict.Error()
	case errors.Is(err, ErrOpsInvalid):
		return ErrOpsInvalid.Error()
	case errors.Is(err, ErrOpsTransition):
		return ErrOpsTransition.Error()
	case errors.Is(err, ErrOpsPolicy):
		return ErrOpsPolicy.Error()
	}
	return ""
}

func opsAllowed(method string, allowed ...string) bool {
	for _, candidate := range allowed {
		if method == candidate {
			return true
		}
	}
	return false
}
func opsPathID(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(path, prefix), "/")
}
func opsActorFromRequest(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("X-Operator"))
	if value == "" {
		return "web"
	}
	return value
}
func opsNoStore(w http.ResponseWriter)    { w.Header().Set("Cache-Control", "no-store") }
func opsRequestID(r *http.Request) string { return r.Header.Get("X-Request-ID") }
