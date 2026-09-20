package contextual

import "slices"

type SourceRecord struct {
	ID                           string
	Kind                         Class
	Asset, SourceDigest, Lineage string
}
type Metadata struct {
	Binding         Binding
	Consent         Consent
	Version, Digest string
	Sources         []SourceRecord
	Omissions       []string
}

func (b *Bundle) Info() Metadata {
	if b == nil {
		return Metadata{}
	}
	m := b.metadata
	m.Consent.Classes = slices.Clone(m.Consent.Classes)
	m.Sources = slices.Clone(m.Sources)
	m.Omissions = slices.Clone(m.Omissions)
	return m
}
