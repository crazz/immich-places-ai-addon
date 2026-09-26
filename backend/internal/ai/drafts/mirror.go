package drafts

type MirrorSelection struct {
	Direction  bool     `json:"direction"`
	Precision  bool     `json:"precision"`
	Place      bool     `json:"place"`
	Languages  []string `json:"languages"`
	Provenance bool     `json:"provenance"`
	Model      bool     `json:"model"`
}
