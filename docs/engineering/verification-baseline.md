# Verification baseline

Recorded 18 September 2026 before enforcement implementation.

## Checkout and tools

- Planning checkpoint: `4fbb57d172f2c720f984e0f894852fda72417224`, branch `codex/engineering-foundation`.
- Application source is unchanged from `5e70c6165777949c9d8b50ede3b2768bcaa5df87` at this checkpoint.
- Native macOS arm64 verification: Go 1.25.14, Node.js 22.23.2 and Bun 1.4.2. Temporary toolchain archives were verified against publisher checksums/package integrity.
- Dependencies installed with the committed Bun lockfile frozen; neither dependency manifest nor lockfile changed.

## Actual results

| Check | Result |
|---|---|
| Backend `go test -race ./...` | Passed; package test execution reported 77.236 seconds, excluding initial dependency download/compilation |
| Backend `go vet ./...` | Passed |
| Backend `go build` | Passed; output written outside the repository |
| Frontend ESLint | Failed: five errors, three warnings |
| TypeScript `tsc --noEmit` | Passed |
| Frontend production build | Passed with inherited deprecation warnings |
| Go formatting | Five files differ from gofmt output |
| Frontend unit/browser suites | Not available at this baseline |

ESLint errors: an unused `useSelection` import in `useMapViewModel.ts` (three rule reports), an unnecessary nullable-boolean comparison in `PhotoCardMenu.tsx`, and the boolean name `coordinatesChanged` in `selectionStateHelpers.ts`.

Formatting differences: `backend/gpxService.go`, `backend/handlersGPX.go`, `backend/handlersGPX_test.go`, `backend/handlersLibraries_test.go`, and `backend/syncService_test.go`.

Inherited warnings: unused hook-lint suppression in `useDawarich.ts`, image optimization guidance in `HeaderTitle.tsx`, and a hook dependency warning in `useOverviewLayerReconcile.ts`. The production build also reports the Next.js middleware convention and Node.js module registration deprecations. These are recorded findings, not waived future failures or a claim of warning-free output.

## Follow-up and evidence boundary

The first enforcement change owns the focused lint-error and formatting corrections needed for a green initial gate. Feature implementation and broad refactoring have not started at this baseline. The engineering standards remain the source of truth for required future checks.

Raw local logs for this run are under `/tmp/immich-ai-validation-20260918/`; they are temporary execution evidence, not repository fixtures. No live Immich/provider compatibility, browser journeys, container build, performance benchmark or AI quality evaluation was established by these checks.

GitNexus refreshed to the planning checkpoint with 4,838 nodes and 12,177 edges. Its process extraction reported truncation; an absent graph flow is not proof that the path is absent. The pre-checkpoint change analysis reported documentation/tooling files and no affected application symbols.

The checkpoint preserves imported Markdown hard breaks and original skill text; Git's generic whitespace check reported those trailing spaces. No source-code whitespace failure was concealed by that observation.

## F01 enforcement follow-up — 18 September 2026

The earlier sections remain the pre-enforcement record. The implemented F01 command is `bun run check --base 5e70c61`, using the same Node.js 22.23.2, Go 1.25.14 and Bun 1.4.2 toolchains. Frozen Bun installation completed without dependency changes. The full shared command passed all eleven gates after the bounded corrections below:

| Gate | Actual follow-up result |
|---|---|
| Checker regression suite | 59 tests passed, no skipped/todo tests: 28 size, 20 dependency and 11 verification/formatting tests |
| File-size ratchet | Passed for 290 handwritten files against the explicit adoption base |
| Dependencies | Passed: 198 frontend files, 365 runtime edges and 50 backend Go files |
| Go formatting | Passed for 51 Go files, including the parser helper |
| ESLint | Passed with zero errors; the same three inherited warnings remain visible |
| Route types and TypeScript | Next.js `typegen` and `tsc --noEmit` passed; also verified without pre-existing `.next` or `next-env.d.ts` |
| Go vet | Passed |
| Backend race suite | Passed; package execution reported 66.093 seconds |
| Go and frontend builds | Passed; frontend middleware-convention deprecation remains visible |
| CI configuration | Actionlint 1.7.12 passed; workflow base selection checked with explicit, empty/all-zero, invalid and unavailable-parent revisions |

