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

type Transport struct{ Endpoint string }
type Outcome struct {
	CompletionKnown bool
	Code            string
}

func (t *Transport) Send(ctx context.Context, key, asset string, point writepreview.Point) Outcome {
	if _, err := uuid.Parse(asset); err != nil {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	raw, err := json.Marshal(point)
	if err != nil {
		return Outcome{CompletionKnown: true, Code: "not_sent"}
	}
	return t.send(ctx, key, asset, raw)
}

func (t *Transport) send(ctx context.Context, key, asset string, raw []byte) Outcome {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, t.Endpoint+"/api/assets/"+asset, io.NopCloser(strings.NewReader(string(raw))))
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
	if _, err = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10)); err != nil {
		return Outcome{Code: "outcome_unknown"}
	}
	completed := (response.StatusCode >= 200 && response.StatusCode < 300 && response.StatusCode != 202) || (response.StatusCode >= 400 && response.StatusCode < 500 && response.StatusCode != 408)
	return Outcome{CompletionKnown: completed, Code: "response_received"}
}
