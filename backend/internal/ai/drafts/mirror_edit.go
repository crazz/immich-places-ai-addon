package drafts

import (
	"bytes"
	"encoding/json"
	"io"
	"slices"

	"golang.org/x/text/language"
)

func parseMirrorSelection(raw []byte) (*MirrorSelection, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	if len(raw) > 4096 {
		return nil, ErrInvalid
	}
	selection := &MirrorSelection{Languages: []string{}}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, ErrInvalid
	}
	fields := map[string]any{"direction": &selection.Direction, "precision": &selection.Precision, "place": &selection.Place, "languages": &selection.Languages, "provenance": &selection.Provenance, "model": &selection.Model}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		target := fields[key]
		if err != nil || !ok || target == nil || decoder.Decode(target) != nil {
			return nil, ErrInvalid
		}
		delete(fields, key)
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, ErrInvalid
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, ErrInvalid
	}
	if (!selection.Direction && !selection.Precision && !selection.Place && !selection.Provenance && len(selection.Languages) == 0) || (selection.Model && !selection.Provenance) || len(selection.Languages) > 8 {
		return nil, ErrInvalid
	}
	for i, tag := range selection.Languages {
		parsed, err := language.Parse(tag)
		if err != nil || parsed.String() != tag || slices.Contains(selection.Languages[:i], tag) {
			return nil, ErrInvalid
		}
	}
	slices.Sort(selection.Languages)
	return selection, nil
}
