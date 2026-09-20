package results

import (
	"encoding/json"
	"sort"
)

func semanticFailure(findings []Finding) error {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Code < findings[j].Code
	})
	result := &Failure{Category: "semantic_violation", Findings: []Finding{}}
	for _, finding := range findings {
		if len(result.Findings) == 20 {
			break
		}
		if len(result.Findings) > 0 && result.Findings[len(result.Findings)-1] == finding {
			continue
		}
		result.Findings = append(result.Findings, finding)
		encoded, err := json.Marshal(result)
		if err != nil || len(encoded) > 4096 {
			result.Findings = result.Findings[:len(result.Findings)-1]
			break
		}
	}
	return result
}
