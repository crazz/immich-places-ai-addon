package results

import (
	"bytes"
	"encoding/json"
)

type Proposal struct{ data []byte }

func (p Proposal) MarshalJSON() ([]byte, error) {
	if len(p.data) == 0 {
		return nil, failure("uninitialized", "proposal", "/")
	}
	return bytes.Clone(p.data), nil
}

func (p Proposal) PolicyVersion() (string, error) {
	document, err := p.Data()
	if err != nil {
		return "", failure("uninitialized", "proposal", "/")
	}
	if document.SchemaVersion == "2.0" {
		return "analysis-result-v2", nil
	}
	return "analysis-result-v1", nil
}

// Data returns an independent typed copy; numeric tokens retain exact precision.
func (p Proposal) Data() (Document, error) {
	var document Document
	if err := json.Unmarshal(p.data, &document); err != nil {
		return Document{}, failure("uninitialized", "proposal", "/")
	}
	return document, nil
}
