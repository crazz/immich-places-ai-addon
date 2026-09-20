package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"immich-places-backend/internal/ai/contextual"
)

func TestAIContextSelectedAlbumUsesExactCurrentMembership(t *testing.T) {
	f := newAIImageFixture(t)
	album := selectionID(90)
	selectionSQL(t, f.db, "INSERT INTO albums (userID,immichID,albumName,updatedAt) VALUES (?,?,?,'')", testUserID, album, "stale label")
	selectionSQL(t, f.db, "INSERT INTO albumAssets VALUES (?,?,?)", testUserID, album, selectionA)
	searches, labels := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "private-immich-image-key" {
			t.Error("missing owner credential")
		}
		switch r.Method + " " + r.URL.Path {
		case "GET /api/assets/" + selectionA:
			_ = json.NewEncoder(w).Encode(f.metadata())
		case "POST /api/search/metadata":
			searches++
			var body struct {
				ID       string   `json:"id"`
				AlbumIDs []string `json:"albumIds"`
				Size     int      `json:"size"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || body.ID != selectionA || len(body.AlbumIDs) != 1 || body.AlbumIDs[0] != album || body.Size != 1 {
				t.Error("membership search broadened scope")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"assets": map[string]any{"items": []any{map[string]string{"id": selectionA}}}})
		case "GET /api/albums/" + album:
			labels++
			_ = json.NewEncoder(w).Encode(map[string]string{"id": album, "albumName": "Current mountain trip", "description": "private description"})
		default:
			t.Error("unrelated lookup or mutation", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	f.service.endpoint = server.URL
	req := contextRequest(t, f)
	req.Consent.Classes = []contextual.Class{contextual.AlbumLabel}
	req.Consent.AlbumID = album
	b, err := (&aiContextPreparer{images: f.service}).prepare(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := b.Projection()
	if searches != 1 || labels != 1 || !strings.Contains(string(p), "Current mountain trip") || strings.Contains(string(p), "private description") || strings.Contains(string(p), album) {
		t.Fatalf("incorrect selected album projection: %s; reads %d/%d", p, searches, labels)
	}
}

func TestAIContextAlbumLossNeverSubstitutesAnotherAlbum(t *testing.T) {
	for _, cause := range []string{"local-membership", "upstream-membership", "forbidden", "foreign", "transport"} {
		t.Run(cause, func(t *testing.T) {
			f := newAIImageFixture(t)
			album := selectionID(90)
			if cause != "foreign" {
				selectionSQL(t, f.db, "INSERT INTO albums (userID,immichID,albumName,updatedAt) VALUES (?,?,?,'')", testUserID, album, "private old label")
			}
			if cause != "local-membership" && cause != "foreign" {
				selectionSQL(t, f.db, "INSERT INTO albumAssets VALUES (?,?,?)", testUserID, album, selectionA)
			}
			labels := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/assets/" + selectionA:
					_ = json.NewEncoder(w).Encode(f.metadata())
				case "/api/search/metadata":
					if cause == "forbidden" {
						http.Error(w, "private denial", 403)
						return
					}
					if cause == "transport" {
						http.Error(w, "private failure", 503)
						return
					}
					_, _ = w.Write([]byte(`{"assets":{"items":[]}}`))
				default:
					labels++
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			f.service.endpoint = server.URL
			req := contextRequest(t, f)
			req.Consent.Classes, req.Consent.AlbumID = []contextual.Class{contextual.AlbumLabel}, album
			b, err := (&aiContextPreparer{images: f.service}).prepare(context.Background(), req)
			if cause == "foreign" || cause == "transport" {
				if err == nil || b != nil {
					t.Fatal("unresolved/foreign album accepted")
				}
			} else if err != nil || b == nil || len(b.Info().Sources) != 0 || len(b.Info().Omissions) != 1 {
				t.Fatal("lost membership did not omit optional label", err)
			}
			if labels != 0 {
				t.Fatal("album labels read outside current scope")
			}
		})
	}
}
