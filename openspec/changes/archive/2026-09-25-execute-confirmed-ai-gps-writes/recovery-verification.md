# CH19 recovery follow-up verification

Date: 25 September 2026. Base: `990a82cb28cee425e87c6105244bf37e198e6d4d`; branch: `codex/plan-ai-selection-and-validation`. This record covers the three reviewed recovery defects and tasks 6.1, 6.2 and 7.1. The earlier [implementation record](implementation-verification.md) remains historical evidence for tasks 1–5.

OpenSpec Plus implementation, spec-compliance, TDD and code-quality reviews were performed inline and sequentially under the user's existing preference. All 19 implementation tasks are complete; all 13 cumulative verification gates pass. Recovery fixes were committed as `9cc39d1` and CH19 was archived on 25 September 2026 after explicit user authorization. No deployment or live mutation was performed; ordinary writing remains disabled and GATE-02 remains unverified.

## Changes and scenario evidence

| Review finding / scenarios | Implementation | Maintained regression evidence |
|---|---|---|
| R1: late history replaces newer confirmation; R07 | `WriteOperation.tsx` fences initial history selection and error publication against a newer confirmation, including the detail request already in flight. | `WriteOperationRecovery.test.tsx`: `retains an unresolved confirmation when initial saved history arrives late`; `keeps the newly acknowledged operation when an older history detail arrives late`; `preserves confirmation feedback when the obsolete initial history request fails`. |
| R2: expired repeated identity blocks recovery; R08–R09, W02–W03 | Initial and repeated confirmations share definitive-rejection handling. Missing lookup and ambiguous repeat retain the exact identity; a definitive rejection permits a fresh explicit comparison/confirmation. | Same UI file: `allows a fresh explicit confirmation after a repeated identity is definitively expired`; `retains the exact pending identity after missing lookup and an ambiguous repeat failure`. |
| R3: accepted retry depends on fresh upstream eligibility; W27 | `aiWriteRetry.go` reads the accepted generation within current owner/installation scope before upstream I/O. It returns the current operation without another allocation. | `aiWriteRetryRecovery_test.go`: `TestAIWriteRetryReplayAfterSuccessIsLocal`. |
| W28: durable offline recovery and privacy | SQLite reopening, expiry and disablement do not prevent local recovery; foreign owners, obsolete installations and unaccepted generations do not gain recovery authority. | `TestAIWriteRetryReplaySurvivesReopenAndAuthorityLoss`; `TestAIWriteRetryReplayRetainsOwnerInstallationAndGenerationScope`. |
| W16–W17, W27: concurrent replay versus new retry | The reservation transaction checks the accepted generation before applying the fresh-read result. A losing duplicate recovers another request's accepted retry even when its delayed read fails or observes changed GPS. New retries still require fresh eligibility. | `TestAIWriteRetryConcurrentReplayRecoversAfterReadLosesEligibility`; `TestAIWriteUnacceptedRetryStillRequiresFreshEligibility`; existing `TestAIWriteExplicitRetryIsIdempotentBoundedAndRequiresKnownCompletion`. |

Backend tests use real temporary SQLite, the production migration chain and loopback HTTP fixtures. Request counts, durable attempts/generations and channel barriers supply evidence; timing sleeps are not used as concurrency proof. UI tests exercise actual request/response validation with controlled fetch results and user actions.

## TDD and review ledger

Ignored local evidence is retained under `out/checks/ch19-recovery/`, including the detailed `workflow.md` checklist.

- UI tests 1–3 each failed for the specific history race, then passed after adding the corresponding version check. Evidence: `ui-01-red.log` through `ui-03-green.log`.
- UI test 4 failed because the expired identity remained stored, then passed after sharing rejection handling. Its fresh submission is a distinct explicit user action. Evidence: `ui-04-red.log`, `ui-04-green.log`.
- UI test 5 immediately passed as characterization of the preserved ambiguous branch. Evidence: `ui-05-characterization.log`.
- Go test 1 failed with HTTP 409 after a successful accepted retry. The local generation lookup made it pass. Evidence: `go-01-red.log`, `go-01-green.log`.
- Go test 2 immediately passed after that fix, proving persistence across reopen plus expiry/disablement/offline recovery. Evidence: `go-02-characterization.log`.
- Go test 3 failed with HTTP 409 for both controlled concurrent outcomes: changed GPS and unavailable readback. Moving fresh eligibility rejection after the transaction's duplicate check made both pass. Evidence: `go-03-red.log`, `go-03-green.log`.
- Go tests 4–5 characterize preserved private scope and fresh retry eligibility. The initial test-4 attempt had a fixture compile error (`Database` versus its raw SQL handle); the test was corrected before execution. That error is not counted as a behavioral RED. Evidence: `go-04-characterization-fixed.log`, `go-05-characterization.log`.

