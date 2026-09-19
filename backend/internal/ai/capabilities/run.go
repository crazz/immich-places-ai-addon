package capabilities

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"
)

//go:embed fixtures/*.jpg
var fixtureFS embed.FS

var bundledFixtures = []struct {
	file  string
	color string
	shape string
}{
	{"fixtures/blue-circle.jpg", "blue", "circle"},
	{"fixtures/red-square.jpg", "red", "square"},
	{"fixtures/green-triangle.jpg", "green", "triangle"},
}

func LoadFixture(index int) (Fixture, error) {
	if index < 0 || index >= len(bundledFixtures) {
		return Fixture{}, fmt.Errorf("unknown fixture index %d", index)
	}
	item := bundledFixtures[index]
	raw, err := fixtureFS.ReadFile(item.file)
	if err != nil {
		return Fixture{}, err
	}
	return Fixture{Color: item.color, Shape: item.shape, JPEG: raw}, nil
}

func FixedSelector(index int) FixtureSelector {
	return func() Fixture {
		fixture, err := LoadFixture(index)
		if err != nil {
			panic(err)
		}
		return fixture
	}
}

func DefaultSelector() FixtureSelector {
	return FixedSelector(0)
}

type DispatchFunc func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error)

// AdmitFunc rechecks session and exact current revision before each probe dispatch.
type AdmitFunc func(ctx context.Context, ownerID, profileID string, revision int) error

type Runner struct {
	Dispatch DispatchFunc
	Admit    AdmitFunc
	Protocol ProbeProtocol
	Select   FixtureSelector
	Clock    Clock
}

type probeMode string

const (
	probeImage  probeMode = "image"
	probeJSON   probeMode = "json"
	probeStrict probeMode = "strict"
)

type probeOutcome struct {
	Observation   Observation
	ReportedModel string
	StopSequence  bool
	Dispatched    bool
	Err           error
}

func (r *Runner) Run(ctx context.Context, req ProbeRequest) (Report, error) {
	clock, fixture, imageURL, report, err := r.begin(req)
	if err != nil {
		return Report{}, err
	}

	if err := ctx.Err(); err != nil {
		return r.finishInterrupted(report, err, clock)
	}
	image := r.runProbe(ctx, req, fixture, imageURL, probeImage)
	r.recordProbe(&report, &report.Observations.Image, image)
	if image.Err != nil {
		return r.finishProbeError(report, req, clock, image.Err)
	}
	if image.StopSequence || image.Observation.Status != StatusSupported {
		return r.finish(report, req, clock, nil)
	}

	if err := ctx.Err(); err != nil {
		return r.finishInterrupted(report, err, clock)
	}
	jsonOut := r.runProbe(ctx, req, fixture, imageURL, probeJSON)
	r.recordProbe(&report, &report.Observations.JSON, jsonOut)
	if jsonOut.Err != nil {
		return r.finishProbeError(report, req, clock, jsonOut.Err)
	}
	if jsonOut.StopSequence {
		return r.finish(report, req, clock, nil)
	}

	if err := ctx.Err(); err != nil {
		return r.finishInterrupted(report, err, clock)
	}
	strict := r.runProbe(ctx, req, fixture, imageURL, probeStrict)
	r.recordProbe(&report, &report.Observations.Strict, strict)
	if strict.Err != nil {
		return r.finishProbeError(report, req, clock, strict.Err)
	}
	return r.finish(report, req, clock, nil)
}

func (r *Runner) recordProbe(report *Report, slot *Observation, out probeOutcome) {
	if out.Dispatched || out.Observation.Status != "" {
		*slot = out.Observation
	}
	if out.Dispatched {
		report.InputMayBeConsumed = true
	}
	if out.ReportedModel != "" {
		report.ReportedModel = &out.ReportedModel
	}
}

func (r *Runner) begin(req ProbeRequest) (Clock, Fixture, string, Report, error) {
	if r.Protocol == nil {
		return nil, Fixture{}, "", Report{}, errors.New("probe protocol is required")
	}
	if r.Dispatch == nil {
		return nil, Fixture{}, "", Report{}, errors.New("dispatch is required")
	}
	clock := r.Clock
	if clock == nil {
		clock = realClock{}
	}
	selectFixture := r.Select
	if selectFixture == nil {
		selectFixture = DefaultSelector()
	}
	fixture := selectFixture()
	report := Report{
		AttemptID:         req.AttemptID,
		ProfileID:         req.ProfileID,
		Revision:          req.Revision,
		ProtocolVersion:   ProtocolVersion,
		PolicyFingerprint: req.PolicyFingerprint,
		Lifecycle:         "failed",
		StartedAt:         req.StartedAt,
		DeadlineAt:        req.DeadlineAt,
		RequestedModel:    req.Model,
		Observations:      EmptyObservations(),
		Compatibility:     "failed",
		Applicable:        false,
	}
	return clock, fixture, DataURL(fixture.JPEG), report, nil
}

