## Why

Reviewed direction, translations and provenance remain useful locally even when Immich cannot display them as standard fields. CH22 adds an optional, explicitly reviewed metadata mirror for supported installations without making standard-field success depend on the mirror.

## What Changes

- Offer a default-off mirror choice with a disclosure that authorized asset readers may see the exported information.
- Preview a compact selected record of reviewed direction, place precision, translations and minimal provenance for the analyzed photo.
- Preserve local authoritative history and unrelated upstream metadata, rejecting unsupported capability or changed mirror baselines.
- Record optional metadata as a separate verified step, keeping successful standard-field writes intact if mirroring fails.
- Recover or explicitly retry only the incomplete mirror step; never repeat completed standard writes or propagate the mirror to unreviewed stack members.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-results-and-review`: Add selected mirror contents, visibility disclosure and independent partial-write status.
- `ai-immich-writeback`: Add capability-gated, exact optional metadata steps to durable confirmed operations.

## Impact

Prerequisite: CH20; apply after CH21 in this batch so the combined plan preserves exact stack GPS scope while metadata remains per-image. Traceability: FR-07–08, FR-11–14 and AC-08/AC-09/AC-12.

Affected areas are draft review, capability-aware previews, confirmed step execution, the versioned Immich metadata adapter, audit persistence and AI status UI. The existing Go/Next.js/SQLite deployment remains sufficient; no new library or architectural departure is proposed. Route, permissions, namespace preservation and readback require separate authorized live evidence.

Non-goals: local-data export/deletion controls, deleting a previously written mirror, mandatory mirroring, raw prompt/context export, secrets, neighboring-photo coordinates, EXIF direction editing, sidecar completion guarantees, codex-proxy changes and automatic writeback. Unsupported installations retain the full local record and supported standard-field workflow.
