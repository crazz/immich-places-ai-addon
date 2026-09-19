package selection

import (
	"errors"
	"github.com/google/uuid"
)

var ErrInvalid = errors.New("invalid selection")
var ErrLimit = errors.New("selection limit exceeded")

type Scope struct {
	AlbumID      string `json:"albumID,omitempty"`
	FolderPath   string `json:"folderPath,omitempty"`
	TagID        string `json:"tagID,omitempty"`
	StartDate    string `json:"startDate,omitempty"`
	EndDate      string `json:"endDate,omitempty"`
	View         string `json:"view"`
	GPSFilter    string `json:"gpsFilter"`
	HiddenFilter string `json:"hiddenFilter"`
}

type Input struct {
	Mode     string   `json:"mode"`
	AssetIDs []string `json:"assetIDs"`
	Scope    *Scope   `json:"scope"`
}

type Request struct {
	Input
	RequestedCount int
	DuplicateCount int
}

func Normalize(input Input, maxAssets int) (Request, error) {
	if input.Mode != "explicit" || len(input.AssetIDs) == 0 || input.Scope == nil {
		return Request{}, ErrInvalid
	}
	if len(input.AssetIDs) > 10000 {
		return Request{}, ErrLimit
	}
	result := Request{Input: input, RequestedCount: len(input.AssetIDs)}
	result.AssetIDs = make([]string, 0, len(input.AssetIDs))
	seen := make(map[string]bool)
	for _, raw := range input.AssetIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return Request{}, ErrInvalid
		}
		canonical := id.String()
		if seen[canonical] {
			result.DuplicateCount++
			continue
		}
		seen[canonical] = true
		result.AssetIDs = append(result.AssetIDs, canonical)
		if len(result.AssetIDs) > maxAssets {
			return Request{}, ErrLimit
		}
	}
	scope, err := normalizeScope(*input.Scope)
	if err != nil {
		return Request{}, err
	}
	result.Scope = &scope
	return result, nil
}
