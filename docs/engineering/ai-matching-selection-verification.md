# CH06 all-matching selection verification

Implemented on top of CH05 commit `781deb24ec307fd58465d87dde7af6a0f4b0c4ae`. Work used sequential inline OpenSpec Plus apply/TDD and self-review, without subagents. The owner authorized completing, archiving and committing each change before the next.

## Contract and scenario evidence

The existing selection transaction, quota checks, migration 021, identity binding, cleanup and retained-ID reader are reused. The new pure collector accepts an ordered candidate enumerator and retains at most the eligible cap; root adapters own SQL, deadlines and publication. Query manifests add historical counts and aggregate exclusions, while explicit wire behavior remains unchanged.

| CH06 scenario | Automated evidence |
|---|---|
| Matching assets span several pages | `TestAISelectionMatchingWholeCatalogOrderedRoundTrip`: 205 photos, exact descending order, protected HTTP and identical retained read |
| Combined scope preserves catalog boundaries | `TestAISelectionMatchingScopeParityAndSuppression`: album/tag, recursive case-sensitive folder, GPS, inclusive source-local capture dates, offsets, invalid/absent dates and retained-fact parity |
| Discovery does not expose suppressed assets | Same scope fixture: foreign owner, hidden library and stack children absent from both membership and counts |
| Ambiguous query input is rejected | `TestDecodeRejectsAllMatchingAssetIDPresence`, `TestAISelectionMatchingRejectsAmbiguousHTTPInput`: null/empty IDs, paging/cursor/count/sort/unknown filter, unavailable references and invalid scopes |
| Mixed candidates have reconcilable counts | `TestCollectMatchingClassifiesWholeStream`, `TestAISelectionMatchingCountsEmptyAndMixed`: four matched, two eligible, two distinct exclusion reasons |
| Empty and excluded-only remain explainable | Same SQLite counts fixture: exact zero/hidden-only summaries, no empty resource |
| Relations do not multiply assets | Scope fixture includes extra album/tag memberships; exact owner-qualified relation predicates yield each asset once |
| Exact and over-limit differ | `TestAISelectionMatchingExactAndOverLimitHTTP`: 500 retained; 501 returns 413 with completed exact counts, no ID subset and no new header/items |
| Exclusions do not consume allowance | `TestCollectMatchingLimitsRetentionButFinishesCounts`; `TestAISelectionMatchingHundredThousandCatalog` completes ordinary 100000-match scans with 10/500 eligible |
| Interrupted enumeration has no partial success | Pure failure/fact-budget tests; `TestAISelectionMatchingInterruptedScanDiscardsCounts` and `TestAISelectionMatchingFailedPublicationLeavesNothing`: cancellation after a row, deadline, scan conversion, storage, catalog, item and deferred-commit faults |
| Concurrent changes cannot split counts/membership | `TestAISelectionMatchingEnumerationUsesOneCatalogState`: barrier-controlled WAL writer between observed rows; old consistent counts/order, failed write promotion, zero publication; next query sees committed state |
| Later additions do not join | `TestAISelectionMatchingReopenPreservesHistoricalMembership`: real database reopen after a new match and an excluded video becoming eligible; unchanged IDs, order, counts and expiry |
| Frozen member loses eligibility | Same reopen fixture: hiding a member makes the complete result stale, with no usable subset; CH05 shared-reader regression suite covers additional authority/scope changes |
| Unauthorized query/resource access fails | `TestAISelectionMatchingAuthorityAndAccountDeletion`: missing session, wrong/missing Origin, disabled AI, foreign and absent resources |
| Lifecycle limits apply to both modes | `TestAISelectionMatchingSharedQuotaExpiryAndEpoch`, account-deletion fixture, and unchanged CH05 capacity/cleanup tests |
| Existing workflows and zero external effects | New normal-proxy all-matching browser journey compares unchanged provider, image-request and Immich-write counters; explicit journey plus AI-disabled gallery/manual/GPX smoke regressions |

No new migration or dependency was introduced. Existing migration-upgrade/reopen and scope tests run alongside query scenarios. The shared dependency gate rejects selection reachability to the writer. CH13 must consume frozen tokens and historical summaries; preview still conveys no analysis or write consent.

## Measurements and verification

The synthetic catalog has 100000 candidates and a configured eligible cap of 500. On this local host, the ordinary final development run completed in 120.92 ms for 10 eligible, 125.53 ms for 500 eligible and 142.65 ms for 100000 eligible with exact overflow. Total allocation churn was respectively 24091296, 25026576 and 51282880 bytes; these totals are not peak retained heap. The collector retains only cap IDs/facts and constant-size policy counters; the SQLite engine owns query execution and temporary storage under the same deadline.

A race-instrumented run hit the production five-second deadline. The first scale assertion incorrectly required success regardless of the permitted timeout outcome. The corrected oracle checks either a complete exact result or `DeadlineExceeded` mapped to 503 with no exact-looking counts and no new persisted rows. It does not skip work, retry or extend production deadlines. Ordinary performance and race failure-path evidence are reported separately; neither establishes NAS/reference-hardware acceptance.

A cancellation test also exposed buffered SQLite iteration completing after cancellation. Explicit context checks during and after enumeration now discard the whole result. The focused failure suite passed after that fix. An initial full-check run found formatting errors in the new browser assertion; these were corrected with scoped ESLint formatting.

All 13 installed gates have passing results for the final implementation: the full run passed 12 gates, then the scoped lint rerun passed after browser assertion formatting was fixed. The initial full command remains recorded as exit 1 for lint; it is not presented as a single clean run. Go race tests passed with combined AI statement coverage **88.47% (1404/1587)**; backend/frontend production builds, all seven browser journeys and `immich-places-ai:ch06-validation` Docker build passed. Three inherited lint warnings remain. OpenSpec strict validation passed all five then-active items, and local links/whitespace checks passed.

GitNexus was rebuilt for this checkout with PDG retained. The complete change response identified 37 changed symbols across 21 files and four affected processes at medium risk, with no partial/truncated response flags. Three flows are selection preview; the additional `Run → Admit` property-name association was checked against the unchanged capability runner/provider source and the passing dependency/zero-call fixtures. The broader graph still has documented unresolved receiver edges and process-index truncation, so source/SQLite/HTTP tests supplement it; graph absence is not an all-clear.

After spec sync/archive and documentation reconciliation, the final complete graph response covers 37 symbols in 29 changed files and three selection-preview processes at medium risk, again without partial/truncated flags. Final OpenSpec validation passed all four remaining active items; the maintained selection spec now has 13 requirements and 39 scenarios.

Inline spec, code-quality and final reviews confirmed all 16 delta scenarios, owner/transaction boundaries, bounded collector retention, exact completed counts, safe deadline failure, immutable reload and explicit compatibility. No unresolved correctness finding remains. Detailed local attempt logs and RED/GREEN/refactor notes live under ignored `out/checks/ch06/`.
