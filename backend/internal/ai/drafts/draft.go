package drafts

import "errors"

var (
	ErrUnavailable = errors.New("draft unavailable")
	ErrInvalid     = errors.New("invalid draft")
	ErrConflict    = errors.New("draft revision changed")
	ErrStorage     = errors.New("draft storage unavailable")
)

type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Baseline struct {
	Description   *DescriptionBaseline `json:"description,omitempty"`
	Latitude      *float64             `json:"latitude"`
	Longitude     *float64             `json:"longitude"`
	ImageIdentity string               `json:"imageIdentity"`
	SourceDigest  string               `json:"sourceDigest"`
	AssetID       string               `json:"assetId"`
	OwnerID       string               `json:"ownerId"`
	Checksum      string               `json:"checksum"`
	Type          string               `json:"type"`
	ObservedAt    string               `json:"observedAt"`
	Status        string               `json:"status"`
}

type DescriptionBaseline struct {
	Presence string `json:"presence"`
	Value    string `json:"value"`
}

type Description struct {
	Language      string  `json:"language"`
	Status        string  `json:"status"`
	Text          *string `json:"text"`
	Basis         string  `json:"basis"`
	Stale         bool    `json:"stale"`
	FactsRevision int     `json:"factsRevision"`
	UserSupplied  bool    `json:"userSupplied"`
}

type Draft struct {
	Mirror               *MirrorSelection `json:"mirror,omitempty"`
	PrimaryLanguage      string           `json:"primaryLanguage,omitempty"`
	DescriptionPolicy    string           `json:"descriptionPolicy,omitempty"`
	FactsRevision        int              `json:"factsRevision"`
	Heading              *float64         `json:"heading"`
	HeadingMethod        string           `json:"headingMethod,omitempty"`
	HeadingUncertainty   *float64         `json:"headingUncertainty,omitempty"`
	HeadingStale         bool             `json:"headingStale"`
	HeadingUserSupplied  bool             `json:"headingUserSupplied"`
	HeadingFactsRevision int              `json:"headingFactsRevision"`
	RadiusStale          bool             `json:"radiusStale"`
	Descriptions         []Description    `json:"descriptions"`

	CandidateID          *string  `json:"candidateId"`
	Radius               *float64 `json:"radius"`
	RadiusBasis          string   `json:"radiusBasis"`
	ID                   string   `json:"id"`
	AnalysisID           string   `json:"analysisId"`
	AssetID              string   `json:"assetId"`
	Revision             int      `json:"revision"`
	State                string   `json:"state"`
	Camera               *Point   `json:"camera"`
	Fields               []string `json:"fields"`
	OriginalSourceDigest string   `json:"originalSourceDigest"`
	Baseline             Baseline `json:"baseline"`
	UpdatedAt            string   `json:"updatedAt"`
}

type Observation struct {
	ID                    string   `json:"id"`
	DraftID               string   `json:"draftId"`
	Revision              int      `json:"revision"`
	Baseline              Baseline `json:"baseline"`
	ExpiresAt             string   `json:"expiresAt"`
	PreviewURL            string   `json:"previewUrl"`
	OriginalSourceMatches bool     `json:"originalSourceMatches"`
}
