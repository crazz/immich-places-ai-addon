package main

import (
	"encoding/json"
	"immich-places-backend/internal/ai/results"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

func TestCompiledValidatorIsConcurrentAndNeverLoadsModelURLs(t *testing.T) {
	var calls atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); w.WriteHeader(500) }))
	defer server.Close()
	v, err := results.New()
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../docs/ai-locate/examples/ai-analysis-result.unknown.json")
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(fixture, &document); err != nil {
		t.Fatal(err)
	}
	document["warnings"] = []string{server.URL + "/do-not-fetch; untrusted prose"}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	ctx := results.Context{Mode: results.Visual, Completion: results.Complete, Languages: []string{"en"}, PrimaryLanguage: "en"}
	var group sync.WaitGroup
	errors := make(chan error, 24)
	for i := 0; i < 24; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			result, err := v.Validate(data, ctx)
			if err == nil {
				_, err = json.Marshal(result)
			}
			errors <- err
		}()
	}
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("schema or prose triggered an external request")
	}
}
