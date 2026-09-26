package writeback

import "immich-places-backend/internal/ai/writepreview"

type MirrorOutcome struct {
	AssetID    string                       `json:"assetId"`
	Step       string                       `json:"step"`
	Status     string                       `json:"status"`
	Code       string                       `json:"code"`
	Attempts   int                          `json:"attempts"`
	Generation int                          `json:"generation"`
	Observed   *writepreview.MirrorBaseline `json:"observed"`
	Verified   bool                         `json:"verified"`
	Noop       bool                         `json:"noop"`
	Settled    bool                         `json:"settled"`
	Events     []MirrorEvent                `json:"events"`
}

type MirrorEvent struct {
	Code       string `json:"code"`
	At         string `json:"at"`
	Attempt    int    `json:"attempt"`
	Generation int    `json:"generation"`
}
