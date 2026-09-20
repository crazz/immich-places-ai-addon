package main

import "immich-places-backend/internal/ai/review"

func aiResultFilters(q review.Query) (string, []any) {
	where := ""
	args := []any{}
	for _, filter := range []struct{ column, value string }{{"h.executionState", q.State}, {"a.outcome", q.Outcome}, {"h.assetID", q.Asset}, {"h.jobID", q.Job}, {"h.albumID", q.Album}} {
		if filter.value != "" {
			where += " AND " + filter.column + "=?"
			args = append(args, filter.value)
		}
	}
	if q.StartDate != "" {
		where += " AND h.captureDay>=?"
		args = append(args, q.StartDate)
	}
	if q.EndDate != "" {
		where += " AND h.captureDay<=?"
		args = append(args, q.EndDate)
	}
	if q.Undated {
		where += " AND h.captureDay IS NULL"
	}
	return where, args
}
