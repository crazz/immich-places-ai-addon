# CH21 verification plan

Status: **planned, not executed**. No application behavior, database migration, provider call or live Immich mutation was performed while preparing this change. Scenario IDs below are planned automated cases, not claims of existing test names or passing results. Implementation records the actual test names, commands and outcomes after execution.

## Traceability

The proposal links the owning PRD FR/NFR/AC obligations. Every requirement below maps to scenarios; every scenario maps to task IDs and an automated boundary. Full MODIFIED blocks retain prerequisite scenarios intentionally. Preserve the rest of the maintained capability and all predecessor regression suites when this change is synchronized.

| Requirement | Scenario IDs |
|---|---|
| Freeze only explicitly reviewed stack GPS targets | S01, S02, S03, S04, S20 |
| Bind independent target baselines and exact field scope | S05, S06, S07 |
| Reserve and recover each exact target without overlap | S08, S09, S10 |
| Preserve partial completion through expiry and lifecycle changes | S11, S12, S13, S14 |
| Preserve earlier approvals when adding stack support | S15, S16 |
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
| Review an exact stack manifest and per-target outcomes | S17, S18, S19 |
| Edit camera decisions without inventing precision or writable fields | D05, D06, D07 |
| Protect a draft revision while its GPS write can still act | R01, R02 |

| Scenario / planned case | Tasks | Planned automated boundary |
|---|---|---|
| S01 — Select an exact subset | 1.1 | Go target policy + bounded fresh stack HTTP + private SQLite observation |
| S02 — Retain the single-photo default | 1.1 | Go target policy + bounded fresh stack HTTP + private SQLite observation |
| S03 — Reject inaccessible or incomplete scope | 1.1 | Go target policy + bounded fresh stack HTTP + private SQLite observation |
| S04 — Bound and deduplicate target selection | 1.1 | Go target policy + bounded fresh stack HTTP + private SQLite observation |
| S20 — Reject stack scope without a GPS decision | 1.1 | Go target policy + bounded fresh stack HTTP + private SQLite observation |
| S05 — Preserve sibling descriptions and unrelated fields | 1.2 | Membership/source-race HTTP fixture + immutable SQLite target plans |
| S06 — Observe changed membership | 1.2 | Membership/source-race HTTP fixture + immutable SQLite target plans |
| S07 — Stop a changed target independently | 1.2 | Membership/source-race HTTP fixture + immutable SQLite target plans |
| S08 — Reject an overlapping target atomically | 2.1 | Real SQLite competing overlapping manifests across owners/versions |
| S09 — Recover independent target outcomes | 2.2 | SQLite per-target recovery/replay + exact outgoing request counts |
| S10 — Retry only an eligible exact target | 2.2 | SQLite per-target recovery/replay + exact outgoing request counts |
| S11 — Finish partially at approval expiry | 3.1 | Clock/barrier-driven SQLite expiry/edit/deletion/disablement races |
| S12 — Edit or disable while a member can act | 3.1 | Clock/barrier-driven SQLite expiry/edit/deletion/disablement races |
| S13 — Delete or rotate with unresolved targets | 3.1 | Clock/barrier-driven SQLite expiry/edit/deletion/disablement races |
| S14 — Refresh only verified targets | 3.2 | SQLite catalog/sync fences + RTL private partial-outcome ordering |
| S15 — Upgrade while an old write is unresolved | 4.1 | Populated old-plan migration/reopen with unresolved guard |
| S16 — Keep unverified stack writes unavailable | 4.2 | Operator stack-capability admission fixture + documented live gate |
| P01 — Preview one coarse camera decision | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P02 — Reject an incomplete or expanded decision | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P03 — Observe missing partial and zero GPS | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P04 — Detect an intervening GPS change | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P05 — Distinguish image change from unrelated metadata | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P06 — Fail without a fabricated baseline | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P07 — Retain the same plan across reload | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P08 — Race preview creation with an edit | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P09 — Reject plan substitution | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P10 — Display an unchanged GPS pair | 1.2 | Preview SQLite/HTTP regressions for every selected member |
| P14 — Review without a map | 1.3 | Preview RTL/browser full-scope and private response regression |
| P15 — Preserve manual pending coordinates | 1.3 | Preview RTL/browser full-scope and private response regression |
| P16 — Handle late private preview replies | 1.3 | Preview RTL/browser full-scope and private response regression |
| W01 — Confirm and recover the approved operation | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W02 — Reconcile a lost confirmation response | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W03 — Reject stale or substituted approval | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W04 — Fail before durable approval | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W05 — Write exact GPS while preserving other fields | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W06 — Preserve valid zero and omit unsupported fields | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W07 — Prevent hidden retries or fallback | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W08 — Reject unsupported compatibility | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W09 — Detect fresh GPS or source conflict | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W10 — Lose authority after confirmation | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W11 — Verify an unchanged plan | 2.1 | Approval/adapter/authority regressions with v3 manifests |
| W14 — Recover remote success with lost response | 2.2 | Per-target readback/retry SQLite + request-count regression |
| W15 — Keep unresolved outcomes honest | 2.2 | Per-target readback/retry SQLite + request-count regression |
| W16 — Retry a known completed unsuccessful attempt | 2.2 | Per-target readback/retry SQLite + request-count regression |
| W17 — Reject exhausted stale or concurrent retries | 2.2 | Per-target readback/retry SQLite + request-count regression |
| W18 — Detect a different readback value | 2.2 | Per-target readback/retry SQLite + request-count regression |
| W27 — Recover a retry after its successful execution | 2.2 | Per-target readback/retry SQLite + request-count regression |
| W28 — Recover an accepted retry after restart and disablement | 2.2 | Per-target readback/retry SQLite + request-count regression |
| W19 — Observe rounded GPS or mismatched success | 3.2 | Verified target refresh and sync-fence regression |
| W20 — Recover local failure after remote success | 3.2 | Verified target refresh and sync-fence regression |
| W21 — Preserve verified GPS against an older sync | 3.2 | Verified target refresh and sync-fence regression |
| W22 — Handle an unavailable local catalog row | 3.2 | Verified target refresh and sync-fence regression |
| W23 — Disable while work is active | 3.1 | Multi-target lifecycle/privacy regression |
| W24 — Preserve audit through restart and cleanup | 3.1 | Multi-target lifecycle/privacy regression |
| W25 — Delete or rotate during an operation | 3.1 | Multi-target lifecycle/privacy regression |
| W26 — Keep unverified live writes disabled | 4.2 | Operator capability fixture + live evidence boundary |
| S17 — Select and confirm members accessibly | 1.3 | Vitest/RTL + built keyboard stack selection/confirmation journey |
| S18 — Preserve manual work across the full target set | 1.3 | Vitest/RTL + built keyboard stack selection/confirmation journey |
| S19 — Retain private partial history on reload | 3.2 | SQLite catalog/sync fences + RTL private partial-outcome ordering |
| D05 — Choose a coarse or ambiguous proposal | 1.3 | Draft geometry/scope and keyboard regression |
| D06 — Supply an absent camera point | 1.3 | Draft geometry/scope and keyboard regression |
| D07 — Validate geometry and field scope | 1.3 | Draft geometry/scope and keyboard regression |
| R01 — Edit before dispatch reservation | 3.1 | Real SQLite all-target revision guard regression |
| R02 — Edit while a write is unresolved | 3.1 | Real SQLite all-target revision guard regression |

