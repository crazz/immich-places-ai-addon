# CH22 implementation verification

Date: 26 September 2026. Base: `0c5bc0c0a4e2fb46b21efa332921ed968dc49209`. Status: all nine implementation tasks complete; checkpoint and whole-change reviews clean; engineering gates passed; maintained specs synchronized. No commit, deployment or live Immich mutation was performed. The change remains active and unarchived.

## Implemented boundaries

Optional metadata requires independent installation/profile capability, explicit field selection and an asset-reader disclosure. A canonical `mirror-preview-v4` plan binds the immutable standard target matrix and one analyzed-photo export under `immich-places-ai-addon`. Current reviewed direction, precision, place, up to eight languages and minimal provenance are allowlisted; private context, secrets, raw responses and arbitrary URLs are excluded. Payloads are bounded to 64 KiB. Complete list reads distinguish absence from unavailable/malformed data, and replacement requires exact locally verified ownership.

Migrations 043–045 preserve old approval bytes and add independent metadata steps, events, stable random records, lifecycle exclusions and combined history. Standard fields must be verified and settled before a separately reserved metadata attempt. Authority, source, standard values and namespace are checked again immediately before the exact one-item PUT. Source-consistent semantic readback precedes publication. Each step allows at most two explicitly authorized attempts. Unknown sender completion remains unsettled even after matching observation; accepted retry generations replay locally. Publication failure recovers by reading, including after restart and dispatch disablement. No successful standard write is repeated by metadata repair.

The UI provides default-off selection, independent stale-field review, exact expandable comparison, accessible disclosure, separate step outcomes and retained generation identity. It preserves manual pending GPS and fences stale account/result/history responses. Custom metadata does not promise native Immich display or file/EXIF updates.

## Scenario evidence

Root Go paths below are under `backend/`; pure policy paths are under `backend/internal/ai/`; UI paths are under `src/features/ai/`. All named cases are executable, not planned test names. Retained predecessor suites remain active.

