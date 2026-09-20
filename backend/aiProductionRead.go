package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"immich-places-backend/internal/ai/jobs"
)

type aiJobProgress struct {
	ID            string              `json:"id"`
	Configuration jobs.Configuration  `json:"configuration"`
	CreatedAt     int64               `json:"createdAt"`
	Canceled      bool                `json:"canceled"`
	Blocked       bool                `json:"blocked"`
	Counts        map[string]int      `json:"counts"`
	Items         []aiJobItemProgress `json:"items"`
	Usage         aiJobUsage          `json:"usage"`
}
type aiJobItemProgress struct {
	ID       string  `json:"id"`
	AssetID  string  `json:"assetId"`
	State    string  `json:"state"`
	Failure  string  `json:"failure,omitempty"`
	Attempts int     `json:"attempts"`
	Calls    int     `json:"calls"`
	ResultID *string `json:"resultId"`
	Outcome  string  `json:"outcome,omitempty"`
}
type aiJobUsage struct {
	Calls           int    `json:"calls"`
	ReservedTokens  int64  `json:"reservedTokens"`
	ReportedStatus  string `json:"reportedStatus"`
	CostStatus      string `json:"costStatus"`
	InputReported   *int64 `json:"inputReported"`
	OutputReported  *int64 `json:"outputReported"`
	TotalReported   *int64 `json:"totalReported"`
	EstimatedMicros *int64 `json:"estimatedMicros"`
	Currency        string `json:"currency,omitempty"`
}

func (p *aiProductionJobs) progress(ctx context.Context, owner, id string) (aiJobProgress, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := p.store.db.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return aiJobProgress{}, jobs.ErrStorage
	}
	defer tx.Rollback()
	if err = p.store.currentInstallation(ctx, tx); err != nil {
		return aiJobProgress{}, err
	}
	result, err := p.readProgress(ctx, tx, owner, id)
	if err != nil {
		return aiJobProgress{}, aiJobFailure(ctx, err)
	}
	if err = tx.Commit(); err != nil {
		return aiJobProgress{}, aiJobFailure(ctx, err)
	}
	return result, nil
}
func (p *aiProductionJobs) readProgress(ctx context.Context, tx *sql.Tx, owner, id string) (aiJobProgress, error) {
	result := aiJobProgress{ID: id, Counts: map[string]int{}, Items: []aiJobItemProgress{}, Usage: aiJobUsage{ReportedStatus: "unknown", CostStatus: "unknown"}}
	var request, policyJSON string
	err := tx.QueryRowContext(ctx, `SELECT a.requestJSON,j.createdAt,j.cancelRequested,j.blocked,j.calls,a.policyJSON FROM ai_jobs j JOIN ai_job_admissions a ON a.userID=j.userID AND a.jobID=j.id WHERE j.userID=? AND j.id=? AND j.installationID=?`, owner, id, p.store.binding).Scan(&request, &result.CreatedAt, &result.Canceled, &result.Blocked, &result.Usage.Calls, &policyJSON)
	if err == sql.ErrNoRows {
		return aiJobProgress{}, jobs.ErrDenied
	}
	if err != nil {
		return aiJobProgress{}, err
	}
	var req jobs.Admission
	if len(request) > 32<<10 || json.Unmarshal([]byte(request), &req) != nil {
		return aiJobProgress{}, jobs.ErrStorage
	}
	result.Configuration = req.Configuration
	var policy jobs.ExecutionPolicy
	if len(policyJSON) > 8192 || json.Unmarshal([]byte(policyJSON), &policy) != nil {
		return aiJobProgress{}, jobs.ErrStorage
	}
	if policy.Currency != "" {
		result.Usage.CostStatus = "estimated"
		result.Usage.Currency = policy.Currency
	}

	var attempts, complete, reported int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE(sum(inputReserved+outputReserved),0),sum(inputReported),sum(outputReported),sum(totalReported),count(*),COALESCE(sum(inputReported IS NOT NULL AND outputReported IS NOT NULL),0),COALESCE(sum(inputReported IS NOT NULL OR outputReported IS NOT NULL OR totalReported IS NOT NULL),0) FROM ai_job_usage WHERE userID=? AND jobID=?`, owner, id).Scan(&result.Usage.ReservedTokens, &result.Usage.InputReported, &result.Usage.OutputReported, &result.Usage.TotalReported, &attempts, &complete, &reported); err != nil {
		return aiJobProgress{}, err
	}
	if reported > 0 {
		result.Usage.ReportedStatus = "partial"
	}
	if attempts > 0 && complete == attempts {
		result.Usage.ReportedStatus = "complete"
	}

	if result.Usage.CostStatus == "estimated" {
		var estimate int64
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE(sum(estimatedMicros),0) FROM ai_job_usage WHERE userID=? AND jobID=?`, owner, id).Scan(&estimate); err != nil {
			return aiJobProgress{}, err
		}
		result.Usage.EstimatedMicros = &estimate
	}
	rows, err := tx.QueryContext(ctx, `SELECT i.id,i.assetID,i.state,i.failure,i.attempts,i.calls,a.id,COALESCE(a.outcome,'') FROM ai_job_items i LEFT JOIN ai_analyses a ON a.userID=i.userID AND a.jobID=i.jobID AND a.itemID=i.id WHERE i.userID=? AND i.jobID=? ORDER BY i.position LIMIT 501`, owner, id)
	if err != nil {
		return aiJobProgress{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var item aiJobItemProgress
		if err := rows.Scan(&item.ID, &item.AssetID, &item.State, &item.Failure, &item.Attempts, &item.Calls, &item.ResultID, &item.Outcome); err != nil {
			return aiJobProgress{}, err
		}
		result.Items = append(result.Items, item)
		result.Counts[item.State]++
		result.Counts["total"]++
	}
	if len(result.Items) > 500 {
		return aiJobProgress{}, jobs.ErrStorage
	}
	return result, rows.Err()
}
