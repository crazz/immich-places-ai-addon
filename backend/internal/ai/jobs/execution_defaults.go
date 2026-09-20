package jobs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

const DefaultExecutionVersion = "execution-default-v1"

// Resolve preserves explicit operator restrictions and supplies ordinary launch settings otherwise.
func (p ExecutionPolicies) Resolve(binding ExecutionBinding) (ExecutionPolicy, string, bool) {
	if policy, id, ok := p.Find(binding); ok {
		return policy, id, true
	}
	for _, policy := range p.entries {
		if policy.Binding.Owner == binding.Owner && policy.Binding.Profile == binding.Profile {
			return ExecutionPolicy{}, "", false
		}
	}
	policy := ExecutionPolicy{
		Version: DefaultExecutionVersion, Binding: binding, OutputField: "max_tokens",
		MaxInputTokens: 100000, MaxOutputTokens: 16384,
		MaxRequestBytes: 15 << 20, MaxImageBytes: 10 << 20, Context: true,
	}
	if !policy.Valid() {
		return ExecutionPolicy{}, "", false
	}
	data, _ := json.Marshal(policy)
	digest := sha256.Sum256(data)
	return policy, hex.EncodeToString(digest[:]), true
}
