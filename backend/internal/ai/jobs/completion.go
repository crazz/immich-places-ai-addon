package jobs

import "immich-places-backend/internal/ai/results"

type Completion struct {
	Proposal                                                results.Proposal
	SourceDigest, ImageDigest, PromptVersion, SchemaVersion string
}

type ResultMetadata struct {
	Installation, Profile, Model                                               string
	Revision                                                                   int
	Languages                                                                  []string
	PrimaryLanguage, SelectionDigest                                           string
	SourceDigest, ImageDigest, PromptVersion, SchemaVersion, ValidationVersion string
}

type AnalysisRecord struct {
	ID, JobID, ItemID, Asset, Outcome string
	CreatedAt                         int64
	Payload                           []byte
	Metadata                          ResultMetadata
}

func (AnalysisRecord) String() string   { return "stored AI analysis" }
func (AnalysisRecord) GoString() string { return "stored AI analysis" }
