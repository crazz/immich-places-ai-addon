package translations

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/language"
)

var ErrInvalid = errors.New("invalid translation request")

type Request struct {
	Key                string   `json:"key"`
	DraftID            string   `json:"draftId"`
	Revision           int      `json:"revision"`
	FactsRevision      int      `json:"factsRevision"`
	ProfileID          string   `json:"profileId"`
	ProfileRevision    int      `json:"profileRevision"`
	Basis              string   `json:"basis"`
	BasisKind          string   `json:"basisKind"`
	Languages          []string `json:"languages"`
	Confirmed          bool     `json:"confirmed"`
	ParentID           string   `json:"parentId,omitempty"`
	MaxTokens          int64    `json:"maxTokens,omitempty"`
	MaxEstimatedMicros *int64   `json:"maxEstimatedMicros,omitempty"`
}

func Normalize(req Request) (Request, string, error) {
	if req.MaxTokens < 0 || req.MaxTokens > 8000000000 || (req.MaxEstimatedMicros != nil && *req.MaxEstimatedMicros < 0) || len(req.ParentID) > 128 || !utf8.ValidString(req.ParentID) {
		return Request{}, "", ErrInvalid
	}
	if !req.Confirmed || req.Revision < 1 || req.FactsRevision < 1 || req.ProfileRevision < 1 || strings.TrimSpace(req.Basis) == "" || len(req.Basis) > 16<<10 || !utf8.ValidString(req.Basis) || (req.BasisKind != "scene" && req.BasisKind != "candidate") || len(req.Languages) < 1 || len(req.Languages) > 8 {
		return Request{}, "", ErrInvalid
	}
	for _, id := range []string{req.Key, req.DraftID, req.ProfileID} {
		if strings.TrimSpace(id) == "" || len(id) > 128 || !utf8.ValidString(id) {
			return Request{}, "", ErrInvalid
		}
	}
	req.Languages = slices.Clone(req.Languages)
	seen := map[string]bool{}
	for i, tag := range req.Languages {
		parsed, err := language.Parse(tag)
		if err != nil || tag == "" || len(tag) > 64 || strings.Contains(tag, "_") || parsed == language.Und || seen[parsed.String()] {
			return Request{}, "", ErrInvalid
		}
		req.Languages[i] = parsed.String()
		seen[parsed.String()] = true
	}
	slices.Sort(req.Languages)
	raw, err := json.Marshal(req)
	if err != nil {
		return Request{}, "", ErrInvalid
	}
	digest := sha256.Sum256(raw)
	return req, hex.EncodeToString(digest[:]), nil
}
