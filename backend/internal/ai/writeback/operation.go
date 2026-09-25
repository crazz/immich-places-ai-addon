package writeback

import "immich-places-backend/internal/ai/writepreview"

type Confirmation struct {
	PreviewID string `json:"previewId"`
	Digest    string `json:"digest"`
	Key       string `json:"idempotencyKey"`
}

type Event struct {
	Code    string `json:"code"`
	At      string `json:"at"`
	Attempt int    `json:"attempt"`
}

type Operation struct {
	ID         string            `json:"id"`
	Plan       writepreview.Plan `json:"plan"`
	Digest     string            `json:"digest"`
	Status     string            `json:"status"`
	Code       string            `json:"code"`
	ApprovedAt string            `json:"approvedAt"`
	Attempts   int               `json:"attempts"`
	Generation int               `json:"generation"`
	Observed   *writepreview.GPS `json:"observed"`
	Verified   bool              `json:"verified"`
	Refreshed  bool              `json:"refreshed"`
	Noop       bool              `json:"noop"`
	Settled    bool              `json:"settled"`
	Events     []Event           `json:"events"`
}

type Failure string

func (f Failure) Error() string { return string(f) }
