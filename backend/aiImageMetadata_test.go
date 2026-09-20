package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAIImageUpstreamEligibilityMustBeCurrentAndComplete(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"wrong asset", func(m map[string]any) { m["id"] = selectionB }},
		{"hidden", func(m map[string]any) { m["visibility"] = "hidden" }},
		{"locked", func(m map[string]any) { m["visibility"] = "locked" }},
		{"unknown visibility", func(m map[string]any) { m["visibility"] = "future" }},
		{"trashed", func(m map[string]any) { m["isTrashed"] = true }},
		{"missing trash flag", func(m map[string]any) { delete(m, "isTrashed") }},
		{"null trash flag", func(m map[string]any) { m["isTrashed"] = nil }},
		{"video", func(m map[string]any) { m["type"] = "VIDEO" }},
		{"missing revision", func(m map[string]any) { delete(m, "updatedAt") }},
		{"bad revision", func(m map[string]any) { m["updatedAt"] = "private-location" }},
		{"bad checksum", func(m map[string]any) { m["checksum"] = "private-source" }},
		{"missing owner", func(m map[string]any) { delete(m, "ownerId") }},
		{"stack child", func(m map[string]any) {
			m["stack"] = map[string]any{"id": selectionID(30), "primaryAssetId": selectionB, "assetCount": 2}
		}},
		{"incomplete stack", func(m map[string]any) { m["stack"] = map[string]any{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newAIImageFixture(t)
			m := f.metadata()
			tc.edit(m)
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if strings.HasSuffix(r.URL.Path, "/"+selectionA) {
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(m); err != nil {
						t.Error(err)
					}
					return true
				}
				return false
			}
			p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
			calls, _ := f.counts()
			if err == nil || p != nil || calls != 1 {
				t.Fatalf("invalid metadata allowed preview: calls=%d err=%v", calls, err)
			}
			if strings.Contains(err.Error(), "private") {
				t.Fatal("metadata disclosed")
			}
		})
	}
}

func TestAIImageRejectsChangedSourceRevisionAndEligibility(t *testing.T) {
	for _, field := range []string{"updatedAt", "checksum", "ownerId", "visibility", "type", "stack"} {
		t.Run(field, func(t *testing.T) {
			f := newAIImageFixture(t)
			var metadataReads atomic.Int32
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path != "/api/assets/"+selectionA {
					return false
				}
				m := f.metadata()
				if metadataReads.Add(1) == 2 {
					switch field {
					case "updatedAt":
						m[field] = "2026-09-20T12:00:01Z"
					case "checksum":
						m[field] = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{2}, 20))
					case "ownerId":
						m[field] = selectionID(98)
					case "visibility":
						m[field] = "hidden"
					case "type":
						m[field] = "VIDEO"
					case "stack":
						m[field] = map[string]any{"id": selectionID(30), "primaryAssetId": selectionA, "assetCount": 2}
					}
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(m); err != nil {
					t.Error(err)
				}
				return true
			}
			p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
			if p != nil {
				p.Release()
			}
			calls, _ := f.counts()
			if err == nil || p != nil || calls != 3 {
				t.Fatalf("changed source published or retried: calls=%d err=%v", calls, err)
			}
		})
	}
}

func TestAIImageRejectsAmbiguousEligibilityMetadata(t *testing.T) {
	for _, field := range []string{"isTrashed", "IsTrashed"} {
		t.Run(field, func(t *testing.T) {
			f := newAIImageFixture(t)
			m := f.metadata()
			m["isTrashed"] = true
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			body := strings.TrimSuffix(string(data), "}") + `,"` + field + `":false}`
			f.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if r.URL.Path == "/api/assets/"+selectionA {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(body))
					return true
				}
				return false
			}
			p, err := f.service.prepare(context.Background(), testUserID, f.store.binding, selectionA)
			if p != nil {
				p.Release()
			}
			calls, _ := f.counts()
			if err == nil || p != nil || calls != 1 {
				t.Fatalf("ambiguous eligibility accepted: calls=%d err=%v", calls, err)
			}
		})
	}
}
