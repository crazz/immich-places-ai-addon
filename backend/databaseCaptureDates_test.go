package main

import (
	"context"
	"reflect"
	"slices"
	"testing"
)

func TestCaptureDayCountsPreserveOffsetCalendarDate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	assets := []AssetRow{
		datedAsset("east", "/trip/east.jpg", "2024-01-01T00:30:00+14:00"),
		datedAsset("west", "/trip/west.jpg", "2024-01-01T23:30:00-12:00"),
	}
	if err := db.upsertAssets(ctx, testUserID, assets); err != nil {
		t.Fatal(err)
	}
	counts, err := db.countAssetsByDay(ctx, testUserID, "", "", gpsFilterAll, hiddenFilterVisible, "2023-12-31", "2024-01-02")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(counts, map[string]int{"2024-01-01": 2}) {
		t.Fatalf("source-local counts = %v; want January 1: 2", counts)
	}
	filtered, err := db.getFilteredAssets(ctx, testUserID, "", "", gpsFilterAll, hiddenFilterVisible, "2024-01-01", "2024-01-01", 1, 10)
	if err != nil || len(filtered) != 2 {
		t.Fatalf("January 1 assets = %v, error = %v", filtered, err)
	}
}

func TestCaptureDateRepresentationsAndOrdering(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	dates := map[string]string{
		"before":      "2024-02-28T23:59:59-12:00",
		"date":        "2024-02-29",
		"space":       "2024-02-29 00:00:00",
		"offsetless":  "2024-02-29T23:59:59.999999",
		"east":        "2024-02-29T00:00:00+14:00",
		"west":        "2024-02-29T23:59:59-12:00",
		"after":       "2024-03-01T00:00:00+14:00",
		"dst-summer":  "2024-10-27T02:30:00+02:00",
		"dst-winter":  "2024-10-27T02:30:00+01:00",
		"dst-unknown": "2024-10-27T02:30:00",
	}
	var assets []AssetRow
	for id, date := range dates {
		asset := datedAsset(id, "/trip/"+id+".jpg", date)
		if id == "date" {
			asset.FileCreatedAt = "2025-01-01T00:00:00Z"
		}
		assets = append(assets, asset)
	}
	if err := db.upsertAssets(ctx, testUserID, assets); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, start, end string
		ids              []string
		counts           map[string]int
	}{
		{"leap day", "2024-02-29", "2024-02-29", []string{"date", "west", "space", "offsetless", "east"}, map[string]int{"2024-02-29": 5}},
		{"start only", "2024-03-01", "", []string{"dst-winter", "dst-unknown", "dst-summer", "after"}, map[string]int{"2024-03-01": 1, "2024-10-27": 3}},
		{"end only", "", "2024-02-29", []string{"date", "west", "space", "offsetless", "east", "before"}, map[string]int{"2024-02-28": 1, "2024-02-29": 5}},
		{"DST", "2024-10-27", "2024-10-27", []string{"dst-winter", "dst-unknown", "dst-summer"}, map[string]int{"2024-10-27": 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := db.getFilteredAssets(ctx, testUserID, "", "", gpsFilterAll, hiddenFilterVisible, tc.start, tc.end, 1, 20)
			if err != nil {
				t.Fatal(err)
			}
			var ids []string
			for _, asset := range got {
				ids = append(ids, asset.ImmichID)
				if asset.DateTimeOriginal == nil || *asset.DateTimeOriginal != dates[asset.ImmichID] {
					t.Errorf("capture metadata changed: %+v", asset)
				}
			}
			if !slices.Equal(ids, tc.ids) {
				t.Errorf("ordered IDs = %v, want %v", ids, tc.ids)
			}
			counts, err := db.countAssetsByDay(ctx, testUserID, "", "", gpsFilterAll, hiddenFilterVisible, tc.start, tc.end)
			if err != nil || !reflect.DeepEqual(counts, tc.counts) {
				t.Errorf("counts = %v, want %v; error = %v", counts, tc.counts, err)
			}
		})
	}
}

