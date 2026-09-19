package main

import (
	"encoding/json"
	"testing"
)

func TestAISelectionMatchingExactAndOverLimitHTTP(t *testing.T) {
	db, _ := aiTestHandler(t, true)
	selectionSQL(t, db, `WITH RECURSIVE n(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM n WHERE x<500)
 INSERT INTO assets (userID,immichID,type,originalFileName,fileCreatedAt)
 SELECT ?,printf('aaaaaaaa-0000-4000-8000-%012d',x),'IMAGE','synthetic.jpg','2026-09-19' FROM n`, testUserID)
	h := newAISelectionHandler(db, &Config{AIEnabled: true, ImmichURL: "https://immich.example/api", AIPublicOrigin: aiTestOrigin})
	body := `{"mode":"all-matching","scope":{"view":"all","hiddenFilter":"all"}}`
	rec := aiRequest(h, "POST", "/ai/selection-preview", body, aiTestOrigin, true)
	if rec.Code != 200 {
		t.Fatalf("exact cap: %d %s", rec.Code, rec.Body.String())
	}
	seedAsset(t, db, selectionID(501), nil, nil, "2026-09-19")
	seedAsset(t, db, selectionID(502), nil, nil, "2026-09-19")
	selectionSQL(t, db, "UPDATE assets SET type='VIDEO' WHERE immichID=?", selectionID(502))
	rec = aiRequest(h, "POST", "/ai/selection-preview", body, aiTestOrigin, true)
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &wire); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 413 || string(wire["code"]) != `"SELECTION_LIMIT_EXCEEDED"` || string(wire["matchedCount"]) != "502" || string(wire["eligibleCount"]) != "501" || string(wire["excludedCount"]) != "1" || string(wire["snapshotID"]) != "null" || wire["assetIDs"] != nil || wire["exclusions"] != nil || string(wire["exclusionCounts"]) != `{"unsupported_type":1}` {
		t.Fatalf("overflow must expose exact counts only: %d %s", rec.Code, rec.Body.String())
	}
	var headers, items int
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_snapshots").Scan(&headers); err != nil {
		t.Fatal(err)
	}
	if err := db.db.QueryRow("SELECT count(*) FROM ai_selection_items").Scan(&items); err != nil {
		t.Fatal(err)
	}
	if headers != 1 || items != 500 {
		t.Fatalf("overflow partially published: headers=%d items=%d", headers, items)
	}
}
