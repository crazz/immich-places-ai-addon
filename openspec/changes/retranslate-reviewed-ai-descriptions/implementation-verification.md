# CH17 implementation verification

Implementation base: `0c5bc0c0a4e2fb46b21efa332921ed968dc49209`. OpenSpec Plus Apply and TDD ran inline, as requested; no subagents, commits, archive, deployment or live data calls were used. Verification uses Node 22.23.2, Go 1.25.14 and Bun 1.4.2.

## Scenario evidence

| Scenarios | Actual automated coverage |
|---|---|
| T01–T03 | `translations` request normalization; providerhttp translation request/response tests; `TestAITranslationAdmissionRequiresCurrentPrivateAuthority`; protected translation HTTP admission; reviewed-basis component tests. Requests contain only approved text and instructions. |
| T04–T05 | `TestAITranslationAdmissionIsDurableAndPreservesDraft`, `TestAITranslationAdmissionRejectsSubstitutionAndConcurrentRun`; identical lost-acknowledgement component recovery. |
| T06 | `aiTranslationCapacity_test.go`, `aiTranslationQueue_test.go`, `aiTranslationUsage_test.go`; shared analysis concurrency, pre-send token/cost reservations, request limits, explicit policy violations and storage-failure settlement. |
| T07–T09 | `TestAITranslationExecutionKeepsPartialResultsWithoutRetryOrDraftMutation`; strict codec rejection matrix; protected paged history, SQLite reopen and execution-disabled component history. |
| T10–T12 | `aiTranslationAdoption_test.go`, `aiTranslationWriteGuard_test.go`; exact revision/facts, selected successful languages, geometry/unselected preservation, stale and active-writer rejection. |
| T13–T15 | `aiTranslationRetry_test.go`, `aiTranslationLifecycle_test.go`, `aiTranslationRuntime_test.go`, `aiTranslationAuthority_test.go`; chosen unsuccessful-language retry, cancellation barriers, interrupted recovery, joined shutdown and pre-dispatch/post-resolution/publication fencing. |
| T16–T18 | `aiTranslationHTTP_test.go`, `aiTranslationPrivateLifecycle_test.go`, `aiTranslationIsolation_test.go`, `aiTranslationMigration_test.go`; two-owner isolation, in-flight deletion/rotation, catalog reset, retention, immutable consent, pooled foreign keys and populated 028→029 reopen. Diagnostics use safe failure codes and do not log basis/output. |
| T19 | Built `review.spec.ts` journey “translates reviewed text by keyboard with partial failure and explicit local adoption”: 390-pixel viewport, unavailable map/image, exact two text-only provider calls, selected adoption, reload and zero Immich writes. Screenshot inspected. |
| T20–T21 | `TranslationReview.test.tsx`, `TranslationIntegration.test.tsx` and existing draft-navigation regressions: unsaved keep/discard/navigation, owner/revision/request-generation fencing, stale warning, read-only refresh/history and explicit actions. |

Tests use individual RED→GREEN→refactor cycles for changed behavior. Known existing authorization, migration and write-guard behavior was characterized without manufacturing failures. The local per-test journal is retained in `out/checks/ch17/progress.md`. Cross-task refactoring placed workflow policy/interfaces in the core, SQL/wire composition in root adapters and request coordination in the feature hook.

## Executed checks

- `bun run check --base 0c5bc0c`: eleven gates passed initially, including 59 frontend test files / 155 tests, frontend AI coverage (**93.82% lines, 84.46% branches**) and both production builds. Three inherited ESLint warnings remained; there were no lint errors.
- The Go gate found two prior migration tests hardcoded to schema 28. Both expected-version assertions were advanced to 29; their focused regressions passed. `bun run check --gate go-tests --base 0c5bc0c` then passed the complete race suite with **87.73% AI Go statement coverage (5,690/6,486)**. All thirteen gates have passing results across the recorded runs.
- Initial browser startup was denied by the macOS sandbox before tests could run. The approved outside-sandbox run, `node scripts/checks/smoke.mjs`, passed **3 legacy and 17 AI journeys**, with no retries. All services/data were local synthetic fixtures.
- `openspec validate retranslate-reviewed-ai-descriptions --strict` and `git diff --check`: passed.
- GitNexus was refreshed for this checkout. Cumulative tracked-change analysis reported 13 changed symbols and 17 affected flows, with **critical** aggregate risk, chiefly startup and draft review. Pre-edit impacts and source/call-site review covered the scheduler, installation lifecycle and UI integration; regressions exercise these paths. Process extraction has documented caps, and change detection omits untracked files; this is not complete graph coverage or a clean-risk claim.

## Limits

These checks establish deterministic behavior, persistence and synthetic integration. They do not establish translation quality, live provider text compatibility, reference-NAS performance or live Immich compatibility. CH19 GATE-02 remains pending and write dispatch remains default-off. CH20–CH22 are separate subsequent changes. No container/build/deployment inputs changed in CH17. Preserve additive migration 029 and encryption keys when rolling back binaries; no live rollback was exercised.
