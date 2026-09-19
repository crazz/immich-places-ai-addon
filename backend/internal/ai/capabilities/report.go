package capabilities

import "time"

const ProtocolVersion = "capability-v2"

type ObservationStatus string

const (
	StatusSupported   ObservationStatus = "supported"
	StatusUnsupported ObservationStatus = "unsupported"
	StatusUnverified  ObservationStatus = "unverified"
)

type Observation struct {
	Status ObservationStatus `json:"status"`
	Reason string            `json:"reason,omitempty"`
}

type Observations struct {
	Image  Observation `json:"image"`
	JSON   Observation `json:"json"`
	Strict Observation `json:"strict"`
}

type Report struct {
	AttemptID          string       `json:"attemptID"`
	ProfileID          string       `json:"profileID"`
	Revision           int          `json:"revision"`
	ProtocolVersion    string       `json:"protocolVersion"`
	PolicyFingerprint  string       `json:"policyFingerprint"`
	Lifecycle          string       `json:"lifecycle"`
	StartedAt          time.Time    `json:"startedAt"`
	DeadlineAt         time.Time    `json:"deadlineAt"`
	CompletedAt        *time.Time   `json:"completedAt,omitempty"`
	RequestedModel     string       `json:"requestedModel"`
	ReportedModel      *string      `json:"reportedModel,omitempty"`
	Observations       Observations `json:"observations"`
	Compatibility      string       `json:"compatibility"`
	Applicable         bool         `json:"applicable"`
	InputMayBeConsumed bool         `json:"inputMayBeConsumed,omitempty"`
	Usage              *Usage       `json:"usage,omitempty"`
	ProviderSizeLimit  *int         `json:"providerSizeLimit,omitempty"`
	TokenLimitSupport  *bool        `json:"tokenLimitSupport,omitempty"`
}

type Usage struct {
	PromptTokens     *int `json:"promptTokens,omitempty"`
	CompletionTokens *int `json:"completionTokens,omitempty"`
	TotalTokens      *int `json:"totalTokens,omitempty"`
}

func EmptyObservations() Observations {
	return Observations{
		Image:  Observation{Status: StatusUnverified},
		JSON:   Observation{Status: StatusUnverified},
		Strict: Observation{Status: StatusUnverified},
	}
}

func CompatibilitySummary(obs Observations) string {
	if hasSystemicFailure(obs) {
		return "failed"
	}
	if obs.Image.Status == StatusUnsupported {
		return "unsupported"
	}
	if obs.Image.Status == StatusSupported &&
		obs.JSON.Status == StatusUnsupported && obs.Strict.Status == StatusUnsupported {
		return "unsupported"
	}
	if obs.Image.Status == StatusSupported && obs.Strict.Status == StatusSupported &&
		(obs.JSON.Status == StatusSupported || obs.JSON.Status == StatusUnsupported) {
		return "strict-schema sample compatible"
	}
	if obs.Image.Status == StatusSupported && obs.JSON.Status == StatusSupported &&
		obs.Strict.Status == StatusUnsupported {
		return "json-only compatible"
	}
	if obs.Image.Status == StatusSupported {
		return "incomplete"
	}
	return "incomplete"
}

func hasSystemicFailure(obs Observations) bool {
	return isSystemicReason(obs.Image.Reason) || isSystemicReason(obs.JSON.Reason) || isSystemicReason(obs.Strict.Reason)
}
