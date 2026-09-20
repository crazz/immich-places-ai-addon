package review

type Entry struct {
	ID              string  `json:"id"`
	JobID           string  `json:"jobId"`
	AssetID         string  `json:"assetId"`
	AnalysisID      *string `json:"analysisId"`
	TerminalAt      string  `json:"terminalAt"`
	Mode            string  `json:"mode"`
	Model           string  `json:"model"`
	Label           *string `json:"label"`
	ExecutionState  string  `json:"executionState"`
	ProposalOutcome *string `json:"proposalOutcome"`
	ReviewState     string  `json:"reviewState"`
	WriteState      string  `json:"writeState"`
	Failure         string  `json:"failure,omitempty"`
	CaptureDay      *string `json:"captureDay"`
	AlbumID         *string `json:"albumId"`
	AlbumLabel      *string `json:"albumLabel"`
	SourceAvailable bool    `json:"sourceAvailable"`
}

type Page struct {
	Items      []Entry `json:"items"`
	NextCursor string  `json:"nextCursor,omitempty"`
}
