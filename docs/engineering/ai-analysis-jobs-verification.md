# CH11 durable job verification

Implemented on top of CH09 commit `bc2439b4e833a0a29e6281a02d85846b96416c4c`. Work follows sequential inline OpenSpec Plus planning/apply/TDD and self-review without subagents. The [caller guide](../ai-analysis-jobs.md) defines internal execution, persistence and lifecycle limits. No public job route, startup worker, real provider dispatch or Immich mutation is connected.

## Scenario traceability

All seven delta requirements and 25 scenarios map to automated tests. Persistence tests use real temporary file-backed SQLite, production migrations and actual transactions; worker tests wrap the real store with a synthetic executor.

| Scenario | Evidence |
|---|---|
| Valid submission survives reopening | `TestAIJobSubmissionSurvivesReopenWithExactMembership`: exact normalized configuration, order/deduplication and queued membership |
| Invalid or foreign submission is rejected atomically | `TestAIJobInvalidSubmissionLeavesNoPartialWork`: invalid limits, consent/digest/languages/IDs, disabled authority, foreign profile and stale installation |
| Insertion failure rolls back membership | `TestAIJobInsertionFailureRollsBackHeaderAndItems`: item insertion, deferred commit and canceled-operation rollback |
| Concurrent identical submissions converge | `TestAIJobConcurrentIdenticalSubmissionsConverge`: twelve concurrent submissions through pooled connections |
| Changed request conflicts | `TestAIJobChangedRequestConflictsWithoutReplacingOriginal`: asset, language, revision and limit changes preserve the original |
| Explicit new run preserves earlier results | `TestAIJobExplicitNewRunPreservesEarlierResult`: distinct IDs and retained exact historical profile/model revision |
| Competing claims respect capacity and ownership | `TestAIJobCompetingClaimsRespectOwnerAndGlobalCapacity`: twelve competing workers and three owners |
| Future retry is not claimable | `TestAIJobFutureRetryWaitsUntilDue`: controlled clock before and exactly at the due time |
| Reservation cannot exceed item or job budget | `TestAIJobReservationsCannotExceedItemOrJobBudget`: concurrent reservations; `TestAIJobExhaustedBudgetTerminalizesUndispatchedItems`: no stranded queued work |
| Heartbeat requires the exact active lease | `TestAIJobHeartbeatRequiresExactUnexpiredLease`: exact/foreign/expired/replaced authority; `TestAIJobWorkerRenewsActiveLeaseAndCompletes`: successful monitor renewal and completion |
| Restart recovers expired work and fences old worker | `TestAIJobRestartRecoveryFencesOldLeaseAndReservation`: database reopen, replacement token, stale guard/finalization rejection and a new reservation requirement |
| Repeated interruptions reach a terminal bound | `TestAIJobRepeatedInterruptionsStopAtAttemptBound`: finite claims with zero or three actual reservations |
| Cancellation rejects late completion | `TestAIJobCancellationRejectsLateResultsAndPreservesCompleted`: mixed queued/running/completed membership and expiry reconciliation |
| Worker observes cancellation and heartbeat failure | `TestAIJobWorkerCancellationAndHeartbeatFailureJoinExecutor`: parent cancellation, heartbeat fault, canceled job and closed tick channel; `TestAIJobWorkerGuardsCannotOutliveAttemptContext`: detached callback contexts cannot regain authority |
| Valid unknown completion is durable | `TestAIJobUnknownCompletionIsAtomicAndDurable`: canonical unknown payload, exact metadata and reopen; `TestAIJobWorkerCompletesSyntheticAttemptThroughDurableGuards`: end-to-end synthetic attempt |
| Invalid result or failed insertion cannot partially succeed | `TestAIJobInvalidOrFailedResultNeverPartiallySucceeds`: invalid proposal/languages/digests/versions, absent reservation and failed result insertion |
| Terminal result cannot be overwritten | `TestAIJobTerminalAnalysisCannotBeOverwritten`: reused completed lease and direct SQL update rejection |
| History is private and survives source removal | `TestAIJobHistoryRemainsPrivateAfterSourceRemoval`: owner isolation, catalog deletion and independently loaded payload/metadata |
| Transient failure retries with bounded delay | `TestAIJobTransientFailureSchedulesBoundedRetry`: retained reservation, clamped retry-after/jitter and due-time admission |
| Permanent failure leaves unrelated items usable | `TestAIJobPermanentFailureDoesNotStopUnrelatedItems`: next item remains executable |
| Systemic failure blocks remaining job work | `TestAIJobSystemicFailureBlocksCallsAndLatePublication`: active and queued work fenced, unrelated job unaffected |
| Installation rotation invalidates unfinished work | `TestAIJobInstallationRotationFencesUnfinishedWorkAtomically`: existing binding rotation cancels old work and retains analyses; `TestAIJobFailedInstallationRotationRollsBackInvalidation`: shared transaction rollback |
| Account deletion removes private lifecycle data | `TestAIJobAccountDeletionCascadesOnlyOwnedPrivateHistory`: real FK cascades and other-owner preservation |
| Explicit cleanup is bounded and terminal only | `TestAIJobCleanupIsBoundedPrivateAndTerminalOnly`: old/recent/active/foreign history; `TestAIJobCleanupRejectsInvalidBoundsAndWorksWhenAIDisabled`: cutoff/batch validation and explicit disabled-AI cleanup |
| Existing database upgrades safely | `TestAIJobUpgradePreservesDataAndPoolConstraints`: upgrade from 021, reopen, preserved catalog/provider/snapshot data and four held FK-enabled connections |

