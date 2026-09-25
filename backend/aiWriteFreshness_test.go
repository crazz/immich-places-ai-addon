package main

import (
	"context"
	"testing"
	"time"
)

func TestAIWriteFreshChangesCannotSend(t *testing.T) {
	for _, kind := range []string{"gps", "image", "hidden", "trashed", "key", "expiry", "disabled", "installation"} {
		t.Run(kind, func(t *testing.T) {
			w := newAIWriteFixture(t)
			switch kind {
			case "gps":
				w.meta["exifInfo"] = map[string]float64{"latitude": 1, "longitude": 2}
			case "image":
				w.meta["checksum"] = "AgICAgICAgICAgICAgICAgICAgI="
			case "hidden":
				w.meta["visibility"] = "hidden"
			case "trashed":
				w.meta["isTrashed"] = true
			case "key":
				key := "changed"
				if err := w.f.db.updateImmichAPIKey(context.Background(), testUserID, &key); err != nil {
					t.Fatal(err)
				}
			case "expiry":
				w.f.now = w.f.now.Add(5 * time.Minute)
			case "disabled":
				w.enabled = false
			case "installation":
				selectionSQL(t, w.f.db, "UPDATE ai_installation_identity SET id='new' WHERE singleton=1")
			}
			_ = w.writer.runOne(context.Background(), testUserID, w.op.ID)
			if sends, _ := w.counts(); sends != 0 {
				t.Fatal("obsolete authority sent", sends)
			}
			if kind == "gps" || kind == "image" {
				if op := w.status(t); op.Status != "conflict" {
					t.Fatal(op)
				}
			}
			if kind == "key" {
				return
			}
		})
	}
}
