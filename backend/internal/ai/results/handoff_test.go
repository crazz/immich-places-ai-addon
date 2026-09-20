package results

import (
	"bytes"
	"testing"
)

func TestCanonicalSchemaHandoffOwnsItsBytes(t *testing.T) {
	first := CanonicalSchema()
	if !bytes.Equal(first, []byte(canonicalSchema)) {
		t.Fatal("transport schema differs from canonical")
	}
	first[0] = 'x'
	if !bytes.Equal(CanonicalSchema(), []byte(canonicalSchema)) {
		t.Fatal("caller changed retained schema")
	}
}

func TestLanguageHandoffUsesCanonicalContextRules(t *testing.T) {
	tags, primary, err := NormalizeLanguages([]string{"uk", "pt-br"}, "pt-BR")
	if err != nil || primary != "pt-BR" || len(tags) != 2 || tags[1] != "pt-BR" {
		t.Fatal("normalization failed")
	}
	for _, languages := range [][]string{nil, {"en", "EN"}, {"private text"}, {"en"}} {
		if _, _, err := NormalizeLanguages(languages, "uk"); err == nil {
			t.Fatal("invalid context accepted")
		}
	}
}
