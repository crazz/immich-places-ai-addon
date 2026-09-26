# CH20 implementation verification

Date: 26 September 2026. Base: `0c5bc0c0a4e2fb46b21efa332921ed968dc49209`. Status: implemented, verified and synchronized; all nine tasks and thirteen engineering gates complete. No commit, deployment, live provider request or private Immich mutation was performed.

## Implemented boundaries

Independent revisioned GPS/primary-description selection, preserve/replace/owned append, exact source observations and canonical v2 previews flow through the existing confirmed writer. Additive migrations 030–035 preserve physical v1 tables/bytes while adding private description baselines, v2 previews/operations, field outcomes, monotonic verification evidence and append lineage. Shared preview quota, target exclusion, writer runtime, OS lock and authority fences cover both versions. Description capability admission is separate and default-off.

The review found and corrected two safety defects with observed failing regressions: malformed unselected GPS no longer hides valid description evidence, and verification evidence now commits before fallible local catalog refresh. A later external revert cannot make a previously successful field retryable. Three bounded read-only backend review checkpoints concluded compliant. The frontend helper used the frozen DTO contract and main reviewed its completed checkpoint. Both helpers used Astra Extra High; the tool exposes no speed multiplier.

## Scenario evidence

Tests below are actual executable tests, not proposed fixtures. Root Go paths are under `backend/`; core paths are under `backend/internal/ai/`. Existing prerequisite tests remain active.

| Scenarios | Executed evidence |
|---|---|
| C01–C03, D05–D07 | `drafts/description_selection_test.go`: `TestStageReviewedDescriptionWithoutCamera`, `TestDescriptionSelectionRejectsUnreadyTextAndScope`; existing draft geometry/revision tests; `DescriptionWriteFields.test.tsx` and `DescriptionWriteReview.test.tsx` |
| C04–C06 | `aiDraftDescriptionBaseline_test.go`, `aiDescriptionBaseline_test.go`, `aiWriteDescriptionPreview_test.go`, `writepreview/description_preview_test.go`; exact text baseline and API/component validation tests |
| C07–C08 | `writepreview/description_test.go`: first append and repeated append with changed language; `TestAIVerifiedAppendOwnsStableLineageAcrossReopenAndNewDescriptions` retains private verified ownership and exact surrounding text |
| C09–C10 | `TestManagedAppendRejectsUnownedTamperedAndOverLimitText`, `TestAIDescriptionBaselineRejectsLossyAndMalformedText`; malformed/oversized durable-plan rejection; UI size/readiness validation |
| C11 | `TestAIStandardCombinedSocketWritesExactFieldsAndReconcilesEachOutcome`: one exact three-field request; `TestAIDescriptionOnlyDispatchWritesOneExactFieldAndPreservesGPS`: description-only request and unchanged catalog GPS |
| C12–C14 | Combined socket test covers response loss and both partial directions; `aiWriteStandardRecovery_test.go`, `aiWriteStandardRetry_test.go`, `aiWriteStandardEvidence_test.go`, `aiWriteEvidenceFailure_test.go`, `aiWriteFieldIndependence_test.go`, `aiWriteStandardReconcile_test.go`; real SQLite/reopen and no successful-field replay |
| C13 remaining-field exclusion | `TestAIStandardUnknownSenderExcludesAnotherOwnersRemainingV1Plan` holds the shared guard until definite completion; only a newly reviewed remaining GPS plan can then be confirmed |
| C15 | `TestAIStandardUpgradePreservesPopulatedV1ApprovalsAndGuards` compares exact plan bytes, digest, complete operation projection, attempts, generation, audit and guards for populated queued/verified/unknown schema-028 state across 035/reopen |
| C16 | `TestAIStandardNewDescriptionCannotOverlapRetainedUnknownV1Owner`; reverse-version remaining-field test; `TestAIStandardConfirmationRejectsUnknownAlteredAndOversizedPlans` rejects changed authority and >1 MiB with no mutation |
| C17 | `writeback/capability_test.go`, `aiWriteCapabilityConfig_test.go`, `TestAIStandardCapabilityChangesCannotAuthorizeConfirmationOrDispatch`; actual-installation-bound synthetic browser opt-in; ordinary/live dispatch remains gated |
| C18–C20 | `DescriptionWriteReview.test.tsx`, `StandardWriteOutcome.test.tsx`; new built `writeback.spec.ts` description journey verifies keyboard/full exact text, no camera, pending manual GPS, unchanged Missing GPS and retained history after disablement |
| R01–R02 | `aiWriteStandardRevision_test.go`, race-enabled `TestAIStandardDraftEditAndReservationHaveOneAtomicWinner`: exactly one edit/reservation winner and correct guard/revision state |
| P01–P10, P14–P16 | Maintained preview pure/SQLite/HTTP/component/browser suites plus description preview/API/late-response coverage |
| P11–P13 | `TestAIStandardPreviewQuotaAndExpiryAreSharedAcrossVersionsAndReopen`: shared ten-preview capacity, exact expiry boundary and protected history in both formats; maintained bounded cleanup fixture |
| W01–W11 | Maintained CH19 approval/transport/authority/no-op tests plus v2 exact confirmation/reload projection, local idempotency, capability fences, exact socket payload and malformed authority rejection |
| W14–W22, W27–W28 | Maintained CH19 recovery/retry/publication/sync tests plus v2 recovery, explicit retry identity, partial readback, unavailable-field evidence and catalog-failure/external-revert protection |
| W23–W26 | Maintained account/installation/disablement fixtures plus v2 capability confirmation/final-send checks, private cross-owner history, populated upgrade/audit and explicit default-off configuration |

