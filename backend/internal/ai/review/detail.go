package review

import (
	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
)

type Detail struct {
	Entry      Entry                `json:"entry"`
	Proposal   *results.Document    `json:"proposal"`
	Provenance *jobs.ResultMetadata `json:"provenance"`
}