func TestCaptureDatesWithoutCalendarDay(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	assets := []AssetRow{
		datedAsset("empty", "/trip/empty.jpg", ""),
		datedAsset("invalid", "/trip/invalid.jpg", "2024-02-30T12:00:00Z"),
		datedAsset("garbage", "/trip/garbage.jpg", "not-a-date"),
		datedAsset("missing", "/trip/missing.jpg", ""),
	}
	assets[3].DateTimeOriginal = nil
	if err := db.upsertAssets(ctx, testUserID, assets); err != nil {
		t.Fatal(err)
	}
	for _, bounds := range [][2]string{{"", ""}, {"2024-01-01", ""}, {"", "2024-12-31"}, {"2024-01-01", "2024-12-31"}} {
		t.Run(bounds[0]+"/"+bounds[1], func(t *testing.T) {
			want := 0
			if bounds == [2]string{} {
				want = len(assets)
			}
			got, err := db.getFilteredAssets(ctx, testUserID, "", "", gpsFilterAll, hiddenFilterVisible, bounds[0], bounds[1], 1, 20)
			if err != nil || len(got) != want {
				t.Errorf("undated membership = %d, want %d; error = %v", len(got), want, err)
			}
			counts, err := db.countAssetsByDay(ctx, testUserID, "", "", gpsFilterAll, hiddenFilterVisible, bounds[0], bounds[1])
			if err != nil || len(counts) != 0 {
				t.Errorf("undated buckets = %v; error = %v", counts, err)
			}
		})
	}
}

func TestCaptureDatesAcrossCatalogConsumers(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	assets := []AssetRow{
		datedAsset("valid", "/trip/valid.jpg", "2024-02-29T23:30:00-12:00"),
		datedAsset("invalid", "/trip/invalid.jpg", "2024-02-30T12:00:00Z"),
		datedAsset("empty", "/trip/empty.jpg", ""),
	}
	for i := range assets {
		assets[i].Latitude, assets[i].Longitude = ptr(52.0), ptr(21.0)
	}
	if err := db.upsertAssets(ctx, testUserID, assets); err != nil {
		t.Fatal(err)
	}
	if err := db.upsertAlbum(ctx, testUserID, "trip", "Trip", nil, 3, "", nil); err != nil {
		t.Fatal(err)
	}
	if err := db.replaceAlbumAssets(ctx, testUserID, "trip", []string{"valid", "invalid", "empty"}); err != nil {
		t.Fatal(err)
	}
	for _, bounds := range [][2]string{{"2024-02-01", "2024-03-01"}, {"2024-02-01", ""}, {"", "2024-03-01"}} {
		t.Run(bounds[0]+"/"+bounds[1], func(t *testing.T) {
			for _, gps := range []string{gpsFilterAll, gpsFilterWithGPS} {
				var albums []AlbumRow
				var err error
				if gps == gpsFilterWithGPS {
					albums, err = db.getAlbumsWithGPSCount(ctx, testUserID, bounds[0], bounds[1])
				} else {
					albums, err = db.getAlbumsByGPSFilter(ctx, testUserID, gps, bounds[0], bounds[1])
				}
				if err != nil || len(albums) != 1 || albums[0].FilteredCount != 1 || albums[0].NoGPSCount != 0 {
					t.Errorf("albums %s = %+v; error = %v", gps, albums, err)
				}
			}
			for _, album := range []string{"", "trip"} {
				markers, err := db.getMapMarkers(ctx, testUserID, album, "", bounds[0], bounds[1], nil, 10)
				if err != nil || len(markers) != 1 || markers[0].ImmichID != "valid" {
					t.Errorf("markers %q = %+v; error = %v", album, markers, err)
				}
				total, err := db.countMapMarkers(ctx, testUserID, album, "", bounds[0], bounds[1], nil)
				if err != nil || total != 1 {
					t.Errorf("marker total %q = %d; error = %v", album, total, err)
				}
			}
			folder, total, err := db.getFolderAssets(ctx, testUserID, "/trip", gpsFilterAll, hiddenFilterVisible, "", bounds[0], bounds[1], 1, 10)
			if err != nil || total != 1 || len(folder) != 1 || folder[0].ImmichID != "valid" {
				t.Errorf("folder = %+v, total = %d; error = %v", folder, total, err)
			}
			tree, err := db.getFolderTree(ctx, testUserID, gpsFilterAll, hiddenFilterVisible, "", bounds[0], bounds[1])
			if err != nil {
				t.Fatal(err)
			}
			if node := findNode(tree.Children, "/trip"); node == nil || node.AssetCount != 1 {
				t.Errorf("folder tree = %+v", tree)
			}
		})
	}
}
