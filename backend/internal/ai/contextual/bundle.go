package contextual

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const Version = "context-v1"

var ErrInvalid = errors.New("invalid context request")

type Class string

const (
	Capture    Class = "capture_time"
	AlbumLabel Class = "selected_album"
	Hint       Class = "user_hint"
	Neighbors  Class = "nearby_locations"
)

type Binding struct {
	Owner, Installation, Asset, Selection, Profile, SourceDigest string
	Revision                                                     int
}
type Consent struct {
	Binding Binding
	Version string
	Classes []Class
	AlbumID string
}
type Album struct {
	ID, Label string
	Member    bool
}
type Candidate struct {
	Asset, CaptureTime  string
	SourceDigest        string
	Lineage             string
	Latitude, Longitude float64
	Accessible          bool
}
type Input struct {
	Binding           Binding
	Consent           Consent
	CaptureTime, Hint string
	Album             *Album
	Candidates        []Candidate
	Window            time.Duration
}
type Source struct {
	Asset        string        `json:"-"`
	SourceDigest string        `json:"-"`
	ID           string        `json:"id"`
	Kind         Class         `json:"kind"`
	Text         string        `json:"text,omitempty"`
	Time         *CaptureStamp `json:"time,omitempty"`
	Location     *Coordinate   `json:"location,omitempty"`
	Lineage      string        `json:"lineage,omitempty"`
}
type Projection struct {
	Version string   `json:"version"`
	Status  string   `json:"status"`
	Sources []Source `json:"sources"`
}
type Bundle struct {
	projection []byte
	metadata   Metadata
}

func Build(in Input) (*Bundle, error) {
	if err := ValidateConsent(in.Binding, in.Consent); err != nil {
		return nil, err
	}
	if in.Window <= 0 || in.Window > 24*time.Hour || len(in.Candidates) > 64 {
		return nil, ErrInvalid
	}
	if in.Consent.Has(Hint) && (len(in.Hint) > 2000 || !utf8.ValidString(in.Hint)) {
		return nil, ErrInvalid
	}
	if in.Consent.Has(AlbumLabel) && in.Album != nil && (len(in.Album.Label) > 256 || !utf8.ValidString(in.Album.Label)) {
		return nil, ErrInvalid
	}
	if in.Consent.Has(AlbumLabel) && (in.Consent.AlbumID == "" || len(in.Consent.AlbumID) > 128 || (in.Album != nil && in.Album.ID != in.Consent.AlbumID)) {
		return nil, ErrInvalid
	}
	sources := []Source{}
	for _, class := range in.Consent.Classes {
		switch class {
		case Capture:
			stamp, _ := ParseCapture(in.CaptureTime)
			sources = append(sources, Source{ID: "capture", Kind: Capture, Time: &stamp})
		case Hint:
			sources = append(sources, Source{ID: "hint", Kind: Hint, Text: in.Hint})
		case AlbumLabel:
			if in.Album != nil && in.Album.Member {
				sources = append(sources, Source{ID: "album", Kind: AlbumLabel, Text: in.Album.Label})
			}
		case Neighbors:
			sources = append(sources, neighborSources(in)...)
		}
	}
	status := "ready"
	if len(sources) == 0 {
		status = "empty"
	}
	data, err := json.Marshal(Projection{Version: Version, Status: status, Sources: sources})
	if len(data) > 16<<10 {
		return nil, ErrInvalid
	}
	if err != nil {
		return nil, ErrInvalid
	}
	consent := in.Consent
	consent.Classes = slices.Clone(consent.Classes)
	metadata := Metadata{Binding: in.Binding, Consent: consent, Version: Version, Sources: []SourceRecord{}, Omissions: []string{}}
	for _, source := range sources {
		metadata.Sources = append(metadata.Sources, SourceRecord{ID: source.ID, Kind: source.Kind, Lineage: source.Lineage, Asset: source.Asset, SourceDigest: source.SourceDigest})
	}
	for _, class := range consent.Classes {
		if !slices.ContainsFunc(sources, func(s Source) bool { return s.Kind == class }) {
			metadata.Omissions = append(metadata.Omissions, string(class)+"_unavailable")
		}
	}
	if consent.Has(Neighbors) {
		metadata.Omissions = append(metadata.Omissions, "bounded_discovery")
	}
	encoded, _ := json.Marshal(struct {
		Metadata   Metadata
		Projection json.RawMessage
	}{metadata, data})
	digest := sha256.Sum256(encoded)
	metadata.Digest = hex.EncodeToString(digest[:])
	return &Bundle{projection: data, metadata: metadata}, nil
}

func (c Consent) Has(class Class) bool { return slices.Contains(c.Classes, class) }

func ValidateConsent(binding Binding, consent Consent) error {
	if consent.Version != Version || consent.Binding != binding || binding.Revision < 1 {
		return ErrInvalid
	}
	for _, value := range []string{binding.Owner, binding.Installation, binding.Asset, binding.Selection, binding.Profile, binding.SourceDigest} {
		if strings.TrimSpace(value) == "" || len(value) > 128 || !utf8.ValidString(value) {
			return ErrInvalid
		}
	}
	seen := map[Class]bool{}
	for _, class := range consent.Classes {
		if seen[class] || (class != Capture && class != Hint && class != AlbumLabel && class != Neighbors) {
			return ErrInvalid
		}
		seen[class] = true
	}
	return nil
}
func (b *Bundle) Projection() ([]byte, error) {
	if b == nil || b.metadata.Version != Version {
		return nil, ErrInvalid
	}
	return append([]byte(nil), b.projection...), nil
}

func (*Bundle) String() string   { return "consented context" }
func (*Bundle) GoString() string { return "consented context" }
