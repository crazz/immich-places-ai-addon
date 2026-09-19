package selection

import (
	"errors"
	"testing"
)

func TestDecodeRejectsCaseAliases(t *testing.T) {
	for _, data := range []string{
		`{"Mode":"explicit","assetIDs":["aaaaaaaa-0000-4000-8000-000000000001"],"scope":{"view":"all"}}`,
		`{"mode":"explicit","Mode":"explicit","assetIDs":["aaaaaaaa-0000-4000-8000-000000000001"],"scope":{"view":"all"}}`,
		`{"mode":"explicit","assetIDs":["aaaaaaaa-0000-4000-8000-000000000001"],"scope":{"view":"all","GPSFilter":"all"}}`,
	} {
		if _, err := Decode([]byte(data)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("ambiguous field spelling accepted: %v", err)
		}
	}
}
