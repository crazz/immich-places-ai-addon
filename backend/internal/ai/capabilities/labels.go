package capabilities

import (
	"errors"
	"strings"
)

func ParsePlainLabels(text string) (color, shape string, err error) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	if len(lines) != 2 {
		return "", "", errors.New("plain labels must be exactly two lines")
	}
	color = strings.TrimSpace(strings.TrimPrefix(lines[0], "color:"))
	shape = strings.TrimSpace(strings.TrimPrefix(lines[1], "shape:"))
	if !strings.HasPrefix(lines[0], "color:") || !strings.HasPrefix(lines[1], "shape:") || color == "" || shape == "" {
		return "", "", errors.New("plain labels must use color and shape lines")
	}
	if strings.Contains(color, ":") || strings.Contains(shape, ":") {
		return "", "", errors.New("plain labels contain unexpected fields")
	}
	return strings.ToLower(color), strings.ToLower(shape), nil
}

func EvaluateImageObservation(fixture Fixture, text string) Observation {
	color, shape, err := ParsePlainLabels(text)
	if err != nil {
		return Observation{Status: StatusUnverified, Reason: "invalid_output"}
	}
	if color == fixture.Color && shape == fixture.Shape {
		return Observation{Status: StatusSupported}
	}
	return Observation{Status: StatusUnverified, Reason: "wrong_fixture_facts"}
}
