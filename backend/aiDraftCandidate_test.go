package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"immich-places-backend/internal/ai/jobs"
	"immich-places-backend/internal/ai/results"
)

func draftProposalFixture(t *testing.T, change func(*results.Document)) (*aiJobFixture, string) {
	t.Helper()
	f, id := draftFixture(t)
	old, err := f.store.ReadAnalysis(context.Background(), testUserID, id)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.store.Cancel(context.Background(), testUserID, old.JobID); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("internal/ai/results/testdata/ai-analysis-result.synthetic.json")
	if err != nil {
		t.Fatal(err)
	}
	var document results.Document
	if err = json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	change(&document)
	f.input.Key = "draft-proposal"
	f.input.Languages = []string{"en", "uk"}
	if _, err = f.store.Submit(context.Background(), f.input); err != nil {
		t.Fatal(err)
	}
	lease, ok, err := f.store.Claim(context.Background(), jobs.DefaultPolicy())
	if err != nil || !ok {
		t.Fatal(err)
	}
	if err = f.store.Reserve(context.Background(), lease); err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	validator, err := results.New()
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := validator.Validate(raw, results.Context{Mode: results.Visual, Completion: results.Complete, Languages: f.input.Languages, PrimaryLanguage: "en"})
	if err != nil {
		t.Fatal(err)
	}
	completion := aiJobCompletion(t)
	completion.Proposal = proposal
	id, err = f.store.Complete(context.Background(), lease, completion)
	if err != nil {
		t.Fatal(err)
	}
	return f, id
}

func TestAIDraftExplicitAmbiguousCandidateStagesCoarsePoint(t *testing.T) {
	f, id := draftProposalFixture(t, func(d *results.Document) {
		d.Outcome = "ambiguous"
		d.SelectedCandidateID = nil
		radius := json.Number("5000")
		d.Candidates[0].CameraLocation.EstimatedRadiusM = &radius
		other := d.Candidates[0]
		other.ID = "candidate-2"
		d.Candidates = append(d.Candidates, other)
	})
	rec := aiRequest(draftHandler(f), "POST", "/ai/results/"+id+"/draft", `{"candidateId":"candidate-2"}`, aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatalf("candidate acceptance %d %s", rec.Code, rec.Body.String())
	}
	var value struct {
		ID          string   `json:"id"`
		CandidateID *string  `json:"candidateId"`
		Radius      *float64 `json:"radius"`
	}
	if json.Unmarshal(rec.Body.Bytes(), &value) != nil || value.CandidateID == nil || *value.CandidateID != "candidate-2" || value.Radius == nil || *value.Radius != 5000 {
		t.Fatal("candidate choice lost", value)
	}
	rec = draftPatch(f, value.ID, `"1"`, `{"state":"staged","fields":["gps"]}`)
	if rec.Code != 200 {
		t.Fatal("coarse point blocked", rec.Code, rec.Body.String())
	}
}

func TestAIDraftCandidateChangeInvalidatesFactsWithoutMovingSubjectIntoCamera(t *testing.T) {
	f, id := draftProposalFixture(t, func(d *results.Document) {
		other := d.Candidates[0]
		other.ID = "subject-only"
		other.CameraLocation = nil
		other.CameraDirection = nil
		d.Candidates = append(d.Candidates, other)
	})
	value := acceptedDraft(t, f, id)
	rec := draftPatch(f, value.ID, `"1"`, `{"candidateId":"subject-only"}`)
	if rec.Code != 200 {
		t.Fatalf("candidate change %d %s", rec.Code, rec.Body.String())
	}
	if json.Unmarshal(rec.Body.Bytes(), &value) != nil || value.Camera != nil || value.CandidateID == nil || *value.CandidateID != "subject-only" || !value.HeadingStale || !value.RadiusStale || value.FactsRevision != 2 {
		t.Fatal("candidate transition invented support", value)
	}
}
