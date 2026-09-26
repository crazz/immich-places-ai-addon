# CH20 verification plan

Status: **implemented, verified and synchronized**. The [implementation verification](implementation-verification.md) records actual test names, commands, failures/fixes and results. The traceability below preserves the original planned acceptance boundaries. No live provider request or private Immich mutation was performed.

## Traceability

The proposal links the owning PRD FR/NFR/AC obligations. Every requirement below maps to scenarios; every scenario maps to task IDs and an automated boundary. Full MODIFIED blocks retain prerequisite scenarios intentionally. Preserve the rest of the maintained capability and all predecessor regression suites when this change is synchronized.

| Requirement | Scenario IDs |
|---|---|
| Select one reviewed standard description independently of GPS | C01, C02, C03 |
| Compare exact selected description baselines | C04, C05, C06 |
| Update one owned managed append block without duplication | C07, C08, C09, C10 |
| Reconcile selected standard fields independently | C11, C12, C13, C14 |
| Preserve historical approvals across standard-field upgrades | C15, C16, C17 |
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
| Review independent standard-field choices and outcomes | C18, C19, C20 |
| Edit camera decisions without inventing precision or writable fields | D05, D06, D07 |
| Protect a draft revision while its GPS write can still act | R01, R02 |

| Scenario / planned case | Tasks | Planned automated boundary |
|---|---|---|
| C01 — Save a description without camera coordinates | 1.1 | Go draft field policy + protected HTTP + RTL |
| C02 — Preserve text during GPS-only saving | 1.1 | Go draft field policy + protected HTTP + RTL |
| C03 — Reject unusable description selection | 1.1 | Go draft field policy + protected HTTP + RTL |
| C04 — Review empty Unicode and multiline baselines | 1.2 | Real SQLite observations + authorized metadata HTTP fixture |
| C05 — Detect a changed description | 1.2 | Real SQLite observations + authorized metadata HTTP fixture |
| C06 — Distinguish unrelated changes from unavailable reads | 1.2 | Real SQLite observations + authorized metadata HTTP fixture |
| C07 — Append while preserving existing text | 2.1 | Go append policy + exact outgoing/readback HTTP + SQLite lineage |
| C08 — Repeat append or change primary language | 2.1 | Go append policy + exact outgoing/readback HTTP + SQLite lineage |
| C09 — Reject tampered or unowned blocks | 2.2 | Go block/UTF-8 bounds + HTTP zero-mutation fixtures |
| C10 — Reject oversized text without truncation | 2.2 | Go block/UTF-8 bounds + HTTP zero-mutation fixtures |
| C11 — Verify a combined exact update | 3.1 | Local socket exact payload/count + durable SQLite reservation |
| C12 — Recover a lost description response | 3.2 | HTTP readback failure/mixed-state fixture + SQLite generation recovery |
| C13 — Retain a mixed standard-field outcome | 3.2 | HTTP readback failure/mixed-state fixture + SQLite generation recovery |
| C14 — Preserve replay identity and attempt limits | 3.2 | HTTP readback failure/mixed-state fixture + SQLite generation recovery |
| C15 — Reopen old approval and unknown sender state | 4.1 | Populated v1 upgrade/reopen + cross-owner/version guard races |
| C16 — Reject cross-version overlap or altered plans | 4.1 | Populated v1 upgrade/reopen + cross-owner/version guard races |
| C17 — Keep unverified description support disabled | 4.2 | Operator capability admission fixture + documented live gate |
| P01 — Preview one coarse camera decision | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P02 — Reject an incomplete or expanded decision | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P03 — Observe missing partial and zero GPS | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P04 — Detect an intervening GPS change | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P05 — Distinguish image change from unrelated metadata | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P06 — Fail without a fabricated baseline | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P07 — Retain the same plan across reload | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P08 — Race preview creation with an edit | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P09 — Reject plan substitution | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P10 — Display an unchanged GPS pair | 1.2 | Existing preview SQLite/HTTP regression, extended to description plans |
| P14 — Review without a map | 1.3 | Existing preview RTL/browser regression, full text and selected GPS overlaps |
| P15 — Preserve manual pending coordinates | 1.3 | Existing preview RTL/browser regression, full text and selected GPS overlaps |
| P16 — Handle late private preview replies | 1.3 | Existing preview RTL/browser regression, full text and selected GPS overlaps |
| W01 — Confirm and recover the approved operation | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W02 — Reconcile a lost confirmation response | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W03 — Reject stale or substituted approval | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W04 — Fail before durable approval | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W05 — Write exact GPS while preserving other fields | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W06 — Preserve valid zero and omit unsupported fields | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W07 — Prevent hidden retries or fallback | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W08 — Reject unsupported compatibility | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W09 — Detect fresh GPS or source conflict | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W10 — Lose authority after confirmation | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W11 — Verify an unchanged plan | 3.1 | Existing approval/adapter/authority SQLite + HTTP regression for v1 and v2 |
| W14 — Recover remote success with lost response | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W15 — Keep unresolved outcomes honest | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W16 — Retry a known completed unsuccessful attempt | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W17 — Reject exhausted stale or concurrent retries | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W18 — Detect a different readback value | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W27 — Recover a retry after its successful execution | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W28 — Recover an accepted retry after restart and disablement | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W19 — Observe rounded GPS or mismatched success | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W20 — Recover local failure after remote success | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W21 — Preserve verified GPS against an older sync | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W22 — Handle an unavailable local catalog row | 3.2 | Existing CH19 readback/refresh/retry regressions for v1 and v2 |
| W23 — Disable while work is active | 4.2 | Existing lifecycle/privacy fixtures plus description capability admission |
| W24 — Preserve audit through restart and cleanup | 4.2 | Existing lifecycle/privacy fixtures plus description capability admission |
| W25 — Delete or rotate during an operation | 4.2 | Existing lifecycle/privacy fixtures plus description capability admission |
| W26 — Keep unverified live writes disabled | 4.2 | Existing lifecycle/privacy fixtures plus description capability admission |
| C18 — Review and confirm text with keyboard input | 1.3 | Vitest/RTL + built description confirmation/recovery journey |
| C19 — Preserve manual GPS during a text-only operation | 1.3 | Vitest/RTL + built description confirmation/recovery journey |
| C20 — Keep field outcomes and newer review separate | 1.3 | Vitest/RTL + built description confirmation/recovery journey |
| D05 — Choose a coarse or ambiguous proposal | 1.1 | Draft policy/RTL geometry and explicit field-selection regression |
| D06 — Supply an absent camera point | 1.1 | Draft policy/RTL geometry and explicit field-selection regression |
| D07 — Validate geometry and field scope | 1.1 | Draft policy/RTL geometry and explicit field-selection regression |
| R01 — Edit before dispatch reservation | 4.1 | Real SQLite edit-versus-send reservation regression |
| R02 — Edit while a write is unresolved | 4.1 | Real SQLite edit-versus-send reservation regression |

