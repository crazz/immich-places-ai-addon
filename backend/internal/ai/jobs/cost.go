package jobs

func (p ExecutionPolicy) EstimatedMicros(output int64) *int64 {
	if !p.Valid() || p.InputMicrosPerMillion == nil || output < 1 || output > p.MaxOutputTokens {
		return nil
	}
	value := (p.MaxInputTokens*(*p.InputMicrosPerMillion) + output*(*p.OutputMicrosPerMillion) + 999999) / 1000000
	return &value
}