Additional tests verify the 100-item recovery bound and systemic block preservation, five-second transaction contexts, safe worker failure routing without hidden retry and pure concurrency/lease/retry-policy bounds. There are 35 focused CH11 test functions, including table-driven failure cases. Characterization tests passed immediately where behavior was already implemented; new or corrected behavior followed observed RED/GREEN cycles.

## Verification results

All thirteen shared gates passed against CH09 base `bc2439b`: checker tests, source-size ratchet, dependencies, Go formatting, lint, type generation, TypeScript, Go vet, race-instrumented Go tests/coverage, Go build, frontend unit tests, frontend build and browser smoke. The initial run passed twelve gates and found one stale migration-test expectation (21 instead of the newly added 22). After correcting that assertion, the complete backend race/coverage gate passed; no production fix or skipped test was needed. Final size/format checks were also repeated after the test edit.

Backend AI statement coverage was **89.03% (2678/3008)** against the 80% floor. All 28 frontend tests and seven browser journeys passed. Lint reported the same three inherited warnings and no errors. The backend Docker image built successfully with the jobs package and migration. No dependency manifest changed.

Fresh staged GitNexus analysis returned all **341/341 changed symbols across 64 files and 31 affected flows**, with no partial or truncated result. Aggregate risk was **CRITICAL**. Reviewed flows cover normalization, canonical validation, transaction contexts, lease authority, reservations, cancellation, failure policy and cleanup. Source confirms that store/worker construction remains test-only; the live integration is the existing installation binder calling atomic invalidation, plus normal migration application. The full backend race suite covers that integration and legacy migration behavior. Cycle enumeration was complete with zero circular imports.

An incremental GitNexus refresh reached its generated database size limit; a successful full index rebuild restored the pre-commit review. Index inference still reports receiver/cross-language gaps and bounded flow enumeration, so returned change analysis was checked against actual callers and tests. Complete change-query output is not an exhaustive runtime graph claim.

## Limits and remaining evidence

No private photographs, credentials or live provider calls are used. The lifecycle code has no external transport or writer dependency. Existing manual/GPX and AI-disabled behavior remains part of the shared browser regression suite; no new launch UI is claimed.

CH12 must supply production selection/consent/access admission, consistent worker policy and heartbeat scheduling, provider failure translation, usage/token/cost accounting and public submission/cancellation integration. Real-model quality, full-schema live compatibility, reference-NAS 500-asset submission latency and load behavior remain unmeasured. No automatic retention period is chosen. Later drafts/write audits must protect referenced results from cleanup.
