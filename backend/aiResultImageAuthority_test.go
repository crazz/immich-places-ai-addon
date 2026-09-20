package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAIResultThumbnailDeniesLostLocalAndUpstreamAccess(t *testing.T) {
	for _, kind := range []string{"hidden", "removed", "hidden-library", "foreign", "upstream-forbidden", "upstream-hidden", "changed-key", "changed-installation"} {
		t.Run(kind, func(t *testing.T) {
			f, p, req := productionFixture(t)
			productionSession(t, f)
			ctx := context.Background()
			job, err := p.submit(ctx, testUserID, req)
			if err != nil {
				t.Fatal(err)
			}
			if err = p.store.Cancel(ctx, testUserID, job.ID); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "hidden":
				selectionSQL(t, f.image.db, `UPDATE assets SET isHidden=1 WHERE userID=?`, testUserID)
			case "removed":
				selectionSQL(t, f.image.db, `DELETE FROM assets WHERE userID=?`, testUserID)
			case "hidden-library":
				selectionSQL(t, f.image.db, `INSERT INTO libraries(libraryID,name,isHidden) VALUES('private','Private',1)`)
				selectionSQL(t, f.image.db, `UPDATE assets SET libraryID='private' WHERE userID=?`, testUserID)
			case "foreign":
				job.ID = selectionB
			}
			f.image.handle = func(w http.ResponseWriter, r *http.Request) bool {
				if kind == "upstream-forbidden" {
					http.Error(w, "private upstream text", 403)
					return true
				}
				if kind == "upstream-hidden" {
					m := f.image.metadata()
					m["visibility"] = "locked"
					w.Header().Set("Content-Type", "application/json")
					if err := json.NewEncoder(w).Encode(m); err != nil {
						t.Error(err)
					}
					return true
				}
				if strings.Contains(r.URL.Path, "thumbnail") {
					if kind == "changed-key" {
						key := "revoked-key"
						if err := f.image.db.updateImmichAPIKey(ctx, testUserID, &key); err != nil {
							t.Error(err)
						}
					}
					if kind == "changed-installation" {
						selectionSQL(t, f.image.db, `UPDATE ai_installation_identity SET id=?`, selectionB)
					}
				}
				return false
			}
			h := newAIResultHandler(&aiResultStore{jobs: p.store}, f.image.service)
			before, _ := f.image.counts()
			rec := aiRequest(h, "GET", "/ai/jobs/"+job.ID+"/items/"+job.Items[0].ID+"/thumbnail", "", "", true)
			if rec.Code != 404 || strings.HasPrefix(rec.Header().Get("Content-Type"), "image/") || strings.Contains(rec.Body.String(), "private upstream") {
				t.Fatal("unauthorized image returned", rec.Code)
			}
			reads, _ := f.image.counts()
			if (kind == "hidden" || kind == "removed" || kind == "hidden-library" || kind == "foreign") && reads != before {
				t.Fatal("local denial reached upstream", reads)
			}
			if f.hits.Load() != 0 {
				t.Fatal("image read called provider")
			}
		})
	}
}
