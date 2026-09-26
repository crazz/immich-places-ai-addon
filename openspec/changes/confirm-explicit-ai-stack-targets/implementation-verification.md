# CH21 implementation verification

Date: 26 September 2026. Base: `0c5bc0c0a4e2fb46b21efa332921ed968dc49209`. Status: implemented, verified and synchronized; all nine tasks and thirteen engineering gates complete. No commit, deployment, live provider request or private Immich mutation was performed.

## Implemented boundaries

The analyzed photo remains the default. Explicit stack review records independently authorized source identities and nullable GPS baselines for at most 50 selected members. An immutable `stack-preview-v3` manifest binds the exact target/field matrix, draft revision, installation, policy and five-minute validity. Siblings receive only the approved GPS pair; the analyzed photo may also carry its independently approved CH20 description.

Additive migrations 036–042 retain private reviews, previews, v3 approvals, independent target attempts/generations/events/field outcomes, opaque deletion exclusions, typed append lineage and mixed history. Confirmation acquires every target guard atomically or none. The existing bounded writer executes one exact target at a time, with at most two explicitly authorized attempts per target and read-only recovery of ambiguous senders. Verified fields survive catalog failures; only source-consistent verified GPS refreshes existing local rows. Old v1/v2 approvals retain their bytes and authority.

Three bounded backend safety/spec review checkpoints concluded compliant after missing W17 and mixed S14 acceptance evidence was added. Main reviewed the frozen frontend checkpoint and final integration. The built browser exposed a missing composition-root route; a failing UI regression exposed late initial history replacing an explicitly inspected operation. Both were fixed and verified. Stack summaries now show attempt budgets only on individual targets. The two helpers used Astra Extra High; the delegation tool exposes no speed multiplier.

## Scenario evidence

Root Go paths are under `backend/`; core paths are under `backend/internal/ai/`; frontend paths are under `src/features/ai/`. These are executable tests, not planned test names. Existing prerequisite suites remain active.