The first full enforcement run failed `TestDoIncrementalSyncFallsBackToFull` during temporary-directory cleanup. Source tracing showed asynchronous frequent-location enrichment could outlive five existing full/incremental sync test fixtures. Those tests now register a wait-group join before mock-server/database cleanup. They were extracted into `backend/syncServiceLifecycle_test.go` with every assertion retained; the original file's size ceiling decreased from 1580 to 1467 lines. The five exact cases then passed one fixed `-race -count=20` run (23.109 seconds), followed by the passing full command. Production sync behavior was unchanged; the failure was not hidden by retries, sleeps or disabled work.

The original five frontend lint errors and five gofmt differences are corrected. The check workflow uses read-only repository permissions on pull requests/pushes, immutable action references and frozen dependency installation; no remote workflow execution or branch-protection configuration is claimed by this local verification. Generated build/index/coverage state and temporary test fixtures are excluded from source control.

Execution evidence remains under `/tmp/immich-ai-validation-20260918/`: `verification-shared-gate-01.log` preserves the failure, `verification-lifecycle-green.log` records the bounded fixture verification, and `verification-shared-gate-02.log` records the passing shared command. The per-test RED/GREEN/refactor record is `verification-tdd.md`. A final test-fixture portability refactor isolates the missing-Go regression from the host's executable paths without changing its assertions; its focused tests and lint passed separately.

The final integration review found inconsistent handling of repeated comparison-base options between focused and shared commands. All three now use one parser and reject repeated options. Two additional CLI regressions passed their individual RED/GREEN/refactor cycles, preserving all existing assertions. The final shared command exited 0 with all eleven gates passing: 61 checker tests (29 size, 21 dependency, 11 verification/formatting), 291 handwritten files and 199 frontend dependency inputs with 368 runtime edges. The unchanged backend race suite reused Go's valid cached result; the fresh 66.093-second run above remains its execution evidence. Final review has no outstanding findings. Evidence: `cli-alignment-tdd.md`, `cli-alignment-cli-matrix.log` and `f01-final-shared-gate.log` in the same temporary log directory.

F02 frontend unit/component coverage and F03 deterministic browser/application journeys remain pending. AI implementation is absent and AI coverage is unmeasured. These results do not establish live Immich/provider compatibility, model quality, container-build compatibility or performance acceptance.

## F02 frontend testing follow-up — 18 September 2026

F02 installs Vitest/React Testing Library with seven known-behavior tests for pagination logic, pointer selection, keyboard activation and loading controls. Application code is unchanged. The user chose ordinary characterization against existing code, so these tests may pass immediately; no mutations or manufactured failures were used. New runner behavior followed test-first development: the existing required-gate test failed because the frontend gate was absent, then passed after adding it. A focused failure-propagation test covers the runner's existing error contract.

The full shared command `bun run check --base fc3f9ae` passed all twelve gates on the pinned toolchains: 62 checker tests, seven frontend tests, the 296-file size check, 203 frontend dependency inputs/370 runtime edges/50 Go files, formatting, lint, route types/TypeScript, Go vet, the cached backend race result and both builds. The three inherited lint warnings and middleware deprecation remain visible. The first cumulative run found ESLint scanning generated coverage-report JavaScript; the exact generated `coverage/` output is now excluded under the existing coding-standard policy. The corrected full command exited 0.

Coverage includes 171 frontend application files, including untested files. Initial legacy coverage is 0.60% lines and 0.68% branches; no legacy floor is imposed. Native thresholds enforce 80% lines and branches for future `src/features/ai/` code. Synthetic configuration preflight verified untested AI failure, covered AI success and absent AI scope; actual AI code is absent and unmeasured. Browser journeys, broad legacy UI coverage, backend AI coverage, live compatibility and model quality remain outside this evidence.

Frozen Bun installation reported no changes. Application dependency declarations and existing resolved package versions are retained, except for the planned development-only Node type upgrade. Vite's native alias resolution avoids an extra plugin. GitNexus reported low impact for the runner and no indexed import cycles; unresolved configuration-constant references and truncated global flow extraction were checked against source, and do not establish absence of other paths. All implementation and final review were performed in this task following the user's no-subagent instruction.

Local evidence is under `/tmp/immich-ai-validation-20260918/`: `f02-slice1-review.md`, `f02-runner-red.log`, `f02-runner-green.log`, `f02-runner-final.log`, `f02-final-shared-gate.log` (initial failure) and `f02-final-shared-gate-02.log` (all gates pass). Reports are generated under ignored `coverage/`. No remote CI execution or deployment is claimed.
