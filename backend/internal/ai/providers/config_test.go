package providers

import (
	"strings"
	"testing"
)

func TestValidateConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name  string
		edit  func(*Input)
		valid bool
	}{
		{"valid", func(*Input) {}, true},
		{"anonymous local", func(v *Input) { v.BaseURL = "http://localhost:8080/v1"; v.Secret = nil }, true},
		{"empty name", func(v *Input) { v.Name = " " }, false},
		{"long name", func(v *Input) { v.Name = strings.Repeat("x", 81) }, false},
		{"empty model", func(v *Input) { v.Model = " " }, false},
		{"long model", func(v *Input) { v.Model = strings.Repeat("x", 201) }, false},
		{"relative URL", func(v *Input) { v.BaseURL = "/v1" }, false},
		{"bad URL", func(v *Input) { v.BaseURL = "https://%" }, false},
		{"bad scheme", func(v *Input) { v.BaseURL = "ftp://example.com" }, false},
		{"missing host", func(v *Input) { v.BaseURL = "https:///v1" }, false},
		{"URL credentials", func(v *Input) { v.BaseURL = "https://secret@example.com" }, false},
		{"URL query", func(v *Input) { v.BaseURL = "https://example.com?key=secret" }, false},
		{"URL fragment", func(v *Input) { v.BaseURL = "https://example.com#secret" }, false},
		{"long URL", func(v *Input) { v.BaseURL = "https://example.com/" + strings.Repeat("x", 2048) }, false},
		{"long secret", func(v *Input) { secret := strings.Repeat("x", 4097); v.Secret = &secret }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			secret := " key bytes "
			input := Input{Config: Config{Name: " Provider ", BaseURL: " https://example.com/v1/ ", Model: " vision "}, Enabled: true, Secret: &secret}
			tc.edit(&input)
			got, err := Validate(input)
			if (err == nil) != tc.valid {
				t.Fatalf("valid = %v, error = %v", tc.valid, err)
			}
			if tc.valid && (got.Name != "Provider" || got.Model != "vision") {
				t.Errorf("configuration not normalized: %+v", got.Config)
			}
			if tc.name == "valid" && (got.BaseURL != "https://example.com/v1" || *got.Secret != secret) {
				t.Error("URL normalization or exact secret preservation failed")
			}
		})
	}
}