| Scenarios | Executed acceptance evidence |
|---|---|
| M01–M03 | `TestAIMirrorSelectionCannotSilentlyDowngradeToStandardOnlyPreview`, `TestMirrorPlanRejectsMissingDisclosureExpandedScopeAndUnownedBaseline`; independent capability, draft scope and fixture admission tests; unsupported UI recovery offers explicit deselection |
| M04 | `TestMirrorExportContainsOnlyExplicitReviewedContent`, `TestMirrorSelectedCurrentSubsetExcludesStaleAndUnselectedContent`, `TestMirrorUnknownDirectionNeverBorrowsCandidateOrSubjectBearing`, `TestMirrorRetainedHeadingKeepsItsOwnMethodAndUncertainty` |
| M05 | `TestMirrorExportRejectsStaleInvalidOrOversizedSelection`, `TestMirrorExportAllowsEightLanguagesAndRejectsNineBeforeSizeLimit`, `TestAIMirrorCombinedPreviewFailureRetainsLocalContentAndCreatesNoPlan` |
| M06 | `TestAIMirrorChoiceUsesExistingPrivateRevisionStorageAcrossReopen`, `TestAIMirrorApprovalRejectsExportThatDoesNotMatchSavedRevision`, `MirrorDraft.test.tsx`, `MirrorSelection.test.tsx`; edits invalidate stale review and prior approval |
| M07 | `TestAIMirrorReplacesVerifiedOwnedNamespaceWithNewlyReviewedContent`, `TestAIMirrorCombinedStackWritesOnlyAnalyzedNamespaceAndPreservesSiblingMetadata`; exact replacement, stable ID, unrelated key preservation and zero sibling metadata reads/writes |
| M08 | `TestAIMirrorBaselineUsesAuthorizedCompleteListAndNeverTreats400AsAbsent`, `TestMirrorNamespaceRequiresCompleteUnambiguousObjectList`, `TestAIMirrorOwnershipRequiresExactApprovedAndObservedValue`, `TestAIMirrorRecordRejectsMissingAndUnverifiedStepLineage` |
| M09 | `TestMirrorSemanticComparisonPreservesArraysTextAndNumericPrecision`, `TestMirrorNamespaceRejectsUnboundedNumericWork`, `TestAIMirrorReadbackRejectsSourceChangedDuringNamespaceRead` |
| M10, M20 | `TestAIMirrorFailureRetainsCatalogRefreshExactDraftAndPartialHistoryAfterReopen`, `TestAIMirrorHistoryNeverReportsStandardOnlySuccess`, `MirrorOutcome.test.tsx`, `MirrorStandardOutcome.test.tsx`; real SQLite exact draft/refresh/history retained across reopen and truthful description-only no-op/retry scope |
| M11 | `TestAIMirrorRecoversLocalPublicationFailureAfterRestartWithoutResending`, `TestAIMirrorUnknownSenderStaysUnsettledAndRecoveryNeverResends`; evidence distinguishes observed state from sender causality |
| M12 | `TestAIMirrorRetryOnlyResendsMetadataAndAcceptedReplayIsLocal`, `TestAIMirrorRetryRejectsStaleExpiredUnavailableExhaustedAndCoalescesConcurrent`, `TestAIMirrorRecoveryHTTPBindsOwnerTargetStepAndGeneration`; known-complete explicit retry, concurrent coalescing and local replay |
| M13 | Retry safety and unknown-sender fixtures cover unavailable readback, two-attempt exhaustion, expiry, bounded scheduled observations, explicit readonly reconciliation and replay after disabled reopen |
| M14 | `TestAIMirrorActualPartialConflictingAndUnknownStandardReadbackBlocksMetadata`, `TestAIMirrorSendRechecksStandardFieldsImmediatelyBeforeMutation`; actual partial GPS/description, conflicting and ambiguous standard outcomes make zero metadata sends |
| M15 | `TestAIMirrorMigrationExtendsPreviewStorageWithoutChangingActiveV3`, `TestAIMirrorStepUpgradeNeverCreatesOrAdmitsStepsForOldApprovals`, `TestOlderPlanVersionsCannotAcquireMetadataAuthority`, `TestAIWriteUpgradeFromPrerequisitesAndPooledOwnerConstraints`; populated v1–v3 bytes/guards/audit and reopen |
| M16 | `TestAIMirrorRevokedAuthorityBetweenStepsNeverSends`: enablement, credential, installation, source, namespace, capability and expiry; retained standard success |
| M17 | `TestAIMirrorCleanupAndDeletionRetainOtherOwnersExactHistoryAndUnknownGuard`, `TestAIMirrorDeletionRetainsUnknownSenderAndKnownCompletionCleansOnlyItsGuard`; ordinary cleanup, two owners, cascade privacy, opaque unknown exclusion and safe late completion |
| M18, W26 | `scripts/checks/ai-mirror-fixture.test.mjs`, capability policy tests, operator guide; independent synthetic-only opt-in and separate live evidence boundary |
| M19, P14 | `MirrorSelection.test.tsx`, `MirrorComparison.test.tsx`; built keyboard/mobile/failed-map metadata journey checks disclosure, exact text, target and step order with zero preview mutations |
| M20, P15 | Built description-only mirror journey verifies failed first PUT then explicit retry, only one standard write, preserved local text/manual GPS/Missing GPS membership and disabled reload; mobile screenshot inspected |
| M21, P16 | `MirrorOutcome.test.tsx`, `MirrorComparison.test.tsx`, `mirrorApi.test.ts`, retained history tests; late status, confirmation, history and recovery replies cannot replace a newer private view or discard unresolved identity |
| D05–D07 | Maintained draft geometry and CH20 description readiness tests; mirror stale direction/precision review and strict selection tests; no precision threshold, invented camera or expanded field authority |
| R01–R02 | `TestAIMirrorDraftRevisionCancelsOnlySettledUnstartedMirror`; active metadata freezes revision, safe pending work cancels atomically, unsaved UI edits remain available |
| P01–P02 | Maintained exact staged preview suites plus `TestMirrorPlanBindsExactExportDisclosureAndAnalyzedStandardStep`, `TestAIMirrorHTTPPreviewRequiresDisclosureAndKeepsExactStackScope` |
| P03–P06 | Retained source/GPS/description/stack baseline suites plus complete namespace parsing, source bracketing, unavailable/foreign rejection and no fabricated baseline |
| P07–P10 | `TestAIMirrorPreviewFreezesSelectedExportAndDisclosureAcrossReopen`, `TestAIMirrorPreviewReloadIsStrictlyLocal`, approval-integrity tests and `TestAIMirrorOwnedUnchangedValueUsesVerifiedStandardAndMetadataNoops`; unchanged steps use zero mutation attempts |
| W01–W04 | `TestAIMirrorConfirmationPersistsOneBlockedStepAndReplaysLocally`, immutable approval validation; retained CH19–CH21 exact idempotency, overlap rollback, authority and commit-failure regressions |
| W05–W08 | `TestMetadataTransportSendsExactlyOneApprovedNamespaceItem`, `TestMetadataTransportNeverRetriesRedirectsOrAmbiguousFailures`; exact unicode payload, redirects, dropped connection, deadline, oversized response, 429/500/202/408, one request only |
| W09–W11 | Fresh standard/authority/source checks, `TestAIMirrorStandardStepVerifiesBeforeMirrorAndRetainsGuard`, no-op and lifecycle fixtures |
| W14–W18, W27–W28 | Metadata retry/recovery/HTTP/safety suites plus retained standard/stack regressions; no successful-standard resend or uncertain/exhausted reset |
| W19–W22 | Source-consistent mirror readback, exact local retention and read-only publication recovery; retained numeric GPS tolerance, sync drain and missing catalog suites |
| W23–W25 | Revision/revocation/cleanup/deletion/reopen fixtures; `TestAIMirrorDowngradeRefusesToDiscardRetainedState`, `TestAIMirrorLifecycleDowngradeCannotDiscardRetainedStepProtection` preserve audit and exclusions |

