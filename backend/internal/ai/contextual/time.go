package contextual

import "time"

type CaptureStamp struct {
	Status string `json:"status"`
	Value  string `json:"value,omitempty"`
	Day    string `json:"day,omitempty"`
}

func ParseCapture(raw string) (CaptureStamp, time.Time) {
	if raw == "" {
		return CaptureStamp{Status: "missing"}, time.Time{}
	}
	if len(raw) > 128 {
		return CaptureStamp{Status: "invalid"}, time.Time{}
	}
	if instant, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return CaptureStamp{Status: "offset", Value: raw, Day: raw[:10]}, instant
	}
	for _, layout := range []string{"2006-01-02T15:04:05.999999999", "2006-01-02 15:04:05.999999999", "2006-01-02"} {
		if _, err := time.Parse(layout, raw); err == nil {
			return CaptureStamp{Status: "local_unknown", Value: raw, Day: raw[:10]}, time.Time{}
		}
	}
	return CaptureStamp{Status: "invalid"}, time.Time{}
}
