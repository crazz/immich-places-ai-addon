package capabilities

import (
	"errors"
)

var (
	ErrModeUnsupported  = errors.New("response mode unsupported")
	ErrAuthentication   = errors.New("authentication failure")
	ErrUnavailableModel = errors.New("unavailable model")
	ErrRateLimit        = errors.New("rate limit")
	ErrPolicy           = errors.New("policy failure")
	ErrNetwork          = errors.New("network failure")
	ErrServer           = errors.New("server failure")
	ErrRefusal          = errors.New("capability response refused")
	ErrToolResponse     = errors.New("capability response requested tools")
	ErrTruncation       = errors.New("capability response finish reason is not successful")
)

type ModeUnsupportedError struct {
	Mode string
}

func (e *ModeUnsupportedError) Error() string {
	return "response mode unsupported"
}

func (e *ModeUnsupportedError) Is(target error) bool {
	return target == ErrModeUnsupported
}

func MapDispatchFailure(err error, mode string) (obs Observation, stop bool, mapped bool) {
	var unsupported *ModeUnsupportedError
	if errors.As(err, &unsupported) {
		if unsupported.Mode == "" {
			unsupported.Mode = mode
		}
		return Observation{Status: StatusUnsupported, Reason: "unsupported_mode"}, false, true
	}
	if errors.Is(err, ErrModeUnsupported) {
		return Observation{Status: StatusUnsupported, Reason: "unsupported_mode"}, false, true
	}
	switch {
	case errors.Is(err, ErrAuthentication):
		return Observation{Status: StatusUnverified, Reason: "authentication"}, true, true
	case errors.Is(err, ErrUnavailableModel):
		return Observation{Status: StatusUnverified, Reason: "unavailable_model"}, true, true
	case errors.Is(err, ErrRateLimit):
		return Observation{Status: StatusUnverified, Reason: "rate_limit"}, true, true
	case errors.Is(err, ErrPolicy):
		return Observation{Status: StatusUnverified, Reason: "policy"}, true, true
	case errors.Is(err, ErrNetwork):
		return Observation{Status: StatusUnverified, Reason: "network"}, true, true
	case errors.Is(err, ErrServer):
		return Observation{Status: StatusUnverified, Reason: "server"}, true, true
	default:
		return Observation{}, false, false
	}
}

// ClassifyParseFailure maps typed assistant-parse failures to safe observation reasons.
func ClassifyParseFailure(err error) string {
	switch {
	case errors.Is(err, ErrToolResponse):
		return "tool_response"
	case errors.Is(err, ErrRefusal):
		return "refusal"
	case errors.Is(err, ErrTruncation):
		return "truncation"
	default:
		return "invalid_output"
	}
}

func isSystemicReason(reason string) bool {
	switch reason {
	case "authentication", "unavailable_model", "rate_limit", "policy", "network", "server":
		return true
	default:
		return false
	}
}

// finishErrorAfterMappedFailure decides whether a mapped probe failure ends the Run with an
// error or completes with the observation already recorded.
//
// - unsupported mode: continue/finish without error (caller may still run later probes)
// - policy: rethrow so the report ends as a failed lifecycle rather than an ordinary observation
// - other systemic categories: finish without error (report carries the reason)
// - anything else: rethrow
func finishErrorAfterMappedFailure(obs Observation, err error) error {
	if obs.Status == StatusUnsupported {
		return nil
	}
	if obs.Reason == "policy" {
		return err
	}
	if obs.Status == StatusUnverified && isSystemicReason(obs.Reason) {
		return nil
	}
	return err
}
