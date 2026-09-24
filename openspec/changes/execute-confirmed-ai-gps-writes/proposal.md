## Why

After CH18 the user can inspect an exact GPS plan but cannot apply it. CH19 completes the first single-photo save with confirmation, durable audit, conflict detection and verified recovery together, so a lost response or restart cannot turn into an unapproved or blind repeated write.

## What Changes

- Confirm a current unexpired CH18 preview using its stored digest and a private idempotency key, recording the exact approved plan durably before any dispatch.
- Write only the approved camera latitude/longitude to exactly one authorized photo, preserving every other field and avoiding implicit stack expansion.
- Recheck source and baseline, serialize overlapping AI writes, disable hidden mutation retries and verify upstream GPS before reporting success or refreshing the local catalog.
- Recover interrupted/ambiguous operations through readback before any possible bounded explicit retry; retain visible unresolved, conflict and failed outcomes with their audit.
- Show confirmation, progress and verified outcome in AI Results, keeping original analyses and draft revisions independent of write status.
- Coordinate active writes with draft edits, account/installation changes, feature disablement and cleanup; introduce no new provider or precision-readiness requirements.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-immich-writeback`: Add durable confirmation and exact single-asset GPS execution/reconciliation to the capability introduced by prerequisite CH18. Its preview requirements remain unchanged.
- `ai-results-and-review`: Add active-write revision protection and verified write status to retained draft/result review.

## Impact

Prerequisites: CH16 and CH18 applied, verified and synchronized, plus implemented CH11 durable job/recovery patterns. The CH18 capability is planned at proposal time; CH19 adds distinct requirements and must not be applied against a missing prerequisite. No codex-proxy work is required.

Affected areas are focused `backend/internal/ai/writeback` workflows, exact Immich mutation/read adapters, owner-qualified SQLite operations/audit and migration, startup/shutdown wiring, draft revision guards, local catalog refresh and `src/features/ai/` confirmation/status. Reuse existing Go/Next.js/SQLite dependencies and deployed services. No architectural departure is proposed.

Non-goals: bulk/stack scope (CH21), description writes (CH20), direction or extended metadata (CH22), GPS clearing/rollback, automatic approval, guaranteed atomicity against external clients, provider repairs and new retention policy. Synthetic verification proves the protocol; live route/readback compatibility remains GATE-02 and requires an explicitly authorized disposable fixture before enabling real writes. The supplied analysis photographs are not authorized mutation fixtures.
