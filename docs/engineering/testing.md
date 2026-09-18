# Testing rules

**Status:** Adopted 17 September 2026. Binding for new AI functionality and affected existing behavior. F01 repository checks, F02 frontend unit/component testing and F03 application smoke tests are installed. Go AI coverage enforcement remains pending. AI code is absent and its coverage is unmeasured. A documented requirement is not evidence of a passing check.

This document owns test strategy, coverage and verification gates. Read it with [architecture](architecture.md), [coding standards](coding-standards.md), and the relevant [PRD scenarios](../ai-locate/PRD.md). Product acceptance and live compatibility remain distinct from ordinary regression tests.

## Test layers

| Layer | Required approach | What it must prove |
|---|---|---|
| Go domain/workflows | Standard Go `testing`, table-driven cases where appropriate, small fakes at I/O seams | Validation, decisions, state transitions, budgets, conflicts and invariants |
| SQLite repositories/migrations | Real temporary file-backed SQLite with the existing driver and real Goose migrations | Actual queries, constraints, tenant scope, transactions, upgrades and persistence after reopen |
| Provider/Immich adapters and handlers | Local fake HTTP servers (`httptest`), real adapter serialization and handler paths | Methods, exact IDs/fields, headers, bounded bodies, failures, timeouts and readback |
| Frontend pure logic/components/hooks | Vitest + React Testing Library | User-visible behavior, accessible interactions, revision/stale states and error rendering |
| Complete application journeys | Playwright against the built frontend and real Go backend; external provider/Immich are deterministic local fakes | Auth/proxy/API/database/UI integration and the required acceptance workflows |
| Live compatibility and model evaluation | Separate opt-in suite using explicitly authorized disposable/synthetic or owner-approved fixtures | Real endpoint/version support, actual model quality and performance that mocks cannot prove |

