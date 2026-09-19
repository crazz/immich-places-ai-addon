package capabilities

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

func EvaluateJSONObservation(fixture Fixture, text string) Observation {
	color, shape, err := ParseSyntheticSample(text)
	if err != nil {
		return Observation{Status: StatusUnverified, Reason: "invalid_output"}
	}
	if color == fixture.Color && shape == fixture.Shape {
		return Observation{Status: StatusSupported}
	}
	return Observation{Status: StatusUnverified, Reason: "wrong_fixture_facts"}
}

func ParseSyntheticSample(text string) (color, shape string, err error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", "", errors.New("empty sample")
	}
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	tok, err := decoder.Token()
	if err != nil {
		return "", "", err
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return "", "", errors.New("sample must be one JSON object")
	}
	fields := make(map[string]string, 2)
	for decoder.More() {
		keyTok, err := decoder.Token()
		if err != nil {
			return "", "", err
		}
		key, ok := keyTok.(string)
		if !ok {
			return "", "", errors.New("object key must be a string")
		}
		if _, exists := fields[key]; exists {
			return "", "", errors.New("duplicate object key")
		}
		if key != "color" && key != "shape" {
			return "", "", errors.New("unknown field")
		}
		var value string
		if err := decoder.Decode(&value); err != nil {
			return "", "", err
		}
		if value == "" {
			return "", "", errors.New("empty field value")
		}
		fields[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return "", "", err
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return "", "", errors.New("extra JSON after sample")
	}
	colorVal, hasColor := fields["color"]
	shapeVal, hasShape := fields["shape"]
	if !hasColor || !hasShape {
		return "", "", errors.New("color and shape strings are required")
	}
	return colorVal, shapeVal, nil
}