The exact per-test RED/GREEN or existing-behavior characterization record is retained in `out/checks/ch20/progress.md`; frontend checkpoint details are in `out/checks/ch20/frontend-progress.md`. Tests of existing behavior were allowed to pass immediately under the adopted testing standard. No failures were skipped or retries added to the harness.

## Verification results

Commands use the pinned Node 22.23.2, Go 1.25.14 and Bun 1.4.2 runtime, loaded with `source out/checks/ch19/env.sh`. Shared gates use `bun run check --gate <gate> --base 0c5bc0c`. Gates were run individually to avoid repeating the long backend race/coverage suite; this does not omit a gate.

| Gate | Actual result |
|---|---|
| Checker tests | PASS |
| Size ratchet | PASS |
| Dependencies | PASS |
| gofmt | PASS |
| ESLint | PASS; three pre-existing legacy warnings |
| Route type generation | PASS |
| TypeScript | PASS |
| Go vet | PASS |
| Go race tests and AI coverage | PASS: all packages; root 707.689 s; AI statements 87.77% (6107/6958) |
| Frontend tests and coverage | PASS: 63 files, 176 tests; AI lines 94.26% (1314/1394), branches 86.38% (2283/2643) |
| Go build | PASS |
| Frontend production build | PASS; existing middleware convention deprecation notice |
| Complete browser smoke | PASS: 3 legacy journeys and 18 AI journeys; focused new description journey PASS (11.2 s) |

Logs are retained under `out/checks/ch20/`. Browser Chromium runs outside the default sandbox because its macOS launch is blocked there; every external service in these tests is a loopback synthetic fixture. The new narrow-screen screenshot `description-mobile.png` was inspected: exact text wraps within the dialog and selected fields/confirmation remain readable.

The initial full backend gate exposed a fixture scheduling error: an original analysis job retained a queued item and could be claimed instead of the second account's new job. Both cross-owner fixtures now cancel that unrelated analysis work without touching the unresolved writer approval or guard. Five race/coverage repetitions passed. The full root package also reached Go's default ten-minute aggregate limit while actively migrating a later fixture. The rerun uses command-local `GOFLAGS=-timeout=20m` with the same required race/coverage gate; no application deadline, test assertion, automatic retry or checker configuration changed.

GitNexus initially returned corrupt paths after an incremental refresh. A full forced rebuild repaired the graph; exact CLI checks replaced unusable results. Refreshed change analysis identified critical shared draft/decoder/writer/UI impact (304 changed symbols, 114 flows); maintained and new regressions cover those boundaries. New untracked files, interface/receiver and process-extraction gaps were checked against source/tests and are not treated as clean zero impact. Strict validation of this change passed. Main-spec synchronization adds six requirements and updates twelve across `ai-immich-writeback` and `ai-results-and-review`, preserving existing scenarios. The change remains unarchived.

## Separate live gate

Synthetic checks verify application behavior, not live description endpoint/version/rights compatibility or sidecar/UI behavior. CH19 live GATE-02 remains open. No capability attestation was installed on a live service, and the owner-approved private photo was not modified. The [operator guide](../../../docs/ai-description-writes.md) documents exact admission and recovery rules.
