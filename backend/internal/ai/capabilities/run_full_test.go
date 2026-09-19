package capabilities_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"immich-places-backend/internal/ai/capabilities"
)

type recordingProtocol struct {
	imageBodies  []string
	jsonBodies   []string
	strictBodies []string
}

func (p *recordingProtocol) EncodeImageProbe(model, instruction, imageDataURL string) ([]byte, error) {
	body := `{"model":"` + model + `","mode":"image","instruction":` + quote(instruction) + `,"image":true}`
	p.imageBodies = append(p.imageBodies, body)
	return []byte(body), nil
}

func (p *recordingProtocol) EncodeJSONProbe(model, instruction, imageDataURL string) ([]byte, error) {
	body := `{"model":"` + model + `","mode":"json","response_format":{"type":"json_object"}}`
	p.jsonBodies = append(p.jsonBodies, body)
	return []byte(body), nil
}

func (p *recordingProtocol) EncodeStrictProbe(model, instruction, imageDataURL string) ([]byte, error) {
	body := `{"model":"` + model + `","mode":"strict","response_format":{"type":"json_schema","strict":true}}`
	p.strictBodies = append(p.strictBodies, body)
	return []byte(body), nil
}

func (p *recordingProtocol) ParseAssistantText(body []byte) (string, string, error) {
	text := string(body)
	if strings.HasPrefix(text, "TEXT:") {
		return strings.TrimPrefix(text, "TEXT:"), "reported-model", nil
	}
	if strings.HasPrefix(text, "JSON:") {
		return strings.TrimPrefix(text, "JSON:"), "reported-model", nil
	}
	return "", "", nil
}

func quote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `'`) + `"`
}

func TestRunObservesFullCompatibility(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			hits.Add(1)
			n := hits.Load()
			switch n {
			case 1:
				return []byte("TEXT:color: blue\nshape: circle"), nil
			case 2, 3:
				return []byte(`JSON:{"color":"blue","shape":"circle"}`), nil
			default:
				t.Fatalf("unexpected dispatch %d", n)
				return nil, nil
			}
		},
	}
	report, err := runner.Run(context.Background(), capabilities.ProbeRequest{
		OwnerID:           "user-1",
		ProfileID:         "profile-1",
		Revision:          2,
		Model:             "gpt-5.6-sol",
		PolicyFingerprint: "fp",
		AttemptID:         "attempt-1",
		StartedAt:         time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC),
		DeadlineAt:        time.Date(2026, 9, 19, 12, 2, 0, 0, time.UTC),
		Applicability: capabilities.ApplicabilityContext{
			AIEnabled: true, ProfileEnabled: true, ActiveRevision: 2,
			CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: "fp",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 3 {
		t.Fatalf("hits = %d, want 3", hits.Load())
	}
	if len(protocol.imageBodies) != 1 || len(protocol.jsonBodies) != 1 || len(protocol.strictBodies) != 1 {
		t.Fatalf("probe encodings image=%d json=%d strict=%d", len(protocol.imageBodies), len(protocol.jsonBodies), len(protocol.strictBodies))
	}
	if report.Observations.Image.Status != capabilities.StatusSupported ||
		report.Observations.JSON.Status != capabilities.StatusSupported ||
		report.Observations.Strict.Status != capabilities.StatusSupported {
		t.Fatalf("observations = %+v", report.Observations)
	}
	if report.Compatibility != "strict-schema sample compatible" {
		t.Fatalf("compatibility = %q", report.Compatibility)
	}
	serialized, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	lowered := strings.ToLower(string(serialized))
	if strings.Contains(lowered, "geolocation") || strings.Contains(lowered, "guarantee") {
		t.Fatalf("report must not claim geolocation accuracy or guaranteed schema enforcement: %s", serialized)
	}
	if report.Usage != nil || report.ProviderSizeLimit != nil || report.TokenLimitSupport != nil {
		t.Fatalf("omitted provider metadata must stay unknown: usage=%v size=%v token=%v", report.Usage, report.ProviderSizeLimit, report.TokenLimitSupport)
	}
}

func TestRunPreservesJSONOnlyCompatibility(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			switch n {
			case 1:
				return []byte("TEXT:color: blue\nshape: circle"), nil
			case 2:
				return []byte(`JSON:{"color":"blue","shape":"circle"}`), nil
			case 3:
				return nil, &capabilities.ModeUnsupportedError{Mode: "strict"}
			default:
				t.Fatalf("unexpected dispatch %d body=%s", n, body)
				return nil, nil
			}
		},
	}
	report, err := runner.Run(context.Background(), baseProbeRequest())
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 3 {
		t.Fatalf("hits = %d, want 3 (no replacement request)", hits.Load())
	}
	if report.Observations.Image.Status != capabilities.StatusSupported ||
		report.Observations.JSON.Status != capabilities.StatusSupported ||
		report.Observations.Strict.Status != capabilities.StatusUnsupported {
		t.Fatalf("observations = %+v", report.Observations)
	}
	if report.Compatibility != "json-only compatible" {
		t.Fatalf("compatibility = %q", report.Compatibility)
	}
	if report.RequestedModel != "gpt-5.6-sol" {
		t.Fatalf("model changed to %q", report.RequestedModel)
	}
}

func TestRunStrictSchemaWithoutJSONMode(t *testing.T) {
	protocol := &recordingProtocol{}
	var hits atomic.Int32
	runner := &capabilities.Runner{
		Protocol: protocol,
		Select:   capabilities.FixedSelector(0),
		Clock:    fixedClock{now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)},
		Dispatch: func(ctx context.Context, ownerID, profileID string, revision int, body []byte) ([]byte, error) {
			n := hits.Add(1)
			switch n {
			case 1:
				return []byte("TEXT:color: blue\nshape: circle"), nil
			case 2:
				return nil, &capabilities.ModeUnsupportedError{Mode: "json"}
			case 3:
				return []byte(`JSON:{"color":"blue","shape":"circle"}`), nil
			default:
				t.Fatalf("unexpected dispatch %d", n)
				return nil, nil
			}
		},
	}
	report, err := runner.Run(context.Background(), baseProbeRequest())
	if err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 3 {
		t.Fatalf("hits = %d, want 3", hits.Load())
	}
	if report.Observations.Image.Status != capabilities.StatusSupported ||
		report.Observations.JSON.Status != capabilities.StatusUnsupported ||
		report.Observations.Strict.Status != capabilities.StatusSupported {
		t.Fatalf("observations = %+v", report.Observations)
	}
	if report.Compatibility != "strict-schema sample compatible" {
		t.Fatalf("compatibility = %q", report.Compatibility)
	}
}

func baseProbeRequest() capabilities.ProbeRequest {
	return capabilities.ProbeRequest{
		OwnerID:           "user-1",
		ProfileID:         "profile-1",
		Revision:          2,
		Model:             "gpt-5.6-sol",
		PolicyFingerprint: "fp",
		AttemptID:         "attempt-1",
		StartedAt:         time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC),
		DeadlineAt:        time.Date(2026, 9, 19, 12, 2, 0, 0, time.UTC),
		Applicability: capabilities.ApplicabilityContext{
			AIEnabled: true, ProfileEnabled: true, ActiveRevision: 2,
			CurrentProtocolVersion: capabilities.ProtocolVersion, CurrentPolicyFingerprint: "fp",
		},
	}
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }
