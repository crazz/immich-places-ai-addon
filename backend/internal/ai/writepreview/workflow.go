package writepreview

import (
	"context"
	"time"
)

type DraftReader interface {
	ReadDraft(context.Context, string, string) (Snapshot, error)
}

type MetadataReader interface {
	ReadMetadata(context.Context, Snapshot) (Metadata, error)
}

type PreviewStore interface {
	Publish(context.Context, Snapshot, Preview, []byte) error
}

func Create(ctx context.Context, drafts DraftReader, metadata MetadataReader, store PreviewStore, owner, draftID string, revision int, id string, now func() time.Time) (Preview, error) {
	snapshot, err := drafts.ReadDraft(ctx, owner, draftID)
	if err != nil {
		return Preview{}, err
	}
	if err := Validate(snapshot, revision); err != nil {
		return Preview{}, err
	}
	current, err := metadata.ReadMetadata(ctx, snapshot)
	if err != nil {
		return Preview{}, err
	}
	if err := Compare(snapshot, current); err != nil {
		return Preview{}, err
	}
	preview, raw, err := Build(snapshot, current, id, now())
	if err != nil {
		return Preview{}, err
	}
	if err = store.Publish(ctx, snapshot, preview, raw); err != nil {
		return Preview{}, err
	}
	return preview, nil
}
