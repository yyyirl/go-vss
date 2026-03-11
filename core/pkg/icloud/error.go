package icloud

import (
	"encoding/json"
	"fmt"
	"time"
)

//go:generate ../bin/stringer -type=ErrorCode -linecomment -output=error_string.go

// ErrorCode specifies the server error that occurred.
type ErrorCode uint8

// All available error codes.
const (
	Unknown                  ErrorCode = iota // UNKNOWN
	AccessDenied                              // ACCESS_DENIED
	AtomicError                               // ATOMIC_ERROR
	AuthenticationFailed                      // AUTHENTICATION_FAILED
	AuthenticationRequired                    // AUTHENTICATION_REQUIRED
	BadRequest                                // BAD_REQUEST
	Conflict                                  // CONFLICT
	Exists                                    // EXISTS
	InternalError                             // INTERNAL_ERROR
	NotFound                                  // NOT_FOUND
	QuotaExceeded                             // QUOTA_EXCEEDED
	Throttled                                 // THROTTLED
	TryAgainLater                             // TRY_AGAIN_LATER
	ValidatingReferenceError                  // VALIDATING_REFERENCE_ERROR
	ZoneNotFound                              // ZONE_NOT_FOUND
)

var errorCodeDescriptions = map[ErrorCode]string{
	Unknown:                  "An unknown error occurred.",
	AccessDenied:             "You don't have permission to access the endpoint, record, zone, or database.",
	AtomicError:              "An atomic batch operation failed.",
	AuthenticationFailed:     "Authentication was rejected.",
	AuthenticationRequired:   "The request requires authentication but none was provided.",
	BadRequest:               "The request was not valid.",
	Conflict:                 "The recordChangeTag value expired. (Retry the request with the latest tag.)",
	Exists:                   "The resource that you attempted to create already exists.",
	InternalError:            "An internal error occurred.",
	NotFound:                 "The resource was not found.",
	QuotaExceeded:            "If accessing the public database, you exceeded the app's quota. If accessing the private database, you exceeded the user's iCloud quota.",
	Throttled:                "The request was throttled. Try the request again later.",
	TryAgainLater:            "An internal error occurred. Try the request again.",
	ValidatingReferenceError: "The request violates a validating reference constraint.",
	ZoneNotFound:             "The zone specified in the request was not found.",
}

// Description of the error code.
func (ec ErrorCode) Description() string {
	return errorCodeDescriptions[ec]
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// ErrorCode from the string representation the server returns.
func (ec *ErrorCode) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	switch s {
	case AccessDenied.String():
		*ec = AccessDenied
	case AtomicError.String():
		*ec = AtomicError
	case AuthenticationFailed.String():
		*ec = AuthenticationFailed
	case AuthenticationRequired.String():
		*ec = AuthenticationRequired
	case BadRequest.String():
		*ec = BadRequest
	case Conflict.String():
		*ec = Conflict
	case Exists.String():
		*ec = Exists
	case InternalError.String():
		*ec = InternalError
	case NotFound.String():
		*ec = NotFound
	case QuotaExceeded.String():
		*ec = QuotaExceeded
	case Throttled.String():
		*ec = Throttled
	case TryAgainLater.String():
		*ec = TryAgainLater
	case ValidatingReferenceError.String():
		*ec = ValidatingReferenceError
	case ZoneNotFound.String():
		*ec = ZoneNotFound
	default:
		return fmt.Errorf("unknown error code %q", s)
	}

	return nil
}

// Error is the generic error response returned on non 2xx HTTP status codes.
type Error struct {
	// Reason for the error.
	Reason string `json:"reason"`
	// RetryAfter specifies the suggested time to wait before trying the
	// operation again. If not set, the operation can't be retried.
	RetryAfter time.Duration `json:"retryAfter"`
	// Code is the server error code.
	Code ErrorCode `json:"serverErrorCode"`
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("API error: %s, retry after %s", e.Reason, e.RetryAfter)
	}
	return fmt.Sprintf("API error: %s", e.Reason)
}

// UnmarshalJSON implements json.Unmarshaler. It is in place to unmarshal the
// some values returned by the server into proper Go types.
func (e *Error) UnmarshalJSON(b []byte) error {
	type LocalError Error
	localError := struct {
		*LocalError

		RetryAfter string `json:"retryAfter"`
	}{
		LocalError: (*LocalError)(e),
	}

	if err := json.Unmarshal(b, &localError); err != nil {
		return err
	}

	// If the "retry after" duration is not specified, parsing it is omitted.
	var err error
	if s := localError.RetryAfter; s != "" {
		e.RetryAfter, err = time.ParseDuration(s)
	}

	return err
}
