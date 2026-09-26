# CH17 verification plan

Status: **implemented and verified**; see [actual results](implementation-verification.md). The plan below records the original preparation scope. No application behavior, database migration, provider call or live Immich mutation was performed while preparing this change. Scenario IDs below are planned automated cases, not claims of existing test names or passing results. Implementation records the actual test names, commands and outcomes after execution.

## Traceability

The proposal links the owning PRD FR/NFR/AC obligations. Every requirement below maps to scenarios; every scenario maps to task IDs and an automated boundary. Full MODIFIED blocks retain prerequisite scenarios intentionally. Preserve the rest of the maintained capability and all predecessor regression suites when this change is synchronized.

| Requirement | Scenario IDs |
|---|---|
| Explicitly authorize translation of reviewed facts | T01, T02, T03 |
| Preserve bounded private translation runs | T04, T05, T06 |
| Retain independent language outcomes | T07, T08, T09 |
| Adopt translations only through current draft review | T10, T11, T12 |
| Retry and cancel translation without hidden regeneration | T13, T14, T15 |
| Isolate translation history and lifecycle | T16, T17, T18 |
| Make translation review accessible and explicit | T19, T20, T21 |

| Scenario / planned case | Tasks | Planned automated boundary |
|---|---|---|
| T01 — Generate from corrected facts | 1.1 | Go policy + protected HTTP + provider payload fixture |
| T02 — Reject absent consent or invalid input | 1.1 | Go policy + protected HTTP + provider payload fixture |
| T03 — Retain scene-only and offline review | 1.1 | Go policy + protected HTTP + provider payload fixture |
| T04 — Recover a lost submission acknowledgement | 1.1, 4.1 | Real SQLite admission/reopen + HTTP identity recovery |
| T05 — Reject substitution and concurrent admission | 1.1, 4.1 | Real SQLite admission/reopen + HTTP identity recovery |
| T06 — Enforce execution budgets | 1.2, 4.2 | Go budget policy + bounded concurrent provider fixture |
| T07 — Keep partial language success | 1.2, 2.1 | Provider codec/HTTP + SQLite language outcomes |
| T08 — Reject invalid translation output | 1.2, 2.1 | Provider codec/HTTP + SQLite language outcomes |
| T09 — Reload retained outcomes | 2.1, 4.1 | SQLite reopen + RTL retained private history |
| T10 — Adopt selected suggestions | 3.1 | Go draft policy + real SQLite competing edits/write reservation |
| T11 — Reject stale or concurrently edited adoption | 3.1 | Go draft policy + real SQLite competing edits/write reservation |
| T12 — Respect active write authority | 3.1 | Go draft policy + real SQLite competing edits/write reservation |
| T13 — Retry only the chosen unsuccessful language | 2.2, 1.2 | SQLite run recovery + provider call counts/cancellation barriers |
| T14 — Cancel or restart an active run | 2.2, 1.2 | SQLite run recovery + provider call counts/cancellation barriers |
| T15 — Invalidate provider authority before dispatch | 2.2, 1.2 | SQLite run recovery + provider call counts/cancellation barriers |
| T16 — Deny foreign translation operations | 4.1 | Two-owner SQLite/HTTP lifecycle + log redaction fixtures |
| T17 — Delete during a provider call | 4.1 | Two-owner SQLite/HTTP lifecycle + log redaction fixtures |
| T18 — Preserve retained runs without private log output | 4.1 | Two-owner SQLite/HTTP lifecycle + log redaction fixtures |
| T19 — Complete keyboard translation review | 3.2, 2.1 | Vitest/RTL + built Playwright translation journey |
| T20 — Fence private late replies | 3.2, 2.1 | Vitest/RTL + built Playwright translation journey |
| T21 — Inspect without dispatch | 3.2, 2.1 | Vitest/RTL + built Playwright translation journey |

## Fixtures and execution discipline

Text-only provider replies for two languages, refusal, wrong language, malformed/oversized JSON, delayed response, cancellation and unknown completion; populated draft/translation databases and a shared analysis/translation concurrency fixture. These translation-specific cases must be added during implementation; the general HTTP/SQLite/RTL/browser harnesses already exist.

Use OpenSpec Plus TDD for changed behavior, one meaningful failing assertion before its production change. Existing-behavior characterization is allowed to pass immediately against unchanged code; never manufacture a failure. Use injected clocks and explicit concurrency barriers. Persistence, isolation, migration, audit and transaction claims require real temporary file-backed SQLite and normal Goose migrations, including reopen. Provider/Immich call assertions use local HTTP servers and the real serialization/handler path. No retries hide failing browser tests.

Each vertical slice includes its mapped policy, persistence, API and UI checks. The final apply verification uses `bun run check --base <actual-implementation-base>` with Node 22.23.2, Go 1.25.14 and Bun 1.4.2: checker tests, size/import rules, gofmt, lint, type generation/types, Go vet/race tests, frontend unit coverage, both builds and built-browser acceptance. Preserve 80% Go AI statements and 80% AI TypeScript lines/branches. Run relevant container builds when build/package/deployment inputs change. Extend affected import rules rather than omitting new AI adapters from coverage.

Run pre-edit GitNexus impact on each changed symbol and pre-commit `detect_changes` against the correct checkout. UNKNOWN/partial/lower-bound results need source/test confirmation; they are not clean impact evidence. Check strict OpenSpec validation, ordered delta composition, scenario/test mapping, local links and whitespace before completing the change. Sync only after implementation and verification; this planning phase leaves maintained specs unchanged.

Retain the existing legacy auth/manual/GPX journeys and AI-disabled regression, plus an affected built AI journey. Keep inspection/acceptance at zero Immich writes and preserve CH19 confirmation identity, late-history, retry-generation and local-refresh regressions throughout. Populate upgrade fixtures with queued, completed and unresolved old state; an empty-database migration test is not enough. No test deletes audit to simplify recovery.

## Live and rollout boundary

Ordinary tests use synthetic data and deterministic local services. Live provider text compatibility, Immich field/scope/metadata compatibility, language quality and performance require separately recorded evidence; an offline pass does not establish them. CH19 live GATE-02 is still pending while this batch is planned. New field/scope capabilities remain default-off and require their own authorized version/profile/rights/readback evidence. Document the precise tested behavior and sidecar/UI limits; do not reuse a GPS pass as a metadata or stack pass.

The user request here authorizes planning only. Do not dispatch private text/images, activate a writer, deploy, restore a database, clear GPS or mutate the selected CH19 fixture as part of this plan. A live gate must present the exact proposed operation and use the normal authenticated durable confirmation path. Preserve backups, encryption keys and unresolved guards during any later authorized rollout.