## Fixtures and execution discipline

Existing descriptions with Unicode, empty/absent text, exact whitespace, first and repeated managed blocks, foreign/edited/duplicated markers, combined GPS/text partial readback and lost responses; populated v1 operation/guard upgrades. Extend the existing synthetic Immich fixture; a passing GPS-only fixture is insufficient.

Task 4.1 also extends maintained P11–P13 with mixed-version previews: one shared owner capacity, expiry across versions, reopen at the limit, and cleanup that preserves retained approvals/audit. These requirements remain in the maintained spec without a delta; successor versions retain the same regression boundary.

Use OpenSpec Plus TDD for changed behavior, one meaningful failing assertion before its production change. Existing-behavior characterization is allowed to pass immediately against unchanged code; never manufacture a failure. Use injected clocks and explicit concurrency barriers. Persistence, isolation, migration, audit and transaction claims require real temporary file-backed SQLite and normal Goose migrations, including reopen. Provider/Immich call assertions use local HTTP servers and the real serialization/handler path. No retries hide failing browser tests.

Each vertical slice includes its mapped policy, persistence, API and UI checks. The final apply verification uses `bun run check --base <actual-implementation-base>` with Node 22.23.2, Go 1.25.14 and Bun 1.4.2: checker tests, size/import rules, gofmt, lint, type generation/types, Go vet/race tests, frontend unit coverage, both builds and built-browser acceptance. Preserve 80% Go AI statements and 80% AI TypeScript lines/branches. Run relevant container builds when build/package/deployment inputs change. Extend affected import rules rather than omitting new AI adapters from coverage.

Run pre-edit GitNexus impact on each changed symbol and pre-commit `detect_changes` against the correct checkout. UNKNOWN/partial/lower-bound results need source/test confirmation; they are not clean impact evidence. Check strict OpenSpec validation, ordered delta composition, scenario/test mapping, local links and whitespace before completing the change. Sync only after implementation and verification; this planning phase leaves maintained specs unchanged.

Retain the existing legacy auth/manual/GPX journeys and AI-disabled regression, plus an affected built AI journey. Keep inspection/acceptance at zero Immich writes and preserve CH19 confirmation identity, late-history, retry-generation and local-refresh regressions throughout. Populate upgrade fixtures with queued, completed and unresolved old state; an empty-database migration test is not enough. No test deletes audit to simplify recovery.

## Live and rollout boundary

Ordinary tests use synthetic data and deterministic local services. Live provider text compatibility, Immich field/scope/metadata compatibility, language quality and performance require separately recorded evidence; an offline pass does not establish them. CH19 live GATE-02 is still pending while this batch is planned. New field/scope capabilities remain default-off and require their own authorized version/profile/rights/readback evidence. Document the precise tested behavior and sidecar/UI limits; do not reuse a GPS pass as a metadata or stack pass.

The original preparation authorized planning only; the subsequent user instruction authorized implementation of CH17 → CH20 → CH21 → CH22. This does not authorize dispatching private text/images, activating a live writer, deployment, database restoration, clearing GPS or mutating the selected CH19 fixture. A live gate must present the exact proposed operation and use the normal authenticated durable confirmation path. Preserve backups, encryption keys and unresolved guards during any later authorized rollout.
