package drafts

import (
	"bytes"
	"encoding/json"
	"immich-places-backend/internal/ai/results"
	"math"
	"reflect"
)

type DescriptionEdit struct {
	Text   *string `json:"text"`
	Review bool    `json:"review"`
}

type Edit struct {
	Mirror            json.RawMessage            `json:"mirror"`
	PrimaryLanguage   *string                    `json:"primaryLanguage"`
	DescriptionPolicy *string                    `json:"descriptionPolicy"`
	CandidateID       *string                    `json:"candidateId"`
	Heading           json.RawMessage            `json:"heading"`
	ReviewHeading     bool                       `json:"reviewHeading"`
	ReviewPrecision   bool                       `json:"reviewPrecision"`
	Descriptions      map[string]DescriptionEdit `json:"descriptions"`
	State             string                     `json:"state"`
	Camera            json.RawMessage            `json:"camera"`
	Fields            json.RawMessage            `json:"fields"`
}

func ParsePoint(raw json.RawMessage) (*Point, error) {
	if bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	var input struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || input.Latitude == nil || input.Longitude == nil || !validNumber(*input.Latitude, 90) || !validNumber(*input.Longitude, 180) {
		return nil, ErrInvalid
	}
	return &Point{Latitude: *input.Latitude, Longitude: *input.Longitude}, nil
}
func validNumber(value, limit float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= -limit && value <= limit
}

func Apply(current Draft, edit Edit, document *results.Document) (Draft, error) {
	wasStaged := current.State == "staged"
	changed := false
	if len(edit.Camera) > 0 {
		point, err := ParsePoint(edit.Camera)
		if err != nil {
			return Draft{}, err
		}
		changed = !reflect.DeepEqual(point, current.Camera)
		current.Camera = point
	}
	if edit.CandidateID != nil {
		if document == nil || len(edit.Camera) > 0 {
			return Draft{}, ErrInvalid
		}
		selected, err := FromProposal(Draft{}, *document, edit.CandidateID)
		if err != nil {
			return Draft{}, err
		}
		changed = !reflect.DeepEqual(current.CandidateID, edit.CandidateID) || !reflect.DeepEqual(current.Camera, selected.Camera)
		current.Camera = selected.Camera
		current.CandidateID = edit.CandidateID
	} else if changed {
		current.CandidateID = nil
	}
	if changed {
		current.FactsRevision++
		current.HeadingStale = current.Heading != nil
		current.RadiusStale = current.Radius != nil
		current.Descriptions = append([]Description{}, current.Descriptions...)
		for i := range current.Descriptions {
			if current.Descriptions[i].Basis == "candidate" {
				current.Descriptions[i].Stale = true
			}
		}
	}

	if len(edit.Fields) > 0 {
		var fields []string
		if json.Unmarshal(edit.Fields, &fields) != nil || !validFields(fields) {
			return Draft{}, ErrInvalid
		}
		changed = changed || !reflect.DeepEqual(current.Fields, fields)
		current.Fields = fields
	}
	if edit.PrimaryLanguage != nil {
		changed = changed || current.PrimaryLanguage != *edit.PrimaryLanguage
		current.PrimaryLanguage = *edit.PrimaryLanguage
	}
	if edit.DescriptionPolicy != nil {
		if *edit.DescriptionPolicy != "preserve" && *edit.DescriptionPolicy != "replace" && *edit.DescriptionPolicy != "managed_append" {
			return Draft{}, ErrInvalid
		}
		changed = changed || current.DescriptionPolicy != *edit.DescriptionPolicy
		current.DescriptionPolicy = *edit.DescriptionPolicy
	}
	if len(edit.Heading) > 0 || edit.ReviewHeading {
		if len(edit.Heading) > 0 {
			var heading *float64
			if json.Unmarshal(edit.Heading, &heading) != nil || (heading != nil && (!validNumber(*heading, 360) || *heading < 0 || *heading >= 360)) {
				return Draft{}, ErrInvalid
			}
			current.Heading = heading
			current.HeadingUserSupplied = true
		}
		current.HeadingStale = false
		current.HeadingFactsRevision = current.FactsRevision
		changed = true
	}
	if len(edit.Descriptions) > 0 {
		current.Descriptions = append([]Description{}, current.Descriptions...)
		for language, correction := range edit.Descriptions {
			found := false
			for i := range current.Descriptions {
				d := &current.Descriptions[i]
				if d.Language != language {
					continue
				}
				found = true
				if correction.Text != nil {
					if len(*correction.Text) > 16<<10 {
						return Draft{}, ErrInvalid
					}
					d.Text = correction.Text
					d.Status = "complete"
					d.UserSupplied = true
				} else if !correction.Review || d.Status != "complete" {
					return Draft{}, ErrInvalid
				}
				d.Stale = false
				d.FactsRevision = current.FactsRevision
			}
			if !found {
				return Draft{}, ErrInvalid
			}
		}
		changed = true
	}
	if edit.ReviewPrecision {
		current.RadiusStale = false
		changed = true
	}
	if len(edit.Mirror) > 0 {
		selection, err := parseMirrorSelection(edit.Mirror)
		if err != nil {
			return Draft{}, err
		}
		changed = changed || !reflect.DeepEqual(selection, current.Mirror)
		current.Mirror = selection
	}
	if changed && current.State == "staged" {
		current.State = "draft"
	}
	switch edit.State {
	case "":
	case "draft", "rejected":
		current.State = edit.State
	case "staged":
		if !Ready(current) {
			return Draft{}, ErrInvalid
		}
		if !wasStaged || !changed {
			current.State = "staged"
		}
	default:
		return Draft{}, ErrInvalid
	}
	current.Revision++
	return current, nil
}
