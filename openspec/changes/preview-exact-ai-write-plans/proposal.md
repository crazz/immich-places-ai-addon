## Why

A saved camera draft is not sufficient authority to change Immich: its revision, target and current GPS may have changed since review. CH18 lets the user inspect an exact fresh before/after GPS plan before CH19 adds a separate confirmation and execution step.

## What Changes

- Produce a private, immutable, expiring GPS preview for one staged CH16 draft revision and its analyzed asset.
- Reread currently authorized source metadata and GPS, preserving absent/partial values and detecting changes against the reviewed baseline.
- Bind exact coordinates, target, fields, source identity, owner, installation, revision and expiry into a server-generated plan and digest.
- Show an accessible numeric before/after comparison and clear conflict, expired and unavailable states, while retaining the editable draft.
- Reject stale or expanded scope and keep previews entirely free of provider calls and Immich mutations.

## Capabilities

### New Capabilities

- `ai-immich-writeback`: Introduce exact, revision-bound, read-only GPS write previews; execution is added separately by CH19.

### Modified Capabilities

None.

## Impact

Prerequisite: apply, verify and synchronize CH16 `persist-revisioned-ai-review-drafts` first. CH18 must ship a useful inspection boundary without a mutation executor. CH19 consumes its stored plans; it cannot be applied before this contract exists.

Affected areas are a focused preview workflow under `backend/internal/ai/`, root read-only Immich/SQLite/HTTP adapters, an additive migration and AI-only preview UI under `src/features/ai/`. Reuse the adopted Go/Next.js/SQLite architecture and current dependencies. No architectural departure or new library is proposed.

Non-goals: confirmation dispatch, Immich writes, descriptions or heading as writable fields, metadata mirroring, stack/batch expansion, GPS clearing, translation, accuracy thresholds and codex-proxy changes. Fresh read access does not claim verified write permission or live mutation compatibility.
