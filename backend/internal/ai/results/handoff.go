package results

func CanonicalSchema() []byte { return []byte(canonicalSchema) }

func ResearchSchema() []byte { return []byte(researchSchema) }

func NormalizeLanguages(languages []string, primary string) ([]string, string, error) {
	validated, err := validateContext(Context{Mode: Visual, Languages: languages, PrimaryLanguage: primary})
	if err != nil {
		return nil, "", err
	}
	tags := make([]string, len(languages))
	for i, raw := range languages {
		tags[i], _ = normalizeLanguage(raw)
	}
	return tags, validated.primary, nil
}
