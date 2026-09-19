## Context

See [proposal.md](proposal.md) for motivation and prerequisites. CH05 is planned alongside this change; its [design](../archive/2026-09-20-freeze-explicit-ai-selections/design.md) owns the selection policy, protected routes, immutable storage, installation identity, expiry and quotas. Implement and sync CH05 before applying this delta. CH04 already provides the shared source-local date predicates.

Design grounding at `86b44ac`: GitNexus `context` traces `buildAssetFilter` to `getFilteredAssets`, `countFilteredAssets`, `countAssetsByDay` and `getFolderAssets`; folder queries additionally apply a recursive path range. Direct source verification confirms owner-qualified album/tag relations, hidden-library suppression, stack-primary filtering, GPS/visibility predicates and deterministic `fileCreatedAt DESC, immichID DESC` ordering. Existing paginated count/list operations do not provide the atomic freeze required here. The graph's unavailable process-resource detail and lower-bound interface edges were checked against these adapters and source; graph absence is not evidence of no consumers.

## Goals / Non-Goals

**Goals:** extend the same selection service with a complete-query resolver, bounded memory and exact counts from a consistent SQLite read; reuse owner/installation authority, persistence and lifecycle rather than introduce a second kind of resource.

**Non-Goals:** changes to gallery page selection, gallery count endpoints, image decoding, an analysis launcher, provider calls, queue implementation or write permission. This change does not promise a NAS latency benchmark from local synthetic measurements.

## Decisions

### One mode-specific input and one shared snapshot contract

Extend CH05's `POST /ai/selection-preview` input union with `mode: "all-matching"` and required normalized `scope`. Reject `assetIDs`, pagination, cursors, client counts and sorting inputs in this mode, including explicit empty values. Keep the same JSON size/duplicate-key rules, authentication, exact mutation origin and AI-enabled guards. `GET /ai/selections/{id}` remains mode-independent and private.

The all-matching result contains `matchedCount`, `eligibleCount`, `excludedCount`, `exclusionCounts`, ordered eligible IDs and either the retained snapshot ID/expiry or null when empty/oversized. For compatibility with common snapshot accounting, set `requestedCount` and `uniqueCount` to `matchedCount`, and `duplicateCount` to zero; document that these query values describe discovered unique candidates, not user-supplied IDs. Aggregate reasons replace the explicit mode's per-input exclusions; the response union and persisted manifest discriminator make this difference explicit. No excluded-ID array is materialized. CH05's per-input accounting and errors remain unchanged.

### Reuse policy while distinguishing discovery from eligibility

The root SQLite adapter composes the owner-qualified catalog predicates verified above, including album/tag existence and ownership checks and the exact CH04 capture-date helper. Reuse or narrowly extract those predicates after GitNexus impact analysis; do not implement a second date parser or copy filter SQL into the domain. Folder range normalization remains the existing recursive, case-sensitive component-boundary convention. Use `EXISTS` or equivalent distinct candidate projection for relational filters so joins cannot multiply assets.

Discovery suppresses hidden libraries and stack children, as gallery queries do. It includes matching videos and individually hidden assets if the normalized visibility scope allows them; those pass through CH05's same classifier and contribute aggregate exclusions. Thus an explicit known stack-child ID can receive CH05's `stack_child` exclusion, while all-matching discovery never exposes/counts stack children. Neither mode expands a stack or overrides AI's hidden-asset policy. Unknown/foreign album or tag scope retains CH05's owner-safe failure, never an unscoped fallback.

### Stream one consistent transaction instead of looping over pages

The domain remains in `backend/internal/ai/selection/`. Add a narrow candidate-enumeration port implemented in root `aiSelection*.go` adapters. The SQLite adapter streams only policy/membership fields in `fileCreatedAt DESC, immichID DESC` order inside the same transaction used to calculate counts, enforce quotas and insert the snapshot. Keep transaction ownership in the adapter; neither SQL nor a concrete database type enters the AI package. If projection/order extraction would grow inherited oversized files, place the selection-specific responsibility in focused new files within the standards' ratchet.