Vitest with React Testing Library and Playwright are documented in the official [Next.js Vitest guide](https://nextjs.org/docs/app/guides/testing/vitest) and [Playwright guide](https://nextjs.org/docs/app/guides/testing/playwright). The installed Next.js 16.1.6 guidance directs async Server Component verification toward end-to-end tests; do not force unsupported rendering into a unit-test harness.

## Test-writing rules

1. Follow the repository's OpenSpec Plus TDD workflow for new or changed behavior: one failing test observed for the intended reason, minimal production change, green, then assess refactoring before the next test. Tests added for known existing behavior are characterization tests: assert the known user-visible contract against unchanged application code, and allow them to pass on their first run. Do not manufacture a failure, mutate a temporary copy, or require mutation testing for this work. This user-approved distinction takes precedence over a skill's blanket RED-first rule. Every OpenSpec behavior scenario still maps to at least one automated test. For documentation-only changes, verify content, references and applicable configuration; do not add artificial application tests.
2. Tests assert externally meaningful behavior. Multiple assertions are welcome when they describe one outcome. For security and writeback, asserting the absence of a forbidden call or the exact outgoing payload is part of the behavior contract.
3. Mock at nondeterministic/external boundaries. Do not mock the repository when claiming SQL/tenant/transaction correctness, or replace every collaborator with mocks just to test wiring.
4. Fast tests need no internet, credentials, public map tiles or private photographs. Use synthetic fixtures and deterministic local HTTP services. Separate live suites explicitly; “not run” is not a pass.
5. Control time for leases/backoff/expiry. Prefer fake clocks, barriers and channels over sleeps. Isolate databases, files and mutable state; close them through test cleanup. Do not depend on test order or global timezone.
6. Do not introduce skipped/focused tests, ignored failing assertions or automatic retries that conceal flakiness. Investigate a flaky critical-path test before using it as a release gate.
7. Before changing ambiguous inherited behavior, add a focused characterization test. Preserve manual/GPX behavior and document deliberate changes rather than rewriting expected values to make a failure disappear.
8. Co-locate focused tests with their production owner. Apply the file-size rule to handwritten test code too. Avoid huge shared fixtures that hide which inputs matter.
9. Run affected tests during development and the full required gates before merge. A failed or unavailable gate is reported explicitly; do not claim success based on tool invocation alone.

## Coverage policy

Required floor for the new AI scope: at least **80% statement coverage for new AI Go code**, and **80% line and branch coverage for new AI TypeScript code**. Measure the new scope explicitly, including untested files; do not impose an invented passing percentage on the existing repository without measuring it first. Exclusions must identify genuinely generated code or data.

These percentages are minimum checks, not proof of quality. Every specified scenario and safety invariant below must be covered regardless of the percentage. No meaningless assertions, tests that mirror implementation, or broad snapshots solely to raise coverage. Report which failure paths remain unverified.

The frontend coverage command uses Vitest's V8 provider and includes untested TypeScript application files. Native thresholds enforce 80% lines and branches for `src/features/ai/` when that code exists. Tests and declaration-only files are excluded; legacy application files remain measured without an invented percentage floor. No AI files means no AI measurement, even when the harness succeeds. Filtered development runs do not establish complete coverage; the shared gate always runs the full suite.

## Mandatory acceptance/failure scenarios

- One user cannot access another user's provider, job, image/context, result, draft or write operation.
- Accepting/editing/retranslating a proposal causes zero Immich mutations.
- Every mutation uses exactly the approved assets, fields and revision; stale approval and changed stack membership cannot expand it.
- Existing GPS/descriptions changed by another client produce a conflict; append is repeat-safe.
- A lost response or local persistence failure after upstream success triggers reconciliation before resend; optional metadata failure preserves standard-field success.
- Restart, lease expiry, cancellation, provider disablement and revoked access prevent stale/unapproved dispatch or result publication.
- Missing/partial GPS, zero coordinates, offset-boundary and offsetless timestamps, undated assets and reversed ranges follow the documented rules.
- Unknown/ambiguous/malformed/refused/truncated model output cannot manufacture a camera point, heading, source or writable authority.
- Location corrections invalidate dependent direction/descriptions and previous approvals; failed translations remain visible and retry independently.
- Context/image payloads obey consent and size rules; disallowed hosts, redirects and oversized images fail without leaking credentials or private data.
- AI disabled preserves existing browsing, manual placement and GPX workflows; completed results remain distinct from current Missing GPS membership.
- Migration/reopen, retained audit, account deletion and export/deletion policy work on real SQLite fixtures.

Use the [Go race detector](https://go.dev/doc/articles/race_detector) on the backend suite and exercise competing claims/writes deliberately. A clean race run only covers executed paths; it does not prove the state machine or database protocol correct by itself.

## Verification gates and tooling status

| Check | Scope |
|---|---|
| File size and baseline ratchet | All tracked handwritten source/test files plus new files; validate explicit exclusions |
| Import/dependency rules | New AI core cannot import I/O adapters or writer from analysis; frontend feature boundaries have no cycles |
| Format/lint/type checks | Go formatting and `go vet`; existing ESLint rules plus selected AI rules; TypeScript typecheck and frontend formatting |
| Backend suite | From `backend`: `go test -race ./...`, including real SQLite fixtures and handler/adapter tests |
| Frontend suite and coverage | Vitest/RTL with coverage for the new AI scope |
| Build | Frontend production build and Go build; verify container builds when build context, packages or deployment files change |
| Acceptance journeys | Small deterministic Playwright suite for the affected completed workflows, expanded as capabilities arrive |
| OpenSpec checks | Structural validation plus requirement/scenario/test traceability; neither replaces application tests |

`bun run check` is the shared local/CI entry point. Its thirteen gates run checker regression tests, the size ratchet, source-based dependency checks, gofmt, existing ESLint, Next.js route type generation, TypeScript, Go vet, race-enabled Go tests, frontend unit/component tests with coverage, both application builds and browser smoke tests. Every invoked gate reports `RUN` and `PASS` or `FAIL` with its error; a missing tool, nonzero exit or termination fails the command. Failed route generation blocks TypeScript checking; either failed application build blocks smoke execution. Independent gates still run so their results remain visible.

The workflow in [checks.yml](../../.github/workflows/checks.yml) runs these commands on pull requests and pushes with read-only repository permissions and immutable action references. It does not publish images or configure remote branch protection. Existing release publishing is unchanged. CI installs the pinned Playwright Chromium runtime and retains browser failure reports for seven days. Go AI coverage enforcement remains pending.

Keep live-provider/Immich integration, quality evaluation and heavy reference-NAS benchmarks separate from ordinary PR checks. Require the applicable evidence before release, and rerun when relevant behavior/version/configuration changes. No ordinary test run sends private data or mutates the user's real library.

## Available commands and remaining setup

Use **Node.js 22.23.2, Go 1.25.14 and Bun 1.4.2**, the versions pinned in CI and verified locally. Put their binaries on `PATH`, install frontend dependencies with `bun install --frozen-lockfile`, and run `go mod download` from `backend/`. Keep `bun.lock`, `backend/go.mod` and `backend/go.sum` unchanged. CI checks that installation did not modify them; the Go vet/test/build gates use `-mod=readonly` and CI sets `GOTOOLCHAIN=local`.

Run these commands from the repository root. npm can run the same scripts, for example `npm run check -- --base 5e70c61`; Bun supplies the frozen dependency installation.

| Command | Purpose |
|---|---|
| `bun run check --base 5e70c61` | All thirteen installed gates against the initial adoption base |
| `bun run check` | Same gates using the available main branch's merge base |
| `bun run check --help` | List gate names and arguments |
| `bun run check --gate go-tests --base HEAD` | Only the existing backend race suite |
| `bun run check:tools --base HEAD` | Node built-in checker regression suite |
| `bun run check:size --base 5e70c61` | Handwritten file sizes and historical exception ratchet |
| `bun run check:dependencies --base 5e70c61` | Runtime cycles, local import resolution and AI structural boundaries |
| `bun run check:format` | Read-only gofmt check of tracked and non-ignored new Go files |
| `bun run check:types --base HEAD` | Next.js `typegen`, then `tsc --noEmit`, including a clean checkout |
| `bun run test:unit` | Offline Vitest/RTL suite, without watch mode |
| `bun run test:unit src/shared/components/PaginationFooter.test.tsx` | Focused component tests for development feedback |
| `bun run test:unit:coverage` | Full frontend suite with text, HTML and JSON summary coverage reports |
| `bun run check --gate frontend-tests --base HEAD` | The same full frontend coverage run through the shared runner |
| `node node_modules/@playwright/test/cli.js install --with-deps chromium` | Install the pinned browser and OS dependencies; run once after dependency installation or a Playwright upgrade |
| `bun run test:smoke --base HEAD` | Build both applications, then run all three browser smoke journeys |
| `bun run check --gate smoke --base HEAD` | Same build prerequisites and browser suite through shared verification |
| `node node_modules/@playwright/test/cli.js test manual.spec.ts` | Focused browser development run against already-built current artifacts |
| `bun run lint` / `bun run build` | Existing ESLint and Next.js production build |

Any gate can be selected with `--gate`: `checker-tests`, `size`, `dependencies`, `gofmt`, `lint`, `typegen`, `types`, `go-vet`, `go-tests`, `go-build`, `frontend-tests`, `frontend-build`, `smoke`. Focused checks do not establish a full verification pass. Backend `make test`, `make vet` and `go test -race ./...` remain available.

Frontend tests are co-located as `src/**/*.test.ts` or `.test.tsx`, use explicit Vitest imports, and run in isolated jsdom environments with DOM cleanup after each test. The initial seven tests cover pagination boundaries and numbered-button pointer, keyboard and loading behavior. Empty suites, focused-only tests, assertion failures and coverage failures fail the command; retries are disabled. Full browser journeys and proxy/auth integration use the separate smoke suite.

### Browser smoke environment

The three `tests/e2e/*.spec.ts` journeys exercise authentication/browsing, manual placement and GPX import. Playwright runs Chromium with one worker and no retries against the production Next.js standalone output, a compiled Go backend, normal migrations and a fresh temporary SQLite directory. Each test owns its synthetic account/key. Loopback ports 3080 (frontend), 8089 (backend) and 8090 (fake Immich) must be free; existing servers are never reused. Startup is bounded to 30 seconds per server, tests to 30 seconds each and the suite to three minutes. Graceful shutdown has an eight-second fallback; normal teardown removes temporary data.

Browser map tiles are synthetic. Unexpected browser destinations fail the test, and a local deny proxy prevents backend external calls. Known optional Nominatim requests after a save/reload are explicitly denied to exercise the existing offline label fallback and recorded in fixture evidence. Unknown external destinations and unknown Immich fixture routes fail verification. Application APIs, auth cookies, migrations, GPX matching and confirmed location writes remain real. No private photos, credentials or live library are used.

Failed runs retain screenshots and traces under `test-results/` and an HTML report under `playwright-report/`; synthetic upstream writes and blocked-geocoding attempts are attached as JSON. Both directories are generated by Playwright and excluded from handwritten-source checks and Git. Inspect the report with `node node_modules/@playwright/test/cli.js show-report`. A direct Playwright invocation does not rebuild artifacts; use the shared smoke command for complete verification. Browser installation needs network access; test execution uses local services only. An uninstalled browser fails visibly.

These checks characterize existing non-AI behavior, exact unstacked asset writes and persisted readback after browser reload. They do not establish database recovery after process restart, all stack semantics, live compatibility, AI flag enforcement or AI acceptance scenarios; add those as their capabilities arrive.

### Comparison bases

Use an explicit base for review and reproducibility. `--base` must resolve to an available Git commit. Without it, the checker tries the merge base with `origin/main`, then local `main`; if neither exists it fails and asks for `--base`. Use `5e70c61` for this initial adoption review, or the actual target/base commit for subsequent changes. `HEAD` is useful for focused execution but does not compare already committed branch changes with their review target.

CI fetches complete history. Pull requests use the event's base SHA; pushes use the event's previous SHA. When a new branch push has an empty/all-zero previous SHA, CI uses `HEAD^`, a real parent commit. If that parent or the supplied base is unavailable, verification fails rather than substituting the current tree. A force-push whose previous revision cannot be fetched therefore needs its comparison history restored before it can pass.

### Formatting and enforcement boundaries

`check:format` never writes files. Correct reported Go paths with `gofmt -w <paths>` and rerun it. Existing ESLint rules enforce the frontend's formatting/naming conventions; use a scoped `eslint --fix <paths>` for mechanical corrections and review the diff. The gate retains three inherited warnings recorded in the [historical baseline](verification-baseline.md); it does not suppress them or relax application lint rules.

Go build output is under ignored `out/checks/`; Next.js output, `next-env.d.ts`, TypeScript build metadata, coverage output, `.gitnexus/` and the local OpenSpec Plus update timestamp are ignored. Temporary regression fixtures are cleaned up. Generated output must not be committed; an existing tracked `backend/coverage.out` remains explicitly classified by its producer in the coding standard.

The [engineering tooling prerequisite](../ai-locate/OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite) records the installed F01–F03 harnesses. Backend AI coverage enforcement must arrive with the first AI Go code. Use the documented commands and recorded results; no generic `npm test` or unexecuted verification pass is implied.

Build the smallest meaningful harness first; add feature scenarios with their owning changes. Import and coverage rules must grow with real packages. An absent AI package has no measured coverage and cannot count as a successful coverage gate. Existing lint/tests/build failures discovered during setup must be reported and resolved or explicitly recorded as inherited limitations, not silently suppressed.

After any implementation change, report commands actually executed, results, applicable scenarios covered and remaining limitations. A missing runtime/toolchain is an unavailable check, not a pass. Live suites require explicitly authorized disposable/synthetic or owner-approved resources; no ordinary test run sends private data or mutates the user's real library.
