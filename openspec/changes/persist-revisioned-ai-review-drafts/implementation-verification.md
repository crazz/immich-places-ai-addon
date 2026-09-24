# CH16 implementation verification

Status: implemented, verified and synchronized on 24 September 2026. All ten tasks are complete. Active change retained without archive; no live compatibility or deployment claim.

Base: `ff271974fb7622156cbb765b57897a8e977e769e`, branch `codex/plan-ai-selection-and-validation`, current checkout. Inline implementation and self-review, no subagents. Unrelated `.brooks-lint-history.json` is excluded.

## Workflow

- [x] Preflight: artifacts, engineering standards, relevant product/design/reconciliation/roadmap and installed commands read. GitNexus bound to this checkout at the base commit.
- [x] Dependency analysis: sequential slices 1 → 2 → 3 → 4; persistence/HTTP/editor boundaries overlap. New draft package and migration; existing history, cleanup and UI integration files confirmed in Git history.
- [x] Slice 1: durable acceptance, review states and private lifecycle.
- [x] Slice 2: explicit camera/content correction and isolated accessible editing.
- [x] Slice 3: revision conflicts, uncertain saves and response fencing.
- [x] Slice 4: current-source observations and acknowledgement.
- [x] Whole-change spec then quality self-review, full gates and spec synchronization; recorded with the CH16 implementation commit.

Conventions applied throughout: English; 500-line handwritten limit; focused pure `internal/ai/drafts` policy, root SQL/HTTP adapters and `src/features/ai` UI; immutable originals; exact owner/installation scope; no provider or Immich mutation authority; no manual/GPX pending-state changes; explicit null/zero; bounded DTOs; one test at a time; real file-backed SQLite for persistence; synthetic external fixtures; safe errors; current-revision preconditions; measured AI coverage and full installed gates against the base.

## Development evidence

Test logs are retained locally under ignored `out/checks/ch16/`. Commands source `out/checks/ch16/env.sh` for pinned Node 22.23.2, Go 1.25.14 and Bun 1.4.2. Initial Go invocation exposed missing module-cache metadata; dependency setup is required before a meaningful RED result. Setup failure is not recorded as a behavioral RED.

| Test | Before → after | RED / GREEN | Refactor assessment |
|---|---|---|---|
| `TestAIDraftAcceptSurvivesReopenWithoutChangingAnalysis` | 0 → 1 in `backend/aiDraftAcceptance_test.go` | `d01-red.log`: route 404; `d01-green.log`: PASS | No extraction: focused SQL adapter and draft data contract. |

## GitNexus

Initial graph commit and checkout match the base; no incomplete index reasons. `newAIResultHandler` and `aiResultStore` impact LOW (composition-root caller). `scanAIResultEntry` impact HIGH: detail/list and downstream thumbnail/HTTP flows; regression verification required. `PurgeBefore` impact UNKNOWN from unresolved receiver calls; source confirms four calls in existing cleanup tests, with no production caller. These limitations are retained rather than treated as absence of impact.

## Final gates and scenario map

All thirteen installed gates passed against the recorded base. Final results and retained log names appear below. No live provider or Immich calls were authorized or performed.

D01 setup used a fresh module cache and the installed Homebrew Go 1.25.14 because temporary caches had been pruned. The final original-payload assertion reads SQLite directly: the legacy execution store intentionally denies ReadAnalysis when execution is disabled. D02 draft test was removed until D01 completed, then reintroduced as characterization of the implemented unique acceptance boundary.

D02 characterization (1 → 2 acceptance tests): PASS, repeated acceptance returns the unique snapshot. Refactor assessment: unique aggregate lookup is minimal. HTTP protection (2 → 3): missing origin returned 200 (RED), origin/authentication/strict bodies PASS (GREEN). Refactor assessment: retain existing strict decoder; mutation guard will be shared when PATCH is introduced. Composition-root `main` impact UNKNOWN is the Go executable entrypoint, verified in source; changed only draft route registration and configured origin.

