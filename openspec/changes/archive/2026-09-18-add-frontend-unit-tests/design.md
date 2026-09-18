## Context

F01 provides one local/CI verification runner, pinned Node.js 22.23.2 and Bun 1.4.2, a frozen frontend lockfile, and source/lint/type/build checks. The frontend remains Next.js 16.1.6 with React 19.2.3 and strict TypeScript. The adopted testing standard already selects Vitest and React Testing Library. There is no AI frontend package or measured AI coverage yet.

The pagination helper and numbered pagination buttons provide small existing pure-logic and accessible interaction boundaries without application contexts or external services. Their current behavior is the characterization contract; this change does not redesign pagination or repair unrelated controls.

This design follows the adopted architecture, testing and coding standards. It introduces development tooling only, with no architecture departure, application API, migration, stored-data ownership or deployment change.

## Goals / Non-Goals

**Goals:**

- Give contributors a reproducible, offline frontend unit/component suite with focused execution and isolated state.
- Demonstrate meaningful regression detection against existing pure logic and accessible component interactions while preserving application behavior.
- Measure frontend coverage including untested new-AI code, enforce the adopted AI line/branch floor when that scope exists, and distinguish absent scope from measured success.
- Make the frontend suite part of the existing shared local/CI command and document its boundaries and actual verification evidence.

**Non-Goals:**

- Browser journeys, async Server Component integration and application smoke testing; these belong to F03.
- AI product implementation, backend coverage tooling, live services or private fixtures.
- Application dependency upgrades, broad legacy refactoring, blanket legacy coverage targets or accessibility changes unrelated to the harness.

## Decisions

### One isolated frontend test environment

Use a single Vitest project with jsdom for pure logic and React component tests. The small common environment keeps initial configuration understandable; its DOM startup cost is acceptable for this scope. Test discovery is explicit for co-located TypeScript test files under the frontend source tree, separate from the existing Node checker tests and future Playwright suite.

Use explicit Vitest imports, React Testing Library, its user-event helpers and jest-dom's Vitest integration. Shared setup owns DOM cleanup after each test. Tests own any additional mocks, temporary state and restoration. React transformation uses the maintained React Vite plugin; Vite's native TypeScript path resolution reuses the existing aliases. This avoids a redundant plugin, a second alias map and a custom module loader.

The development dependency set is Vitest and its V8 coverage provider 5.0.1, Vite 8.3.0, the React plugin 6.1.1, jsdom 30.1.0, React Testing Library 16.3.3, DOM Testing Library 10.4.2, jest-dom 7.0.1 and user-event 14.6.7. Node declarations move to the compatible 22.x line (22.20.3). These versions were installed and loaded together in a disposable preflight using the pinned Node runtime and existing React/TypeScript versions. Application framework versions remain unchanged. The repository lockfile is the reproducible installation authority.

### Small, behavior-focused characterization

Co-locate focused tests with the existing pagination helper and component. Exercise the helper's present bounded page-window behavior and the numbered buttons through accessible role/name queries and user interactions, including the disabled loading state. Avoid implementation snapshots, CSS assertions and invented semantics for currently unnamed icon buttons.

These tests characterize known existing behavior against unchanged application code and may pass on their first run. The user explicitly chose this approach over controlled regressions in disposable copies; the adopted testing standard records that distinction. Use meaningful inputs, interactions and expected outcomes without manufacturing a failure or requiring mutation testing. Report the tests as characterization coverage, not newly implemented behavior or a completed RED/GREEN cycle. New verification-runner behavior follows the normal one-test-at-a-time RED/GREEN/REFACTOR cycle.

### Native coverage policy

Use Vitest's V8 provider with explicit frontend source inclusion, so unimported files remain visible. Exclude test/setup code and declaration-only/generated inputs by their genuine role; do not exclude application files simply because they are difficult to cover. Report actual legacy coverage without a new legacy percentage floor.

Apply native glob thresholds to the new AI frontend scope under `src/features/ai/`: at least 80% lines and 80% branches. An isolated preflight verified that an untested synthetic AI file fails both thresholds and that covering its branches passes. With no AI files, there is no AI measurement; neither a successful harness run nor an aggregate legacy percentage is an AI coverage pass. Documentation and completion reports state this limitation explicitly. No custom coverage parser or synthetic AI package is needed.

Focused runs are for development feedback. The full, unfiltered suite with coverage is the shared verification gate; a selected test file cannot establish complete coverage. Keep reports under the existing ignored coverage output.

### Shared execution and failure reporting

Expose a non-watch frontend unit command and a separate full coverage command. Extend the existing verification gate catalog with a frontend test gate that invokes the same full Vitest coverage run through the selected Node runtime. The existing CI workflow already uses the shared runner and frozen installation; keep that single integration path rather than adding a competing test job.

A missing tool, empty discovered suite, focused test, assertion failure, coverage failure or terminated process must produce a failing command. Disable retries and focused-only tests. Do not introduce skipped tests, relax lint/type checks, suppress inherited warnings or convert test failures into successful output. Runner regression tests establish invocation and failure propagation; real harness executions establish configuration validity.

The suite is offline and uses synthetic inputs, with no provider, Immich, map-tile or credential dependency. Only development/build outputs change on disk. DOM and mock cleanup prevent cross-test state leakage. Neither the harness nor its verification fixtures own application persistence or remote mutations.

### Verification and documentation ownership

The testing standard owns usable commands, coverage interpretation and remaining limits. Update its status, the OpenSpec context, related engineering status references and the roadmap together after verification; retain F03 and backend AI coverage as pending. Append actual evidence to the engineering baseline without rewriting historical results.

Validation uses the new focused suite and coverage checks, meaningful runner regressions, size/dependency checks, existing lint/types, and the full shared command including Go race tests and both builds. Confirm the frozen lockfile installs with the pinned toolchain. Apply the required GitNexus impact/change analysis and cycle checks; graph uncertainty remains a reported limitation verified against source and tests. No remote CI, live compatibility or application journey pass is implied by a local run.

## Risks / Trade-offs

- jsdom does not prove browser rendering, Next.js proxy/auth integration or async Server Component behavior. F03 owns those application boundaries.
- Initial legacy coverage will be small. Its purpose is a working regression harness and truthful measurement; feature changes must add their own scenarios and satisfy the adopted AI floor.
- New test dependencies add a substantial development-only tree. Pin compatible versions, retain the frozen lockfile and verify imports, lint, types and builds together before adoption.
- Already-correct code can make a newly written characterization test immediately green. Assert the known user-visible contract with meaningful inputs and interactions; preserve application behavior and report the test history honestly.
- Native coverage thresholds cannot turn an absent package into evidence. Explicit status reporting preserves the distinction until AI code exists.
