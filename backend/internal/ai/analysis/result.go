package analysis

import (
	"slices"

	"immich-places-backend/internal/ai/images"
	"immich-places-backend/internal/ai/results"
)

type Metadata struct {
	Image                                                                   images.Metadata
	ProfileID, Model, Format, PromptVersion, SchemaVersion, PrimaryLanguage string
	Revision                                                                int
	Languages                                                               []string
	Usage                                                                   *Usage
}
type Result struct {
	proposal results.Proposal
	metadata Metadata
}

func (r *Result) Proposal() results.Proposal {
	if r == nil {
		return results.Proposal{}
	}
	return r.proposal
}
func (r *Result) Info() Metadata {
	if r == nil {
		return Metadata{}
	}
	m := r.metadata
	m.Languages = slices.Clone(m.Languages)
	if m.Usage != nil {
		u := *m.Usage
		u.PromptTokens = copyCount(u.PromptTokens)
		u.CompletionTokens = copyCount(u.CompletionTokens)
		u.TotalTokens = copyCount(u.TotalTokens)
		m.Usage = &u
	}
	return m
}
func copyCount(v *int) *int {
	if v == nil {
		return nil
	}
	n := *v
	return &n
}

func newResult(proposal results.Proposal, metadata Metadata) *Result {
	result := &Result{proposal: proposal, metadata: metadata}
	result.metadata = result.Info()
	return result
}

func (Result) String() string   { return "validated Visual result" }
func (Result) GoString() string { return "validated Visual result" }
