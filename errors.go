package smsgo

import (
	"errors"
	"fmt"
)

// FieldError is a per-field validation error item.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error is the standardized error returned by the SDK for non-2xx responses
// and transport failures. It implements the error interface.
type Error struct {
	// Status is the HTTP status code (0 for network/transport failures).
	Status int
	// Code is a stable error code (e.g. validation_error, insufficient_balance,
	// rate_limited). For transport failures it is "network_error".
	Code string
	// Message is a human-readable description.
	Message string
	// Details is the raw response body (parsed JSON, raw string, or nil).
	Details any
	// FieldErrors carries per-field detail on validation_error (422).
	FieldErrors []FieldError
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Status > 0 {
		return fmt.Sprintf("smsgo: %s (status %d, code %s)", e.Message, e.Status, e.Code)
	}
	return fmt.Sprintf("smsgo: %s (code %s)", e.Message, e.Code)
}

// AsError extracts a [*Error] from err, if the chain contains one.
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

// httpCodeName maps an HTTP status to a stable error code, mirroring the Node SDK.
func httpCodeName(status int) string {
	switch status {
	case 400:
		return "bad_request"
	case 401:
		return "unauthorized"
	case 402:
		return "insufficient_balance"
	case 409:
		return "provider_out_of_stock"
	case 422:
		return "validation_error"
	case 429:
		return "rate_limited"
	case 503:
		return "payment_unavailable"
	default:
		return fmt.Sprintf("http_%d", status)
	}
}
