package providers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

type FailureCategory string

const (
	FailureLimit      FailureCategory = "LIMIT"
	FailureTimeout    FailureCategory = "TIMEOUT"
	FailureCanceled   FailureCategory = "CANCELED"
	FailureUpstream   FailureCategory = "UPSTREAM"
	FailureConnection FailureCategory = "CONNECTION"
	FailureEncoding   FailureCategory = "ENCODING"
)

func NewRequestID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(buf[:])
}

type TransportFailure struct {
	Category  FailureCategory
	RequestID string
	Status    int
	Code      string
	cause     error
}

func (f *TransportFailure) Error() string {
	if f == nil {
		return "provider transport failure"
	}
	return fmt.Sprintf("provider %s failure", f.Category)
}

func (f *TransportFailure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.cause
}

func AsTransportFailure(err error) (*TransportFailure, bool) {
	var failure *TransportFailure
	if errors.As(err, &failure) {
		return failure, true
	}
	return nil, false
}

func NewTransportFailure(category FailureCategory, requestID string, status int, cause error) *TransportFailure {
	return &TransportFailure{
		Category:  category,
		RequestID: requestID,
		Status:    status,
		cause:     cause,
	}
}
