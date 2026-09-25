## Context

See [proposal.md](proposal.md). CH16 supplies local draft revisions and reviewed baselines; CH18 supplies stored exact previews. Both are planning dependencies at `0fb8eaf`, not existing write behavior. CH11 provides reusable patterns for short SQLite transactions, injected clocks, bounded claims and stale-worker fencing; its analysis retry policy is not a mutation policy.

GitNexus resolved the manual chain from `useLocationAssignment` to `saveAssetLocationsWithRetry`; source confirms a second frontend save pass. `backend/immichClient.go` constructs a retrying client with `RetryMax = 3`, and `backend/handlers.go` expands manual location targets. Neither is the AI writer. Read `aiJobClaim.go`, `aiJobRecovery.go` and `aiJobAccount_test.go` during apply after fresh graph tracing; this design takes their architectural pattern, not unverified implementation compatibility. The graph has unresolved Go references, so the identified boundaries were checked in source and impact must be rerun before edits.

Follow [architecture](../../../../docs/engineering/architecture.md), [testing](../../../../docs/engineering/testing.md), [coding standards](../../../../docs/engineering/coding-standards.md), ADR-07 and [technical design §11](../../../../docs/ai-locate/TECHNICAL_DESIGN.md#11-immich-writeback-contract). No architectural departure or new library is proposed.

## Goals / Non-Goals

**Goals:** One durable approval produces at most its exact single-photo GPS operation, with explicit attempt accounting and recoverable observed outcomes. Approval, dispatch, verification and catalog refresh are separate durable facts.

**Non-Goals:** No promise of exactly-once upstream execution, remote transactions, atomic compare-and-swap or automatic rollback. No reuse of analysis retry budgets or the manual writer. No expanded field/scope support, even when the upstream API supports it.

## Decisions

### 1. A separately gated writer with consumer-owned interfaces

Add `internal/ai/writeback` for confirmation, dispatch/reconciliation policy and pure state transitions. It consumes preview/draft reads, an operation repository, current asset authority, a GPS-only mutation port and a GPS readback port. Clock, IDs and runtime scheduling are injected. Root `aiWrite*.go` adapters own SQL and HTTP integration; a focused `internal/aiadapters/immichwrite` transport owns exact external serialization. Analysis, draft and preview workflows receive no mutation port and must not import this writer.

Use the existing backend process and local SQLite. `AI_WRITE_ENABLED` defaults to false and, together with existing global AI enablement, gates new confirmations and every mutation reservation/dispatch. This is an operator rollout/kill switch, not a provider readiness policy. Local draft/preview/operation reads remain available, and already-started writes may be reconciled read-only while dispatch is disabled. Supported adapter/version configuration is checked on the server, without synthetic mutation of ordinary photos.

Alternative: call the legacy location handler from AI Results. Rejected because it adds stack scope and hidden retries. Alternative: add a queue service. Rejected by the adopted deployment and unnecessary for a single SQLite writer runtime.

### 2. Confirmation commits exact authority before work starts

| Route | Contract |
|---|---|
| `POST /ai/write-operations` | Accept only preview ID, matching digest and a stable application idempotency key; record approval and enqueue locally. |
| `GET /ai/write-operations/{id}` | Private durable plan, progress, safe error and observed outcome, with no upstream side effect. |
| `GET /ai/write-operations/by-key/{key}` | Reconcile a lost submission acknowledgement in the current owner/installation scope. |
| `POST /ai/write-operations/{id}/reconcile` | Request bounded read-only verification, including an operation whose automatic read attempts stopped. |
| `POST /ai/write-operations/{id}/retry` | Explicit bounded retry of the same approved GPS step only when reconciliation established eligibility. |

All routes use session/origin protection, no-store, strict bounded JSON and safe errors. Confirmation bodies are at most 4 KiB. Use 404 for unavailable/foreign references, 409 for changed digest/revision/active target/idempotency conflict, 410 for expired unconsumed previews, 400 for malformed input, and 503 for storage/disabled dispatch. An existing matching idempotency result is still readable/reconcilable after expiry or disablement; it does not create new dispatch authority.

In one short transaction, check installation, account, writer enablement, exact staged revision and preview expiry/digest; allocate the per-target guard; consume the preview exactly once; insert operation, immutable approved plan and approval audit. Unique `(owner, installation, idempotencyKey)` binds the key to one plan digest. Same key/same plan returns the existing operation; same key/different plan fails. A preview also has a unique operation reference so another key cannot consume it twice. A crash or commit failure here must result in either no operation/no mutation or the single recoverable durable operation. A response to the browser is not a prerequisite for the worker to find that committed record.

The browser creates and retains its key before POST using the existing HTTP-compatible request-ID helper; do not assume `crypto.randomUUID` exists on the NAS's HTTP origin. An uncertain submission keeps that key and uses lookup before any repeated POST. Disable the confirmation control while unresolved; never generate a replacement key automatically.

Initial history loading is subordinate to a newer confirmation in the same mounted draft view. Fence its automatic selection and error publication against that newer action; a lifetime abort guard alone only covers navigation. Clear a pending identity only when its matching operation is recovered or its own submission receives a definitive rejection. Initial and repeated confirmation share that rejection policy; absent lookup results and transport/storage uncertainty retain the same identity. This refinement preserves the existing private state and request boundaries without a new state library or persistence format.

### 3. Separate immutable audit from mutable progress

Add `ai_write_operations`, `ai_write_targets` and append-only bounded `ai_write_events`. Each operation stores owner/installation, preview and draft revision, immutable canonical plan/digest, approval time, policy version and mutable summary status. One target row stores approved before/intended GPS, reviewed-image identity, attempt reservations, current lease/generation and observed readbacks. Events record approval, reservation, dispatch outcome, verification, conflict and local-refresh outcome with time and safe codes. Record only the fields needed for the decision; no raw HTTP body, credentials, image bytes, user hints or full model answer.

Use account-qualified foreign keys and a separate unique active-target guard keyed by `(installationID, assetID)`, not by account: two addon accounts must not write the same upstream photo concurrently. Foreign-owner contention returns only a generic busy/conflict response. A guard survives lease expiry and unresolved outcomes; it is not released merely because a worker disappeared. Protect operation-linked previews and draft/source history from ordinary cleanup. Completed history remains immutable even when a new draft revision or operation is created.

Suggested internal states: `queued`, `writing`, `verifying`, `retryable`, `succeeded`, `conflict`, `failed`, `canceled`. `verifying` can include `reconciliation_required` or `local_refresh_pending`; it is not successful. Store dispatch count and independently bounded read attempts. Unknown remote outcomes retain the target guard; a terminal no-dispatch failure/cancel can release it. Definite observed success/conflict can release it only after prior senders are quiescent. There is no partial multi-field state because this change has one GPS step.

### 4. Fresh checks and durable reservation fence each dispatch

A worker acquires the target guard and a generation-bound lease, reads fresh authorized metadata outside a transaction, compares source identity and exact GPS before-values, then rechecks local owner/installation/credential identity/draft revision/enablement in the reservation transaction. A changed baseline returns `IMMICH_CONFLICT`; a changed image requires renewed review. Known revoked/hidden/trashed/removed sources prevent dispatch. Do not overwrite current GPS merely because the draft came from Missing GPS. The approved target remains one asset even if stack membership changes.

If the preview already described an unchanged pair, verify it and record a no-op without sending a mutation. Otherwise commit the attempt reservation and `writing` state before calling the mutation adapter. Each attempt owns a generation token, deadline and exact approved bytes. Recheck local cancellation/enablement/authority immediately before sending. Errors after reservation are conservatively ambiguous unless the adapter proves no request was sent. Never hold a SQL transaction across the network. Expired approval permits readback/reconciliation but no new mutation; a new plan/confirmation is needed for later work.

Coordinate CH16 edits through an injected local revision guard: an edit/rejection before the next dispatch reservation cancels queued or safely retryable approval atomically and invalidates its preview; while a reserved attempt can still act it returns `WRITE_IN_PROGRESS` until the outcome is settled. The UI preserves unsaved edits for later. This makes the irreversible boundary explicit: an already-sent request cannot be recalled by editing the draft. The draft workflow does not import writer policy; the root repository implements the guard using the same transaction boundary.

### 5. Exact single-asset transport with no hidden resend

The initial adapter profile targets the pinned Immich v3.2.2 single-asset PATCH contract documented in the reconciliation, subject to GATE-02 verification on the actual deployed version. Use `PATCH /api/assets/{approvedAssetID}` with exactly `latitude` and `longitude`; preserve descriptions, direction, timestamps, ratings, favorite state and custom metadata by omitting them. Use current server-held credentials and the configured Immich origin. Do not use bulk endpoints or enumerate stack members. Unsupported versions remain unavailable until a profile is documented and tested; never try PUT or another endpoint automatically after failure.

Use a dedicated non-retrying HTTP transport, 20-second mutation deadline, redirects disabled and bounded response drain. Do not install retry middleware or an upstream idempotency/replay header without a verified Immich contract; application idempotency stays local. Ensure the transport cannot replay a potentially sent body automatically. Test actual request counts on dropped connections, redirects and 429/5xx responses rather than merely inspecting a retry setting. Readback uses CH18's bounded one-attempt read transport with a 10-second deadline. Returned success status/body is adapter-profile specific and covered by the fixture; it does not replace readback.

### 6. Reconcile before any possible retry

After any potentially sent request, move to `verifying` and read current authorized source/GPS. Compare image identity before coordinate outcome. The versioned readback tolerance is absolute `1e-7` degrees per coordinate, with no rounding of the submitted values; this is numerical API comparison, not claimed geographic accuracy. Before-value conflict detection remains exact as in CH18. An intended-value match means the desired state was observed, not proof this operation caused it. Record whether a mutation was attempted or the result was a no-op.

| Readback or recovery observation | Outcome |
|---|---|
| Intended pair matches and source is unchanged | Record verified GPS, then update local state; do not resend. |
| Baseline pair remains unchanged | Record the observation; retry is possible only after sender/completion and authority checks below. |
| Different GPS or changed source | Conflict/renewed review; preserve observed values and do not overwrite them. |
| Read unavailable, access lost or identity cannot be established | Remain unresolved with a safe reason; do not infer success or resend. |

There are at most two mutation attempts per confirmed operation, and the second requires explicit user Retry, a fresh unchanged-baseline read, current draft/authority, remaining approval validity and proof that the earlier sender cannot still act. No automatic mutation retry occurs. A local timeout alone does not prove upstream completion. If the adapter cannot establish that a potentially accepted request has finished, remain `verifying` and offer Check status; an unchanged read alone is insufficient for resend. A definitely unsent request or definitive completed rejection with unchanged GPS can become `retryable`. This deliberately narrows the technical design's “unchanged baseline permits retry” to cases with known completion; it does not promise automatic recovery from every timeout.

Automatic read reconciliation is bounded to three attempts per recovery cycle with persisted due times (1, 5 and 15 seconds), then pauses visibly. An explicit Check status starts another bounded read-only cycle. Retry requests themselves are idempotent per operation generation and cannot reserve concurrent second attempts. Read failure never consumes an additional mutation allowance or resets the durable count.

Resolve an already-accepted retry generation from the current owner/installation-qualified durable state before performing upstream eligibility reads. It returns the current operation even after completion, expiry, restart or disablement and grants no new authority. Only an unaccepted retry performs the fresh baseline/authority checks; retain a second duplicate-generation check in its reservation transaction so concurrent submissions still allocate at most one retry. Keep network I/O outside transactions.

Leases fence local persistence, not an already-sent remote request. Keep an in-process per-target sender guard until the goroutine finishes. Only one write-runtime process may own a local SQLite installation: acquire an OS-backed exclusive runtime lock for dispatch, and fail closed if unavailable or unsupported. Retain the durable target guard across process death; restart first reconciles potentially sent attempts and never treats an expired lease as permission to resend. Readback may continue while another worker is uncertain, but a new mutation cannot. External requests that outlive a crashed process remain ambiguous unless completion is established. These rules use the existing local-host deployment, not distributed leases pretending to fence Immich.

### 7. Publish verified outcomes and refresh the catalog safely

After intended GPS is observed, persist the observation, operation outcome and exact local asset GPS refresh in one short transaction where possible. If local persistence fails, the prior durable `writing`/`verifying` reservation remains recoverable; read again before any later update and never resend because the catalog is stale. Update only the approved asset in the authorized owner scope. Do not touch other users' catalog rows or fabricate a deleted local asset; if its row disappeared, mark refresh pending/unavailable for normal authorized resync while retaining verified upstream outcome.

Coordinate publication using a narrow root adapter over `SyncService.pauseUserSync` and `releaseUserSyncLock`. GitNexus context and source confirm that this boundary cancels/drains an existing per-owner sync and reserves its lock; existing tests cover reservation and timeout. Acquire it with a five-second deadline before the final readback, hold it through the short local publication transaction, then release on every path. A drain timeout leaves local refresh pending and performs no catalog update. Once drained, no older same-owner sync can publish after the verified pair; future syncs read fresh state. Do not hold a database transaction while draining or reading upstream, and do not pause sync for the full uncertain mutation lifetime. Impact and focused regression tests must cover the existing settings/API-key caller. This is reuse of the existing service boundary, not a new sync system.

AI Results displays operation status separately from analysis/review, links to the exact approved revision and shows before/intended/observed GPS and safe errors. Refresh missing-GPS counts and markers only after verified values are durably available locally. A photo may leave Missing GPS while its AI result and audit remain visible. Preserve drafts on conflict/failure; do not mark a newer revision “saved” because an older one succeeded. Keyboard/narrow-screen/tile-failure behavior and private-response fencing follow CH16/CH18. Pending manual work is preserved; known same-photo pending edits block confirmation. No AI operation is routed through the manual save helper.

### 8. Lifecycle, rollout and verification are part of the first release

Allocate migration 028 after rechecking the prerequisite tip. Test fresh databases, 025→CH16→CH18→CH19 upgrades, repeat migration, pooled foreign keys, uniqueness, transaction failure and reopen/recovery on file-backed SQLite. User deletion removes only that user's private operation data and fences late publication; target guards for a still-active sender cannot be repurposed as authority for another account. Retain only an opaque busy guard until that sender is quiescent, with no private payload after deletion. Installation rotation prevents all old dispatch and makes old private records unavailable under the new binding. Never use a new installation's credentials to reconcile an old origin.

Global/write disablement and shutdown stop claims/reservations and cancel local I/O; they cannot undo requests already received upstream. Preserve the ambiguous reservation for later read-only reconciliation. Do not start a new send simply on re-enable. Removing a provider profile does not alter a previously reviewed draft; writes depend on Immich authority, not provider availability. Ordinary cleanup must retain active reservations, linked plans and audit; CH23/CH24 own future deletion/retention changes.

Use two global writer slots and one per owner, at most 100 recovery candidates per sweep and bounded shutdown. Inject clocks and barriers in tests; no sleeping races. Keep every handwritten source/test file within 500 lines by separating approval, state policy, repository, transport, reconciliation, runtime and UI responsibilities. Full acceptance, persistence, race, boundary and build gates in [verification-plan.md](verification-plan.md) must pass before enabling dispatch. No slice may expose a temporarily unconfirmed/unlogged writer.

Before rollout, take a consistent SQLite backup, deploy with `AI_WRITE_ENABLED=false`, and run deterministic checks. GATE-02 requires the actual Immich version, method/payload, before/after readback and preservation of unrelated fields on an explicitly authorized disposable fixture. No live mutation is authorized by this proposal or the three analysis photos. Enable only after that evidence; otherwise report compatibility as unverified and keep dispatch disabled. Rollback disables dispatch, drains/reconciles outstanding work and retains new tables/audit. Do not run destructive down migrations or claim that binary rollback reverses Immich coordinates.

## Risks / Trade-offs

- Another client can change GPS between preflight read and PATCH → serialize this addon's writers and detect observed conflicts, but explicitly retain the upstream non-atomic race limitation.
- Remote completion may be unknowable after a timeout → keep an unresolved operation/target guard and allow read-only checks; availability is sacrificed instead of guessing a safe resend.
- API readback is not proof of sidecar completion or causal ownership → report verified API GPS and do not claim filesystem completion or exact-once execution.
- Account deletion can race a request already received by Immich → prevent further sends and private-data recreation, but do not promise remote rollback.
- Local catalog sync can return older data → coordinate publication at the existing sync boundary and prove the race with a controlled fixture before release.
