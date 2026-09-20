package providerhttp

import (
	"encoding/json"
	"immich-places-backend/internal/ai/results"
)

func researchGenerationSchema() ([]byte, error) {
	var schema map[string]json.RawMessage
	if err := json.Unmarshal(results.ResearchSchema(), &schema); err != nil {
		return nil, err
	}
	var required []string
	if err := json.Unmarshal(schema["required"], &required); err != nil {
		return nil, err
	}
	// Strict generation requires every property; readers also accept omitted optional sources.
	schema["required"], _ = json.Marshal(append(required, "sources"))
	return json.Marshal(schema)
}
