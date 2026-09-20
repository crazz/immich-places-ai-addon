package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"immich-places-backend/internal/ai/contextual"
)

func (s *aiContextPreparer) neighbors(ctx context.Context, req aiContextRequest, capture string) ([]contextual.Candidate, error) {
	_, target := contextual.ParseCapture(capture)
	if target.IsZero() {
		return nil, nil
	}
	if req.Window <= 0 || req.Window > 24*time.Hour {
		return nil, contextual.ErrInvalid
	}
	rows, err := s.images.db.db.QueryContext(ctx, `SELECT immichID FROM assets WHERE userID=? AND immichID!=? AND type='IMAGE' AND isHidden=0 AND (stackPrimaryAssetID IS NULL OR stackPrimaryAssetID='' OR stackPrimaryAssetID=immichID) AND latitude IS NOT NULL AND longitude IS NOT NULL AND julianday(dateTimeOriginal) BETWEEN julianday(?) AND julianday(?)`+hiddenLibraryFilter+` ORDER BY ABS(julianday(dateTimeOriginal)-julianday(?)),immichID LIMIT 64`, req.Binding.Owner, req.Binding.Asset, target.Add(-req.Window).Format(time.RFC3339Nano), target.Add(req.Window).Format(time.RFC3339Nano), target.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errAIImageUpstream
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, errAIImageUpstream
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, errAIImageUpstream
	}
	out := []contextual.Candidate{}
	for _, id := range ids {
		lineage := "unknown"
		if s.lineage != nil {
			lineage, err = s.lineage(ctx, req.Binding.Owner, id)
			if err != nil {
				return nil, errAIImageUpstream
			}
		}
		if lineage == "ai" {
			continue
		}
		authority, err := s.images.authorize(ctx, req.Binding.Owner, req.Binding.Installation, id)
		if err != nil {
			continue
		}
		data, available, err := s.read(ctx, authority.key, "/api/assets/"+id, nil)
		if err != nil {
			return nil, err
		}
		if !available {
			continue
		}
		candidate, err := aiContextNeighbor(data, id)
		clear(data)
		if errors.Is(err, errAIImageDenied) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if current, err := s.images.authorize(ctx, req.Binding.Owner, req.Binding.Installation, id); err != nil || current != authority {
			continue
		}
		out = append(out, candidate)
	}
	return out, nil
}

func aiContextNeighbor(data []byte, id string) (contextual.Candidate, error) {
	digest, err := aiImageSourceDigest(data, id)
	if err != nil {
		return contextual.Candidate{}, err
	}
	exif, err := aiContextExif(data)
	if err != nil {
		return contextual.Candidate{}, err
	}
	c := contextual.Candidate{Asset: id, Lineage: "unknown"}
	if exif == nil || exif.Latitude == nil || exif.Longitude == nil || exif.DateTimeOriginal == nil {
		return c, nil
	}
	c.Accessible = true
	c.Latitude = *exif.Latitude
	c.Longitude = *exif.Longitude
	c.CaptureTime = *exif.DateTimeOriginal
	encoded, err := json.Marshal(struct {
		Base                string
		Latitude, Longitude float64
		CaptureTime         string
	}{digest, c.Latitude, c.Longitude, c.CaptureTime})
	if err != nil {
		return contextual.Candidate{}, errAIImageUpstream
	}
	sum := sha256.Sum256(encoded)
	c.SourceDigest = hex.EncodeToString(sum[:])
	return c, nil
}
