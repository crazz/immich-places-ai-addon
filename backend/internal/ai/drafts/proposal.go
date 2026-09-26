package drafts

import "immich-places-backend/internal/ai/results"

func FromProposal(value Draft, document results.Document, choice *string) (Draft, error) {
	value.FactsRevision = 1
	value.HeadingFactsRevision = 1
	value.Descriptions = []Description{}
	for _, d := range document.Descriptions {
		stale := d.Basis == "candidate" && d.CandidateID != nil && choice != nil && *d.CandidateID != *choice
		value.Descriptions = append(value.Descriptions, Description{Language: d.Language, Status: d.Status, Text: d.Text, Basis: d.Basis, FactsRevision: 1, Stale: stale})
	}
	if choice == nil {
		choice = document.SelectedCandidateID
	}
	if choice == nil {
		if document.Outcome == "ambiguous" {
			return Draft{}, ErrInvalid
		}
		return value, nil
	}
	for _, candidate := range document.Candidates {
		if candidate.ID != *choice {
			continue
		}
		value.CandidateID = choice
		if candidate.CameraDirection != nil {
			heading, err := candidate.CameraDirection.AzimuthDeg.Float64()
			if err != nil {
				return Draft{}, ErrInvalid
			}
			value.Heading = &heading
			value.HeadingMethod = candidate.CameraDirection.Method
			if candidate.CameraDirection.UncertaintyDeg != nil {
				uncertainty, err := candidate.CameraDirection.UncertaintyDeg.Float64()
				if err != nil {
					return Draft{}, ErrInvalid
				}
				value.HeadingUncertainty = &uncertainty
			}
		}
		if candidate.CameraLocation != nil {
			lat, err := candidate.CameraLocation.Latitude.Float64()
			if err != nil {
				return Draft{}, ErrInvalid
			}
			lon, err := candidate.CameraLocation.Longitude.Float64()
			if err != nil {
				return Draft{}, ErrInvalid
			}
			value.Camera = &Point{Latitude: lat, Longitude: lon}
			value.RadiusBasis = candidate.CameraLocation.RadiusBasis
			if candidate.CameraLocation.EstimatedRadiusM != nil {
				radius, err := candidate.CameraLocation.EstimatedRadiusM.Float64()
				if err != nil {
					return Draft{}, ErrInvalid
				}
				value.Radius = &radius
			}
		}
		return value, nil
	}
	return Draft{}, ErrInvalid
}
