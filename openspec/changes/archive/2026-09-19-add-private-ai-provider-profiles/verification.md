# CH01 verification — 19 September 2026

Implemented against `b7c470f8a9c4610706c98195592f239353233fe2` on `codex/engineering-foundation`. Planning, implementation, tests, spec review and quality review were performed inline in this task, following the user's no-subagent instruction. Scope is private configuration only; no provider transport or AI analysis is installed.

## Scenario evidence

| Spec scenario | Automated evidence |
|---|---|
| Create and reopen a private profile | `TestAIProviderCreateListAndReopen`: real file-backed SQLite, ciphertext validation, reopen, redacted serialization and second-user isolation |
| Edit current/stale revisions | `TestAIProviderRevisionAndCredentialLifecycle`, `TestAIProviderConcurrentEditsHaveOneWinner`: immutable prior configuration, exact CAS revision, one competing winner |
| Disable a logical profile | Lifecycle test checks logical disablement joined across historical versions; owner-isolation test preserves another user's same-ID profile; HTTP journey and UI/browser tests confirm authoritative disabled state |
| Another user references a profile | Reopen/list isolation and HTTP journey: other-owner and missing IDs return the same 404 without disclosure |
| Retain or replace a credential | Lifecycle test decrypts each retained/replacement ciphertext and checks public `hasSecret` only |
| Change destination without treatment | `TestAIProviderFailedEditsPreserveRevision`: rejected destination edit leaves active revision and version count unchanged; UI explains explicit replacement/removal |
| Remove credentials/account deletion | Lifecycle test clears every version's secret, preserves unrelated profile material, then verifies owning-account cascades; `TestAIProviderIsolationUsesOwnerForSameProfileID` checks another owner's revision and decryptable key survive disablement, erasure and deletion |
| Reuse corrupt/plaintext material | Failed-edit test verifies fail-closed reuse and no extra revision |
| Disabled installation preserves workflows | Default-off configuration test, disabled listing/mutation tests and three real browser legacy journeys with `AI_ENABLED=false` |
| Reject unauthenticated/invalid origins | HTTP protection/failure tables cover missing/invalid/expired sessions, missing/null/foreign/duplicate origins, with no persisted mutation |
| Validate bounded configuration; save offline | Pure validation, strict HTTP/size/content-type/unknown-field tests; HTTP journey's local provider counts exactly zero requests; browser journey confirms zero Immich writes |
| Create/edit/disable through Settings | `ProviderSettings.test.tsx` plus `tests/e2e/providers.spec.ts`, using real proxy, session, backend and SQLite |
| Cancel/fail safely | UI test discards cancelled/closed keys, clears submitted keys, blocks pending/duplicate submits and preserves non-secret edits on conflict without automatic retry |
| Disabled/empty/unavailable UI | UI test verifies loading, failure, explicit reload, disabled and empty states |

Additional persistence checks cover fresh migration 018, upgrade from version 17 without catalog loss, repeated migration, per-connection foreign keys/busy timeout, relative data directories containing URI-reserved characters, and an injected SQLite version-insert failure that rolls back both revision advancement and credential erasure. Coverage checker tests include uncovered AI adapters, below-floor/missing/malformed measurements, failed/terminated test processes and missing fresh output.

## Verification results

All thirteen required gates have passing results across the full run and focused corrections:

- 69 checker tests; source size/ratchet; dependency checks; Go formatting; ESLint; route type generation; TypeScript; Go vet.
- Complete race-enabled backend suite with all-package instrumentation: root package 78.410 seconds, provider core 1.390 seconds. AI Go statement coverage **91.56% (141/154)**, against an 80% floor. Legacy coverage is measured without an invented floor.
- 12 frontend tests; AI coverage **100% lines (86/86)** and **96.43% branches (81/84)**, including the complete AI source inventory. Whole-app legacy coverage is much lower and is not represented by these AI percentages.
- Both production builds; three default-off legacy browser journeys (8.5 seconds) and one enabled provider journey (2.9 seconds). No browser retries or skipped tests.
- Backend Docker image `immich-places-backend:ch01-check` built successfully, including `internal/`; Compose validated with synthetic values. No image was published.
- Strict OpenSpec validation for the change and maintained capability; delta/main requirements match; Git whitespace checks pass.

