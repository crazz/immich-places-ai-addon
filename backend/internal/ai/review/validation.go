package review

import (
	"encoding/hex"
	"slices"

	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
)

func ValidateRecord(payload []byte, m jobs.ResultMetadata, job jobs.Job, asset, outcome string) (results.Document, error) {
	mode := job.Input.Mode
	if mode == "" {
		mode = "visual"
	}
	if m.SchemaVersion != "1.0" || m.ValidationVersion != "analysis-result-v1" || m.Installation != job.Input.Installation || m.Profile != job.Input.Profile || m.Revision != job.Input.Revision || m.Model != job.Model || m.SelectionDigest != job.Input.SelectionDigest || !slices.Equal(m.Languages, job.Input.Languages) || m.PrimaryLanguage != job.Input.PrimaryLanguage {
		return results.Document{}, ErrUnavailable
	}
	if m.Mode != "" && string(m.Mode) != mode {
		return results.Document{}, ErrUnavailable
	}
	for _, digest := range []string{m.SourceDigest, m.ImageDigest, m.SelectionDigest} {
		data, err := hex.DecodeString(digest)
		if err != nil || len(data) != 32 {
			return results.Document{}, ErrUnavailable
		}
	}
	ctx := results.Context{Mode: results.Mode(mode), Completion: results.Complete, Languages: m.Languages, PrimaryLanguage: m.PrimaryLanguage}
	switch mode {
	case "visual":
		if m.PromptVersion != "visual-v1" || m.Context != nil {
			return results.Document{}, ErrUnavailable
		}
	case "context-assisted":
		if m.PromptVersion != "context-assisted-v1" || m.Context == nil || m.Mode != results.ContextAssisted {
			return results.Document{}, ErrUnavailable
		}
		binding := m.Context.Binding
		if binding.Owner != job.Input.Owner || binding.Installation != m.Installation || binding.Asset != asset || binding.Selection != m.SelectionDigest || binding.Profile != m.Profile || binding.Revision != m.Revision || binding.SourceDigest != m.SourceDigest {
			return results.Document{}, ErrUnavailable
		}
		for _, source := range m.Context.Sources {
			ctx.Sources = append(ctx.Sources, results.Source{ID: source.ID})
		}
	default:
		return results.Document{}, ErrUnavailable
	}
	v, err := results.New()
	if err != nil {
		return results.Document{}, ErrUnavailable
	}
	proposal, err := v.Validate(payload, ctx)
	if err != nil {
		return results.Document{}, ErrUnavailable
	}
	document, err := proposal.Data()
	if err != nil || document.Outcome != outcome {
		return results.Document{}, ErrUnavailable
	}
	return document, nil
}
