## Why

AI analyses must remain discoverable after the source photo gains GPS, leaves the current catalog view or receives another analysis. CH14 provides private durable history with truthful outcome, review and write status, instead of relying on Missing GPS or transient progress state.

## What Changes

- Add an AI Results view and protected paginated history/detail APIs for located, ambiguous, unknown and failed terminal items.
- Filter retained history by capture date, selected album provenance and AI execution/outcome state, independently of current Missing GPS membership.
- Preserve separate runs, immutable proposals and safe failures, with current-access thumbnails or explicit unavailable placeholders.
- Show analysis outcome separately from local review and write state; this change does not invent draft approval or successful writes.
- Keep stored private history readable while AI execution is disabled, without contacting a provider.

## Capabilities

### New Capabilities

- `ai-results-and-review`: private persistent result browsing, immutable detail and independent state presentation.

### Modified Capabilities

None.

## Impact

Depends on the planned [CH12](../archive/2026-09-20-connect-visual-analysis-to-durable-jobs/proposal.md); it can browse Visual history without CH13, while remaining compatible with CH13's Context provenance. Adds focused read policy/DTOs under `backend/internal/ai/review/`, root persistence/HTTP adapters, additive query metadata/indexes and UI under `src/features/ai/`. No new dependency or deployment service is proposed.

CH15 owns map/evidence/language presentation; CH16 owns drafts and CH19 onward owns writes. No editing, acceptance, result deletion/export or automatic retention is included. Covers FR-01/09/13 and NFR-01/03/06 in the [PRD](../../../docs/ai-locate/PRD.md) and REC-05 in the [reconciliation](../../../docs/ai-locate/RECONCILIATION.md).