D03 (0 → 1 transition test): PATCH returned 405 (RED); reject/reopen creates three immutable snapshots and reaccept returns revision 3 (GREEN). Refactor: shared mutation-origin guard extracted for POST/PATCH, existing tests pass. Test body only changes local review state; no external adapter is available to the draft store.

D04 (3 → 4 acceptance tests): existing record validation characterization PASS; invalid IDs/corrupt proposal create no draft. No refactor: reuse validated detail. D15 (0 → 1 lifecycle test): cleanup failed on the retained analysis FK (RED); exclusion of draft-referenced jobs PASS with existing cleanup tests (GREEN). Refactor: one owner-qualified anti-join, no new retention policy.

D16 (1 → 2 lifecycle tests) and D17 (2 → 3) characterize owner/installation fencing and two-account cascades: PASS. Refactor: no duplication requiring a production abstraction; existing transaction/foreign keys enforce the boundary. History projection (4 → 5 acceptance tests): unreviewed/missing draft identity RED, owner-qualified one-row join GREEN with all existing result tests. `review.Entry` impact UNKNOWN was verified against its scanner and detail/list consumers in source.

D06/D09 (1 → 2 transition tests): explicit zero point rejected RED, camera/field editing and unstaging GREEN. Refactor: extracted bounded nullable point parsing from the transition. D05 (0 → 1 candidate tests): explicit choice rejected RED, coarse 5000 m candidate accepted and staged GREEN. Refactor: proposal-to-draft conversion owns candidate mapping. D10 (0 → 1 content tests): missing stale inherited values RED, field-specific factual revision GREEN. Refactor: independent scene-only descriptions retain their original basis. D11 (1 → 2): description edit rejected RED, one-language correction with GPS-only staging GREEN. Refactor: original language set is used for validation, no provider dependency. D07 (2 → 3): valid heading rejected RED, nullable heading and geometry/scope validation GREEN. Refactor: shared finite-range numeric validation. These focused commands run all `TestAIDraft` tests after each change; logs use scenario prefixes under the local evidence directory.

Backend-to-UI dependency refinement: baseline persistence was implemented before the editor consumes it; no slice is marked complete ahead of its UI and review gates. D08 competing revisions: race-enabled PASS with exactly one 200 and one 412, missing precondition 428 and reacceptance reconciling revision 2. Refactor: no additional locking; SQLite serialization is the authority. D13 (0 → 1 baseline tests): absent route RED, real read-adapter observations/acknowledgement GREEN for absent/partial/zero/present GPS. Refactor: source decoding and observation persistence split. D14 (0 → 1 baseline-safety tests): existing fencing characterized across GPS/source/access/key/installation/revision/expiry/deletion/storage failures; PASS, no failed acknowledgement adds a revision. Observation cap (1 → 2): eleventh observation accepted RED; cap ten and bounded expired cleanup GREEN.

UI acceptance (0 → 1 `ResultDraft` test): missing action RED, explicit acceptance and saved revision GREEN. Existing inspection test updated to permit the newly explicit local action while retaining inert original presentation. UI numeric edit (1 → 2): missing numeric control RED, zero-pair save with exact If-Match GREEN. Refactor: separated input validation, API, editor and review lifecycle; synthetic draft fixture extracted into `testing/draft.ts` and focused tests rerun.

## Additional per-test evidence and refactor assessments

All tests below were added one at a time after the preceding test passed. Characterization asserts already implemented behavior; no manufactured failure was required. Local logs retain RED/GREEN and setup failures separately.

