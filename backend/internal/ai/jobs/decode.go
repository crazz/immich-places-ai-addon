package jobs

import (
	"bytes"
	"encoding/json"
)

func DecodeRequest(data []byte, target any) error {
	if len(data) > 32<<10 || len(bytes.TrimSpace(data)) == 0 || bytes.Equal(bytes.TrimSpace(data), []byte("null")) || !unambiguousExecutionJSON(string(data)) {
		return ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return ErrInvalid
	}
	return nil
}
