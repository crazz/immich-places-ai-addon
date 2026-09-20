package providerhttp

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestResearchGenerationSchemaRequiresAnExplicitPossiblyEmptySourceList(t *testing.T) {
	body, err := EncodeResearch("model", "instruction", "data:image/jpeg;base64,eA==", "strict")
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		ResponseFormat struct {
			JSONSchema struct {
				Schema struct {
					Required []string `json:"required"`
				} `json:"schema"`
			} `json:"json_schema"`
		} `json:"response_format"`
	}
	if json.Unmarshal(body, &request) != nil {
		t.Fatal("invalid request")
	}
	if !slices.Contains(request.ResponseFormat.JSONSchema.Schema.Required, "sources") {
		t.Fatal("strict generation schema leaves an optional property")
	}
}
