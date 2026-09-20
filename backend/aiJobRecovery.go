package main

import (
	"context"
	"database/sql"

	"immich-places-backend/internal/ai/jobs"
)

type aiJobExpired struct {
	owner, job, item, token             string
	attempts, calls, jobCalls, maxCalls int
	canceled, blocked                   bool
}

func (s *aiJobStore) Recover(ctx context.Context) (int, error) {
	recovered := 0
	err := s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.current(ctx, tx); err != nil {
			return err
		}
		now := s.now()
		rows, err := tx.QueryContext(ctx, `SELECT i.userID,i.jobID,i.id,i.leaseToken,i.attempts,i.calls,j.calls,j.maxCalls,j.cancelRequested,j.blocked FROM ai_job_items i JOIN ai_jobs j ON j.userID=i.userID AND j.id=i.jobID WHERE j.installationID=? AND i.state='running' AND i.leaseExpiresAt<=? ORDER BY i.leaseExpiresAt,i.id LIMIT 100`, s.binding, now.UnixNano())
		if err != nil {
			return err
		}
		defer rows.Close()
		var expired []aiJobExpired
		for rows.Next() {
			var item aiJobExpired
			if err := rows.Scan(&item.owner, &item.job, &item.item, &item.token, &item.attempts, &item.calls, &item.jobCalls, &item.maxCalls, &item.canceled, &item.blocked); err != nil {
				return err
			}
			expired = append(expired, item)
		}
		if err = rows.Err(); err != nil {
			return err
		}
		if err = rows.Close(); err != nil {
			return err
		}
		for _, item := range expired {
			next := jobs.Interrupted(item.attempts, item.calls, item.jobCalls, item.maxCalls, item.canceled)
			if item.blocked && !item.canceled {
				next = jobs.Transition{State: "blocked", Failure: "blocked"}
			}
			var finished *int64
			if next.State != "retry_wait" {
				value := now.UnixNano()
				finished = &value
			}
			if _, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state=?,failure=?,nextAttemptAt=?,leaseToken=NULL,leaseExpiresAt=NULL,finishedAt=? WHERE userID=? AND jobID=? AND id=? AND leaseToken=?`, next.State, next.Failure, now.Add(next.Delay).UnixNano(), finished, item.owner, item.job, item.item, item.token); err != nil {
				return err
			}
			recovered++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return recovered, nil
}
