package results

import "strings"

func languageFindings(document *Document, ctx validationContext) []Finding {
	findings := []Finding{}
	seen := map[string]bool{}
	for i := range document.Descriptions {
		description := &document.Descriptions[i]
		tag, valid := normalizeLanguage(description.Language)
		if !valid {
			findings = append(findings, Finding{Code: "language_tag", Path: itemPath("descriptions", i, "language")})
		} else {
			if seen[tag] || !ctx.languages[tag] {
				findings = append(findings, Finding{Code: "language_set", Path: itemPath("descriptions", i, "language")})
			}
			seen[tag] = true
			description.Language = tag
		}
		complete := description.Status == "complete" && description.Text != nil && strings.TrimSpace(*description.Text) != "" && description.UnavailableReason == nil
		unavailable := description.Status == "unavailable" && description.Text == nil && description.UnavailableReason != nil && strings.TrimSpace(*description.UnavailableReason) != ""
		if !complete && !unavailable {
			findings = append(findings, Finding{Code: "description_status", Path: itemPath("descriptions", i, "status")})
		}
		if (description.Basis == "candidate") != (description.CandidateID != nil) {
			findings = append(findings, Finding{Code: "description_basis", Path: itemPath("descriptions", i, "candidate_id")})
		}
	}
	for tag := range ctx.languages {
		if !seen[tag] {
			findings = append(findings, Finding{Code: "language_coverage", Path: "/descriptions"})
			break
		}
	}
	return findings
}
