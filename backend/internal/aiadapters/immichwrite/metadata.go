package immichwrite

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/writepreview"
)

func (t *Transport) SendMetadata(ctx context.Context, key string, plan writepreview.Plan) Outcome {
	asset, err := uuid.Parse(plan.TargetID)
	if err != nil || asset.String() != plan.TargetID || plan.Version != "mirror-preview-v4" || plan.Mirror == nil || !writepreview.ValidMirrorPlan(*plan.Mirror, plan.DraftRevision) {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	item := struct {
		Key   string          `json:"key"`
		Value json.RawMessage `json:"value"`
	}{plan.Mirror.Key, plan.Mirror.Value}
	raw, err := json.Marshal(struct {
		Items []any `json:"items"`
	}{[]any{item}})
	if err != nil {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, t.Endpoint+"/api/assets/"+plan.TargetID+"/metadata", io.NopCloser(strings.NewReader(string(raw))))
	if err != nil {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	req.Header.Set("x-api-key", key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "identity")
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: true, MaxResponseHeaderBytes: 16 << 10, TLSHandshakeTimeout: 10 * time.Second}
	defer transport.CloseIdleConnections()
	client := http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(req)
	if err != nil {
		return Outcome{Code: "outcome_unknown"}
	}
	defer response.Body.Close()
	n, err := io.Copy(io.Discard, io.LimitReader(response.Body, (64<<10)+1))
	if err != nil || n > 64<<10 {
		return Outcome{Code: "outcome_unknown"}
	}
	known := (response.StatusCode >= 200 && response.StatusCode < 300 && response.StatusCode != 202) || (response.StatusCode >= 400 && response.StatusCode < 500 && response.StatusCode != 408)
	return Outcome{CompletionKnown: known, Code: "response_received"}
}
