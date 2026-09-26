package writepreview

import "encoding/json"

const MirrorNamespace = "immich-places-ai-addon"

type MirrorBaseline struct {
	Present bool            `json:"present"`
	Value   json.RawMessage `json:"value,omitempty"`
}

func ParseMirrorNamespace(raw []byte) (MirrorBaseline, error) {
	var baseline MirrorBaseline
	value, err := parseMirrorJSON(raw, MirrorListLimit)
	list, ok := value.([]any)
	if err != nil || !ok {
		return baseline, Failure{Code: "METADATA_UNAVAILABLE"}
	}
	keys := map[string]bool{}
	for _, entry := range list {
		item, ok := entry.(map[string]any)
		key, validKey := item["key"].(string)
		object, validValue := item["value"].(map[string]any)
		if !ok || !validKey || key == "" || keys[key] || !validValue {
			return MirrorBaseline{}, Failure{Code: "METADATA_UNAVAILABLE"}
		}
		keys[key] = true
		if key == MirrorNamespace {
			encoded, err := json.Marshal(object)
			if err != nil || len(encoded) > MirrorValueLimit {
				return MirrorBaseline{}, Failure{Code: "METADATA_UNAVAILABLE"}
			}
			baseline = MirrorBaseline{Present: true, Value: encoded}
		}
	}
	return baseline, nil
}
