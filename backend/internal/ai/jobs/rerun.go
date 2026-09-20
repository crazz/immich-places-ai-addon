package jobs

type RerunChoice struct {
	ParentJobID string `json:"parentJobId"`
	Kind        string `json:"kind"`
}
