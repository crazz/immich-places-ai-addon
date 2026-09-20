package results

type Mode string

const (
	Visual          Mode = "visual"
	ContextAssisted Mode = "context-assisted"
)

type Completion string

const (
	Complete     Completion = "complete"
	Refused      Completion = "refused"
	Truncated    Completion = "truncated"
	ToolResponse Completion = "tool_response"
)

type Source struct {
	ID                   string
	ContextExtent        bool
	SourceReportedRadius bool
	ViewpointAlignment   bool
}

type Context struct {
	Mode            Mode
	Completion      Completion
	Languages       []string
	PrimaryLanguage string
	Sources         []Source
}
