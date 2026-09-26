package main

import (
	"context"
	"time"

	"immich-places-backend/internal/ai/writepreview"
)

func (s *aiDraftStore) mirrorSource(ctx context.Context, owner, asset string, reader *aiImagePreparer) (writepreview.MirrorBaseline, aiImageAuthority, error) {
	fail := func() (writepreview.MirrorBaseline, aiImageAuthority, error) {
		return writepreview.MirrorBaseline{}, aiImageAuthority{}, writepreview.Failure{Code: "METADATA_UNAVAILABLE"}
	}
	if reader == nil {
		return fail()
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	authority, err := s.results.imageAuthority(ctx, owner, asset)
	if err != nil {
		return fail()
	}
	raw, _, err := reader.fetch(ctx, authority.key, "/api/assets/"+asset+"/metadata", writepreview.MirrorListLimit)
	if err != nil {
		return fail()
	}
	defer clear(raw)
	baseline, err := writepreview.ParseMirrorNamespace(raw)
	if err != nil {
		return fail()
	}
	if current, err := s.results.imageAuthority(ctx, owner, asset); err != nil || current != authority {
		return fail()
	}
	return baseline, authority, nil
}