| Test / focused behavior | Result and refactor assessment |
|---|---|
| `ResultDraft`: preserve two-tab edits and compare saved revision | RED missing recovery controls → GREEN. Dedicated draft request helper and review lifecycle keep replay explicit. |
| `DraftEditor`: local heading/language corrections with GPS-only staging | RED missing fields → GREEN. Local edit DTO includes only explicitly corrected non-GPS fields. |
| `DraftNavigation`: explicit keep/discard | RED missing prompt → GREEN; corrected effect so typing does not itself prompt. Separate URL restoration from navigation interception. |
| `DraftEditor`: overlapping manual work | RED staging allowed → GREEN. Read-only bridge reports pending IDs; no pending-state mutation API added. |
| `DraftEditor`: isolated map and failed tiles | RED map missing → GREEN. Separate map component reuses construction only, with cleanup on unmount. |
| `DraftEditor`: explicit candidate change | RED missing selector → GREEN. Proposal candidate choice stays separate from inspection focus. |
| `TestAIDraftCandidateChangeInvalidatesFactsWithoutMovingSubjectIntoCamera` | RED candidate PATCH unsupported → GREEN. Reuses validated proposal conversion; subject-only candidate leaves camera absent. |
| `DraftBaseline`: exact partial values and image acknowledgement | RED missing presentation → GREEN. Fixed test transport to use byte response and await Next image load callback; no weakening of image acknowledgement. Source API and component split by responsibility. |
| `TestAIDraftChangedStagedContentRequiresSeparateRestaging` | RED content-plus-stage retained staged state → GREEN. Explicit previous-state guard requires a later staging action. |
| `ResultDraft`: discard after uncertain save at unchanged revision | RED local coordinates survived discard → GREEN. Editor reset generation handles equal-revision comparisons. |
| `ResultDraft`: late acceptance across owner/result changes | Characterization PASS. Existing abort/key boundaries suffice; no new abstraction. |
| `TestAIDraftUpgradeFrom25AndPoolConstraints` | Characterization PASS. Real migrations/reopen/repeat and four concurrent pooled connections enforce owner FKs. |
| `TestAIDraftFailedMigrationRollsBackAndRetainedTablesAllowLegacyReads` | Characterization PASS. Deliberate schema conflict proves rollback; existing legacy analysis read works with new tables retained. No destructive application rollback procedure introduced. |
| `TestAIDraftAcceptanceSerializesWithPurge` | Race-enabled characterization PASS across eight synchronized starts. Existing short SQLite writer transactions serialize acceptance/purge; no extra lock. |
| `TestAIDraftLocalOfflineDecisionsMakeNoUpstreamRequests` | Characterization PASS. Zero calls during local actions; hidden/removed source observations make no unauthorized read. Reused synthetic read transport. |
| `TestAIDraftReanalysisKeepsEachDecisionDistinct` | Characterization PASS. Same-photo runs retain independent draft IDs and revisions. |
| `TestAIDraftBaselineUsesItsExistingSQLiteConnection` | RED nested connection acquisition timed out with pool size one → GREEN. Final access checks use the existing SQL transaction, with no network inside it. |
| `TestReviewOneLanguageRetainsOtherStaleAndUnavailableContent` | Characterization PASS. One review leaves the second language stale, third unavailable and heading stale, without mutating the previous value. |
| Built journey `persists local drafts, reconciles two tabs and reviews baselines without upstream writes` | First launch blocked by sandbox; rerun outside sandbox found a real Leaflet animation-after-removal error. Nonanimated initial viewport and intact remove lifecycle fixed it; rebuilt journey PASS. Local installed Leaflet source and Context7 docs informed the fix. |
| Built journey `keeps manual pending coordinates isolated and asks before discarding unsaved AI edits` | Characterization PASS. AI requests produce zero writes; a subsequent explicit legacy save still sends the original manual pair. |
| `DraftEditor`: controls during pending save | RED editable while saving → GREEN. Disabled/inert fieldset prevents unsaved edits being lost on response. |
| `TestAIDraftAccountDeletionFencesInflightObservation` | Race-enabled characterization PASS. A channel pauses the HTTP read while the account is deleted; no late observation is published. |
| `TestAIDraftBaselineAcknowledgementInvalidatesStaging` | RED baseline retained stage → GREEN. Snapshot and current projection transition together in the transaction. |
| `DraftEditor`: unsaved camera stale feedback | RED estimates remained current → GREEN. Local factual changes immediately label dependent content stale. |
| `DraftEditor`: clearing language review preserves corrected text | RED correction was discarded → GREEN. Only review-only entries are removed. |
| `TestAIDraftStrictBodiesPreconditionsAndSafeStorageFailure` | Duplicate/oversized/malformed requests characterized correctly; explicit null fields incorrectly accepted (RED). Raw bounded field selection now distinguishes omitted from null (GREEN). Storage failure preserves the revision and hides private SQL details. |

