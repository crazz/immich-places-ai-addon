package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type aiWriteRuntime struct {
	store    *aiWriteStore
	lock     *os.File
	stopping atomic.Bool
}

func newAIWriteRuntime(results *aiResultStore, images *aiImagePreparer, syncService *SyncService, cfg *Config) *aiWriteRuntime {
	r := &aiWriteRuntime{}
	r.lock, _ = acquireAIWriteLock(filepath.Join(cfg.DataDir, "ai-write.lock"))
	r.store = &aiWriteStore{drafts: &aiDraftStore{results: results}, images: images, sync: syncService, profile: cfg.AIWriteProfile, enabled: func() bool { return r.lock != nil && !r.stopping.Load() && cfg.AIEnabled && cfg.AIWriteEnabled }}
	r.store.capabilities = cfg.AIWriteCapabilities
	results.writer = r.store
	return r
}

func (r *aiWriteRuntime) sweep(ctx context.Context) error {
	if r.lock == nil {
		return nil
	}
	s := r.store
	rows, err := s.drafts.results.jobs.db.db.QueryContext(ctx, `SELECT userID,id FROM (
 SELECT o.userID,o.id,o.approvedAt FROM ai_all_write_operations o JOIN ai_all_write_targets t ON t.userID=o.userID AND t.installationID=o.installationID AND t.operationID=o.id WHERE o.installationID=? AND (o.status='queued' OR (o.status IN ('writing','verifying') AND t.reads<3 AND t.dueAt<=? AND (t.senderActive=0 OR t.leaseUntil<=?)))
 UNION ALL
 SELECT o.userID,o.id,o.approvedAt FROM ai_stack_write_operations o WHERE o.installationID=? AND (
 EXISTS(SELECT 1 FROM ai_stack_write_targets t WHERE t.userID=o.userID AND t.installationID=o.installationID AND t.operationID=o.id AND (t.status='queued' OR (t.status IN ('writing','verifying') AND t.reads<3 AND t.dueAt<=? AND (t.senderActive=0 OR t.leaseUntil<=?))))
 OR EXISTS(SELECT 1 FROM ai_mirror_write_steps m JOIN ai_stack_write_targets t ON t.userID=m.userID AND t.installationID=m.installationID AND t.operationID=m.operationID AND t.assetID=m.assetID WHERE m.userID=o.userID AND m.installationID=o.installationID AND m.operationID=o.id AND (
  (m.status IN ('blocked','queued') AND t.status='succeeded' AND t.verified=1 AND t.completionKnown=1 AND t.senderActive=0)
  OR (m.status IN ('writing','verifying') AND m.reads<3 AND m.dueAt<=? AND (m.senderActive=0 OR m.leaseUntil<=?)))))
 ) ORDER BY approvedAt,id LIMIT 100`, s.drafts.results.jobs.binding, s.drafts.results.jobs.now().UnixNano(), s.drafts.results.jobs.now().UnixNano(), s.drafts.results.jobs.binding, s.drafts.results.jobs.now().UnixNano(), s.drafts.results.jobs.now().UnixNano(), s.drafts.results.jobs.now().UnixNano(), s.drafts.results.jobs.now().UnixNano())
	if err != nil {
		return err
	}
	var work [][2]string
	for rows.Next() {
		var item [2]string
		if err = rows.Scan(&item[0], &item[1]); err != nil {
			rows.Close()
			return err
		}
		work = append(work, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	queue := make(chan [2]string, len(work))
	for _, item := range work {
		queue <- item
	}
	close(queue)
	var wg sync.WaitGroup
	failures := make(chan error, len(work))
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range queue {
				if ctx.Err() != nil {
					return
				}
				if err := s.runOne(ctx, item[0], item[1]); err != nil {
					failures <- err
				}
			}
		}()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		return err
	}
	return nil
}

func (r *aiWriteRuntime) run(ctx context.Context) {
	if r.lock == nil {
		return
	}
	defer r.lock.Close()
	defer r.stopping.Store(true)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		// Durable state retains any failed sweep; a later bounded sweep can reconcile it.
		_ = r.sweep(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
