# CH22 verification plan

Status: **implemented, reviewed, verified with synthetic fixtures and synchronized**. The original scenario-to-task plan below is retained for traceability. Executed test names, commands, results and the separate live gate are recorded in [implementation-verification.md](implementation-verification.md). No live Immich mutation was performed.

## Traceability

The proposal links the owning PRD FR/NFR/AC obligations. Every requirement below maps to scenarios; every scenario maps to task IDs and an automated boundary. Full MODIFIED blocks retain prerequisite scenarios intentionally. Preserve the rest of the maintained capability and all predecessor regression suites when this change is synchronized.

| Requirement | Scenario IDs |
|---|---|
| Gate optional metadata independently of standard writing | M01, M02, M03 |
| Preview only selected current shareable metadata | M04, M05, M06 |
| Preserve unrelated and externally changed metadata | M07, M08, M09 |
| Execute and recover the mirror as its own approved step | M10, M11, M12, M13, M14 |
| Retain step audit and authority across lifecycle changes | M15, M16, M17, M18 |
| Preview only a staged exact GPS decision | P01, P02 |
| Use fresh authorized source and GPS before-values | P03, P04, P05, P06 |
| Bind an immutable preview to its exact authority and revision | P07, P08, P09, P10 |
| Show a usable accessible comparison without manual-save side effects | P14, P15, P16 |
| Confirm one immutable plan durably and idempotently | W01, W02, W03, W04 |
| Mutate only the approved single-asset GPS pair | W05, W06, W07, W08 |
| Recheck source baseline and current authority before dispatch | W09, W10, W11 |
| Reconcile ambiguous outcomes before any bounded explicit retry | W14, W15, W16, W17, W18, W27, W28 |
| Verify GPS before local success publication | W19, W20, W21, W22 |
| Retain private audit and respect write lifecycle controls | W23, W24, W25, W26 |
| Review optional metadata visibility and partial success | M19, M20, M21 |
| Edit camera decisions without inventing precision or writable fields | D05, D06, D07 |
| Protect a draft revision while its GPS write can still act | R01, R02 |

| Scenario / planned case | Tasks | Planned automated boundary |
|---|---|---|
| M01 — Keep unsupported mirroring optional | 1.1 | Go capability/selection policy + protected HTTP zero-send fixture |
| M02 — Require deliberate supported selection | 1.1 | Go capability/selection policy + protected HTTP zero-send fixture |
| M03 — Reject implicit or mirror-only admission | 1.1 | Go capability/selection policy + protected HTTP zero-send fixture |
| M04 — Preview minimized reviewed content | 1.2 | Go export policy/bounds + SQLite preview revision + RTL |
| M05 — Reject stale or oversized export | 1.2 | Go export policy/bounds + SQLite preview revision + RTL |
| M06 — Invalidate changed metadata selection | 1.2 | Go export policy/bounds + SQLite preview revision + RTL |
| M07 — Create or replace only the owned namespace | 1.3, 2.2 | Metadata HTTP fixture + semantic comparator + SQLite lineage |
| M08 — Reject unavailable or conflicting namespace | 1.3, 2.2 | Metadata HTTP fixture + semantic comparator + SQLite lineage |
| M09 — Compare semantic readback honestly | 1.3, 2.2 | Metadata HTTP fixture + semantic comparator + SQLite lineage |
| M10 — Preserve standard success when mirror fails | 2.1 | SQLite step ordering + HTTP exact namespace/count + RTL partial status |
| M11 — Recover a lost mirror acknowledgement | 2.2 | Dropped-response/readback HTTP + SQLite local-publication recovery |
| M12 — Retry only a known eligible mirror | 3.1 | SQLite step-generation and expiry races + no standard resend HTTP |
| M13 — Keep unknown or expired work read-only | 3.1 | SQLite step-generation and expiry races + no standard resend HTTP |
| M14 — Block mirror after incomplete standard work | 2.1 | SQLite step ordering + HTTP exact namespace/count + RTL partial status |
| M15 — Upgrade and reopen without inferred mirror authority | 4.1 | Populated migration/reopen + two-owner lifecycle/authority races |
| M16 — Revoke authority between steps | 4.1 | Populated migration/reopen + two-owner lifecycle/authority races |
| M17 — Delete or clean up during mirror work | 4.1 | Populated migration/reopen + two-owner lifecycle/authority races |
| M18 — Separate source support from live verification | 4.2 | Operator mirror-capability admission fixture + documented live gate |
| P01 — Preview one coarse camera decision | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P02 — Reject an incomplete or expanded decision | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P03 — Observe missing partial and zero GPS | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P04 — Detect an intervening GPS change | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P05 — Distinguish image change from unrelated metadata | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P06 — Fail without a fabricated baseline | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P07 — Retain the same plan across reload | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P08 — Race preview creation with an edit | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P09 — Reject plan substitution | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P10 — Display an unchanged GPS pair | 1.2 | Versioned preview SQLite/HTTP regressions with optional namespace values |
| P14 — Review without a map | 3.2 | Accessible preview and private response regression |
| P15 — Preserve manual pending coordinates | 3.2 | Accessible preview and private response regression |
| P16 — Handle late private preview replies | 3.2 | Accessible preview and private response regression |
| W01 — Confirm and recover the approved operation | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W02 — Reconcile a lost confirmation response | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W03 — Reject stale or substituted approval | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W04 — Fail before durable approval | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W05 — Write exact GPS while preserving other fields | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W06 — Preserve valid zero and omit unsupported fields | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W07 — Prevent hidden retries or fallback | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W08 — Reject unsupported compatibility | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W09 — Detect fresh GPS or source conflict | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W10 — Lose authority after confirmation | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W11 — Verify an unchanged plan | 2.1 | Approval/adapter/authority regressions with separate metadata reservation |
| W14 — Recover remote success with lost response | 3.1 | Per-step readback/retry SQLite + request-count regression |
| W15 — Keep unresolved outcomes honest | 3.1 | Per-step readback/retry SQLite + request-count regression |
| W16 — Retry a known completed unsuccessful attempt | 3.1 | Per-step readback/retry SQLite + request-count regression |
| W17 — Reject exhausted stale or concurrent retries | 3.1 | Per-step readback/retry SQLite + request-count regression |
| W18 — Detect a different readback value | 3.1 | Per-step readback/retry SQLite + request-count regression |
| W27 — Recover a retry after its successful execution | 3.1 | Per-step readback/retry SQLite + request-count regression |
| W28 — Recover an accepted retry after restart and disablement | 3.1 | Per-step readback/retry SQLite + request-count regression |
| W19 — Observe rounded GPS or mismatched success | 2.2 | Standard refresh preserved through mirror readback/partial outcomes |
| W20 — Recover local failure after remote success | 2.2 | Standard refresh preserved through mirror readback/partial outcomes |
| W21 — Preserve verified GPS against an older sync | 2.2 | Standard refresh preserved through mirror readback/partial outcomes |
| W22 — Handle an unavailable local catalog row | 2.2 | Standard refresh preserved through mirror readback/partial outcomes |
| W23 — Disable while work is active | 4.1 | Step-aware lifecycle/private audit regression |
| W24 — Preserve audit through restart and cleanup | 4.1 | Step-aware lifecycle/private audit regression |
| W25 — Delete or rotate during an operation | 4.1 | Step-aware lifecycle/private audit regression |
| W26 — Keep unverified live writes disabled | 4.2 | Operator mirror-capability fixture + live evidence boundary |
| M19 — Inspect disclosure and exact export accessibly | 3.2, 1.2 | Vitest/RTL + built optional mirror partial-success journey |
| M20 — Review a partial result without losing local data | 3.2, 1.2 | Vitest/RTL + built optional mirror partial-success journey |
| M21 — Fence stale private mirror responses | 3.2, 1.2 | Vitest/RTL + built optional mirror partial-success journey |
| D05 — Choose a coarse or ambiguous proposal | 1.2 | Draft selection/geometry regression with optional mirror fields |
| D06 — Supply an absent camera point | 1.2 | Draft selection/geometry regression with optional mirror fields |
| D07 — Validate geometry and field scope | 1.2 | Draft selection/geometry regression with optional mirror fields |
| R01 — Edit before dispatch reservation | 4.1 | Real SQLite revision protection across both steps |
| R02 — Edit while a write is unresolved | 4.1 | Real SQLite revision protection across both steps |

