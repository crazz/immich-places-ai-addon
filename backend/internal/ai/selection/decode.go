package selection

import (
	"bytes"
	"encoding/json"
	"io"
	"unicode/utf8"
)

const MaxBytes = 1 << 20

func Decode(data []byte) (Input, error) {
	if len(data) > MaxBytes {
		return Input{}, ErrLimit
	}
	if !utf8.Valid(data) {
		return Input{}, ErrInvalid
	}
	tokens := json.NewDecoder(bytes.NewReader(data))
	if err := uniqueObjectKeys(tokens, 0); err != nil {
		return Input{}, ErrInvalid
	}
	if _, err := tokens.Token(); err != io.EOF {
		return Input{}, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input *Input
	if err := decoder.Decode(&input); err != nil || input == nil {
		return Input{}, ErrInvalid
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return Input{}, ErrInvalid
	}
	for key := range object {
		if key != "mode" && key != "assetIDs" && key != "scope" {
			return Input{}, ErrInvalid
		}
	}
	var scope map[string]json.RawMessage
	if err := json.Unmarshal(object["scope"], &scope); err != nil || scope == nil {
		return Input{}, ErrInvalid
	}
	for key, value := range scope {
		switch key {
		case "view", "albumID", "folderPath", "tagID", "gpsFilter", "hiddenFilter", "startDate", "endDate":
		default:
			return Input{}, ErrInvalid
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return Input{}, ErrInvalid
		}
		if key == "albumID" && input.Scope.View != "album" || key == "folderPath" && input.Scope.View != "folder" {
			return Input{}, ErrInvalid
		}
	}
	return *input, nil
}

func uniqueObjectKeys(decoder *json.Decoder, depth int) error {
	if depth > 16 {
		return ErrInvalid
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	keys := map[string]bool{}
	for decoder.More() {
		if delim == '{' {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || keys[name] {
				return ErrInvalid
			}
			keys[name] = true
		}
		if err := uniqueObjectKeys(decoder, depth+1); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
