package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestCatalogDateRangesRejectReversedBounds(t *testing.T) {
	handlers, mux := newTestHandlers(t)
	mux.HandleFunc("GET /assets/day-counts", handlers.handleGetAssetDayCounts)
	for _, endpoint := range []string{
		"/assets", "/assets/day-counts", "/assets/missing-location-count",
		"/albums", "/folders", "/folders/assets", "/map-markers",
	} {
		t.Run(endpoint, func(t *testing.T) {
			req := withTestUser(httptest.NewRequest("GET", endpoint+"?path=/trip&startDate=2024-03-01&endDate=2024-02-29", nil))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s; want 400", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCatalogDayCountsReturnSourceLocalDays(t *testing.T) {
	handlers, mux := newTestHandlers(t)
	mux.HandleFunc("GET /assets/day-counts", handlers.handleGetAssetDayCounts)
	db := handlers.db.(*Database)
	if err := db.upsertAssets(context.Background(), testUserID, []AssetRow{
		datedAsset("east", "/trip/east.jpg", "2024-01-01T00:30:00+14:00"),
		datedAsset("west", "/trip/west.jpg", "2024-01-01T23:30:00-12:00"),
	}); err != nil {
		t.Fatal(err)
	}
	req := withTestUser(httptest.NewRequest("GET", "/assets/day-counts?startDate=2023-12-31&endDate=2024-01-02", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var counts map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &counts); err != nil || rec.Code != http.StatusOK || !reflect.DeepEqual(counts, map[string]int{"2024-01-01": 2}) {
		t.Fatalf("status = %d, counts = %v, error = %v", rec.Code, counts, err)
	}
}

func TestCatalogDateRangeHTTPContracts(t *testing.T) {
	handlers, mux := newTestHandlers(t)
	mux.HandleFunc("GET /assets/day-counts", handlers.handleGetAssetDayCounts)
	for _, endpoint := range []string{
		"/assets", "/assets/day-counts", "/assets/missing-location-count",
		"/albums", "/folders", "/folders/assets", "/map-markers",
	} {
		for _, tc := range []struct {
			name, query string
			auth        bool
			status      int
		}{
			{"same day", "startDate=2024-02-29&endDate=2024-02-29", true, 200},
			{"start only", "startDate=2024-02-29", true, 200},
			{"end only", "endDate=2024-02-29", true, 200},
			{"no bounds", "", true, 200},
			{"bad start", "startDate=2024-02-30&endDate=2024-03-01", true, 400},
			{"bad end", "startDate=2024-01-01&endDate=2024-2-29", true, 400},
			{"timestamp bound", "startDate=2024-02-29T00:00:00Z&endDate=2024-03-01", true, 400},
			{"unauthenticated", "startDate=2024-02-29&endDate=2024-02-29", false, 401},
		} {
			t.Run(endpoint+"/"+tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", endpoint+"?path=/trip&"+tc.query, nil)
				if tc.auth {
					req = withTestUser(req)
				}
				want := tc.status
				if endpoint == "/assets/day-counts" && (tc.name == "start only" || tc.name == "end only" || tc.name == "no bounds") {
					want = http.StatusBadRequest
				}
				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, req)
				if rec.Code != want {
					t.Fatalf("status = %d, body = %s; want %d", rec.Code, rec.Body.String(), want)
				}
				if want >= 400 {
					var body map[string]string
					if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["error"] == "" || len(body) != 1 {
						t.Errorf("error body = %s; decode error = %v", rec.Body.String(), err)
					}
				}
			})
		}
	}
}
