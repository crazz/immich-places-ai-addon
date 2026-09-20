## Context

See [proposal.md](proposal.md), [CH12's admission metadata](../archive/2026-09-20-connect-visual-analysis-to-durable-jobs/design.md), the [PRD history contract](../../../docs/ai-locate/PRD.md#fr-09--keep-reviewable-result-history), and adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md) and [coding standards](../../../docs/engineering/coding-standards.md). At `23c6ca6`, GitNexus finds `ReadAnalysis` with incomplete receiver callers; targeted source/test search confirms its private job-store reads. Migration 022 already separates immutable analyses from catalog rows, but exposes neither history queries nor public detail. The current read gate requires AI enabled. The existing thumbnail path checks visibility and uses the owner's upstream client. The shell's catalog readiness cannot be assumed for independent history.

## Goals / Non-Goals

**Goals:** Durable private terminal history, bounded filters/pagination, reloadable immutable detail and truthful independent states, even when GPS/catalog membership changes or AI execution is off.

**Non-Goals:** No draft storage, approval/rejection, write state machine, deletion/export, automatic retention, provider calls or new library. Detailed geometry/evidence/language views arrive in CH15.

## Decisions

### Read ownership and API

Create focused query/DTO rules in `internal/ai/review` with consumer-owned read interfaces. Root `aiResult*.go` adapters query the existing job/item/analysis tables; no parallel result store or monolithic AI store. Separate history-read authority from execution enablement while retaining authenticated local owner and current installation binding. Preserve old-installation records internally but do not expose them as current-installation results. Public reads cannot activate workers or modify job state.

Expose `GET /ai/results` for terminal succeeded, failed and canceled items. Active/blocked work stays in job progress. A history entry uses the stable item ID plus job ID; successful entries additionally contain an analysis ID. `GET /ai/results/{analysisID}` returns successful immutable detail. `GET /ai/jobs/{jobID}/items/{itemID}/result` resolves either successful detail or a safe failed/canceled detail, so technical failures never need fabricated canonical proposals. Every lookup/join includes owner and installation. Use session protection, no-store responses, finite deadlines and indistinguishable missing/foreign responses.

List summaries contain IDs, timestamps, mode/model, a bounded primary-language place/scene label when present, execution state, nullable proposal outcome, review/write projections, safe failure, available thumbnail reference and a current-source availability flag. Do not include full proposals, context bundles or raw errors in list responses. Detail keeps the validated canonical document separate from server provenance; use stored schema/validation version and reject corrupt or unsupported records safely without invoking a provider or rewriting them.

### Retained filters and pagination

Use descending terminal time plus job/item ID as a deterministic order. Default page size 30, maximum 100, with opaque versioned owner/installation/filter-bound keyset cursors. Add a monotonic terminal-history sequence in the read projection, assigned transactionally with terminal publication, and capture its maximum in the first-page cursor. This watermark excludes later completions even when their timestamps tie or the clock moves backward. Invalid or mismatched cursors fail explicitly. Fetch limit plus one rather than counting or loading all analyses; no full payload JSON scan for pagination. New terminal entries appear on refresh, not in the middle of a continued page sequence.

Filter by terminal execution state, proposal outcome, asset/run, recorded capture-date bounds and selected album at launch. The album control is explicitly labeled **Selected album at launch**: it uses CH12's frozen selected-album provenance, not all current album memberships. Timeline runs without that provenance remain visible without the filter. This bounded, stable interpretation preserves reproducibility after source removal and does not invent historical album membership. Date filters use CH04's source-local recorded day from admission metadata, no upload-time fallback. Undated items remain in unbounded history and an explicit undated filter, and are excluded from bounded date ranges; reject reversed ranges.

Add query columns/indexes with an ordered migration where needed. Derive immutable filter projections from persisted admission metadata at item completion, including failed/canceled items. Backfill only from retained authoritative job metadata; leave older missing provenance explicitly unknown instead of reading today's catalog and calling it historical. Verify the actual query plans on representative synthetic data. Current GPS and the Missing GPS filter never constrain history membership.

### Independent state and current image access

Keep `executionState` separate from `proposalOutcome`: `succeeded` may mean `located`, `ambiguous` or `unknown`; failed/canceled has no proposal outcome. CH14 has no draft/write records, so expose `reviewState=unreviewed` and `writeState=not_requested`, not approved/saved guesses. Future owners can extend these projections from actual durable records. A new run appears as a separate row grouped by asset only in presentation; it never replaces a prior analysis or changes its review/write state. Capture original run/model/mode provenance even if profiles are later edited or removed.

History content belongs to the local owner and survives source removal/access loss. New thumbnail/preview retrieval still requires current owner/installation and current local/upstream asset access. Reuse the authorized image proxy only after confirming its current policy is sufficient; add a narrow read-only adapter if required, not a mutation client. Suppress thumbnails for hidden, removed or inaccessible sources with an unavailable placeholder. Never use a stored provider URL or cached private image from another account. If current asset metadata exists, label it current rather than silently replacing the retained analysis provenance.

### UI integration and failure handling

Add an AI Results navigation surface and list/detail state under `src/features/ai/`, using a small shell integration that requires an authenticated local user and reachable backend but does not wait for successful catalog sync. Preserve catalog filter/selection state when switching views. Store only navigation IDs/filter values in URL state, not private context, descriptions or results. Clear private requests/cache on logout/account change, abort superseded loads and ignore late responses. Use validated DTOs and escaped plain text for all model/provider content.

Provide loading, empty, filtered-empty, stale/offline, safe-error and source-unavailable states with retry and keyboard-operable pagination. Returning from detail restores list filters and position. No polling triggers reanalysis; explicit refresh is enough for this terminal list. Render basic immutable detail and safe failure information now; CH15 consumes the same detail without redefining it. No “accept” or “save” action appears before its owning change.

### Verification, migration and rollout

Use real SQLite tests for owner/installation isolation, stable pagination, source deletion/GPS changes, old/new runs, failed history, provenance backfill and disabled-execution reads. API fixtures cover bad filters/cursors, missing/foreign IDs, corrupt records and current thumbnail authority. RTL/browser journeys cover history independent of Missing GPS, reload/back navigation, zero provider calls/writes, private state reset and manual/GPX regression. Run shared apply gates and Docker packaging for new core code; do not claim measured NAS performance from fixtures.

Apply/sync CH12 before CH14; CH13 is optional for Visual browsing and only supplies additional mode provenance. CH14 creates the main `ai-results-and-review` capability; CH15 must apply/sync afterward. Rollback hides the new surface and disables execution before restoring a compatible backup when necessary. Do not destructively downgrade retained history or silently adopt proposed retention defaults.

## Risks / Trade-offs

- Older results may lack date/album provenance; explicit unknown data is preferable to false historical precision.
- Album-at-launch differs from current album membership; the control label and empty state must make that distinction clear.
- Retained local results can outlive upstream access; current images stay separately authorized and local deletion remains owned by CH23/CH24.
- Detail payloads may be large or corrupt; enforce existing bounds and safe record-unavailable behavior without crashing the list.
