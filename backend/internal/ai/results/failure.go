package results

type Finding struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type Failure struct {
	Category string    `json:"category"`
	Findings []Finding `json:"findings"`
}

func (f *Failure) Error() string { return f.Category }

func failure(category, code, path string) error {
	return &Failure{Category: category, Findings: []Finding{{Code: code, Path: path}}}
}
