package results

func outcomeFindings(document Document) []Finding {
	valid := true
	switch document.Outcome {
	case "located":
		valid = false
		if document.SelectedCandidateID != nil {
			for _, candidate := range document.Candidates {
				if candidate.ID == *document.SelectedCandidateID && candidate.CameraLocation != nil {
					valid = true
					break
				}
			}
		}
	case "ambiguous":
		valid = len(document.Candidates) >= 2 && document.SelectedCandidateID == nil
	case "unknown":
		valid = document.SelectedCandidateID == nil
	}
	if !valid {
		return []Finding{{Code: "outcome_selection", Path: "/selected_candidate_id"}}
	}
	return nil
}
