# CH13 batch workflow verification

Implemented against `cac9613a2cdd7fdd10a588126419f5f62b8e3627` using sequential inline OpenSpec Plus apply/TDD and self-review, without subagents. The [workflow guide](../ai-batch-workflow.md) records the public behavior. New persistence tests use real temporary SQLite with application migrations and bounded loopback Immich/provider fixtures.

## Scenario traceability

All eight delta requirements and 24 scenarios are mapped below. Existing CH11/12 tests remain evidence for unchanged shared transactional and lifecycle behavior.

| Scenario | Automated evidence |
|---|---|
| Distinguish page from all matching | [Intention tests](../../src/features/ai/selectionIntent.test.ts), [preview tests](../../src/features/ai/BatchLaunch.test.tsx), [browser workflows](../../tests/e2e/workflows.spec.ts) |
| Explain exclusions and limits | [Preview component](../../src/features/ai/BatchLaunch.test.tsx), [over-limit response](../../src/features/ai/selectionApi.test.ts), [gallery scope](../../src/features/ai/GalleryIntegration.test.tsx) |
| Invalidate changed preview inputs | [Preview invalidation](../../src/features/ai/BatchLaunch.test.tsx), [gallery mapping](../../src/features/ai/GalleryIntegration.test.tsx), browser matching-to-page consent reset |
| Launch Visual without hidden context | [Admission builder](../../src/features/ai/launchAdmission.test.ts), [launch form](../../src/features/ai/LaunchForm.test.tsx), browser Visual execution |
| Choose bounded Context-assisted inputs | [Core choice validation](../../backend/internal/ai/jobs/context_validation_test.go), [Context admission](../../backend/aiProductionContextAdmission_test.go), launch form and browser Context reanalysis |
| Reset confirmation after changes | Launch form, preview invalidation and browser matching-to-page reset |
| Persist contextual completion across reopen | [Completion](../../backend/aiProductionContextCompletion_test.go), [reopened result](../../backend/aiProductionContextReopen_test.go), [restarted retry](../../backend/aiProductionContextRestart_test.go) |
| Reject invented or changed evidence | [Stored source authority](../../backend/aiProductionContextEvidence_test.go), [changed retry](../../backend/aiProductionContextRetry_test.go), [publication freshness](../../backend/aiProductionContextFreshness_test.go), [post-resolution authority](../../backend/aiProductionContextDispatchAuthority_test.go) |
| Preserve Visual isolation during upgrade | [Visual execution](../../backend/aiProductionExecution_test.go), [migration retention](../../backend/aiProductionMigration_test.go), [explicit Context policy](../../backend/aiProductionContextPolicy_test.go), [core consent](../../backend/internal/ai/jobs/context_admission_test.go) |
| Resume observing a mixed job | [Progress panel](../../src/features/ai/JobProgressPanel.test.tsx), [URL restoration](../../src/features/ai/AIWorkspace.test.tsx), browser reload and mixed cancellation |
| Lose connectivity or switch identity | [Polling lifecycle](../../src/features/ai/useJobProgress.test.tsx), [progress DTO consistency](../../src/features/ai/jobValidation.test.ts) |
| Reconcile ambiguous submission | [No automatic POST retry](../../src/features/ai/jobApi.test.ts), launch form and browser lost-acknowledgement reconciliation |
| Cancel mixed work | Progress panel, browser keyboard cancellation and [durable cancellation](../../backend/aiJobCancellation_test.go) |
| Retry failed members as a new run | Progress member choices, [fresh linked run](../../backend/aiProductionRerun_test.go), [parent state boundaries](../../backend/aiProductionRerunAuthority_test.go) |
| Reanalyze with new choices | Browser Context reanalysis, launch form and fresh linked-run tests |
| Reject invalid lineage or changed eligibility | Parent ownership/membership/eligibility boundaries in rerun authority tests |
| Complete the workflow using a keyboard | Browser keyboard preview/consent/submit/cancel and progress announcement/focus test |
| Preserve the write boundary and disabled behavior | Browser execution-disablement read, zero-write fixture assertions, existing manual/GPX journeys, [disabled cancel API](../../backend/aiProductionHTTP_test.go) |
| Valid submission survives reopening | [Production reopen/scale](../../backend/aiProductionScale_test.go), restarted Context retry |
| Invalid or foreign submission is rejected atomically | [Production admission rejection](../../backend/aiProductionAdmissionValidation_test.go), Context admission and rerun authority tests |
| Insertion failure rolls back membership | [Atomic admission failure](../../backend/aiProductionAtomic_test.go) |
| Admit without waiting for inference | [Production admission](../../backend/aiProductionAdmission_test.go), Context admission and freeze-before-dispatch tests |
| Reject stale or incomplete authority | Production admission rejection, Context post-resolution authority and rerun authority tests |
| Retry an admitted key after snapshot expiry | [Durable idempotency](../../backend/aiProductionIdempotency_test.go), explicit same-key UI reconciliation |

## Verification status

All thirteen shared checks have passing evidence: the initial cumulative run passed eleven, and its dependency/lint failures were corrected and rerun successfully. Final frontend changes were rechecked with strict types, size, dependencies, lint, full coverage (57 tests), production builds and all ten browser journeys. The unchanged backend race suite passed with 88.22% AI statement coverage (3790/4296); the backend Docker build passed. Both frontend AI coverage floors pass. GitNexus was force-rebuilt after stale entries were found: the complete pre-commit analysis reports 490 changed symbols across 94 files, 88 affected processes and CRITICAL aggregate risk, with zero import cycles. Context/rerun taint reports contain no findings; unsupported callback/field/implicit flows and lower-bound process discovery are covered separately by source review and behavioral tests, not treated as a proof of safety. Evidence and per-test RED/GREEN or characterization records are retained locally under ignored `out/checks/ch13-apply/`.

The browser harness creates disposable accounts and provider profiles through the real application. A test-only loopback control process reads its temporary SQLite database using Node 22's built-in read-only `DatabaseSync`, binds an explicit synthetic execution policy to the actual installation/owner/revision, and restarts the backend with that policy. It writes no database rows and adds no production endpoint or dependency. The original one-pixel fixture had an invalid PNG IDAT checksum; a correctly encoded one-pixel fixture now exercises the strict Go image decoder as well as the browser.

Synthetic canonical responses establish integration, consent, accounting and persistence behavior. They do not establish model quality, compatibility with a live provider or NAS capacity. No private image or real provider was used. No Immich writer or draft acceptance path is added.
