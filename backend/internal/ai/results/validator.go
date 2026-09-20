package results

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed ai-analysis-result.v1.schema.json
var canonicalSchema string

const schemaID = "urn:immich-places-ai-addon:analysis-result:1.0"

type Validator struct{ schema *jsonschema.Schema }

type rejectingLoader struct{}

func (rejectingLoader) Load(string) (any, error) {
	return nil, failure("unavailable", "schema_resource", "/")
}

func New() (*Validator, error) { return compileSchema(canonicalSchema) }

func compileSchema(data string) (*Validator, error) {
	var document any
	decoder := json.NewDecoder(strings.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil, failure("unavailable", "schema_compile", "/")
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.UseLoader(rejectingLoader{})
	if err := compiler.AddResource(schemaID, document); err != nil {
		return nil, failure("unavailable", "schema_compile", "/")
	}
	schema, err := compiler.Compile(schemaID)
	if err != nil {
		return nil, failure("unavailable", "schema_compile", "/")
	}
	return &Validator{schema: schema}, nil
}

func (v *Validator) Validate(data []byte, ctx Context) (Proposal, error) {
	if v == nil || v.schema == nil {
		return Proposal{}, failure("unavailable", "validator", "/")
	}
	trusted, err := validateContext(ctx)
	if err != nil {
		return Proposal{}, err
	}
	if ctx.Completion != Complete {
		return Proposal{}, failure("incomplete_response", "completion", "/")
	}
	document, err := parse(data)
	if err != nil {
		return Proposal{}, err
	}
	if object, ok := document.(map[string]any); ok {
		if version, present := object["schema_version"].(string); present && version != "1.0" {
			return Proposal{}, failure("unsupported_schema", "version", "/schema_version")
		}
	}
	if err := v.schema.Validate(document); err != nil {
		return Proposal{}, failure("schema_violation", "schema", "/")
	}
	var typed Document
	if err := json.Unmarshal(data, &typed); err != nil {
		return Proposal{}, failure("schema_violation", "typed_value", "/")
	}
	findings := append(outcomeFindings(typed), referenceFindings(typed, trusted)...)
	findings = append(findings, geometryFindings(typed, trusted)...)
	findings = append(findings, languageFindings(&typed, trusted)...)
	if len(findings) != 0 {
		return Proposal{}, semanticFailure(findings)
	}
	encoded, err := json.Marshal(typed)
	if err != nil {
		return Proposal{}, failure("unavailable", "result_encoding", "/")
	}
	if len(encoded) > maxBytes {
		return Proposal{}, failure("limit_exceeded", "serialized_bytes", "/")
	}
	return Proposal{data: encoded}, nil
}
