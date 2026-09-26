## Context

See [proposal](proposal.md). Apply after CH20 and CH21 in this batch. CH19 currently has no metadata reader or mutation step; it only verifies GPS. The pinned [v3.2.2 controller](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/controllers/asset.controller.ts), [DTO](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/dtos/asset.dto.ts) and [service](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/services/asset.service.ts) expose GET metadata and `PUT /api/assets/{id}/metadata` with an `items` array of `{key, value}` objects, checking asset-update authority. That path is independent of standard updates. Missing-key lookup is a bad-request response in the pinned service, so a generic 400 must not be interpreted as an empty baseline.

GitNexus and source grounding for the writer are recorded in the batch plan and CH20/CH21 designs. No existing general metadata writer is assumed. Follow adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md), [coding standards](../../../docs/engineering/coding-standards.md), ADR-07, REC-01/REC-05 and technical design §11.4–11.5.

## Goals / Non-Goals

**Goals:** add one optional, independently recoverable namespace step for the analyzed photo while retaining local authoritative data and successful standard steps.

**Non-Goals:** arbitrary metadata editing, deleting upstream metadata, mirroring private context or raw model output, editing EXIF heading, modifying originals or proving sidecar completion. No new service, library or generic orchestration engine.

## Decisions

### D1. Bind capability and selected export content

Introduce explicit operator capability enablement for the configured adapter/installation separately from GPS and description support. Mirror is default-off and remains unavailable when unsupported, unverified or permission checks fail. Read-only checks may establish version and read access; only separately authorized evidence establishes the live mutation profile. Saving a profile or opening review never sends a test write.

A draft mirror selection identifies the analyzed photo and individually included current fields: reviewed direction/method/uncertainty, reviewed precision, chosen place text, selected current language texts and minimal provenance. The preview displays the exact export and asset-reader visibility disclosure. Omit unavailable/stale fields unless independently reviewed; unknown direction stays null/omitted and is never replaced by subject bearing. A correction invalidates affected exports and the prior preview. No mirror is copied to a stack sibling.

The namespace value uses schema version 1, application identity, a stable random per-asset mirror record ID, reviewed revision references without local account IDs, selected reviewed content and bounded provenance (analysis mode, model name if selected, review time, user/model-origin flags). Exclude credentials, prompts, hints, neighboring coordinates, raw provider response and arbitrary source URLs. At most eight chosen languages and 64 KiB of canonical UTF-8 JSON; reject overflow rather than silently dropping approved content. Output validation and canonicalization live in focused `internal/ai/writepreview`/`writeback` policy; adapters alone own I/O.

### D2. Freeze a namespace-specific baseline and plan

Read `GET /api/assets/{id}/metadata` through the bounded authorized reader and extract exactly `immich-places-ai-addon`. A successfully decoded complete list without that key means absent; transport failure, invalid/duplicate keys, malformed values or a truncated/oversized response means unavailable. Preserve all unrelated keys and compare only this namespace semantically: object key order is irrelevant, array order and string contents are not. Reject duplicate JSON keys and non-finite/unsupported values.

Add v4 plans with an optional metadata step after the analyzed photo's standard step. Require at least one standard field selected for a new combined confirmation; unchanged standard fields may verify as no-op before the mirror. Mirror-only repair is available only as a retry of an already-approved incomplete metadata step, not a new general export workflow. This keeps CH22 bounded to the roadmap's optional second step.

The digest includes exact export bytes, namespace key, absent/present before-value, source identity, selected fields, capability profile and disclosure acknowledgement. An existing recognized mirror must match its last owned, verified value for replacement. A foreign, incompatible or user-edited namespace becomes a visible conflict; do not adopt it merely because the key matches. Deselecting mirror changes the draft/plan and requires a fresh preview; the server never silently downgrades an approved combined plan.

### D3. Add step state under the shared confirmed writer

Extend CH21 target state with independently durable standard and metadata step records. The analyzed target retains its exclusion until both steps are settled and no sender may still act. Only after its standard step is verified (including verified no-op) can its metadata step reserve. Failure on another stack member does not undo this target's standard success, but an unresolved or failed standard step on the analyzed target prevents its mirror.

Before the metadata attempt, recheck current owner/installation, credential/profile, enablement, draft revision, source identity and namespace baseline. Standard-field bookkeeping changes do not change image identity. Send exactly one item under the approved key using a dedicated non-retrying PUT in `internal/aiadapters/immichwrite`, with the existing 20-second attempt deadline, bounded response and no redirects/fallback. No other namespace or standard field appears in that request.

Read the namespace back and compare semantic content and unchanged source. A timeout remains unresolved until readback; desired content establishes observed state, not sender causality. Only an incomplete metadata step with known prior-sender completion, unchanged baseline, valid unexpired approval and remaining two-attempt budget can be explicitly retried. Retry identifiers are per target/step/generation; accepted-generation replay is local and cannot allocate another attempt. Completed standard steps are never dispatched during mirror repair. After approval expiry, retain the incomplete audit and require a new reviewed combined plan if the user wants another mirror attempt.

### D4. Preserve partial success and local authority

Report standard and metadata state separately, including unsupported, not selected, blocked, unresolved, failed and verified. The operation is fully successful only when every selected step is verified; a failed optional step produces a partial result without rolling back verified GPS/description or dropping local translations. Each step retains exact private before/intended/observed values, attempt/generation and timestamps. Local storage failure after upstream success invokes read-only reconciliation, never a repair write. A successful metadata HTTP response alone is insufficient.

The UI shows compact selected contents with an expandable exact comparison, keyboard-operable consent and an explicit retry action for the incomplete step. No map is required to understand it. Existing identity recovery and account/result response fencing remain. Unsupported metadata leaves standard-only planning and all local records usable.

### D5. Migration, rollout and verification

Allocate an additive migration after CH21 for step records and mirror ownership lineage. Old v1/v2/v3 approval bytes remain unchanged and decode without a metadata choice; no migration defaults a mirror on. Preserve active legacy guards and history, composite ownership, ordinary-cleanup protection, cancellation/deletion fences and installation invalidation. Audit readers project old operations as a single standard step without fabricating metadata attempts.

The [verification plan](verification-plan.md) covers payload minimization, namespace collision/conflict, semantic readback, partial-success recovery, restart and deletion using pure tests, real SQLite and local HTTP, plus accessible component and built-browser journeys. Apply follows Plus TDD, pre-edit GitNexus impact and all thirteen installed gates with the actual base and coverage floors. Live GATE-02 must separately verify the real metadata route, permissions, preservation and readback; no capability claim follows solely from source inspection or GPS success.

Back up before migration, preserve encryption material and initially keep metadata dispatch disabled. Rollback stops new sends, retains unresolved steps and requires a compatible binary; it does not remove mirrors, clear GPS or restore old database snapshots over newer audit.

## Risks / Trade-offs

- Asset readers may see exported information → default-off selection, explicit content preview and data minimization.
- The namespace may contain another installation's or user's edits → fail closed and preserve it; no automatic merge or takeover.
- Standard and metadata calls are not atomic → separate step states and no repeat of successful standard fields during recovery.
- Valid custom JSON may not appear in Immich UI or files → describe it as an optional API mirror; local direction/translations remain authoritative.
