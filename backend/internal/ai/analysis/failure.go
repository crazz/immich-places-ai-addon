package analysis

import (
	"context"
	"errors"
)

var ErrInvalidRequest = errors.New("invalid Visual request")
var ErrDenied = errors.New("Visual authority unavailable")
var ErrBudget = errors.New("Visual dispatch budget unavailable")
var ErrUnsupportedFormat = errors.New("Visual response format unsupported")
var ErrUpstream = errors.New("Visual provider unavailable")
var ErrLimit = errors.New("Visual attempt limit exceeded")
var ErrInvalidResponse = errors.New("invalid or incomplete Visual response")

func dispatchFailure(err error) error {
	for _, safe := range []error{context.Canceled, context.DeadlineExceeded, ErrDenied, ErrBudget, ErrLimit, ErrUnsupportedFormat, ErrUpstream} {
		if errors.Is(err, safe) {
			return safe
		}
	}
	return ErrUpstream
}
