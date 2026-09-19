package capabilities

import (
	"encoding/base64"
	"time"
)

type Fixture struct {
	Color string
	Shape string
	JPEG  []byte
}

type FixtureSelector func() Fixture

type ProbeRequest struct {
	OwnerID           string
	ProfileID         string
	Revision          int
	Model             string
	PolicyFingerprint string
	AttemptID         string
	StartedAt         time.Time
	DeadlineAt        time.Time
	Applicability     ApplicabilityContext
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type ProbeProtocol interface {
	EncodeImageProbe(model, instruction, imageDataURL string) ([]byte, error)
	EncodeJSONProbe(model, instruction, imageDataURL string) ([]byte, error)
	EncodeStrictProbe(model, instruction, imageDataURL string) ([]byte, error)
	ParseAssistantText(body []byte) (text, reportedModel string, err error)
}

func ImageInstruction() string {
	return "Describe the single visible geometric shape. Reply with exactly two lines in this form and nothing else:\ncolor: <color>\nshape: <shape>"
}

func JSONInstruction() string {
	return "Describe the single visible geometric shape. Reply with one JSON object that has string fields color and shape only."
}

func StrictInstruction() string {
	return "Describe the single visible geometric shape. Reply with one JSON object that matches the requested schema with string fields color and shape only."
}

func DataURL(jpeg []byte) string {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpeg)
}