## Review and execution discipline

Backend checkpoints A and B passed separate spec and quality reviews. C's first spec review requested five missing acceptance combinations; six focused tests then passed against unchanged production. C spec re-review and code-quality review are clean. Main reviewed the frozen frontend against the full proposal/design/scenarios, strict DTOs, independent field scope, stale-response/retry identity, accessibility and manual-GPS boundaries. The frontend helper recorded 31 individual cycles and 134 focused/regression tests. D spec review identified description-only standard no-op/retry labeling and manual-GPS gating gaps. Two focused main-owned regressions drove field-specific labels and selected-GPS-only retry/replay checks; D spec re-review is compliant. D code-quality review is clean with no Critical, Important or Minor findings. Main's built screenshot review found a misleading parent outcome fallback; a component test failed before the two-line display fix and all 16 affected regression tests passed afterward.

Per-test RED/GREEN, characterization and refactor assessments are in `out/checks/ch22/progress.md` and `frontend-progress.md`. A recorded lifecycle-test batching deviation was corrected by separate failure/fix/verification cycles; later tests were added individually. Existing behavior characterization was allowed to pass immediately. No acceptance case is skipped or automatically retried.

The first full browser run exposed an existing polling/action timing race in the translation journey: keyboard input could be sent while its checkbox was still disabled by the final refresh. The trace contains no adoption request. The journey now waits for enabled state and asserts checked state before keyboard submission; application translation code is unchanged. The final built browser run passed all 3 legacy and 20 AI journeys without retries.

## Verification results

