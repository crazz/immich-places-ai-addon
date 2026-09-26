package main

import (
	"encoding/json"
	"strconv"
	"unicode/utf8"

	"immich-places-backend/internal/ai/drafts"
)

func aiDescriptionBaseline(exif json.RawMessage) *drafts.DescriptionBaseline {
	if string(exif) == "null" {
		return &drafts.DescriptionBaseline{Presence: "absent"}
	}
	fields, err := aiImageUniqueObject(exif)
	if err != nil {
		return nil
	}
	raw, exists := fields["description"]
	if !exists {
		return &drafts.DescriptionBaseline{Presence: "absent"}
	}
	if string(raw) == "null" {
		return &drafts.DescriptionBaseline{Presence: "null"}
	}
	var value string
	if !utf8.Valid(raw) || json.Unmarshal(raw, &value) != nil || !aiExactJSONText(raw) || len(value) > 64<<10 {
		return nil
	}
	return &drafts.DescriptionBaseline{Presence: "value", Value: value}
}

// encoding/json replaces unpaired UTF-16 surrogates; a write baseline must not.
func aiExactJSONText(raw []byte) bool {
	for i := 1; i < len(raw)-1; i++ {
		if raw[i] != '\\' {
			continue
		}
		i++
		if raw[i] != 'u' {
			continue
		}
		unit, err := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if err != nil {
			return false
		}
		i += 4
		if unit >= 0xdc00 && unit <= 0xdfff {
			return false
		}
		if unit < 0xd800 || unit > 0xdbff {
			continue
		}
		if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
			return false
		}
		low, err := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
		if err != nil || low < 0xdc00 || low > 0xdfff {
			return false
		}
		i += 6
	}
	return true
}
