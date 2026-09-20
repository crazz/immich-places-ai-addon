package contextual

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestStoredBundleRejectsChangedOrUnboundedProvenance(t *testing.T) {
	b, err := Build(contextInput())
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(b)
	for _, data := range [][]byte{[]byte("null"), []byte("{}"), bytes.ReplaceAll(raw, []byte("Near a bridge"), []byte("changed hint")), []byte(strings.Repeat(" ", 32769)), bytes.ReplaceAll(raw, []byte("context-v1"), []byte("context-v2"))} {
		var restored Bundle
		if json.Unmarshal(data, &restored) == nil {
			t.Fatal("invalid stored provenance accepted")
		}
		if _, err = restored.Projection(); err == nil {
			t.Fatal("failed restore retained authority")
		}
	}
}
