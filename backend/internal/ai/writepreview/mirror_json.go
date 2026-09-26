package writepreview

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MirrorValueLimit = 64 << 10
const MirrorListLimit = 1 << 20

func parseMirrorJSON(raw []byte, limit int) (any, error) {
	if len(raw) == 0 || len(raw) > limit || !mirrorUnicode(raw) {
		return nil, Failure{Code: "METADATA_UNAVAILABLE"}
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	nodes := 0
	value, err := mirrorJSONValue(decoder, 0, &nodes)
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, Failure{Code: "METADATA_UNAVAILABLE"}
	}
	return value, nil
}

func mirrorJSONValue(decoder *json.Decoder, depth int, nodes *int) (any, error) {
	*nodes++
	invalid := Failure{Code: "METADATA_UNAVAILABLE"}
	if depth > 32 || *nodes > 20000 {
		return nil, invalid
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, invalid
	}
	if number, ok := token.(json.Number); ok {
		value, err := number.Float64()
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) || len(number) > 128 {
			return nil, invalid
		}
		if i := strings.IndexAny(string(number), "eE"); i >= 0 {
			exponent, err := strconv.Atoi(string(number[i+1:]))
			if err != nil || exponent < -1024 || exponent > 1024 {
				return nil, invalid
			}
		}
	}
	delim, container := token.(json.Delim)
	if !container {
		return token, nil
	}
	if delim == '{' {
		object := map[string]any{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			key, ok := keyToken.(string)
			if _, duplicate := object[key]; err != nil || !ok || duplicate {
				return nil, invalid
			}
			value, err := mirrorJSONValue(decoder, depth+1, nodes)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		if close, err := decoder.Token(); err != nil || close != json.Delim('}') {
			return nil, invalid
		}
		return object, nil
	}
	if delim == '[' {
		array := []any{}
		for decoder.More() {
			value, err := mirrorJSONValue(decoder, depth+1, nodes)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		if close, err := decoder.Token(); err != nil || close != json.Delim(']') {
			return nil, invalid
		}
		return array, nil
	}
	return nil, invalid
}

// encoding/json substitutes unpaired surrogates; a baseline must retain exact text.
func mirrorUnicode(raw []byte) bool {
	if !utf8.Valid(raw) {
		return false
	}
	quoted := false
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			quoted = !quoted
			continue
		}
		if !quoted || raw[i] != '\\' {
			continue
		}
		i++
		if i >= len(raw) {
			return false
		}
		if raw[i] != 'u' {
			continue
		}
		if i+4 >= len(raw) {
			return false
		}
		unit, err := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if err != nil || (unit >= 0xDC00 && unit <= 0xDFFF) {
			return false
		}
		i += 4
		if unit < 0xD800 || unit > 0xDBFF {
			continue
		}
		if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
			return false
		}
		low, err := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
		if err != nil || low < 0xDC00 || low > 0xDFFF {
			return false
		}
		i += 6
	}
	return true
}
