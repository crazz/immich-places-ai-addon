package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"immich-places-backend/internal/ai/contextual"
)

func TestAIContextRejectsAmbiguousAlbumMembership(t *testing.T) {
	f := newAIImageFixture(t)
	album := selectionID(90)
	selectionSQL(t, f.db, "INSERT INTO albums (userID,immichID,albumName,updatedAt) VALUES (?,?,?,'')", testUserID, album, "label")
	selectionSQL(t, f.db, "INSERT INTO albumAssets VALUES (?,?,?)", testUserID, album, selectionA)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/assets/" + selectionA:
			_ = json.NewEncoder(w).Encode(f.metadata())
		case "/api/search/metadata":
			_, _ = fmt.Fprintf(w, `{"assets":{"items":[],"Items":[{"id":%q}]}}`, selectionA)
		case "/api/albums/" + album:
			_ = json.NewEncoder(w).Encode(map[string]string{"id": album, "albumName": "label"})
		default:
			t.Error("unexpected operation")
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	f.service.endpoint = server.URL
	req := contextRequest(t, f)
	req.Consent.Classes, req.Consent.AlbumID = []contextual.Class{contextual.AlbumLabel}, album
	if b, err := (&aiContextPreparer{images: f.service}).prepare(context.Background(), req); err == nil || b != nil {
		t.Fatal("ambiguous membership admitted")
	}
}
