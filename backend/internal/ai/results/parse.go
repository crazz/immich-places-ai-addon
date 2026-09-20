package results

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

var errJSON = errors.New("invalid JSON")
var errBudget = errors.New("JSON budget exceeded")

const (
	maxBytes  = 1 << 20
	maxDepth  = 32
	maxNodes  = 20000
	maxString = 8192
)

func parse(data []byte) (any, error) {
	if len(data) > maxBytes {
		return nil, failure("limit_exceeded", "bytes", "/")
	}
	if !validUnicode(data) {
		return nil, failure("invalid_json", "unicode", "/")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	nodes := 0
	value, err := decodeValue(decoder, 0, &nodes, "")
	if errors.Is(err, errBudget) {
		return nil, failure("limit_exceeded", "structure", "/")
	}
	if err != nil {
		return nil, failure("invalid_json", "syntax", "/")
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, failure("invalid_json", "trailing_value", "/")
	}
	return value, nil
}

func decodeValue(decoder *json.Decoder, depth int, nodes *int, field string) (any, error) {
	*nodes++
	if *nodes > maxNodes {
		return nil, errBudget
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, errJSON
	}
	delimiter, container := token.(json.Delim)
	if !container {
		if number, ok := token.(json.Number); ok {
			if err := boundedNumber(number); err != nil {
				return nil, err
			}
		}
		if text, ok := token.(string); ok && len(text) > stringLimit(field) {
			return nil, errBudget
		}
		return token, nil
	}
	if depth >= maxDepth {
		return nil, errBudget
	}
	switch delimiter {
	case '{':
		object := map[string]any{}
		for decoder.More() {
			token, err := decoder.Token()
			if err != nil {
				return nil, errJSON
			}
			key, ok := token.(string)
			if !ok {
				return nil, errJSON
			}
			if len(key) > maxString {
				return nil, errBudget
			}
			if _, duplicate := object[key]; duplicate {
				return nil, errJSON
			}
			value, err := decodeValue(decoder, depth+1, nodes, key)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		if close, err := decoder.Token(); err != nil || close != json.Delim('}') {
			return nil, errJSON
		}
		return object, nil
	case '[':
		array := []any{}
		for decoder.More() {
			if len(array) >= collectionLimit(field) {
				return nil, errBudget
			}
			value, err := decodeValue(decoder, depth+1, nodes, field)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		if close, err := decoder.Token(); err != nil || close != json.Delim(']') {
			return nil, errJSON
		}
		return array, nil
	}
	return nil, errJSON
}

func collectionLimit(field string) int {
	switch field {
	case "observations", "evidence_refs", "source_refs":
		return 100
	case "candidates", "uncertainty_notes":
		return 20
	case "descriptions":
		return 10
	case "warnings":
		return 50
	default:
		return maxNodes
	}
}

func stringLimit(field string) int {
	switch field {
	case "id", "selected_candidate_id", "candidate_id", "language", "evidence_refs", "source_refs":
		return 128
	default:
		return maxString
	}
}
