# Testing rules

**Status:** Adopted 17 September 2026. Binding for new AI functionality and affected existing behavior. Required tooling not yet installed is listed below; a documented requirement is not evidence of a passing check.

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

1. Follow the repository's OpenSpec Plus TDD workflow when implementing an OpenSpec change: one failing test observed for the intended reason, minimal production change, green, then assess refactoring before the next test. Every OpenSpec behavior scenario maps to at least one automated test. For documentation-only changes, verify content, references and applicable configuration; do not add artificial application tests.
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

Use one documented local check entry point and the same commands in CI. Pin tool versions and use the committed lockfile. The existing Go Makefile provides test/vet commands, but the new frontend harness, coverage gates, size/import checks and CI wiring still need implementation. Do not claim these checks are active today.

Keep live-provider/Immich integration, quality evaluation and heavy reference-NAS benchmarks separate from ordinary PR checks. Require the applicable evidence before release, and rerun when relevant behavior/version/configuration changes. No ordinary test run sends private data or mutates the user's real library.

## Available commands and remaining setup

The pinned checkout has the following entry points; their existence does not imply they passed in this environment:

| Working directory | Existing command | Purpose |
|---|---|---|
| `backend/` | `make test` | Existing Go test suite |
| `backend/` | `make vet` | Existing Go static checks |
| `backend/` | `go test -race ./...` | Required race-enabled suite when the toolchain is available |
| Repository root | `npm run lint` | Existing ESLint configuration |
| Repository root | `npm run build` | Existing frontend production build |

The [engineering tooling prerequisite](../ai-locate/OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite) must add the shared local/CI entry point, file-size ratchet, dependency checks, frontend test/coverage harness, deterministic application acceptance harness and required CI jobs. It must pin tool versions and document exact commands and coverage scope. Do not invent a passing `npm test` or a coverage result before those entry points exist.

Build the smallest meaningful harness first; add feature scenarios with their owning changes. Import and coverage rules must grow with real packages. An absent AI package has no measured coverage and cannot count as a successful coverage gate. Existing lint/tests/build failures discovered during setup must be reported and resolved or explicitly recorded as inherited limitations, not silently suppressed.

After any implementation change, report commands actually executed, results, applicable scenarios covered and remaining limitations. A missing runtime/toolchain is an unavailable check, not a pass. Live suites require explicitly authorized disposable/synthetic or owner-approved resources; no ordinary test run sends private data or mutates the user's real library.
