## Context

See [proposal.md](proposal.md) for motivation. CH14 stores immutable validated analyses and terminal history; CH15 only changes inspection focus. `review.Entry` currently reports `unreviewed` and `not_requested`. There is no durable draft or writer. The current migration tip is 025.

Design grounding at `0fb8eaf`: GitNexus queries covered private history, image authority, proposal maps and manual saving. The graph resolved `saveAssetLocationsWithRetry` to its caller in `useLocationAssignment`; some Go constructor results were unresolved and were verified in source. Relevant sources are `backend/aiResultHTTP.go`, `aiResultDetail.go`, `aiResultImage.go`, `aiJobCleanup.go`, `aiJobInstallation.go`, `internal/ai/review/`, and `src/features/ai/ProposalReview.tsx` and `ReviewMap.tsx`. Graph results are discovery evidence, not proof that all callers were enumerated. This planning change edits no source symbols; apply must refresh impact analysis before each existing-symbol edit.

The design follows [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md), [coding standards](../../../docs/engineering/coding-standards.md) and ADR-07. No departure is proposed.

## Goals / Non-Goals

**Goals:** Separate immutable analysis, durable user decisions, transient editor state and future write authority. Provide a stable revision contract that CH18 can consume without introducing a writer into review.

**Non-Goals:** No general workflow engine, provider call, new package dependency or new global pending-coordinate type. Descriptions and direction can be corrected locally, but only GPS is eligible for the later write plan in this batch. No additional precision, confidence or translation-readiness gate for GPS.

## Decisions

### 1. A local aggregate with one draft per retained analysis

Add a focused `internal/ai/drafts` package owning validation and review transitions. It consumes small repository and read-only metadata interfaces, with explicit time/ID inputs. Root `aiDraft*.go` files own SQL, authenticated HTTP adaptation and composition; they do not move SQL into the core. Extend the existing history projection through owner-qualified joins rather than modifying the immutable result payload.

Use `ai_drafts` for the current aggregate and `ai_draft_revisions` for immutable bounded revision snapshots. Unique `(userID, installationID, analysisID)` prevents repeated acceptance from creating duplicate drafts. Each revision records source analysis, exact target `[analyzedAssetID]`, selected fields `[]` or `[gps]`, chosen candidate or explicit user point, camera coordinates, local direction/descriptions with factual basis and stale flags, baseline state, review state and timestamps. The source proposal remains in `ai_analyses`; do not copy its evidence tree into every revision. Revision content is bounded to 128 KiB, individual description text to 16 KiB and language keys to the original requested set. Reuse canonical coordinate/language constraints at the boundary.

Alternative: overwrite the analysis or place data in browser storage/manual pending state. Rejected because it loses lineage, restart durability and server revision authority. Historical revisions are kept for review and future audit; this change adds no time-based pruning policy.

### 2. Explicit actions and optimistic concurrency

Proposed protected routes follow existing `/ai/` conventions:

| Route | Contract |
|---|---|
| `POST /ai/results/{analysisID}/draft` | Explicit acceptance; create revision 1 or return the existing draft unchanged. Include an explicit candidate choice for ambiguous proposals. Unknown/subject-only results can create an incomplete draft without invented camera coordinates. |
| `GET /ai/drafts/{id}` | Current local draft and revision token, scoped to owner and installation. |
| `PATCH /ai/drafts/{id}` | Replace allowed editable fields and/or review state using a required `If-Match` revision token. |
| `POST /ai/drafts/{id}/baseline` | Explicit fresh source/baseline review; prepare a bounded observation or acknowledge an exact observation with the expected draft revision. Never mutate Immich. |

All mutation routes use session and existing origin protection, strict bounded JSON with unknown fields rejected, no-store responses and safe error codes. Identity, target and source-analysis fields are server-owned. Missing revision preconditions return 428; a stale token returns 412 with a reload/review path, never an automatic merge. Foreign/unavailable IDs use indistinguishable 404 responses. Validation uses 400 and temporary storage failure 503. Duplicate create requests return the same ID/revision even after a lost response; later edits are not overwritten by reacceptance. A lost PATCH response is reconciled with GET before replay.

State transitions: explicit accept creates `draft`; `draft` can become `staged` or `rejected`; reopening rejected content produces a new `draft` revision. Any accepted content change to a staged draft returns it to `draft`. Staging requires a finite camera pair and explicit GPS selection, but no fresh network call, nonzero accuracy, heading or description. Rejection preserves the decision and does not delete analysis. Every persisted change, including baseline acknowledgement, increments revision; no-op reacceptance does not. Future previews bind to the revision, so an edit/rejection invalidates them without review importing the writer.

### 3. Baselines distinguish unavailable, absent GPS and changed images

Keep local draft creation/editing available when the source is inaccessible or Immich is offline. New drafts record `baseline=unavailable`; do not delay local acceptance on network success or substitute cached coordinates for a fresh baseline. An explicit source review can obtain current metadata and a currently authorized preview through the existing bounded read transport. A successful observation records nullable latitude and longitude independently, source asset/owner/checksum/type, observation time and an opaque server observation token. Missing/partial GPS and numeric zero are distinct. Do not persist the API key, image bytes or unrelated EXIF/description fields.