| Scenarios | Executed evidence |
|---|---|
| S01 | `TestAIStackSelectsThreeReviewedPhotosFromFive`, `TestAIStackPreviewFreezesReviewedSubsetAndRetainsBytes`; `StackReview.test.tsx`; exact three-of-five built browser journey |
| S02 | `TestAIStackDefaultPreviewStaysSinglePhotoAfterMemberAdded`, `TestStackSelectionDefaultsToAnalyzedPhoto`; default selection component test |
| S03 | `TestAIStackReviewNeverSubstitutesPartialUnavailableScope`, `TestAIStackReviewBoundsConcurrentReadsAndHonorsCancellation`, `TestAIStackReviewFencesAuthorityChangesDuringReads`; protected HTTP/DTO rejection |
| S04, S20 | `TestStackSelectionRejectsUnauthorizedOrUnreadyExpansion`, `TestAIStackHTTPRejectsInvalidSelectionBeforeExternalReads`, `stackWriteTypes.test.ts`, `stackReviewApi.test.ts`; 50-member and GPS-readiness UI cases |
| S05 | `TestBuildStackManifestFreezesExactMemberFieldMatrix`, `TestAIStackExecutesExactIndependentTargetsAndRefreshesEach`, `TestAIStackPartialPrimaryFieldsRemainIndependentAndNeverReplay`; exact local socket bodies and primary-only description comparison |
| S06 | `TestAIStackUnselectedAdditionPreservesReviewedAndSavedScope`, `TestAIStackChangedMemberStopsIndependentlyWithoutExpandingManifest`; immutable saved membership through selected departure |
| S07 | `TestAIStackChangedMemberStopsIndependentlyWithoutExpandingManifest`: GPS/source/access changes stop only affected targets; analyzed-source freshness also fences later sends |
| S08 | `TestAIStackOverlapRollsBackAllNewApprovalAndGuards`, `TestAIStackUnknownSenderExcludesOtherVersionsAndDeletionPreservesOtherOwner`; last-target rejection leaves no partial approval/guards and no foreign disclosure |
| S09 | `TestAIStackRestartReconcilesAmbiguousMemberWithoutResendingSiblings`, `TestAIStackRecoveredGenerationFencesEveryLateTargetWorker`; real SQLite reopen, exact socket counts and stale worker fencing |
| S10 | `TestAIStackRetryAffectsOneTargetAndReplaysLocallyAfterSuccess`, `TestAIStackHTTPRequiresExactTargetForRetryAndPreservesAcceptedReplay`, `StackWriteOperation.test.tsx`; retained exact retry identity |
| S11 | `TestAIStackExpiryStopsOnlyUnstartedTargetsAndRetainsReadback`; UI observation expiry is distinct from immutable plan expiry |
| S12, R01–R02 | `TestAIStackDraftEditCancelsSafeTargetsAndFreezesAnyReservedMember`, `TestAIStackDisableAndShutdownPreserveSuccessAndUnknownSender`; retained unsaved-edit regression |
| S13 | `TestAIStackAccountDeletionRetainsOnlyUnknownOpaqueExclusion`, `TestAIStackInstallationRotationFencesActiveAndUnstartedTargets`, cross-owner deletion fixture |
| S14 | `TestAIStackMixedSuccessConflictAndMissingCatalogRemainIndependentInHistory`, `TestAIStackMissingCatalogTargetRetainsVerificationAndRecoversReadOnly`, `TestAIStackDrainsOlderSyncBeforePublishingEachVerifiedTarget`, `TestAIStackCatalogFailureRetainsVerifiedFieldsAndBlocksReplayAfterRevert` |
| S15 | `TestAIStackUpgradePreservesV1V2UnknownApprovalsAndExcludesOverlap`: populated schema 035→042, exact old bytes/counters/audit/opaque guards, new overlapping approval and reopen |
| S16, W26 | `TestStackCapabilityIsIndependentAndInstallationBound`, `TestAIStackCapabilityMustRemainExactAtConfirmationAndDispatch`; separate installation-bound synthetic opt-in; ordinary live dispatch remains gated |
| S17, P14 | `StackReview.test.tsx` full keyboard matrix with primary-only text, partial/zero GPS, failed tiles and narrow view; built keyboard exact-subset journey and inspected mobile screenshot |
| S18, P15 | `StackManualOverlap.test.tsx`, `StackWriteOperation.test.tsx`: every selected/restored/retried GPS target is checked; unrelated manual state is preserved; built journey retains one unrelated edit |
| S19, P16 | `TestAIStackPartialHistoryAndAuditSurviveCleanupAndDisabledReopen`, mixed publication fixture, `StackWriteOperation.test.tsx`, `WriteOperationHistory.test.tsx`; delayed owner/revision/history/action replies cannot replace the current inspected operation |
| D05–D07 | Maintained draft geometry/scope and description-selection tests plus stack expansion readiness/field-matrix rejection; no precision threshold or subject-coordinate substitution |
| P01–P02 | Maintained exact staged preview tests; `TestAIStackHTTPRequiresExplicitReviewedSelection`, stack manifest validation and no-write assertions |
| P03–P06 | Stack per-member review/safety tests preserve missing/partial/zero baselines, reject changed or unreadable selected fields and distinguish source identity from unrelated metadata |
| P07–P10 | Durable stack preview reopen/bytes, authority races, protected inputs, canonical decoder and no-op fixtures; strict TypeScript reviewed-selection binding |
| W01–W04 | `TestAIStackConfirmationPersistsEveryExactTargetAtomically`, overlap rollback and protected confirmation tests; retained CH19 idempotency, lost acknowledgement, malformed authority and commit-failure regressions |
| W05–W08 | Exact stack/primary socket tests and retained single-target transport tests; capability confirmation/final-send rejection; no hidden retry, redirect or fallback |
| W09–W11 | Changed-member/source/authority tests, capability races and `TestAIStackMatchingSiblingIsVerifiedWithoutAMutationAttempt`; verified unchanged member uses zero mutation attempts |
| W14–W18, W27–W28 | Stack recovery, retry, target-action, generation and partial-field fixtures; `TestAIStackRetryRejectsUnknownStaleExpiredExhaustedAndCoalescesConcurrent`; accepted replay is local after success/restart/expiry/disablement and cannot resend siblings |
| W19–W22 | Maintained numeric readback tolerance tests plus independent stack publication, sync drain, missing catalog and evidence-before-refresh fixtures; no fabricated rows or repair resend |
| W23–W25 | Stack disable/shutdown/deletion/rotation, cross-owner retention and private partial history fixtures; late completion cannot recreate deleted private state |

