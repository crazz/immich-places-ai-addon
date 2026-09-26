package writepreview

import (
	"encoding/json"

	"github.com/google/uuid"
)

func ValidateMirrorOwnership(current MirrorBaseline, recordID string, verified []byte) error {
	conflict := Failure{Code: "METADATA_CONFLICT"}
	id, err := uuid.Parse(recordID)
	if err != nil || id.Version() != 4 || id.String() != recordID {
		return conflict
	}
	if !current.Present {
		if len(current.Value) != 0 {
			return conflict
		}
		return nil
	}
	if !EqualMirrorValue(current.Value, verified) {
		return conflict
	}
	var value MirrorExport
	if json.Unmarshal(current.Value, &value) != nil || value.Application != MirrorNamespace || value.SchemaVersion != 1 || value.RecordID != recordID {
		return conflict
	}
	return nil
}
