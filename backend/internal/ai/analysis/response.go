package analysis

type Usage struct{ PromptTokens, CompletionTokens, TotalTokens *int }
type Response struct {
	Content []byte
	Usage   *Usage
}
