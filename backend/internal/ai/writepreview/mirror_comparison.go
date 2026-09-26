package writepreview

import (
	"encoding/json"
	"math/big"
)

func EqualMirrorValue(left, right []byte) bool {
	a, aErr := parseMirrorJSON(left, MirrorValueLimit)
	b, bErr := parseMirrorJSON(right, MirrorValueLimit)
	_, aObject := a.(map[string]any)
	_, bObject := b.(map[string]any)
	return aErr == nil && bErr == nil && aObject && bObject && equalMirrorJSON(a, b)
}

func equalMirrorJSON(a, b any) bool {
	switch left := a.(type) {
	case map[string]any:
		right, ok := b.(map[string]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for key, value := range left {
			other, present := right[key]
			if !present || !equalMirrorJSON(value, other) {
				return false
			}
		}
		return true
	case []any:
		right, ok := b.([]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for i, value := range left {
			if !equalMirrorJSON(value, right[i]) {
				return false
			}
		}
		return true
	case json.Number:
		right, ok := b.(json.Number)
		if !ok {
			return false
		}
		x, xOK := new(big.Rat).SetString(string(left))
		y, yOK := new(big.Rat).SetString(string(right))
		return xOK && yOK && x.Cmp(y) == 0
	default:
		return a == b
	}
}