`aiImageSourceDigest` currently hashes `updatedAt` and stack facts along with checksum. Preserve the original analysis digest unchanged, but define a separately versioned reviewed-image identity from asset ID, upstream owner, type and checksum for drafts and later writeback. Do not change CH08/CH09 fingerprint semantics. When the original digest no longer matches, show that the source needs renewed review; explicit review of the current authorized image and chosen point can establish a new baseline without rerunning AI. This does not claim that old analysis saw the current image. Store original and reviewed provenance separately.

Persist baseline observations in `ai_draft_baseline_observations`, keyed by an opaque server ID and owner/installation/draft/revision. They expire after five minutes; fresh GPS and image identity must still match before acknowledgement. A stale observation produces no revision. Keep at most ten unexpired observations per draft and remove at most 100 expired observations per cleanup pass, without pruning draft revisions. Acknowledgement is an explicit current-image review and acceptance of displayed GPS before-values. It is also the resolution path when CH18 finds changed baseline GPS. An unavailable observation never clears an existing baseline. CH18 always rereads current metadata regardless of stored baseline age. Historical source access does not authorize a fresh read; use current local visibility plus upstream asset access and recheck credentials/installation before publishing observations.

Alternative: require online metadata for all draft edits. Rejected because it couples private retained review to upstream availability. Alternative: silently refresh baseline in CH18. Rejected because it can conceal intervening GPS changes.

### 4. Invalidation is field-specific, with no extra GPS quality gate

Candidate changes or camera movement mark inherited heading, location-dependent descriptions and inherited radius/basis as stale. Keep old values visibly labeled for comparison; do not transfer their support to the edited point. Scene-only text remains usable if its facts are unchanged. Each requested language keeps its own complete/unavailable status and factual revision. Editing or explicitly reviewing a dependent value binds it to the current camera-facts revision; one language's review does not approve others. Numeric heading accepts `[0,360)` or null and never derives from subject bearing. User-edited values are labeled user supplied, without promoting model evidence to verified evidence.

GPS staging remains possible with stale/unavailable descriptions or heading because those fields are not selected for writing. A modified point may carry unknown uncertainty; no maximum radius or confidence threshold is introduced. Subject points stay contextual, never automatic camera replacements.

### 5. Separate editor and inspection lifecycles

Add draft DTO/API helpers, a draft state hook, editor components and an explicitly editable map under `src/features/ai/`, split by responsibility. Reuse map construction/geometry utilities without installing the legacy manual assignment handlers. Keep CH15 inspection read-only. Numeric camera/heading inputs, errors and review actions work with keyboard input, narrow viewports and failed tiles. Persist only deliberate Save/Stage/Reject actions; leaving unsaved local edits offers discard/keep-editing rather than silently saving. Show original versus draft values and save status.

Account/result changes clear editor content, cancel reads and fence late replies. Reload restores the last server revision. A second tab receives a conflict with its unsaved edits preserved for comparison. Expose an AI-only draft identity to future preview UI; never insert it into `TPendingLocation` or `saveAssetLocationsWithRetry`. If manual pending coordinates exist for the same asset, show the conflict and block AI staging until the user explicitly resolves that local pending choice; do not discard either decision. Manual edits in another client are ultimately handled by CH18/CH19 fresh GPS comparisons.

### 6. Persistence lifecycle, rollout and verification

Allocate migration 026 only after rechecking the tip during apply. Use composite owner-qualified foreign keys and account-delete cascade; do not attach drafts to mutable catalog rows. Protect referenced analyses/jobs from `aiJobStore.PurgeBefore` while a draft revision is retained, and serialize acceptance versus purge so neither an orphan nor silent draft deletion is possible. Catalog reset must retain drafts. Installation rotation makes old drafts inaccessible and unusable; it never migrates authority to a new instance. Local reads/edits remain usable with provider execution disabled.

Migrate fresh and 025 databases through the real runner, test repeat migration/reopen/foreign-key enforcement on the actual connection pool, and verify failure rollback. Back up SQLite consistently before rollout. Roll back application binaries with new tables retained; destructive down migrations are not a recovery strategy. Do not enable any Immich mutation UI in CH16. The apply verification record must map every scenario to tests and actual results; [verification-plan.md](verification-plan.md) defines the required layers and gates.

## Risks / Trade-offs

- Two-tab editing can interrupt a save → explicit revision conflict preserves both the durable value and local unsaved edits; no last-writer-wins fallback.
- Original fingerprints cannot reveal which historical field changed → require renewed current-image review, preserve provenance and avoid asserting image equivalence from an opaque hash.
- Retaining every explicit revision increases storage → bound payloads and do not create revisions for keystrokes/no-op acceptance; CH23/CH24 own deletion/retention policies.
- Metadata reads add authority boundaries → fail closed on fresh reads while retaining purely local history/editing; credentials and raw metadata stay out of ordinary logs.
- Adding draft joins can regress history pagination → preserve the existing terminal-history watermark and prove one-row-per-entry projection on real SQLite fixtures.
