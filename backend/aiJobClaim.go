package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/jobs"
)

func (s *aiJobStore) Claim(ctx context.Context, policy jobs.Policy) (jobs.Lease, bool, error) {
	return s.claim(ctx, policy, false)
}

func (s *aiJobStore) claim(ctx context.Context, policy jobs.Policy, productionOnly bool) (jobs.Lease, bool, error) {
	if !policy.Valid() {
		return jobs.Lease{}, false, jobs.ErrInvalid
	}
	var lease jobs.Lease
	claimed := false
	err := s.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.current(ctx, tx); err != nil {
			return err
		}
		now := s.now()
		var active int
		if err := tx.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM ai_job_items WHERE state='running' AND leaseExpiresAt>?) + (SELECT count(*) FROM ai_translation_items WHERE state='reserved')`, now.UnixNano()).Scan(&active); err != nil {
			return err
		}
		if active >= policy.Global {
			return nil
		}
		err := tx.QueryRowContext(ctx, `SELECT i.userID,i.jobID,i.id,i.assetID,i.attempts,i.calls
   FROM ai_job_items i JOIN ai_jobs j ON j.userID=i.userID AND j.id=i.jobID
   WHERE (?=0 OR EXISTS(SELECT 1 FROM ai_job_admissions a WHERE a.userID=j.userID AND a.jobID=j.id AND a.version='production-v1')) AND j.installationID=? AND j.cancelRequested=0 AND j.blocked=0 AND i.state IN ('queued','retry_wait') AND i.nextAttemptAt<=? AND i.attempts<3 AND i.calls<3 AND j.calls<j.maxCalls
   AND ((SELECT count(*) FROM ai_job_items active WHERE active.userID=i.userID AND active.state='running' AND active.leaseExpiresAt>?) + (SELECT count(*) FROM ai_translation_items translated WHERE translated.userID=i.userID AND translated.state='reserved'))<?
   ORDER BY j.createdAt,j.id,i.position LIMIT 1`, productionOnly, s.binding, now.UnixNano(), now.UnixNano(), policy.PerOwner).Scan(&lease.Owner, &lease.JobID, &lease.ItemID, &lease.Asset, &lease.Attempts, &lease.Calls)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		lease.Token = uuid.NewString()
		lease.Installation = s.binding
		lease.ExpiresAt = now.Add(policy.LeaseDuration)
		lease.Attempts++
		if _, err = tx.ExecContext(ctx, `UPDATE ai_job_items SET state='running',attempts=attempts+1,leaseCalls=0,leaseToken=?,leaseExpiresAt=?,failure='' WHERE userID=? AND jobID=? AND id=?`, lease.Token, lease.ExpiresAt.UnixNano(), lease.Owner, lease.JobID, lease.ItemID); err != nil {
			return err
		}
		job, err := s.readJob(ctx, tx, lease.Owner, lease.JobID)
		if err != nil {
			return err
		}
		lease.Input = job.Input
		lease.Model = job.Model
		claimed = true
		return nil
	})
	if err != nil {
		return jobs.Lease{}, false, err
	}
	return lease, claimed, nil
}
