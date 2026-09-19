package providers

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"
)

type Config struct {
	Name    string `json:"name"`
	BaseURL string `json:"baseURL"`
	Model   string `json:"model"`
}

type Input struct {
	Config
	Enabled bool    `json:"enabled"`
	Secret  *string `json:"secret,omitempty"`
}

func Validate(input Input) (Input, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Model = strings.TrimSpace(input.Model)
	input.BaseURL = strings.TrimRight(strings.TrimSpace(input.BaseURL), "/")
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 80 {
		return Input{}, errors.New("name must contain 1 to 80 characters")
	}
	if input.Model == "" || utf8.RuneCountInString(input.Model) > 200 {
		return Input{}, errors.New("model must contain 1 to 200 characters")
	}
	u, err := url.Parse(input.BaseURL)
	if err != nil || len(input.BaseURL) > 2048 || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return Input{}, errors.New("base URL must be an HTTP(S) URL without credentials, query or fragment, at most 2048 bytes")
	}
	if input.Secret != nil && len(*input.Secret) > 4096 {
		return Input{}, errors.New("secret must be at most 4096 bytes")
	}
	return input, nil
}
