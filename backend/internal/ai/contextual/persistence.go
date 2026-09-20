package contextual

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type storedBundle struct {
	Metadata   Metadata
	Projection json.RawMessage
}

func (b *Bundle) MarshalJSON() ([]byte, error) {
	projection, err := b.Projection()
	if err != nil {
		return nil, err
	}
	return json.Marshal(storedBundle{b.Info(), projection})
}

func (b *Bundle) UnmarshalJSON(data []byte) error {
	*b = Bundle{}
	if len(data) > 32<<10 {
		return ErrInvalid
	}
	var stored storedBundle
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(&stored) != nil || len(stored.Projection) > 16<<10 {
		return ErrInvalid
	}
	m := stored.Metadata
	if m.Version != Version || ValidateConsent(m.Binding, m.Consent) != nil || len(m.Sources) > 9 || len(m.Omissions) > 5 {
		return ErrInvalid
	}
	var projection Projection
	d = json.NewDecoder(bytes.NewReader(stored.Projection))
	d.DisallowUnknownFields()
	if d.Decode(&projection) != nil || projection.Version != Version || len(projection.Sources) != len(m.Sources) || (projection.Status != "ready" && projection.Status != "empty") {
		return ErrInvalid
	}
	seen := map[string]bool{}
	for i, source := range projection.Sources {
		record := m.Sources[i]
		if source.ID == "" || seen[source.ID] || source.ID != record.ID || source.Kind != record.Kind || !m.Consent.Has(source.Kind) {
			return ErrInvalid
		}
		seen[source.ID] = true
	}
	expected := m.Digest
	m.Digest = ""
	encoded, err := json.Marshal(storedBundle{m, stored.Projection})
	if err != nil {
		return ErrInvalid
	}
	digest := sha256.Sum256(encoded)
	if expected != hex.EncodeToString(digest[:]) {
		return ErrInvalid
	}
	b.projection = append([]byte(nil), stored.Projection...)
	b.metadata = stored.Metadata
	return nil
}