Count every discovered candidate and classify it with the shared domain policy. Retain at most the configured eligible cap in memory plus constant-size counters and one row. After exceeding the cap, continue counting without accumulating more IDs. This is O(matching candidates) time and O(configured cap) retained memory. Raw matching count can exceed the cap: the external-input bound in explicit mode and eligible-output bound in query mode solve different problems.

At completion, zero eligible assets returns exact counts without persistence; eligible count above the cap returns `413 SELECTION_LIMIT_EXCEEDED` with exact counts and no snapshot. A positive in-limit result uses CH05's same atomic insertion, manifest-size bound and owner/global quotas. Compute the digest over mode, normalized scope and ordered retained IDs with the same installation/owner/policy binding. Historical aggregate counts are retained unchanged with the manifest.

Apply CH05's five-second resolution deadline and cancellation context to scanning and publication; database contention, cancellation or timeout yields the shared sanitized failure with no counts advertised as exact. Do not silently retry inside the handler. A new explicit preview request may retry against the then-current catalog. Read transaction duration is bounded; use existing WAL/SQLite connection policy, and prove read-to-write contention rolls back cleanly rather than assuming a lock upgrade succeeds.

### Frozen reads never rerun the query

Store the mode and aggregate summary in CH05's versioned manifest and reuse its snapshot/items tables. No additional migration is expected; verify the actual CH05 schema when implementation begins, and allocate an additive migration only if that verified representation needs one. Reading/consuming checks the same current owner, AI enablement, installation/policy version, expiry and eligibility of retained IDs in their original scope. It does not recompute historical counts or discover replacements. An ineligible retained member invalidates the complete snapshot; newly matching or newly excluded nonmembers do not affect it.

Reusing the single resource avoids divergent expiry or cleanup behavior. CH05's 20-live-snapshots-per-owner and 1,000-installation limits count both modes together; expired cleanup, one-minute sweeps, account deletion and installation-epoch changes act on both. No new dependency, external call, timer or cleanup service is needed.

### Verification at the observable boundaries

Use Plus TDD for each new scenario. Pure policy tests compare mode accounting while real file-backed SQLite fixtures prove whole-query membership, owner isolation, filter combinations, exact/over-limit boundaries, duplicate joins, source-local dates and deterministic order. Barrier-controlled concurrent sync and injected scan/commit failures prove consistent counts/publication; reopen tests prove query snapshots remain frozen. Use fake clocks for expiry and explicit fault/context controls rather than sleeps.

Authenticated normal/proxy HTTP fixtures exercise input union rejection, origin, enablement, current scope, shared quota/deletion/error behavior and zero calls to provider/image/Immich-write spies. Reuse explicit-selection and manual/GPX/browser regressions. A synthetic 100,000-candidate fixture with mixed eligibility and a 500-asset cap records wall time, allocations and deadline behavior, including an oversized result; this provides implementation evidence, not the later owner-approved hardware acceptance claim.

Run all 13 installed shared checks and production build appropriate to the final checkout, changed Go package tests including race coverage, AI Go statement coverage at least 80%, relevant frontend coverage if changed, migration/reopen checks inherited from CH05, file-size/dependency checks and GitNexus pre-commit change analysis. Scenario-to-task coverage is explicit in [tasks.md](tasks.md); no live private images or provider access is needed.

## Risks / Trade-offs

- Large matched sets cost scan time even when eligible count is small → stream minimal rows under a deadline; fail completely if exact counting cannot finish. Index tuning must follow measurements and impact analysis, not an unbounded query promise.
- Catalog filtering and AI eligibility have different exclusions → share predicate composition and classifier, document the candidate universe, and test both counts and membership.
- A concurrent SQLite writer can invalidate transaction upgrade → rollback the entire preview and return a retryable sanitized storage error; never return a partially durable token.
- CH05 is not implemented at planning time → reconcile its concrete ports/schema before CH06 implementation and preserve the additive capability delta instead of assuming both can archive in arbitrary order.

### Migration and rollout

Apply after verified CH05 and its spec sync. No independent schema allocation or data rewrite is planned. Deploy with existing installation AI controls, verify explicit and all-matching fixtures together, and preserve the same cleanup operator guidance. On binary rollback, disable AI and remove/expire affected transient selection resources through the established lifecycle before running a version that cannot decode all-matching manifests; never reinterpret them as explicit selections. Existing manual/GPX data is untouched.
