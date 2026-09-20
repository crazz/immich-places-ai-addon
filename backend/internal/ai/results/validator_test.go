package results

import (
	"embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed testdata/*.json
var fixtureFS embed.FS

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := fixtureFS.ReadFile("testdata/ai-analysis-result." + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func visualContext(languages ...string) Context {
	return Context{Mode: Visual, Completion: Complete, Languages: languages, PrimaryLanguage: languages[0]}
}

func TestCanonicalFixturesRemainUnchanged(t *testing.T) {
	validator, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name      string
		languages []string
	}{{"synthetic", []string{"en", "uk"}}, {"unknown", []string{"en"}}} {
		data := fixture(t, tc.name)
		proposal, err := validator.Validate(data, visualContext(tc.languages...))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(proposal)
		if err != nil {
			t.Fatal(err)
		}
		var expected, actual any
		if err := json.Unmarshal(data, &expected); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &actual); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Fatalf("canonical fields changed: %s", encoded)
		}
	}
}
