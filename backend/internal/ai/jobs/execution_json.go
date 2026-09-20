package jobs

import (
	"encoding/json"
	"io"
	"strings"
	"unicode/utf8"
)

func unambiguousExecutionJSON(raw string) bool {
	if !utf8.ValidString(raw) {
		return false
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if !executionJSONValue(decoder, 0) {
		return false
	}
	_, err := decoder.Token()
	return err == io.EOF
}

func executionJSONValue(decoder *json.Decoder, depth int) bool {
	if depth > 8 {
		return false
	}
	token, err := decoder.Token()
	if err != nil {
		return false
	}
	delim, container := token.(json.Delim)
	if !container {
		return true
	}
	if delim != '{' && delim != '[' {
		return false
	}
	seen := map[string]bool{}
	for decoder.More() {
		if delim == '{' {
			key, err := decoder.Token()
			name, ok := key.(string)
			if err != nil || !ok {
				return false
			}
			name = strings.ToLower(name)
			if seen[name] {
				return false
			}
			seen[name] = true
		}
		if !executionJSONValue(decoder, depth+1) {
			return false
		}
	}
	end, err := decoder.Token()
	return err == nil && ((delim == '{' && end == json.Delim('}')) || (delim == '[' && end == json.Delim(']')))
}
