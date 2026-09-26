package writeback

import "immich-places-backend/internal/ai/writepreview"

type TargetOutcome struct {
	AssetID    string            `json:"assetId"`
	Status     string            `json:"status"`
	Code       string            `json:"code"`
	Attempts   int               `json:"attempts"`
	Generation int               `json:"generation"`
	Observed   *writepreview.GPS `json:"observed"`
	Verified   bool              `json:"verified"`
	Refreshed  bool              `json:"refreshed"`
	Noop       bool              `json:"noop"`
	Settled    bool              `json:"settled"`
	Fields     []FieldOutcome    `json:"fields"`
	Events     []Event           `json:"events"`
}
