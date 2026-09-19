package selection

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func Digest(binding, owner, manifest string, facts []string) string {
	data, _ := json.Marshal(struct {
		Version, Binding, Owner, Manifest string
		Facts                             []string
	}{"selection-digest-v1", binding, owner, manifest, facts})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
