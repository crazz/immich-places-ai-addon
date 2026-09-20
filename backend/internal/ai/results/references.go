package results

import (
	"strconv"
	"strings"
)

func itemPath(collection string, index int, field string) string {
	return "/" + collection + "/" + strconv.Itoa(index) + "/" + field
}

func referenceFindings(document Document, ctx validationContext) []Finding {
	findings := []Finding{}
	add := func(code, path string) { findings = append(findings, Finding{Code: code, Path: path}) }
	observations := map[string]bool{}
	provided := map[string]bool{}
	for i, observation := range document.Observations {
		if !validID(observation.ID) || observations[observation.ID] {
			add("observation_id", itemPath("observations", i, "id"))
		}
		if strings.TrimSpace(observation.Text) == "" {
			add("blank_text", itemPath("observations", i, "text"))
		}
		observations[observation.ID] = true
		if observation.Kind == "provided_context" {
			provided[observation.ID] = true
			if ctx.mode != ContextAssisted || len(ctx.sources) == 0 {
				add("context_provenance", itemPath("observations", i, "kind"))
			}
		}
	}
	candidates := map[string]bool{}
	for i, candidate := range document.Candidates {
		if !validID(candidate.ID) || candidates[candidate.ID] {
			add("candidate_id", itemPath("candidates", i, "id"))
		}
		candidates[candidate.ID] = true
		if candidate.CountryCode != nil && !validCountry(*candidate.CountryCode) {
			add("country_code", itemPath("candidates", i, "country_code"))
		}
		if strings.TrimSpace(candidate.PlaceName) == "" {
			add("blank_text", itemPath("candidates", i, "place_name"))
		}
		if strings.TrimSpace(candidate.SupportSummary) == "" {
			add("blank_text", itemPath("candidates", i, "support_summary"))
		}
		sourceSeen := map[string]bool{}
		authorizedSources := 0
		for j, ref := range candidate.SourceRefs {
			_, authorized := ctx.sources[ref]
			if authorized {
				authorizedSources++
			}
			if !validID(ref) || !authorized || sourceSeen[ref] {
				add("source_reference", itemPath("candidates", i, "source_refs")+"/"+strconv.Itoa(j))
			}
			sourceSeen[ref] = true
		}
		seen := map[string]bool{}
		for j, ref := range candidate.EvidenceRefs {
			if !validID(ref) || !observations[ref] || seen[ref] {
				add("evidence_reference", itemPath("candidates", i, "evidence_refs")+"/"+strconv.Itoa(j))
			}
			seen[ref] = true
			if provided[ref] && authorizedSources == 0 {
				add("context_source", itemPath("candidates", i, "source_refs"))
			}
		}
	}
	for i, description := range document.Descriptions {
		if description.CandidateID != nil && !candidates[*description.CandidateID] {
			add("description_reference", itemPath("descriptions", i, "candidate_id"))
		}
	}
	return findings
}