Each test function was added and verified before the next. Refactor assessments retained distinct asynchronous and transaction boundaries, extracted only the shared rejection policy and reused the existing fixtures. Slice reviews ran in spec-compliance → TDD → code-quality order. The whole-change review found no remaining Critical or Important issue: no DTO, schema, dependency, deployment, manual/GPX or provider behavior changed, and network I/O remains outside SQLite transactions.

## Scoped verification

- UI: 24 tests across four files passed; scoped ESLint, TypeScript/typegen, size and dependency checks passed.
- Backend: eight top-level retry/recovery/reopen/revision tests passed with `-race -count=1` in 15.618 seconds; Go vet passed.
- Size: 1,020 handwritten files checked. Changed source/test files contain 96, 252, 151 and 120 physical lines, respectively.
- Dependencies: 361 frontend files, 694 runtime edges and 582 Go files checked.
- Strict OpenSpec validation passed all 12 specs/changes, including eight maintained capabilities. Both maintained specs preserve their existing requirements and add only R07–R09/W27–W28 plus the corresponding requirement refinements. A temporary merge-script duplication was caught in diff review and corrected before validation. Exact block/purpose comparison, local document links and whitespace checks pass.

## GitNexus and source review

Pre-edit upstream impacts were LOW for the operation session, confirmation actions and retry adapter. The refreshed PDG index is bound to the recorded base and contains 50,782 nodes, 135,897 edges and 625 reported flows. All four changed source/test SHA-256 hashes match the index (`graph-source-hashes.json`). MCP still reports an obsolete one-commit staleness warning; the on-disk index commit and source hashes were checked directly.

The graph's process enumeration is bounded and has cross-language field gaps. Retry context reports five unresolved receiver calls; source inspection identifies those as the five direct calls in the new recovery tests, alongside the resolved production route. They are exercised by the race suite. No empty graph result is treated as proof of no caller.

The initial tracked-only change analysis returned all 36 symbols and ten affected flows with HIGH aggregate risk. A disposable Git index/object directory under `/private/tmp` includes both new tests without changing the user's staging state. Complete analysis returned all **103 changed symbols across 13 files and ten affected flows**, with no partial/truncated result (`graph-all.json`). The CLI's display caps its visible symbol list at 15 even when `--limit` is larger; the complete structured result was captured through the same installed GitNexus backend used by its CLI. The affected flows are the protected retry route through source/GPS validation and operation UI through request/deadline/DTO validation. Those boundaries were inspected in source and are covered by targeted regressions and the cumulative suite. This count predates the final verification-record additions; source/test contents are unchanged.

## Cumulative verification

`bun run check --base 990a82cb28cee425e87c6105244bf37e198e6d4d` exited **0**, with all **13 gates passing** on the final source. Full output: `out/checks/ch19-recovery/full-check.log`.

| Gate | Actual result |
|---|---|
| Checker regressions | PASS: 73 tests |
| Source size | PASS: 1,020 files |
| Dependency boundaries | PASS: 361 frontend files, 694 runtime edges, 582 Go files |
| Go format | PASS: 583 files |
| ESLint | PASS; three inherited warnings |
| Route type generation | PASS |
| TypeScript | PASS |
| Go vet | PASS |
| Go race tests and AI coverage | PASS: root package 483.817 seconds; combined AI statement coverage **87.82% (5298/6033)** |
| Go build | PASS |
| Frontend tests and AI coverage | PASS: **141 tests / 56 files**; AI lines **93.75% (1185/1264)**, branches **84.28% (1849/2194)** |
| Frontend build | PASS |
| Built-application browser acceptance | PASS: **3 legacy + 16 AI journeys**, including exact GPS confirmation/lost-ack recovery/history, with no retries |

The backend run was freshly instrumented after the changed Go source/tests. The pure writeback package is exercised through root integration in the combined coverage profile; its standalone 0% line is not the combined AI measurement. Coverage remains above the installed 80% thresholds. The inherited Next.js middleware deprecation notice is unchanged. No container/deployment configuration changed, so the earlier container evidence was not rerun for this recovery-only diff.

## Release limits

Live GATE-02 was not run. Synthetic results do not establish compatibility with a configured live Immich instance. `AI_WRITE_ENABLED=false` and the unsupported empty default profile remain unchanged. No live provider call, private-photo write, deployment or push occurred. The pre-existing `.brooks-lint-history.json` is untouched. The local commit and archive do not authorize ordinary write rollout; the gate still needs a selected disposable fixture and exact approved GPS plan.
