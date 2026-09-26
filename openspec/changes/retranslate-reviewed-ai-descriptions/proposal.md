## Why

Correcting a location can leave its descriptions stale, and a failed translation currently requires more work than the language correction itself. CH17 lets the user regenerate selected languages from reviewed facts while preserving the camera decision and original analysis.

## What Changes

- Add an explicit, bounded translation action for selected languages of a reviewed draft, without another image analysis or geolocation run.
- Show the exact factual basis and language selection before sending text to the chosen private provider.
- Retain independent language outcomes and earlier text; allow failed languages to be retried without regenerating successful languages.
- Detect concurrent edits and changed factual revisions so delayed translations cannot overwrite newer decisions or silently become current.
- Keep generation, review and write approval separate, with accessible progress, cancellation and durable recovery.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-results-and-review`: Add private, revision-bound description regeneration and explicit adoption of generated text into drafts.

## Impact

Prerequisites: implemented CH09 bounded provider integration, CH11–CH12 durable admission patterns and CH16 revisioned drafts. This change can proceed while CH19 live GPS compatibility remains pending. It completes the regeneration portion of FR-08 and AC-08/AC-12; it does not introduce Immich write authority.

Affected areas are AI draft/review and provider workflows, private operation persistence, protected APIs and the AI language editor. Reuse the existing Go/Next.js/SQLite deployment and dependencies; no architectural departure is proposed.

Non-goals: new geolocation, image or neighboring-photo transmission, web research, automatic adoption, description writeback, stack scope, metadata mirroring, codex-proxy changes and new retention policy. Local edits and translated results alone never authorize an Immich mutation.
