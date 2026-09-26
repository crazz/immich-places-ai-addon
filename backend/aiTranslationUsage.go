package main

import (
	"context"
	"database/sql"
	"immich-places-backend/internal/ai/analysis"
	"immich-places-backend/internal/ai/drafts"
	"immich-places-backend/internal/ai/jobs"
)

func (s *aiTranslationStore) reserveUsage(ctx context.Context, owner string, run aiTranslationRun, tag string) error {
	return s.drafts.write(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := s.authorizeItem(ctx, tx, owner, run, tag); err != nil {
			return err
		}
		var tokens, cost int64
		binding := s.drafts.results.jobs.binding
		if tx.QueryRowContext(ctx, `SELECT COALESCE(sum(inputReserved+outputReserved),0),COALESCE(sum(estimatedMicros),0) FROM ai_translation_usage WHERE userID=? AND installationID=? AND runID=?`, owner, binding, run.ID).Scan(&tokens, &cost) != nil {
			return drafts.ErrStorage
		}
		output := min(run.Policy.MaxOutputTokens, 4096)
		allowance := run.Policy.MaxInputTokens + output
		budget := run.Request.MaxTokens
		if budget == 0 {
			budget = allowance * int64(len(run.Request.Languages))
		}
		estimate := run.Policy.EstimatedMicros(output)
		if allowance > budget-tokens {
			return jobs.ErrBudget
		}
		if cap := run.Request.MaxEstimatedMicros; cap != nil && (estimate == nil || *estimate > *cap-cost) {
			return jobs.ErrBudget
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO ai_translation_usage(userID,installationID,runID,language,inputReserved,outputReserved,estimatedMicros) VALUES(?,?,?,?,?,?,?)`, owner, binding, run.ID, tag, run.Policy.MaxInputTokens, output, estimate)
		if err != nil {
			return drafts.ErrStorage
		}
		return nil
	})
}

func (s *aiTranslationStore) recordUsage(ctx context.Context, owner string, run aiTranslationRun, tag string, usage *analysis.Usage) error {
	if usage == nil {
		return nil
	}
	return s.drafts.write(context.WithoutCancel(ctx), func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE ai_translation_usage SET inputReported=?,outputReported=?,totalReported=? WHERE userID=? AND installationID=? AND runID=? AND language=?`, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens, owner, s.drafts.results.jobs.binding, run.ID, tag)
		if err != nil {
			return err
		}
		input, output := run.Policy.MaxInputTokens, min(run.Policy.MaxOutputTokens, 4096)
		violated := (usage.PromptTokens != nil && int64(*usage.PromptTokens) > input) || (usage.CompletionTokens != nil && int64(*usage.CompletionTokens) > output) || (usage.TotalTokens != nil && int64(*usage.TotalTokens) > input+output)
		if run.Policy.Version != jobs.DefaultExecutionVersion && violated {
			_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO ai_execution_policy_violations VALUES(?,?,?)`, owner, run.PolicyID, s.drafts.results.jobs.now().UnixNano())
		}
		return err
	})
}
