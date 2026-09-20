package results

import "encoding/json"

type Document struct {
	SchemaVersion       string        `json:"schema_version"`
	Outcome             string        `json:"outcome"`
	SelectedCandidateID *string       `json:"selected_candidate_id"`
	Observations        []Observation `json:"observations"`
	Candidates          []Candidate   `json:"candidates"`
	Descriptions        []Description `json:"descriptions"`
	Warnings            []string      `json:"warnings"`
}

type Observation struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Text string `json:"text"`
}

type Candidate struct {
	ID               string          `json:"id"`
	PlaceName        string          `json:"place_name"`
	Locality         *string         `json:"locality"`
	CountryCode      *string         `json:"country_code"`
	CameraLocation   *CameraLocation `json:"camera_location"`
	Subject          *Subject        `json:"subject"`
	CameraDirection  *Direction      `json:"camera_direction"`
	SupportSummary   string          `json:"support_summary"`
	EvidenceRefs     []string        `json:"evidence_refs"`
	SourceRefs       []string        `json:"source_refs"`
	UncertaintyNotes []string        `json:"uncertainty_notes"`
}

type Coordinate struct {
	Latitude  json.Number `json:"latitude"`
	Longitude json.Number `json:"longitude"`
}

type CameraLocation struct {
	Coordinate
	Granularity      string       `json:"granularity"`
	EstimatedRadiusM *json.Number `json:"estimated_radius_m"`
	RadiusBasis      string       `json:"radius_basis"`
}

type Subject struct {
	Name     string      `json:"name"`
	Location *Coordinate `json:"location"`
}

type Direction struct {
	AzimuthDeg     json.Number  `json:"azimuth_deg"`
	Reference      string       `json:"reference"`
	UncertaintyDeg *json.Number `json:"uncertainty_deg"`
	Method         string       `json:"method"`
}

type Description struct {
	Language          string  `json:"language"`
	Status            string  `json:"status"`
	Text              *string `json:"text"`
	Basis             string  `json:"basis"`
	CandidateID       *string `json:"candidate_id"`
	UnavailableReason *string `json:"unavailable_reason"`
}
