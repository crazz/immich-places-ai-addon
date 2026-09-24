## Why

AI Results currently supports inspection only: a useful camera estimate cannot become a durable, editable decision. CH16 preserves the user's review as a local revisioned draft so subsequent exact write previews can use deliberate choices without changing Immich during acceptance or editing.

## What Changes

- Add explicit accept, edit, stage and reject actions with durable owner-scoped review state, leaving the original analysis immutable.
- Support candidate selection and camera coordinates through numeric inputs and an isolated editable map, including coarse estimates and unknown uncertainty without a quality threshold.
- Retain exact analyzed-asset targets, GPS field selection, source identity, baseline availability and review provenance; preserve local editing when upstream metadata is unavailable.
- Detect concurrent edits and mark camera-dependent direction, uncertainty and descriptions for review after relevant changes. Local description/direction edits grant no write authority.
- Keep drafts out of manual/GPX pending coordinates and the legacy save path; preserve them across reloads, reanalysis and AI execution disablement.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-results-and-review`: Add explicit durable draft actions and revision-aware editing alongside the existing side-effect-free inspection contract.

## Impact

Prerequisite: implemented CH15 proposal inspection and CH14 private history. CH18 consumes staged drafts; CH19 alone introduces confirmed GPS writes. Research proxy repairs and successful new provider calls are not prerequisites: any retained valid Visual, Context-assisted or Research result can be reviewed.

Affected areas are focused `backend/internal/ai/` review/draft workflows, owner-qualified SQLite adapters and Goose migrations, protected AI routes, and `src/features/ai/` review state and components. Existing Go, Next.js, SQLite and map dependencies are sufficient. No architectural departure or new library is proposed.

Non-goals: Immich mutations, automatic acceptance, bulk or stack targets, translation generation (CH17), description writeback (CH20), extended metadata, retention-policy changes and any codex-proxy change. A staged draft is a local decision, not approval to write. Existing manual and GPX behavior stays on its existing path.
