package review

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var ErrInvalid = errors.New("invalid result query")
var ErrUnavailable = errors.New("result unavailable")

type Query struct {
	Limit                                                         int
	Cursor, State, Outcome, Asset, Job, StartDate, EndDate, Album string
	Undated                                                       bool
}

func ParseQuery(values url.Values) (Query, error) {
	allowed := map[string]bool{"limit": true, "cursor": true, "state": true, "outcome": true, "assetId": true, "jobId": true, "startDate": true, "endDate": true, "albumId": true, "undated": true}
	for key, v := range values {
		if !allowed[key] || len(v) != 1 || len(v[0]) > 2048 {
			return Query{}, ErrInvalid
		}
	}
	q := Query{Limit: 30, Cursor: values.Get("cursor"), State: values.Get("state"), Outcome: values.Get("outcome"), Asset: values.Get("assetId"), Job: values.Get("jobId"), StartDate: values.Get("startDate"), EndDate: values.Get("endDate"), Album: values.Get("albumId"), Undated: values.Get("undated") == "true"}
	if raw, ok := values["limit"]; ok {
		limit, err := strconv.Atoi(raw[0])
		if err != nil || limit < 1 || limit > 100 {
			return Query{}, ErrInvalid
		}
		q.Limit = limit
	}
	if q.State != "" && q.State != "succeeded" && q.State != "failed" && q.State != "canceled" {
		return Query{}, ErrInvalid
	}
	if q.Outcome != "" && q.Outcome != "located" && q.Outcome != "ambiguous" && q.Outcome != "unknown" {
		return Query{}, ErrInvalid
	}
	for _, id := range []string{q.Asset, q.Job, q.Album} {
		if id != "" {
			parsed, err := uuid.Parse(id)
			if err != nil || parsed.String() != id {
				return Query{}, ErrInvalid
			}
		}
	}
	for _, day := range []string{q.StartDate, q.EndDate} {
		if day != "" {
			parsed, err := time.Parse("2006-01-02", day)
			if err != nil || parsed.Format("2006-01-02") != day {
				return Query{}, ErrInvalid
			}
		}
	}
	if q.StartDate != "" && q.EndDate != "" && q.StartDate > q.EndDate {
		return Query{}, ErrInvalid
	}
	if raw := values.Get("undated"); raw != "" && raw != "true" && raw != "false" {
		return Query{}, ErrInvalid
	}
	if q.Undated && (q.StartDate != "" || q.EndDate != "") {
		return Query{}, ErrInvalid
	}
	return q, nil
}
