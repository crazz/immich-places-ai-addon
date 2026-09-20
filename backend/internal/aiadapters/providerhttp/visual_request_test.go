package providerhttp

import (
	"encoding/json"
	"reflect"
	"testing"

	"immich-places-backend/internal/ai/results"
)

func TestVisualEncodingKeepsOneImageAndCanonicalFormat(t *testing.T) {
	for _, format := range []string{"strict", "json"} {
		raw, err := EncodeVisual("bound-model", "controlled instructions", "data:image/jpeg;base64,AA==", format)
		if err != nil {
			t.Fatal(err)
		}
		var payload map[string]any
		if json.Unmarshal(raw, &payload) != nil || len(payload) != 4 {
			t.Fatal("unexpected top-level parameters")
		}
		if payload["model"] != "bound-model" || payload["stream"] != false {
			t.Fatal("wrong model or streaming")
		}
		messages := payload["messages"].([]any)
		content := messages[0].(map[string]any)["content"].([]any)
		if len(messages) != 1 || len(content) != 2 || content[1].(map[string]any)["type"] != "image_url" {
			t.Fatal("wrong content")
		}
		rf := payload["response_format"].(map[string]any)
		if format == "strict" {
			var schema any
			_ = json.Unmarshal(results.CanonicalSchema(), &schema)
			if !reflect.DeepEqual(rf["json_schema"].(map[string]any)["schema"], schema) {
				t.Fatal("schema drift")
			}
		} else if rf["type"] != "json_object" {
			t.Fatal("wrong JSON format")
		}
	}
}

func TestVisualEncodingRejectsUnboundedOrUnsupportedInput(t *testing.T) {
	cases := [][4]string{{"", "prompt", "data:image/jpeg;base64,AA==", "strict"}, {"model", "", "data:image/jpeg;base64,AA==", "strict"}, {"model", "prompt", "https://private/image", "strict"}, {"model", "prompt", "data:image/jpeg;base64,AA==", "text"}, {"model", string(make([]byte, 15<<20)), "data:image/jpeg;base64,AA==", "json"}}
	for _, c := range cases {
		if raw, err := EncodeVisual(c[0], c[1], c[2], c[3]); err == nil || raw != nil {
			t.Fatal("unsafe request accepted")
		}
	}
}
