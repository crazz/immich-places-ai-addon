package main

import (
	"context"
	"reflect"
	"slices"
	"testing"
)

func TestCaptureDateScopesRemainIsolated(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	check := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	check(db.createUser(ctx, "other-user", "other@example.com", "hashed"))
	check(db.upsertLibrary(ctx, "hidden-library", "Hidden", 1))
	check(db.updateLibraryVisibility(ctx, "hidden-library", true))
	for _, user := range []string{testUserID, "other-user"} {
		var assets []AssetRow
		for _, id := range []string{"located", "missing", "hidden", "stack", "library", "no-album", "no-tag"} {
			asset := datedAsset(id, "/trip/"+id+".jpg", "2024-01-01T00:30:00+14:00")
			if id != "missing" {
				asset.Latitude, asset.Longitude = ptr(52.0), ptr(21.0)
			}
			if id == "stack" {
				asset.StackPrimaryAssetID = ptr("located")
			}
			if id == "library" {
				asset.LibraryID = ptr("hidden-library")
			}
			assets = append(assets, asset)
		}
		check(db.upsertAssets(ctx, user, assets))
		check(db.updateAssetHidden(ctx, user, "hidden", true))
		check(db.upsertAlbum(ctx, user, "trip", "Trip", nil, 6, "", nil))
		check(db.replaceAlbumAssets(ctx, user, "trip", []string{"located", "missing", "hidden", "stack", "library", "no-tag"}))
		check(db.upsertTag(ctx, user, "tag", "Tag", "Tag", nil, nil))
		check(db.replaceTagAssets(ctx, user, "tag", []string{"located", "missing", "hidden", "stack", "library", "no-album"}))
	}
	for _, tc := range []struct {
		name, album, tag, gps, hidden string
		ids                           []string
	}{
		{"all", "", "", gpsFilterAll, hiddenFilterVisible, []string{"no-tag", "no-album", "missing", "located"}},
		{"album", "trip", "", gpsFilterAll, hiddenFilterVisible, []string{"no-tag", "missing", "located"}},
		{"tag", "", "tag", gpsFilterAll, hiddenFilterVisible, []string{"no-album", "missing", "located"}},
		{"both", "trip", "tag", gpsFilterAll, hiddenFilterVisible, []string{"missing", "located"}},
		{"located", "trip", "tag", gpsFilterWithGPS, hiddenFilterVisible, []string{"located"}},
		{"missing", "trip", "tag", gpsFilterNoGPS, hiddenFilterVisible, []string{"missing"}},
		{"hidden", "trip", "tag", gpsFilterAll, hiddenFilterHidden, []string{"hidden"}},
		{"any visibility", "trip", "tag", gpsFilterAll, hiddenFilterAll, []string{"missing", "located", "hidden"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assets, err := db.getFilteredAssets(ctx, testUserID, tc.album, tc.tag, tc.gps, tc.hidden, "2024-01-01", "2024-01-01", 1, 20)
			check(err)
			var ids []string
			for _, asset := range assets {
				ids = append(ids, asset.ImmichID)
			}
			if !slices.Equal(ids, tc.ids) {
				t.Errorf("assets = %v, want %v", ids, tc.ids)
			}
			total, err := db.countFilteredAssets(ctx, testUserID, tc.album, tc.tag, tc.gps, tc.hidden, "2024-01-01", "2024-01-01")
			check(err)
			counts, err := db.countAssetsByDay(ctx, testUserID, tc.album, tc.tag, tc.gps, tc.hidden, "2024-01-01", "2024-01-01")
			check(err)
			if total != len(tc.ids) || !reflect.DeepEqual(counts, map[string]int{"2024-01-01": len(tc.ids)}) {
				t.Errorf("total = %d, buckets = %v; want %d", total, counts, len(tc.ids))
			}
		})
	}
	markers, err := db.getMapMarkers(ctx, testUserID, "trip", "tag", "2024-01-01", "2024-01-01", &TViewportBounds{North: 53, South: 51, East: 22, West: 20}, 10)
	check(err)
	if len(markers) != 2 || markers[0].ImmichID != "located" || markers[1].ImmichID != "hidden" {
		t.Errorf("markers must retain per-asset visibility policy: %+v", markers)
	}
	albums, err := db.getAlbumsByGPSFilter(ctx, testUserID, gpsFilterAll, "2024-01-01", "2024-01-01")
	check(err)
	if len(albums) != 1 || albums[0].FilteredCount != 4 || albums[0].NoGPSCount != 1 {
		t.Errorf("album scope = %+v", albums)
	}
	albums, err = db.getAlbumsWithGPSCount(ctx, testUserID, "2024-01-01", "2024-01-01")
	check(err)
	if len(albums) != 1 || albums[0].FilteredCount != 3 || albums[0].NoGPSCount != 1 {
		t.Errorf("located album scope = %+v", albums)
	}
	folder, total, err := db.getFolderAssets(ctx, testUserID, "/trip", gpsFilterAll, hiddenFilterVisible, "tag", "2024-01-01", "2024-01-01", 1, 2)
	check(err)
	if total != 3 || len(folder) != 2 || folder[0].ImmichID != "no-album" || folder[1].ImmichID != "missing" {
		t.Errorf("folder scope/pagination = %+v; total = %d", folder, total)
	}
	tree, err := db.getFolderTree(ctx, testUserID, gpsFilterAll, hiddenFilterVisible, "tag", "2024-01-01", "2024-01-01")
	check(err)
	if node := findNode(tree.Children, "/trip"); node == nil || node.AssetCount != 3 {
		t.Errorf("folder tree scope = %+v", tree)
	}
}
