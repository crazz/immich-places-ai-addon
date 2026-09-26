package writepreview

import (
	"slices"

	"github.com/google/uuid"
)

const MaxStackTargets = 50

func SelectStackTargets(snapshot Snapshot, requested, members []string) ([]string, error) {
	if len(requested) == 0 {
		requested = []string{snapshot.AssetID}
	}
	selected := slices.Clone(requested)
	slices.Sort(selected)
	selected = slices.Compact(selected)
	if len(selected) > MaxStackTargets || !slices.Contains(selected, snapshot.AssetID) {
		return nil, Failure{Code: "INVALID_TARGET_SELECTION"}
	}
	for _, id := range selected {
		if parsed, err := uuid.Parse(id); err != nil || parsed.String() != id {
			return nil, Failure{Code: "INVALID_TARGET_SELECTION"}
		}
	}
	if len(selected) == 1 {
		return selected, nil
	}
	if !slices.Contains(snapshot.Fields, "gps") || snapshot.Camera == nil || !finite(snapshot.Camera.Latitude, 90) || !finite(snapshot.Camera.Longitude, 180) {
		return nil, Failure{Code: "INVALID_TARGET_SELECTION"}
	}
	for _, id := range selected {
		if !slices.Contains(members, id) {
			return nil, Failure{Code: "TARGET_UNAVAILABLE"}
		}
	}
	return selected, nil
}
