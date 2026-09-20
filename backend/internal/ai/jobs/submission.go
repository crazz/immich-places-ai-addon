package jobs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"immich-places-backend/internal/ai/results"
)

var ErrInvalid = errors.New("invalid AI job request")
var ErrDenied = errors.New("AI job unavailable")
var ErrConflict = errors.New("AI job submission conflicts")
var ErrBudget = errors.New("AI job dispatch budget exhausted")
var ErrLease = errors.New("AI job lease unavailable")
var ErrStorage = errors.New("AI job storage unavailable")

type Submission struct {
	Owner, Installation, Key, Profile                string
	Revision                                         int
	AssetIDs, Languages                              []string
	PrimaryLanguage, SelectionDigest, ConsentVersion string
	MaxCalls                                         int
	Mode                                             string
}

type Job struct {
	ID, Model string
	Input     Submission
	Calls     int
	Canceled  bool
	CreatedAt int64
	Items     []Item
}

type Item struct {
	ID, Asset, State, Failure string
	Attempts, Calls           int
	NextAttemptAt             int64
}

func Normalize(input Submission) (Submission, string, error) {
	if len(input.AssetIDs) == 0 || len(input.AssetIDs) > 10000 || input.Revision < 1 || input.ConsentVersion != "visual-v1" || input.MaxCalls < 1 || (input.Mode != "" && input.Mode != "visual") {
		return Submission{}, "", ErrInvalid
	}
	for _, id := range []string{input.Owner, input.Installation, input.Key, input.Profile} {
		if !boundedIdentity(id) {
			return Submission{}, "", ErrInvalid
		}
	}
	if digest, err := hex.DecodeString(input.SelectionDigest); err != nil || len(digest) != 32 {
		return Submission{}, "", ErrInvalid
	}
	input.Mode = "visual"
	tags, primary, err := results.NormalizeLanguages(input.Languages, input.PrimaryLanguage)
	if err != nil {
		return Submission{}, "", ErrInvalid
	}
	input.Languages, input.PrimaryLanguage = tags, primary
	ids := make([]string, 0, len(input.AssetIDs))
	seen := map[string]bool{}
	for _, id := range input.AssetIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return Submission{}, "", ErrInvalid
		}
		id = parsed.String()
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	input.AssetIDs = ids
	if len(ids) > 500 || input.MaxCalls > 3*len(ids) {
		return Submission{}, "", ErrInvalid
	}
	data, err := json.Marshal(input)
	if err != nil {
		return Submission{}, "", ErrInvalid
	}
	digest := sha256.Sum256(data)
	return input, hex.EncodeToString(digest[:]), nil
}

func boundedIdentity(value string) bool {
	if value == "" || len(value) > 128 || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
