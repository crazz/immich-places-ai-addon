package capabilities

// UsageParser is optional telemetry; missing metadata never implies zero usage.
type UsageParser interface {
	ParseUsage(body []byte) *Usage
}

// CombineUsage only totals fields reported for every dispatched probe.
func CombineUsage(a, b *Usage) *Usage {
	if a == nil || b == nil {
		return nil
	}
	out := &Usage{
		PromptTokens:     sumKnownTokens(a.PromptTokens, b.PromptTokens),
		CompletionTokens: sumKnownTokens(a.CompletionTokens, b.CompletionTokens),
		TotalTokens:      sumKnownTokens(a.TotalTokens, b.TotalTokens),
	}
	if out.PromptTokens == nil && out.CompletionTokens == nil && out.TotalTokens == nil {
		return nil
	}
	return out
}

func sumKnownTokens(a, b *int) *int {
	if a == nil || b == nil || *a < 0 || *b < 0 || *a > int(^uint(0)>>1)-*b {
		return nil
	}
	total := *a + *b
	return &total
}
