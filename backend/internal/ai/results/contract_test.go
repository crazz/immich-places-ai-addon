package results

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func requireFailure(t *testing.T, validator *Validator, data []byte, ctx Context, category string) *Failure {
	t.Helper()
	proposal, err := validator.Validate(data, ctx)
	var rejected *Failure
	if !errors.As(err, &rejected) || rejected.Category != category {
		t.Fatalf("wanted %s failure, got %v", category, err)
	}
	if _, err := json.Marshal(proposal); err == nil {
		t.Fatal("failure returned a serializable proposal")
	}
	return rejected
}

func testValidator(t *testing.T) *Validator {
	t.Helper()
	v, err := New()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestTransportCompletionIsAuthoritative(t *testing.T) {
	v := testValidator(t)
	data := fixture(t, "unknown")
	for _, completion := range []Completion{Refused, Truncated, ToolResponse, "", "invented"} {
		ctx := visualContext("en")
		ctx.Completion = completion
		requireFailure(t, v, data, ctx, "incomplete_response")
	}
}

func mutateFixture(t *testing.T, name string, change func(map[string]any)) []byte {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(fixture(t, name), &document); err != nil {
		t.Fatal(err)
	}
	change(document)
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestConstructorRejectsInvalidAndExternalSchemas(t *testing.T) {
	for _, schema := range []string{`{`, `{"type":"invalid-type"}`, `{"$ref":"https://private.invalid/secret"}`, `{"$ref":"file:///private/secret"}`} {
		validator, err := compileSchema(schema)
		var rejected *Failure
		if validator != nil || !errors.As(err, &rejected) || rejected.Category != "unavailable" || strings.Contains(err.Error(), "private") {
			t.Fatalf("unsafe schema construction: %v", err)
		}
	}
}

func TestAmbiguousOrMalformedJSONCannotBecomeAProposal(t *testing.T) {
	v := testValidator(t)
	valid := fixture(t, "unknown")
	for _, data := range [][]byte{
		append(append([]byte{}, valid...), []byte(` {}`)...),
		[]byte(strings.Replace(string(valid), `"outcome": "unknown"`, `"outcome":"located","outcome":"unknown"`, 1)),
		[]byte(strings.Replace(string(valid), `"id": "obs-1"`, `"id":"other","id":"obs-1"`, 1)),
		[]byte(strings.Replace(string(valid), `"language": "en"`, `"language":"uk","language":"en"`, 1)),
		valid[:len(valid)-3], []byte("```json\n" + string(valid) + "\n```"),
	} {
		requireFailure(t, v, data, visualContext("en"), "invalid_json")
	}
}

func TestJSONUnicodeIsNeverRepaired(t *testing.T) {
	v := testValidator(t)
	for _, text := range []string{string([]byte{0xff}), `\uD800`, `\uDC00`, `\uD800\u0041`} {
		data := []byte(strings.Replace(string(fixture(t, "unknown")), "No distinctive geographic feature can be identified reliably.", text, 1))
		requireFailure(t, v, data, visualContext("en"), "invalid_json")
	}
	valid := []byte(strings.Replace(string(fixture(t, "unknown")), "No distinctive geographic feature can be identified reliably.", `\uD83D\uDE00 and \\uD800`, 1))
	if _, err := v.Validate(valid, visualContext("en")); err != nil {
		t.Fatal("valid pair/literal escape rejected", err)
	}
}

func TestInvalidServerContextFailsBeforeProviderParsing(t *testing.T) {
	v := testValidator(t)
	for _, change := range []func(*Context){
		func(c *Context) { c.Mode = "" }, func(c *Context) { c.Mode = "invented" },
		func(c *Context) { c.Languages = nil }, func(c *Context) { c.Languages = []string{"en", "EN"} },
		func(c *Context) { c.Languages = []string{"iw", "he"}; c.PrimaryLanguage = "he" },
		func(c *Context) { c.Languages = []string{"en_US"}; c.PrimaryLanguage = "en_US" },
		func(c *Context) { c.PrimaryLanguage = "uk" },
		func(c *Context) { c.Sources = []Source{{ID: "source"}} },
		func(c *Context) { c.Mode = ContextAssisted; c.Sources = []Source{{ID: " "}} },
		func(c *Context) { c.Mode = ContextAssisted; c.Sources = []Source{{ID: "same"}, {ID: "same"}} },
	} {
		ctx := visualContext("en")
		change(&ctx)
		requireFailure(t, v, []byte(`broken`), ctx, "invalid_context")
	}
}

func TestUnsupportedSchemaVersionHasItsOwnFailure(t *testing.T) {
	data := mutateFixture(t, "unknown", func(m map[string]any) { m["schema_version"] = "2.0" })
	requireFailure(t, testValidator(t), data, visualContext("en"), "unsupported_schema")
}
