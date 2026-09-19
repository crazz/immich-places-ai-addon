# AI selection verification

CH05 implementation in `codex/plan-ai-selection-and-validation`, compared with planning commit `151ae12a972017de24741251062fcd0885ec1cc5`. Verification started 19 September and continued 20 September 2026, Europe/Lisbon. All data is synthetic and local. No NAS performance, private-photo transmission or fresh Immich permission claim is made.

## CH05 scenario traceability

Tests are in root `backend/aiSelection*_test.go`, pure `backend/internal/ai/selection/*_test.go`, the dependency-checker suite and `tests/e2e/providers.spec.ts` unless specified otherwise.

| OpenSpec scenario | Automated evidence |
|---|---|
| Owner creates and reads a selection | `TestAISelectionOwnerRoundTrip`; browser protected-proxy selection journey |
| Missing session or foreign snapshot | `TestAISelectionRequiresSession`, `TestAISelectionOwnerIsolationAndDeletion` |
| Disabled AI or invalid creation origin | `TestAISelectionAuthorityGuards`, `TestAISelectionRejectsDuplicateOrigin`, `TestAISelectionInternalConsumptionHonorsDisabledAI`; browser rejected-origin request |
| Duplicates are one target | `TestExplicitDuplicatesRetainFirstOccurrence`, `TestAISelectionOwnerRoundTrip` |
| Invalid or ambiguous input | `TestExplicitRejectsInvalidInput`, `TestAISelectionRejectsAmbiguousJSON`, `TestDecodeRejectsCaseAliases` |
| Size and batch bounds | `TestExplicitLimitsRejectWholeRequest`, `TestAISelectionRejectsAmbiguousJSON`, `TestAISelectionResponseLimitRejectsWholePreview`, scale test |
| Explicit IDs outside the active scope | `TestAISelectionAlbumTagScope`, `TestAISelectionCatalogExclusions` |
| Recursive folder boundaries | `TestAISelectionRecursiveFolderScope`, `TestInvalidScopeCannotBroadenSelection` |
| Date and GPS edge cases | `TestAISelectionCaptureDatesAndGPS`, `TestAISelectionUsesCaptureTimestampWithoutFileDateFallback`; retained CH04 capture-date tests |
| Invalid or unsupported scope | `TestInvalidScopeCannotBroadenSelection`, `TestAISelectionAlbumTagScope`, strict HTTP input test |
| Mixed eligible and excluded selection | `TestMixedSelectionUsesDeterministicExclusions`, `TestAISelectionCatalogExclusions` |
| Foreign and unavailable IDs are indistinguishable | `TestAISelectionCatalogExclusions`, `TestAISelectionOwnerIsolationAndDeletion` |
| Hidden scope does not override AI policy | `TestAISelectionHiddenOnlyScopeRetainsPolicyAndNoResource` and zero-eligible/mixed catalog tests |
| Catalog changes after preview | `TestAISelectionReopenDoesNotExpandOrRenew` |
| Concurrent sync and atomic failure | `TestAISelectionConcurrentCatalogObservationFailsAtomicallyOnPromotion`, `TestAISelectionPublicationFailureLeavesNoPartialSnapshot`, competing-creators race test |
| Reopen before expiry | `TestAISelectionReopenDoesNotExpandOrRenew` |
| A frozen target becomes ineligible | `TestAISelectionRetainedTargetChangesStaleWholeSnapshot` |
| Expiry, policy or installation change | `TestAISelectionAbsoluteExpiryAndPolicy`, `TestAISelectionInstallationRotation`, `TestAISelectionUsesConfiguredLimits` |
| Quota and cleanup | Owner/global quota, bounded cleanup and pre-creation cleanup tests |
| Expiry while AI is disabled | `TestAISelectionDisabledStartupCleansExpired`, `TestAISelectionPeriodicCleanupFailureRetriesAndCancels` |
| Account deletion and cleanup failure | Owner-isolation/account-cascade and periodic-failure/retry tests |
| Selection is locally verifiable | Browser provider/image-fetch/write counters; local HTTP/SQLite failure and stale tests; selection writer dependency prohibition |
| Existing workflows remain available | Existing AI-disabled auth/browse, manual placement and GPX browser journeys; no launch control in the new proxy journey |

Additional checks cover migration from 020 through production database reopen, owner-qualified foreign keys, manifest digest inconsistency, sanitized retryable errors, configured bounds/TTL/epoch and detached returned manifests.

## Results

- Focused selection Go tests passed; competing creators also passed with the race detector.
- Protected proxy journey passed against production frontend output and the compiled backend, including real session and SQLite. Provider requests, image-download count and Immich-write count remain unchanged by selection operations.
- Synthetic catalog: 100000 owner assets, 500 explicit IDs; one local run took **12.962125 ms**, allocating **2049824 bytes** during preview. The 501-ID request failed as a whole and left no additional snapshot. These are local measurements, not NAS capacity guarantees.
- The full shared run passed twelve gates; Go tests identified the bootstrap fixture's expected migration version still pinned to 20. Updating it to 21 passed the focused test. The final race/AI-coverage rerun passed: **88.07% AI Go statements (1322/1501), above the 80% floor**. After the capture-source correction below, a fresh complete run passed all thirteen shared gates.
- `docker compose --env-file .env.example config --quiet` passed. The real backend Docker build passed using the existing Dockerfile and local `immich-places-ai:ch05-validation` image; no deployment was performed.

Ignored local logs are under `out/checks/ch05/`, including the per-test RED/GREEN/refactor record, scale measurements, proxy checks and full gate output. Existing-behavior characterization tests were allowed to pass immediately; fixture/tooling failures were not counted as successful behavior verification.

The CH06 integration reconciliation caught an initial CH05 adapter mistake: capture scope projected `fileCreatedAt` rather than `dateTimeOriginal`. A regression fixture with independent file/capture dates failed, then passed after correcting the projection. It also proves missing capture dates have no upload-date fallback, changes to file dates do not stale capture scope, and changes to capture dates do. Existing date/GPS fixtures now explicitly populate capture timestamps. The corrected checkout passed a fresh complete thirteen-gate run and Docker build before the CH05 commit was finalized.

## Review and handoff

The implementation keeps policy/DTOs/digest in the focused selection core and HTTP/catalog/transaction/identity/cleanup composition in root AI adapters. No legacy oversized file grew. SQL parameters are bound; dynamic table names are fixed internal constants. The linear publication transaction remains in one function so rollback and quota ownership are explicit; candidate resolution, loading, installation binding and cleanup have separate responsibilities. There are no new external clients or dependencies.

CH06 must reuse the protected API, installation identity, immutable headers/items, retained-ID revalidation, eligibility reasons and retention policy. Its query-mode admission/counting is still separate planned work at this CH05 checkpoint; it must not re-run the original query when reading a frozen snapshot. CH11 must reuse the persisted installation identity and still enforce current authorization and fresh upstream permission before dispatch.