## Fixtures and execution discipline

Custom metadata list reads, absent/foreign/edited namespace, unrelated keys, object-key reordering, invalid/oversized values, standard success followed by metadata failure, lost metadata acknowledgement and step-specific recovery. Add these fixture capabilities during implementation; there is no existing live mirror harness claim.

Use OpenSpec Plus TDD for changed behavior, one meaningful failing assertion before its production change. Existing-behavior characterization is allowed to pass immediately against unchanged code; never manufacture a failure. Use injected clocks and explicit concurrency barriers. Persistence, isolation, migration, audit and transaction claims require real temporary file-backed SQLite and normal Goose migrations, including reopen. Provider/Immich call assertions use local HTTP servers and the real serialization/handler path. No retries hide failing browser tests.

Each vertical slice includes its mapped policy, persistence, API and UI checks. The final apply verification uses `bun run check --base <actual-implementation-base>` with Node 22.23.2, Go 1.25.14 and Bun 1.4.2: checker tests, size/import rules, gofmt, lint, type generation/types, Go vet/race tests, frontend unit coverage, both builds and built-browser acceptance. Preserve 80% Go AI statements and 80% AI TypeScript lines/branches. Run relevant container builds when build/package/deployment inputs change. Extend affected import rules rather than omitting new AI adapters from coverage.

Run pre-edit GitNexus impact on each changed symbol and pre-commit `detect_changes` against the correct checkout. UNKNOWN/partial/lower-bound results need source/test confirmation; they are not clean impact evidence. Check strict OpenSpec validation, ordered delta composition, scenario/test mapping, local links and whitespace before completing the change. Sync only after implementation and verification; the original planning phase left maintained specs unchanged.

Retain the existing legacy auth/manual/GPX journeys and AI-disabled regression, plus an affected built AI journey. Keep inspection/acceptance at zero Immich writes and preserve CH19 confirmation identity, late-history, retry-generation and local-refresh regressions throughout. Populate upgrade fixtures with queued, completed and unresolved old state; an empty-database migration test is not enough. No test deletes audit to simplify recovery.

## Live and rollout boundary

Ordinary tests use synthetic data and deterministic local services. Live provider text compatibility, Immich field/scope/metadata compatibility, language quality and performance require separately recorded evidence; an offline pass does not establish them. CH19 live GATE-02 remains pending after synthetic verification. New field/scope capabilities remain default-off and require their own authorized version/profile/rights/readback evidence. Document the precise tested behavior and sidecar/UI limits; do not reuse a GPS pass as a metadata or stack pass.

The original request authorized planning; subsequent instructions authorized sequential implementation and synthetic verification. These do not authorize private text/image dispatch, live writer activation, deployment, database restoration, GPS clearing or mutation of the selected CH19 fixture. A live gate must present the exact proposed operation and use the normal authenticated durable confirmation path. Preserve backups, encryption keys and unresolved guards during any later authorized rollout.