## Fixtures and execution discipline

Mixed-GPS stacks with readable siblings, hidden/inaccessible/non-image members, bounded oversized selection, membership changes before/after confirmation, cross-version overlapping target lists and independently delayed target senders. Extend the synthetic stack endpoints; do not use a private live stack in ordinary tests.

Use OpenSpec Plus TDD for changed behavior, one meaningful failing assertion before its production change. Existing-behavior characterization is allowed to pass immediately against unchanged code; never manufacture a failure. Use injected clocks and explicit concurrency barriers. Persistence, isolation, migration, audit and transaction claims require real temporary file-backed SQLite and normal Goose migrations, including reopen. Provider/Immich call assertions use local HTTP servers and the real serialization/handler path. No retries hide failing browser tests.

Each vertical slice includes its mapped policy, persistence, API and UI checks. The final apply verification uses `bun run check --base <actual-implementation-base>` with Node 22.23.2, Go 1.25.14 and Bun 1.4.2: checker tests, size/import rules, gofmt, lint, type generation/types, Go vet/race tests, frontend unit coverage, both builds and built-browser acceptance. Preserve 80% Go AI statements and 80% AI TypeScript lines/branches. Run relevant container builds when build/package/deployment inputs change. Extend affected import rules rather than omitting new AI adapters from coverage.

Run pre-edit GitNexus impact on each changed symbol and pre-commit `detect_changes` against the correct checkout. UNKNOWN/partial/lower-bound results need source/test confirmation; they are not clean impact evidence. Check strict OpenSpec validation, ordered delta composition, scenario/test mapping, local links and whitespace before completing the change. Sync only after implementation and verification; this planning phase leaves maintained specs unchanged.

Retain the existing legacy auth/manual/GPX journeys and AI-disabled regression, plus an affected built AI journey. Keep inspection/acceptance at zero Immich writes and preserve CH19 confirmation identity, late-history, retry-generation and local-refresh regressions throughout. Populate upgrade fixtures with queued, completed and unresolved old state; an empty-database migration test is not enough. No test deletes audit to simplify recovery.

## Live and rollout boundary

Ordinary tests use synthetic data and deterministic local services. Live provider text compatibility, Immich field/scope/metadata compatibility, language quality and performance require separately recorded evidence; an offline pass does not establish them. CH19 live GATE-02 is still pending while this batch is planned. New field/scope capabilities remain default-off and require their own authorized version/profile/rights/readback evidence. Document the precise tested behavior and sidecar/UI limits; do not reuse a GPS pass as a metadata or stack pass.

The user request here authorizes planning only. Do not dispatch private text/images, activate a writer, deploy, restore a database, clear GPS or mutate the selected CH19 fixture as part of this plan. A live gate must present the exact proposed operation and use the normal authenticated durable confirmation path. Preserve backups, encryption keys and unresolved guards during any later authorized rollout.