Append ownership additionally survives verified v3→v2 and v2→v3 writes/reopen in `TestAIStackVerifiedAppendOwnershipSurvivesV3V2AndReopen`. Every field's prior verification is monotonic, preventing a later external revert from granting replay of a previously successful combined payload.

Per-test RED/GREEN and characterization records are in `out/checks/ch21/progress.md` and `frontend-progress.md`. Existing behavior tests may pass immediately under the adopted testing standard. No skipped acceptance tests, weakened assertions or automatic retries were introduced.

## Verification results

Pinned runtime: Node 22.23.2, Go 1.25.14, Bun 1.4.2, loaded by `source out/checks/ch19/env.sh`. Shared gates ran individually with the explicit base above. Build/browser/type checks ran from a copied workspace on ports 13080/18089/18090/18091 because another app already owned the normal preview ports. Production source was identical; only copied test harness port literals changed. The running preview was left intact. Chromium ran outside the macOS sandbox; external services remained synthetic loopback fixtures.

| Gate | Actual result |
|---|---|
| Checker tests | PASS: 77 tests, including 30 size cases |
| Size ratchet | PASS: 1213 handwritten files |
| Dependencies | PASS: static frontend and Go boundaries |
| gofmt | PASS |
| ESLint | PASS: three inherited legacy warnings |
| Route type generation | PASS after final UI fixes |
| TypeScript | PASS after final UI fixes |
| Go vet | PASS |
| Go race tests and coverage | PASS: all packages, root 956.339 s; AI statements 87.52% (6872/7852) |
| Frontend tests and coverage | PASS: 69 files, 205 tests; aggregated AI lines 94.65% (1432/1513), branches 87.11% (2650/3042) |
| Go build | PASS, including the final composition-root route |
| Frontend production build | PASS after both history/display fixes; inherited middleware deprecation notice |
| Complete browser smoke | PASS after final rebuild: 3 legacy journeys (11.5 s) + 19 AI journeys (2.0 min) |

The root Go suite retained the command-local `GOFLAGS=-timeout=20m` used by CH20. The mixed S14 fixture was added after this full run compiled and separately passed under race (3.420 s). The only subsequent backend production change was one composition-root route registration, verified by rebuilt full-app browser journeys and Go vet/build. No writer/domain change followed the coverage run.

Gate failures were resolved: the size checker now skips directory symlinks while checking linked source files and retaining unreadable-source failures; lint now honors the already-adopted generated `out/` exclusion. These changes leave the other app's skill symlinks and local launcher untouched. A new test's literal HTTP header was changed to the existing `Object.fromEntries` convention. Browser startup conflicts and sandbox launch failures were infrastructure failures, not behavioral RED evidence; the subsequent missing endpoint and route failures were observed explicitly.

GitNexus refreshed to 14,330 nodes / 43,926 edges. All-local and explicit-base review each found 390 changed symbols and 160 affected flows across the accumulated CH17–CH21 work, with critical shared writer/draft/DTO impact. No import cycles were found (complete enumeration). Go HTTP route/response inference returned no supported route result, not an all-clear; real composed routing, strict DTO tests and browser requests supply that evidence. Process extraction remains capped and some receiver/cross-language edges unresolved; source inspection, new untracked-file review and behavior tests supplement the graph. Logs remain in `out/checks/ch21/`. Strict change validation and all eight maintained specs pass. Synchronization adds six requirements and modifies twelve, preserving every earlier scenario (writeback 63→80; results/review 94→97). The change remains unarchived.

## Separate live gate

Synthetic tests establish application behavior, not deployed Immich version/rights/stack endpoint compatibility, sidecar completion, visual correctness of a sibling's camera position or remote compare-and-swap. Stack capability remains independently default-off and installation/profile/evidence-bound. CH19 live GATE-02 remains pending. The private photo was not changed. See the [operator guide](../../../docs/ai-stack-writes.md).
