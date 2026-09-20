package main

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"immich-places-backend/internal/ai/contextual"
)

func (s *aiContextPreparer) album(ctx context.Context, req aiContextRequest, key string) (*contextual.Album, error) {
	id := req.Consent.AlbumID
	parsed, err := uuid.Parse(id)
	if err != nil || parsed.String() != id {
		return nil, contextual.ErrInvalid
	}
	var owned, member bool
	err = s.images.db.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM albums WHERE userID=? AND immichID=?), EXISTS(SELECT 1 FROM albumAssets WHERE userID=? AND albumID=? AND assetID=?)`, req.Binding.Owner, id, req.Binding.Owner, id, req.Binding.Asset).Scan(&owned, &member)
	if err != nil {
		return nil, errAIImageUpstream
	}
	if !owned {
		return nil, errAIImageDenied
	}
	if !member {
		return nil, nil
	}
	search, _ := json.Marshal(struct {
		ID       string   `json:"id"`
		AlbumIDs []string `json:"albumIds"`
		Size     int      `json:"size"`
		WithExif bool     `json:"withExif"`
	}{req.Binding.Asset, []string{id}, 1, false})
	data, available, err := s.read(ctx, key, "/api/search/metadata", search)
	if err != nil || !available {
		return nil, err
	}
	defer clear(data)
	member, err = aiContextMembership(data, req.Binding.Asset)
	if err != nil || !member {
		return nil, err
	}
	data, available, err = s.read(ctx, key, "/api/albums/"+id, nil)
	if err != nil || !available {
		return nil, err
	}
	defer clear(data)
	var album struct {
		ID    string  `json:"id"`
		Label *string `json:"albumName"`
	}
	_, objectErr := aiImageUniqueObject(data)
	if objectErr != nil || json.Unmarshal(data, &album) != nil || album.ID != id || album.Label == nil {
		return nil, errAIImageUpstream
	}
	return &contextual.Album{ID: id, Label: *album.Label, Member: true}, nil
}

func aiContextMembership(data []byte, asset string) (bool, error) {
	root, err := aiImageUniqueObject(data)
	if err != nil {
		return false, errAIImageUpstream
	}
	assets, err := aiImageUniqueObject(root["assets"])
	if err != nil {
		return false, errAIImageUpstream
	}
	var items []json.RawMessage
	if json.Unmarshal(assets["items"], &items) != nil || items == nil || len(items) > 1 {
		return false, errAIImageUpstream
	}
	if len(items) == 0 {
		return false, nil
	}
	item, err := aiImageUniqueObject(items[0])
	if err != nil {
		return false, errAIImageUpstream
	}
	var id string
	if json.Unmarshal(item["id"], &id) != nil || id != asset {
		return false, errAIImageUpstream
	}
	return true, nil
}