Cross-task refactor assessment: draft validation/proposal conversion, SQL storage, HTTP adaptation, fresh source reading and baseline persistence have distinct owners. UI DTO/request, editor, map, baseline and reconciliation concerns are separate. Mutation origin and revision checks are shared; baseline final authority checks use one transaction helper. No workflow engine, provider dependency, writer import or manual-pending abstraction was introduced. A GitNexus rename attempt could not resolve the local destructured boolean; source-confirmed three local references were updated to satisfy the existing naming rule. Mechanical lint changes were scoped to changed files.

## Scenario traceability

| Scenario | Automated evidence |
|---|---|
| D01 | `TestAIDraftAcceptSurvivesReopenWithoutChangingAnalysis`; built persistence journey |
| D02 | `TestAIDraftAcceptanceIsIdempotent`; `TestAIDraftConcurrentEditsRequireExactRevision`; reject/reopen reacceptance |
| D03 | `TestAIDraftRejectReopenKeepsRevisionHistoryAndReacceptance`; `TestAIDraftReanalysisKeepsEachDecisionDistinct`; built reject/reload/reopen |
| D04 | `TestAIDraftRejectsUnavailableAndCorruptProposals`; existing result failure/cancellation projection tests |
| D05 | `TestAIDraftExplicitAmbiguousCandidateStagesCoarsePoint`; candidate selector component test |
| D06 | `TestAIDraftUnknownRequiresExplicitCameraAndGPSSelection`; subject-only candidate test |
| D07 | `TestAIDraftGeometryAndScopeValidationIsAtomic`; `TestAIDraftStrictBodiesPreconditionsAndSafeStorageFailure`; zero-coordinate component/browser saves |
| D08 | `TestAIDraftConcurrentEditsRequireExactRevision`; conflict component test; built two-tab conflict |
| D09 | `TestAIDraftChangedStagedContentRequiresSeparateRestaging`; baseline unstaging test; uncertain-save same-revision discard component test; built stage/edit/reload |
| D10 | camera/candidate invalidation integration tests; unsaved-stale component test; browser heading feedback |
| D11 | `TestAIDraftReviewsOneLanguageWithoutApprovingOtherFields`; three-language pure test; editor independent correction/review tests |
| D12 | `TestAIDraftLocalOfflineDecisionsMakeNoUpstreamRequests`; acceptance/reopen with unavailable baseline |
| D13 | `TestAIDraftBaselineRetainsExactNullableGPSAndOriginalProvenance`; baseline image-acknowledgement component test; built absent-GPS acknowledgement |
| D14 | `TestAIDraftStaleBaselineCannotReplaceSavedDecision`; observation cap/expiry; baseline pool test |
| D15 | `TestAIDraftProtectsHistoryFromCleanupAndCatalogReset`; acceptance/purge race; schema-025 upgrade/repeat/reopen tests |
| D16 | `TestAIDraftDeniesForeignAndObsoleteScope`; authenticated/origin HTTP tests; baseline scope/visibility/key invalidation |
| D17 | `TestAIDraftAccountDeletionCascadesWithoutLateRecreation`; barrier-controlled in-flight observation deletion |
| D18 | zero numeric input and failed-tile map component tests; 390px keyboard persistence journey and screenshot |
| D19 | manual-overlap component test and built manual-preservation journey; existing manual/GPX smoke regression |
| D20 | `DraftNavigation`; late private-response component test; built keep/discard navigation and two-tab comparison |

