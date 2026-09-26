package writepreview

import (
	"encoding/json"
	"slices"
)

func (p Plan) MarshalJSON() ([]byte, error) {
	type wire Plan
	raw, err := json.Marshal(wire(p))
	if err != nil || p.Version == "gps-preview-v1" || slices.Contains(p.Fields, "gps") {
		return raw, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	delete(fields, "before")
	delete(fields, "intended")
	return json.Marshal(fields)
}
