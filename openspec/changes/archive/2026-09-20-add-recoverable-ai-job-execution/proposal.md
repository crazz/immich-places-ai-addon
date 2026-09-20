## Why

A validated Visual attempt is transient: browser closure, backend restart or a lost response cannot yet be reconciled against durable work. CH11 establishes a bounded, owner-scoped execution lifecycle so later production analysis can preserve results without duplicate application submissions or late-worker overwrites.

## What Changes

- Persist exact job membership and immutable execution configuration with application idempotency, bounded attempts and separate call accounting.
- Claim work with renewable leases, reject stale workers, recover interrupted work and prevent cancellation from publishing late results.
- Preserve immutable validated analyses and independent terminal item outcomes across restarts, source removal and new runs.
- Reuse the existing installation identity, invalidate unfinished work on controlled installation changes, cascade private data on account deletion and provide bounded explicit terminal-history cleanup.
- Exercise the lifecycle with deterministic synthetic execution before connecting real Visual dispatch.

## Capabilities

### New Capabilities

- `ai-analysis-jobs`: durable scoped submission, execution, recovery, cancellation and immutable history.

### Modified Capabilities

None.

## Impact

Adds the AI jobs core, SQLite integration and an additive migration using the existing deployment, installation identity and canonical validator. It requires CH01, CH05 and CH07. No new dependency, public job API, startup worker, production provider call, frontend workflow, draft or Immich mutation is introduced. CH12 owns consent/access admission and real Visual integration; CH13 owns progress/rerun UI. Automatic retention periods and production benchmarking remain release decisions.

Covers FR-02, FR-05, FR-09 and the local lifecycle portion of FR-14; NFR-01/02/03/07 and AC-07 in the [PRD](../../../../docs/ai-locate/PRD.md). Follows [durable execution](../../../../docs/ai-locate/TECHNICAL_DESIGN.md#8-durable-execution-states-and-budgets) and the adopted engineering standards without an architectural departure.