The first full run failed an old smoke-runner test assumption (it still expected the direct Playwright command) and lacked this machine's custom Chromium path. The checker was updated to exercise the new wrapper, then all 69 checker tests passed. Both browser modes passed with `PLAYWRIGHT_BROWSERS_PATH=/tmp/immich-ai-playwright-browsers`. After the final markup readability cleanup, scoped lint, full frontend coverage, size, production builds and both browser modes passed again. Successful backend checks were not needlessly repeated.

Three inherited ESLint warnings, the existing Next.js middleware deprecation and runtime dependency/color warnings remain visible. They are not new failures or suppressed checks. Remote CI execution is not claimed.

## Inline review and graph evidence

Spec compliance review traced every scenario above to executable evidence. Quality review inspected the complete new source and integration diff: user predicates on reads/joins/CAS, transaction ordering, encrypted-only credential reuse, erasure rollback, default-off routing, strict origin/body checks, redacted DTO/errors, frontend mutation behavior, package imports and deployment. UI markup was expanded for readability; the cohesive form/dialog render functions need no additional abstraction solely to shorten them. No blocking findings remain.

Pre-edit GitNexus impact covered existing integration points. `SettingsPanel` and `FilterBar` were HIGH risk and called out before their edits; their actual changes are the AI public entry point and accessible Settings label. The runner, configuration loader, database constructor and session wrapper had narrower impact, verified against source. Unresolved test/configuration references were checked through their actual test runner or startup paths.

The pre-archive implementation snapshot contains 5,548 nodes and 13,626 edges. Full structured `detect_changes` (not the CLI's capped presentation) reports 247 changed symbols across 56 files and 38 affected processes, risk **CRITICAL**, with no partial/truncated result. This large impact includes new declarations/document symbols plus startup, auth, shared parsing and UI paths. Review checked the complete list; the full backend suite, both builds and legacy browser journeys cover the affected integrations. New provider callbacks call only validation, persistence, encryption and JSON helpers; they have no provider/Immich client. Source import checks pass. The cycle enumeration is complete with zero cycles. Final archive/evidence files and the additional same-ID owner-isolation test are included in the pre-commit refresh; that added test also passed with the race detector.

Global execution-flow extraction itself is budget-limited, and cross-language field links are unresolved. Route/shape tools find no `/ai/providers` route in this Go/proxy setup, so those empty results are **not** a passed API-shape check. The actual Go route registrations, frontend DTO guards, API contract tests and real proxy browser journey provide the cross-language evidence. PDG/taint analysis is not claimed. These graph limitations remain visible rather than being treated as proof of no impact.

## Process and limits

Observed failures before implementation include the missing migration, foreign-key/busy-timeout behavior on the second pooled connection, relative SQLite URI handling, coverage/runner gate behavior and inaccessible Settings button. Initial new-module test attempts also included missing-symbol/module compilation failures; these are recorded as such, not misrepresented as semantic RED assertions for every edge case. Additional tests of already implemented behavior passed directly. No manufactured regression or mutation testing was used.

Local logs are under `/tmp/immich-ai-validation-20260918/`, notably `ch01-full-check.log`, `ch01-checker-final.log`, `ch01-ui-final.log`, `ch01-browser-final.log`, `ch01-docker-build.log`, `ch01-detect-changes.json`, `ch01-graph-cycles.json` and the focused RED/GREEN logs. Generated frontend/Go coverage and browser evidence remain ignored.

The operator guide documents installation flags, encryption-key/backup handling, retained historical credentials, all-revision removal, account cascades and additive rollback. No live Immich/provider compatibility, private-photo processing, model quality, dispatch behavior, forensic storage erasure or deployment is claimed. Those require their later owning changes and explicit live authorization.