The UI tests preserve private values only in component state and use deterministic fetch fixtures; Go persistence assertions use real SQLite and real migrations. Baseline reads use local HTTP transport fixtures. The two new built journeys run against the real Go/Next.js/proxy stack and synthetic services, asserting unchanged provider counters and zero AI-triggered Immich writes.

## Inline reviews

The proposal, D01–D20 delta and design were reread for spec compliance; the design and cumulative source changes were then reviewed for quality. These are inline self-checks, not independent reviews. No subagents were used.

Spec-compliance checklist (each slice):

1. Durable acceptance/history/lifecycle — PASS: explicit idempotent acceptance, immutable originals/revisions, current-state joins, owner/installation isolation, cleanup serialization and account cascades match slice 1.
2. Explicit camera/content editing — PASS: absent/zero/coarse geometry, exact analyzed asset/GPS fields, independent stale heading/languages, isolated map, keyboard controls and manual-overlap handling match slice 2.
3. Revision/recovery boundaries — PASS: If-Match 428/412 behavior, preservation/compare before replay, separate restaging and late-response fencing match slice 3.
4. Baseline workflow — PASS: explicit currently authorized reads, nullable exact GPS, separate provenance, five-minute bounded observations, revision/access/source rechecks and renewal match slice 4.
5. Scenario coverage — PASS: all twenty scenario identifiers map to automated evidence above. Real SQLite, local transport and built-browser layers remain distinguished.
6. Scope/design — PASS: no provider calls from draft actions, AI Immich writer, translation generation, stack expansion, proxy repair, new dependency or ADR departure.

TDD discipline checklist:

1. New behavior — PASS: atomic test-first cycles and regression fixes are recorded above; test-environment errors are not behavioral RED results.
2. Existing behavior — PASS: characterization uses the adopted testing-standard exception; no manufactured failure, mutation test or skipped gate.
3. Refactor assessments — PASS: per-test outcomes plus cross-task consolidation are recorded; fixes preserve task boundaries.

Code-quality and whole-change checklist:

1. Shared contracts/naming — PASS: backend JSON and frontend DTOs agree; draft, revision, source and write status remain distinct.
2. Responsibility/decomposition — PASS: pure draft policy, SQL/HTTP/source adapters, and feature-owned editor/map/reconciliation components are separate; no handwritten file exceeds the size policy.
3. Duplication/shared abstractions — PASS: bounded API request handling, origin/revision guards, snapshot persistence, and transaction authority checks are shared within their owners. No speculative workflow abstraction is introduced.
4. Dead/superseded code — PASS: no abandoned implementation, skipped tests or commented-out code remains.
5. Authorization/transaction boundaries — PASS: owner-qualified records, immutable triggers and short SQLite transactions; network reads occur outside transactions and final authority rechecks use the held connection.
6. UI lifecycle/clarity — PASS: old private responses are fenced, map/image resources are released, unsaved changes are explicit, pending controls are disabled, and local estimates are marked stale immediately.
7. Surgical scope — PASS: every source/test change supports CH16; the legacy migration-tip assertion changes from 25 to 26. Manual/GPX implementation is untouched.
8. Project conventions — PASS: English artifacts, adopted package boundaries, scoped formatting/lint, real SQLite and synthetic upstreams. Final gate results below remain the source of verification truth.
9. Artifact gaps — none requiring a product/design change. The 32 KiB existing strict request decoder remains in use; durable snapshots separately cap at 128 KiB. No live rollout or model-quality claim follows from deterministic tests.

Spec synchronization: `openspec-sync-specs` and the required `openspec-plus-spec` review ran inline under the user's no-subagent/ordered-sync authorization. The existing approved seven requirements and twenty scenarios were merged unchanged into `ai-results-and-review`; its maintained Purpose now includes explicit local drafts and preserves inert inspection/no write authority. Main-spec format, requirement uniqueness, Gherkin, proposal/design alignment and unchanged prior scenarios passed review. Strict validation passed all seven main specs. The change remains active; no archive was performed.