Pinned runtime: Node 22.23.2, Go 1.25.14, Bun 1.4.2 via `source out/checks/ch19/env.sh`. Shared gates use the explicit base above. Production build and browser checks run in the copied workspace on ports 13080/18089/18090/18091 because another app owns the normal ports. Production source is identical; only copied harness port literals differ. Chromium runs outside the macOS sandbox. All services and data used by these journeys are synthetic loopback fixtures.

| Gate | Actual result |
|---|---|
| Checker tests | PASS: 78 tests, no skipped cases |
| Size ratchet | PASS: 1310 handwritten files |
| Dependencies | PASS: frontend and Go boundaries |
| gofmt | PASS |
| ESLint | PASS: three inherited warnings; display/test follow-up scoped check also passes |
| Route type generation | PASS |
| TypeScript | PASS: final route generation/types plus the two focused description-only outcome/replay tests |
| Go vet | PASS |
| Go race tests and coverage | PASS: race-enabled root suite 1068.540s; AI statements 87.48% (7662/8759); command-local `GOFLAGS=-timeout=20m` |
| Frontend tests and coverage | PASS after display fix: 76 files/239 tests, AI lines 95.09% (1548/1628), branches 87.78% (3117/3551); excluding test builders: lines 94.95%, branches 87.73% |
| Go build | PASS |
| Frontend production build | PASS after the final display fix; inherited middleware deprecation notice |
| Complete browser smoke | PASS: 3 legacy (8.0s) and 20 AI (2.2min), including final field-scope fixes |

The initial full Go run was intentionally interrupted to supply the previously needed timeout; this was a command configuration correction, not a test failure. No timeout, retry, skip or assertion was weakened. No dependency, packaging or deployment inputs changed, so no container build is required.

GitNexus refreshed to 14,972 nodes / 48,475 edges / 760 reported flows. All-local and explicit-base change review found 522 changed symbols and 171 affected flows across cumulative CH17–CH22 work; the shared writer/draft/DTO risk remains critical. The complete compare result enumerated all 522 symbols and 171 flows without partial/truncated result flags. Index extraction still reports capped process candidates/traces, and unresolved receiver/cross-language edges need source and tests. The real composed browser/API/DTO tests supplement those gaps; untracked files were reviewed directly. Strict OpenSpec validation passes the change and all eight maintained specs after synchronization. The merge adds six requirements and modifies twelve, preserving all 177 predecessor scenarios and 42 untouched requirement blocks. Local-link checks resolve 188 references, all 48 documented Go test names exist, and whitespace checks pass (retaining two intentional Markdown hard breaks). The final read-only whole-change review is clean, with no Critical, Important, Minor or advisory artifact gaps. No production or test changes followed the final passing gates; final document/configuration checks also pass.

## Final evidence locations

Gate logs are retained under `out/checks/ch22/`: `checker-tests-final.log`, `size-final-3.log`, `dependencies-final-2.log`, `gofmt-final.log`, `lint-final-2.log`, `types-final-4.log`, `go-vet-final.log`, `go-tests-final-20m.log`, `frontend-coverage-final-3.log`, `frontend-build-final-3.log` and `smoke-final-3.log`. The Go build produced `out/checks/immich-places-backend`. `review-target-final.log` records the final two focused tests after nullable test-fixture normalization.

`graph-all-complete.json` retains complete returned symbol/process enumeration; `document-checks.json` records scenario preservation, test-name resolution and local links. `openspec-apply-final.json` reports 9/9 tasks and `all_done`. `openspec-specs-final.log` and `openspec-change-synced.log` record strict validation. The final mobile result is `metadata-mobile.png`. These local evidence files accompany the durable scenario mapping above.

## Separate live gate

Synthetic results establish application behavior, not deployed Immich version/permission support, native display, sidecar propagation or remote compare-and-swap protection. Metadata capability remains independently default-off and bound to exact installation/profile evidence. CH19 GATE-02 remains pending; the selected private photo has not been mutated. See the [operator guide](../../../docs/ai-metadata-mirroring.md). The change remains uncommitted and unarchived.
