package translations

type Outcome struct {
	Language string  `json:"language"`
	Status   string  `json:"status"`
	Text     *string `json:"text"`
}