func (r *Runner) finishProbeError(report Report, req ProbeRequest, clock Clock, err error) (Report, error) {
	if errors.Is(err, ErrUnavailable) || errors.Is(err, ErrDisabled) ||
		errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return r.finishInterrupted(report, err, clock)
	}
	return r.finish(report, req, clock, err)
}

func (r *Runner) finishInterrupted(report Report, err error, clock Clock) (Report, error) {
	now := clock.Now()
	report.CompletedAt = &now
	report.Lifecycle = "canceled"
	reason := "canceled"
	if errors.Is(err, context.DeadlineExceeded) {
		report.Lifecycle = "failed"
		reason = "timeout"
	} else if errors.Is(err, ErrUnavailable) || errors.Is(err, ErrDisabled) {
		report.Lifecycle = "canceled"
		reason = "authority_revoked"
	}
	markNextUnverified(&report.Observations, reason)
	report.Compatibility = "failed"
	report.Applicable = false
	return report, err
}

func markNextUnverified(obs *Observations, reason string) {
	switch {
	case obs.Image.Status == StatusUnverified && obs.Image.Reason == "":
		obs.Image = Observation{Status: StatusUnverified, Reason: reason}
	case obs.JSON.Status == StatusUnverified && obs.JSON.Reason == "":
		obs.JSON = Observation{Status: StatusUnverified, Reason: reason}
	case obs.Strict.Status == StatusUnverified && obs.Strict.Reason == "":
		obs.Strict = Observation{Status: StatusUnverified, Reason: reason}
	}
}

func (r *Runner) finish(report Report, req ProbeRequest, clock Clock, err error) (Report, error) {
	now := clock.Now()
	report.CompletedAt = &now
	if err != nil {
		report.Lifecycle = "failed"
		report.Compatibility = "failed"
		report.Applicable = IsApplicable(report, req.Applicability)
		return report, err
	}
	report.Lifecycle = "completed"
	report.Compatibility = CompatibilitySummary(report.Observations)
	report.Applicable = IsApplicable(report, req.Applicability)
	return report, nil
}

func (r *Runner) runProbe(ctx context.Context, req ProbeRequest, fixture Fixture, imageURL string, mode probeMode) probeOutcome {
	body, err := r.encodeProbe(req.Model, imageURL, mode)
	if err != nil {
		return probeOutcome{StopSequence: true, Err: err}
	}
	if r.Admit != nil {
		if err := r.Admit(ctx, req.OwnerID, req.ProfileID, req.Revision); err != nil {
			return probeOutcome{StopSequence: true, Err: err}
		}
	}
	response, err := r.Dispatch(ctx, req.OwnerID, req.ProfileID, req.Revision, body)
	if err != nil {
		out := r.outcomeFromDispatchError(err, mode)
		out.Dispatched = true
		return out
	}
	text, reportedModel, err := r.Protocol.ParseAssistantText(response)
	if err != nil {
		return probeOutcome{
			Observation:   Observation{Status: StatusUnverified, Reason: ClassifyParseFailure(err)},
			ReportedModel: reportedModel,
			StopSequence:  true,
			Dispatched:    true,
		}
	}
	obs := r.evaluateProbe(fixture, text, mode)
	return probeOutcome{
		Observation:   obs,
		ReportedModel: reportedModel,
		StopSequence:  shouldStopAfterObservation(obs, mode),
		Dispatched:    true,
	}
}

func (r *Runner) encodeProbe(model, imageURL string, mode probeMode) ([]byte, error) {
	switch mode {
	case probeImage:
		return r.Protocol.EncodeImageProbe(model, ImageInstruction(), imageURL)
	case probeJSON:
		return r.Protocol.EncodeJSONProbe(model, JSONInstruction(), imageURL)
	case probeStrict:
		return r.Protocol.EncodeStrictProbe(model, StrictInstruction(), imageURL)
	default:
		return nil, fmt.Errorf("unknown probe mode %q", mode)
	}
}

func (r *Runner) evaluateProbe(fixture Fixture, text string, mode probeMode) Observation {
	switch mode {
	case probeImage:
		return EvaluateImageObservation(fixture, text)
	default:
		return EvaluateJSONObservation(fixture, text)
	}
}

func shouldStopAfterObservation(obs Observation, mode probeMode) bool {
	if mode == probeImage {
		return obs.Status != StatusSupported
	}
	return obs.Status != StatusSupported && obs.Status != StatusUnsupported
}

func (r *Runner) outcomeFromDispatchError(err error, mode probeMode) probeOutcome {
	obs, stop, mapped := MapDispatchFailure(err, string(mode))
	if mapped {
		if mode == probeImage {
			stop = true
		}
		return probeOutcome{
			Observation:  obs,
			StopSequence: stop,
			Err:          finishErrorAfterMappedFailure(obs, err),
		}
	}
	return probeOutcome{
		Observation:  Observation{Status: StatusUnverified, Reason: "provider_failure"},
		StopSequence: true,
		Err:          err,
	}
}

func SharedDeadline(now time.Time) time.Time {
	return now.Add(120 * time.Second)
}
