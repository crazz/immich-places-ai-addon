package writepreview

import (
	"encoding/json"

	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/results"
)

type MirrorProvenance struct {
	Mode  string
	Model string
}

type MirrorExport struct {
	SchemaVersion int                 `json:"schemaVersion"`
	Application   string              `json:"application"`
	RecordID      string              `json:"recordId"`
	Review        MirrorReview        `json:"review"`
	Direction     *MirrorDirection    `json:"direction,omitempty"`
	Precision     *MirrorPrecision    `json:"precision,omitempty"`
	Place         *string             `json:"place,omitempty"`
	Descriptions  map[string]string   `json:"descriptions,omitempty"`
	Provenance    *MirrorExportOrigin `json:"provenance,omitempty"`
}

type MirrorReview struct {
	DraftRevision int `json:"draftRevision"`
	FactsRevision int `json:"factsRevision"`
}

type MirrorDirection struct {
	Heading     *float64 `json:"heading"`
	Method      string   `json:"method"`
	Uncertainty *float64 `json:"uncertainty"`
}

type MirrorPrecision struct {
	Radius *float64 `json:"radius"`
	Basis  string   `json:"basis"`
}

type MirrorExportOrigin struct {
	Mode       string            `json:"mode"`
	Model      string            `json:"model,omitempty"`
	ReviewedAt string            `json:"reviewedAt"`
	Fields     map[string]string `json:"fields"`
}

func BuildMirrorExport(d drafts.Draft, document results.Document, choice drafts.MirrorSelection, provenance MirrorProvenance, recordID string) ([]byte, error) {
	if err := validateMirrorExport(d, document, choice, provenance, recordID); err != nil {
		return nil, err
	}
	value := MirrorExport{SchemaVersion: 1, Application: MirrorNamespace, RecordID: recordID, Review: MirrorReview{d.Revision, d.FactsRevision}}
	origins := map[string]string{}
	var candidate results.Candidate
	for _, item := range document.Candidates {
		if d.CandidateID != nil && item.ID == *d.CandidateID {
			candidate = item
		}
	}
	if choice.Direction {
		direction := &MirrorDirection{Heading: d.Heading, Method: "unknown"}
		origins["direction"] = "model"
		if d.Heading != nil && d.HeadingUserSupplied {
			direction.Method = "user_supplied"
			origins["direction"] = "user"
		} else if d.Heading != nil && d.HeadingMethod != "" {
			direction.Method = d.HeadingMethod
			direction.Uncertainty = d.HeadingUncertainty
		}
		value.Direction = direction
	}
	if choice.Precision {
		value.Precision = &MirrorPrecision{d.Radius, d.RadiusBasis}
		origins["precision"] = "model"
	}
	if choice.Place {
		value.Place = &candidate.PlaceName
		origins["place"] = "model"
	}
	if len(choice.Languages) > 0 {
		value.Descriptions = map[string]string{}
		for _, language := range choice.Languages {
			for _, description := range d.Descriptions {
				if description.Language == language && description.Text != nil {
					value.Descriptions[language] = *description.Text
					origins["description:"+language] = "model"
					if description.UserSupplied {
						origins["description:"+language] = "user"
					}
				}
			}
		}
	}
	if choice.Provenance {
		value.Provenance = &MirrorExportOrigin{Mode: provenance.Mode, ReviewedAt: d.UpdatedAt, Fields: origins}
		if choice.Model {
			value.Provenance.Model = provenance.Model
		}
	}
	raw, err := json.Marshal(value)
	if err != nil || len(raw) > MirrorValueLimit {
		return nil, Failure{Code: "MIRROR_TOO_LARGE"}
	}
	return raw, nil
}
