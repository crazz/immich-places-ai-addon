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
	if len(p.data) == 0 {
		return "", failure("uninitialized", "proposal", "/")
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