GitNexus refresh succeeded for this checkout. The final staged change analysis includes new files: 130 changed symbols, 88 affected flows across 61 files, CRITICAL aggregate breadth, with the complete returned symbol listing and no partial/truncated response. All 95 functions and twelve methods listed in the preceding analysis remain represented; property/document-section listing differences do not imply less behavioral impact. The scope covers draft routes, result projection, composition, AI editor/navigation and their callers; all are covered by the installed source checks and behavioral suites. Earlier pre-edit HIGH warnings covered the scanner, transition and baseline operations. Graph import-cycle enumeration is complete and clean (zero cycles). Route shape extraction reports no supported paired Go-route/consumer shape, so HTTP/component/built-browser tests establish the contract instead. Process indexing itself uses bounded entry-point/branch enumeration; absent graph flows are not proof of absence. The local `gitnexus-precommit.json` retains full details.

Environment-only artifacts: the unrelated generated `out/audits/brooks-2026-09-22/probe.config.ts` and `library-cache.test.tsx` entered broad TypeScript/ESLint globbing. They are preserved beside their originals with `.saved` suffixes, outside Git. No application check exclusion was relaxed. The unrelated `.brooks-lint-history.json` remains untracked and unstaged.

## Final executed verification

The final source state passed `bun run check --base ff271974fb7622156cbb765b57897a8e977e769e` with exit 0. All thirteen gates report PASS: checker regressions, size, dependencies, gofmt, lint, route type generation, TypeScript, Go vet, race-enabled Go tests, Go build, frontend coverage tests, frontend production build and built-browser smoke. The complete local log is `out/checks/ch16/full-check-final.log`.

| Check | Actual result |
|---|---|
| Backend race/coverage | PASS across the backend and all eleven internal test packages. AI statement coverage 87.97% (4569/5194), above the 80% floor. The fresh uncached run is also retained in `go-final.log`; the final combined run reused those valid Go results. |
| Frontend unit/component coverage | PASS: 51 test files, 114 tests. AI lines 93.81% (1000/1066), branches 83.91% (1486/1771), both above the 80% floor. |
| Built-browser journeys | PASS: three legacy journeys and fourteen AI journeys, including both new CH16 journeys. Synthetic services, separate temporary SQLite installations, production builds and no retries. |
| Narrow visual review | PASS: the 390px screenshot was inspected. The language-review checkbox alignment found during inspection was corrected and the full production build/browser gate rerun; `draft-mobile.png` records the final layout. |
| Lint and source policies | PASS, zero lint errors; three existing warnings remain visible. Size ratchet and dependency checks pass without new exclusions. |
| Backend container | PASS: `docker build -f backend/Dockerfile -t immich-places-ch16-check backend`; final source build retained in `backend-container-final.log`. Local build only. |
| OpenSpec | PASS: `openspec validate --all --strict --no-interactive`, eleven specs/active changes; all seven delta requirements and twenty scenarios appear verbatim in the maintained capability. |
| Documentation | PASS: changed Markdown local links resolve, main requirement names remain unique, task/scenario mapping covers D01–D20, and `git diff --check` is clean. |

The first combined attempt failed on the stale schema-tip assertion (25 rather than 26) and generated audit scratch files. Both causes were corrected, with the scratch contents preserved; no failure was suppressed or gate weakened. The final combined run above is the closure result. No source changes followed it; final edits only completed the evidence, task and roadmap records.

CH18 may now start in a fresh local task after this change is committed. It must consume the synchronized draft contract and complete its own exact preview work before CH19. No change was archived, pushed or deployed. Live Immich compatibility remains outside this evidence; the future writer stays disabled pending separately authorized disposable-photo verification.
