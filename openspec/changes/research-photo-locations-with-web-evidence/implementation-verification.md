# Implementation verification

Status: implementation complete; 12/12 tasks verified. No deployment or live geolocation-quality claim.

## Execution

Inline, sequential execution selected by the user; no subagents. Planning and apply preflight are complete. The twelve tasks are ordered as four dependent slices: launch/review, hints/references, durable longer execution, compatibility/verification. The task checkboxes were completed after their applicable reviews and gates passed.

Conventions: use focused AI core/UI owners, preserve the read-only boundary and existing dependencies, keep handwritten files within 500 lines, verify persistence with real SQLite, use one test at a time with recorded refactor assessment, preserve v1 semantics, and run GitNexus impact before existing symbol edits. Final verification includes the installed shared gates and scenario mapping; unavailable live evidence must be identified.

GitNexus is bound to this checkout at base `58350f7`. Initial validator/context/geometry changes are CRITICAL; document/validator types are HIGH. The affected paths are analysis, atomic completion and history detail/thumbnail reads. Receiver-typing gaps and inconsistent mode-symbol identities were checked against source references; an empty graph result is not a clean impact verdict.

## TDD record

| Test | Before → after | RED / characterization | GREEN | Refactor assessment |
|---|---|---|---|---|
| `providerhttp.TestResearchResponseWaitPreservesTheFiniteParentDeadline` | `research_deadline_test.go`: 0 → 1 | Missing deadline helper | Provider HTTP package passes | Centralized response wait calculation; parent cancellation/deadline stay authoritative, normal calls keep the one-minute header limit. |
| `results.TestResearchRetains500MeterEstimate` | `research_test.go`: 0 → 1 | `invalid_context` | Results package passes | No extraction needed; schema selection shares existing validation. |
| `results.TestResearchRetainsCoarseCityEstimate` | `research_test.go`: 1 → 2 | `semantic_violation` | Results package passes | Named the legacy-only precision rule; common numeric/schema checks stay shared. |
| `results.TestResearchPreservesExistingValidationAndUncertainty` | `research_test.go`: 2 → 3 | Characterization; corrected expected overflow category to existing `limit_exceeded` | Results package passes | No production change needed: existing outcomes, numeric bounds and legacy semantics apply. |
| `results.TestResearchRetainsAnswerSources` | `research_sources_test.go`: 0 → 1 | `schema_violation` | Results package passes | Answer references use a separate local map; they do not gain input-context authority. |
| `results.TestResearchBoundsAnswerSourceBytesAndIdentities` | `research_sources_test.go`: 1 → 2 | Accepted oversized UTF-8 URL | Results package passes | Byte bound belongs to semantic validation; JSON Schema character counts alone are insufficient. |
| `results.TestResearchKeepsEstimateWithMissingReference` | `research_sources_test.go`: 2 → 3 | `semantic_violation` | Results package passes | Removed unused answer-lookup map: missing optional refs are a presentation concern; duplicates and context authority remain checked. |
| `results.TestResearchRejectsDuplicateAnswerSourceIdentity` | `research_sources_test.go`: 3 → 4 | Duplicate source accepted | Results package passes | Reused existing identifier convention; no separate authority map introduced. |
| `results.TestResearchBoundsSourceTextBytes` | `research_sources_test.go`: 4 → 5 | Oversized UTF-8 text accepted | Results package passes | Kept text byte limits at the same semantic boundary as URL bounds. |
| `results.TestResearchContextStillRequiresAuthorizedInput` | `research_sources_test.go`: 5 → 6 | `semantic_violation` for an authorized hint | Results package passes | Existing provided-context/source binding reused; answer sources never satisfy it. |
| `results.TestResearchIdentifiesItsValidationPolicy` | `research_test.go`: 3 → 4 | Reported v1 policy | Results package passes | Policy derives from the immutable typed document; no duplicate version state. |
| `analysis_test.TestResearchSendsOneImageAndHintAndReturnsCoarseAnswer` | `analysis/research_test.go`: 0 → 1 | Missing Research runner/codec | Analysis/results/providerhttp packages pass | Shared context binding and image request encoder; only prompt, schema and metadata differ by mode. |
| `TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults` | `aiResearchExecution_test.go`: 0 → 1 | Research admission rejected; then fixture callback lacked a replayable request body | End-to-end SQLite/provider/history test passes | Shared bounded encoder and context execution. Fixture now replays its already-read body for payload assertions; UNKNOWN fixture impact was checked against 15 source test files. |
| `TestAIResearchRejectsExplicitContextRestriction` | `aiResearchAuthority_test.go`: 0 → 1 | Research bypassed explicit context restriction | Research and existing admission tests pass | One existing restriction now covers both context-bearing modes at admission and current-authority checks (HIGH impact). |
| `TestAIResearchRejectsNeighborDisclosure` | `aiResearchAuthority_test.go`: 1 → 2 | Research admitted neighbors | Jobs package and focused admission tests pass | Scoped Research-only exclusion preserves the existing Context-assisted option. |
| `ResearchReview: shows Research coordinates and estimated error even with a degraded map` | `ResearchReview.test.tsx`: 0 → 1 | Detail decoder rejected Research | Five frontend files / eight tests pass | Reused review geometry and legacy components; v2 radius basis is accepted only for Research. Unknown graph refs were verified in API/decoder/component source. |
| `ResearchLaunch: starts Research by default with the exact hint and no extra confirmation or hidden context` | `ResearchLaunch.test.tsx`: 0 → 1 | Initial mode was Visual | Four frontend files / ten tests pass | Mode changes clear class selections; typed hint is included explicitly. Existing Visual test now explicitly chooses Visual under the accepted new default. |
| `review.TestResearchReferencesAreInertAndExcludeUnsafeDestinations` | `review/references_test.go`: 0 → 1 | Missing presentation projection | Review package passes | Pure URL projection clones source records; no DNS, page fetch or source authority. |
| `TestAIResearchDetailKeepsCoordinatesWithoutExposingUnsafeURL` | `aiResearchReferences_test.go`: 0 → 1 | Credential-bearing URL exposed in detail | Research / legacy-context detail tests pass | Safe projection applies after immutable validation; database payload stays unchanged, unusable URL becomes an inactive empty destination. |
| `ResearchReview: shows answer-provided references as deliberate links beside the proposed coordinates` | `ResearchReview.test.tsx`: 1 → 2 | Review rejected answer sources | Research/evidence/parser tests pass | Existing evidence panel distinguishes answer references from retained input provenance; links are visible and deliberate with no embedded media. |
| `ResearchReview: keeps coordinates with missing or unsafe references and renders source text inertly` | `ResearchReview.test.tsx`: 2 → 3 | Missing optional reference rejected review | Research/evidence/parser tests pass | URL safety is isolated in a pure presentation helper; missing Research refs stay inactive while legacy source resolution remains strict. |
| `ResearchReview: rejects malformed Research source collections before rendering` | `ResearchReview.test.tsx`: 3 → 4 | Invalid source collection passed the DTO guard | Research/detail tests pass | Added a bounded version-specific collection guard and explicit v1/v2 TypeScript variants. |
| `TestAIResearchCarriesOneTenMinuteDeadlineThroughDispatch` | `aiResearchDeadline_test.go`: 0 → 1 | Worker used Visual timeout; then root analysis adapter shortened the deadline | Composed Research deadline test passes | Mode stays server-selected; all analysis/dispatch guards share the worker's absolute deadline. Transport response-wait validation follows separately. |
| `TestAIResearchTimeoutConfigurationIsFiniteAndDefaultsToTenMinutes` | `aiResearchConfig_test.go`: 0 → 1 | Missing configurable duration | Focused configuration tests pass | Reused worker policy and bounded startup parsing; legacy zero-valued policies retain the default. |
| `TestAIResearchHonorsShorterInstallationDeadline` | `aiResearchDeadline_test.go`: 1 → 2 | Ignored three-minute policy | Research deadline tests pass | Resolved the duration only at the worker; downstream layers preserve its earlier absolute deadline. |
| `TestAIResearchFreezesDisplayedAlbumAndCaptureBeforeCatalogChanges` | `aiResearchContext_test.go`: 0 → 1 | Preview omitted values | SQLite-to-provider test passes | Reused immutable selection JSON in admissions; optional values have existing byte limits and no new table. Research freshness checks retain displayed values while rechecking image/selection authority. |
| `ResearchLaunch: shows optional album and recorded capture values and includes only selected classes` | `ResearchLaunch.test.tsx`: 1 → 2 | No displayed optional values | Two launch tests pass | Kept optional context rendering in a small component; exact metadata remains bound to the existing immutable selection token. |
| `researchSelection: bounds optional displayed context and rejects foreign assets and malformed values` | `researchSelection.test.ts`: 0 → 1 | Malformed preview accepted | Selection decoder tests pass | Added bounded optional-field validation at the existing API boundary. |
| `ResearchReview: keeps coordinates and links when an answer source reuses an input source ID` | `ResearchReview.test.tsx`: 4 → 5 | Source collision rejected the proposal | Research/evidence/parser tests pass | Input and answer source identities stay separate without adding authority; the evidence panel renders both matches. |
| `providerhttp.TestResearchGenerationSchemaRequiresAnExplicitPossiblyEmptySourceList` | `research_schema_test.go`: 0 → 1 | Strict generation left an optional property | Codec/results/analysis tests pass | Wire schema requires an explicit empty-capable list; the immutable reader still accepts omitted optional references. |
| `analysis_test.TestResearchRejectsNeighborAuthorizationBeforeReservation` | `research_authority_test.go`: 0 → 1 | Research dispatched with neighbor consent | Analysis package passes | Added the same Research-only boundary at the direct runner; no new abstraction. |
| `TestAIResearchHistorySurvivesReopenSourceLossAndDisabledExecution` | `aiResearchHistory_test.go`: 0 → 1 | Characterization of shared immutable lifecycle with v2 | Research SQLite reopen test passes | No production change; exact v2 links/error survive, foreign reads fail, account deletion cascades. |
| `TestAIResearchRestartNeverResendsAfterDefaultReservation` | `aiResearchRecovery_test.go`: 0 → 1 | Characterization | Real SQLite recovery/reopen passes | No production change; the existing consumed call remains consumed after interrupted delivery. |
| `TestAIResearchHeartbeatPreservesDeadlineAndCancellationFencesLateOutput` | `aiResearchHeartbeat_test.go`: 0 → 1 | Characterization | Go race / real SQLite heartbeat test passes | Existing renewable lease and cancellation fencing work with Research; no production change or wall-clock wait required. |
| `comparison: keeps matched coarse, unknown and failed outcomes without inventing measured error` | `research-comparison.test.mjs`: 0 → 1 | Report module absent | Node test passes | One pure report function preserves every supplied run and binds matched inputs by digest; no quality filter. |
| `comparison: measures independently known camera error separately from the model estimate` | `research-comparison.test.mjs`: 1 → 2 | Missing independent measurement | Two Node tests pass | Isolated great-circle distance calculation; reference source remains explicit and separate from estimated error. |
| `comparison: rejects unmatched or invalid measurements instead of producing a misleading comparison` | `research-comparison.test.mjs`: 2 → 3 | Invalid/unmatched input accepted | Three Node tests pass | Kept validation beside the pure report; rejects invalid measurements without imposing an accuracy cutoff. |
| `comparison: writes a reproducible JSON report from an explicit local input file` | `research-comparison.test.mjs`: 3 → 4 | CLI emitted no report | Four Node tests pass | Thin CLI calls the pure report and emits safe errors; no network or image reads. |
| `providerhttp.TestResearchTransportKeepsAShortTLSHandshakeLimit` | `research_deadline_test.go`: 1 → 2 | Stalled TLS used the parent response deadline | Provider HTTP race tests pass | Set a ten-second TLS handshake limit independently of the longer response wait; dial bounds remain unchanged. |
| `TestAIResearchIdempotencyBindsExactHintWithoutAdditionalWork` | `aiResearchIdempotency_test.go`: 0 → 1 | Characterization | Research admission test passes | Existing configuration digest binds Research hint and selection; no duplicate work or production change. |
| `TestAIResearchPublicationRollsBackAnswerReferencesAndHistoryTogether` | `aiResearchAtomic_test.go`: 0 → 1 | Characterization | Atomic Research publication test passes | Real SQLite trigger failure rolls back answer/history/state together without replay; existing transaction boundary reused. |
| `TestAIResearchAndLegacyAnswersSurviveHistoryMigrationAndReopen` | `aiResearchMigration_test.go`: 0 → 1 | Characterization | Real SQLite history migration and reopen pass for v1 and v2 | Existing migration rebuilds projections without rewriting either immutable answer; no new SQL migration. |
| `ResearchReview: keeps enormous estimated errors compact and numeric coordinates visible` | `ResearchReview.test.tsx`: 5 → 6 | Huge decimal text lacked compact scientific display | Six Research review tests pass | Added display-only scientific notation for very large values; no result filtering or radius cap. |
| `ResearchReview: shows optional answer links even when the candidate omits a local source reference` | `ResearchReview.test.tsx`: 6 → 7 | Optional link hidden in selected-candidate view | Research/evidence tests pass | Research keeps all answer links available beside candidate-specific context, without inventing a reference association. |

| `comparison: requires known reference uncertainty for measured error and retains that uncertainty` | `research-comparison.test.mjs`: 4 → 5 | Measured distance was reported despite unknown reference uncertainty | Five Node tests pass | Keep reference uncertainty separate; unknown reference uncertainty leaves measured error null without losing proposed coordinates. |

## Inline reviews

These are the implementing agent's self-reviews, not independent reviews. Proposal, design and all five delta specs were reread for the final contract checks. The [scenario mapping](verification-plan.md) assesses all 68 scenarios at their actual test boundaries. Each slice was assessed in spec → TDD → quality order; the final cumulative review follows below.

### Contract and task checks

1. **1.1 — PASS:** Dual schema/mode validation retains 500 m, city/region, enormous, null-error and ambiguous results; v1 fixtures and rules remain valid. Evidence: `results/research*_test.go`, real mixed-version migration test.
2. **1.2 — PASS:** Existing admission, codec, executor and atomic history compose Research with one default dispatch. Evidence: `TestAIResearchPublishesCoarsePrivateHistoryWithOrdinaryDefaults`.
3. **1.3 — PASS:** `CandidateFacts`, proposal alternatives and the Research browser journey keep coordinates/error usable at 390 px, by keyboard and without a map; no write action is added.
4. **2.1 — PASS:** Selection JSON freezes displayed optional values, configuration binds the exact hint, and Research rejects neighbors. Evidence: context, idempotency, admission and direct-runner tests.
5. **2.2 — PASS:** V2 sources are bounded answer content; absent/unresolved/unusable links retain otherwise valid coordinates. Source identity never substitutes for input-context authority.
6. **2.3 — PASS:** Owner-scoped detail and inert explicit links preserve existing evidence and account clearing. Source/input ID collisions and unreferenced answer links remain usable.
7. **3.1 — PASS:** Worker, root adapter, analysis and transport preserve the finite parent deadline; 600-second default and shorter policy are tested. Dial/TLS remain short, normal response waits remain bounded.
8. **3.2 — PASS:** Real SQLite tests cover renewal, cancellation, uncertain delivery/restart, idempotency, atomic publication and owner/account isolation with no automatic default retransmission.
9. **3.3 — PASS:** V1 and v2 answers survive real history migration/reopen without rewriting bytes; disabled execution and missing sources do not remove history. Rollback instructions preserve compatible readers and data.
10. **4.1 — PASS:** Local comparison CLI preserves coarse, unknown and failed runs, binds matching inputs and separates estimated error from distance to an independent reference with known uncertainty.
11. **4.2 — PASS:** Authorized one-pixel request reached the actual proxy and returned a semantically valid v2 answer. Its inspected 120-second upstream limit is recorded, not silently changed.
12. **4.3 — PASS:** Executed gates and per-scenario evidence are recorded below. Private-photo accuracy remains unmeasured, and no deployment is claimed.

Requirement and scenario checks: PASS at each mapped automated boundary in `verification-plan.md`. WR03 proves the instruction and accepted alternative response, not actual model reasoning quality; WR11–12 deliberately distinguish synthetic contract checks from live measured accuracy.

Proposal non-goals, assessed separately:

- Search metadata/tool-event storage — PASS: no telemetry contract or table.
- Independent source or accuracy certification — PASS: references remain answer content and estimated errors remain estimates.
- Automatic GPS/Immich writes — PASS: unchanged dependency boundary and explicit zero-write integration/browser assertions.
- Draft acceptance/write approval — PASS: no acceptance route, button or authority introduced.
- New search infrastructure — PASS: existing endpoint and deployment; no new service, SDK or dependency.
- Multi-photo sequence analysis — PASS: exactly one prepared target image per item.
- Implicit neighboring-image disclosure — PASS: Research denies neighbor context at admission and runner boundaries.
- Guaranteed precision — PASS: coarse/null error and alternatives stay visible.
- Mandatory confidence/error threshold — PASS: only structural/numeric resource bounds apply.

Design decisions, assessed separately:

1. Existing request/answer path — PASS: schema-aware shared image codec, ordinary complete-response parser.
2. Additive v2 — PASS: unchanged embedded v1 schema and separate Research validation policy.
3. Answer references — PASS: immutable answer storage, safe presentation, no source fetching.
4. Package ownership — PASS: focused existing AI owners and root SQL/transport composition.
5. Durable simple launch — PASS: initial Research, one Start, exact displayed context, finite deadline and existing reservations.
6. Immutable storage/uncertainty review — PASS: no new source table or result rewrite; alternatives and numeric fallback retained.
7. Existing test boundaries — PASS: Go/race/SQLite, component and built-browser evidence; live compatibility separate.
8. Compatible rollout — PASS: dual readers, mixed-version migration test and non-destructive rollback guidance.

### TDD and per-slice refactoring checks

- Slice 1 — PASS: individual RED/GREEN cycles above; preserved v1 cases use the project's characterization rule. Shared schema selection, image encoding and context execution eliminate duplicate workflows.
- Slice 2 — PASS: individual source/context/launch tests preceded behavior changes. Safe link handling remains at server/browser presentation boundaries; cross-language helpers intentionally use their native URL parsers. No second source-authority mechanism.
- Slice 3 — PASS: deadline/configuration/TLS tests preceded changes; existing durable lifecycle behavior was characterized with real SQLite. The earlier absolute deadline remains authoritative; no renewal-specific timer extension or new retry layer.
- Slice 4 — PASS: each comparison CLI behavior completed its own RED/GREEN cycle, including reference uncertainty. Independent distance calculation is a small pure helper; CLI is a thin local-file adapter. No further extraction is justified.
- Per-test refactor outcomes — PASS: recorded above; no manufactured failure for unchanged shared behavior, skipped tests or mutation-testing requirement.

### Code-quality checks after contract approval

Each item was checked across production code and tests in all four slices where applicable.

- File responsibility — PASS: Research prompt, frozen context inputs, source projection, deadline policy and comparison report each have a focused owner.
- Decomposition — PASS: shared workflows/codecs reused; no parallel Research execution stack.
- File growth — PASS: 888 handwritten files checked against the inherited ratchet; new files remain below 500 lines.
- Naming — PASS: Research mode and v2 policy names agree across producer, storage and UI; existing wire casing retained.
- Error handling — PASS: structural/authority/timeout failures remain safe technical failures; optional links do not invalidate coordinates.
- Test quality — PASS: real adapters/SQLite/browser boundaries and behavior assertions; narrow fakes only at external seams.
- Clean-code principles — PASS: single responsibility, explicit input/owner/deadline bindings, no hidden mutable service locator.
- Refactoring patterns — PASS: shared image encoding, shared context runner, separate presentation projection; no extraction merely to satisfy line limits.
- Code smells — PASS: no duplicated inference pipeline, source audit subsystem, dead superseded branch or unused Research abstraction found.
- Surgical changes — PASS: all changes trace to the twelve tasks; ESLint adds the new local report scripts to the existing tooling rules without weakening application rules.
- Simplicity — PASS: existing dependencies and Go/Next.js/SQLite deployment preserved; one finite duration setting meets the accepted requirement.
- Speculative features — PASS: no search orchestration, source fetcher, writeback or readiness subsystem.
- Comments — PASS: intent/constraint comments only; no narrated test steps or commented-out implementation.
- Testability — PASS: controllable clocks/transport and file-backed SQLite cover timing and persistence independently.
- Readability — PASS: new TypeScript/report files formatted with repository configuration; existing files changed locally.
- Maintainability — PASS: version selection is explicit and legacy fixtures remain authoritative.
- TODO/FIXME/dead-code markers — PASS: none added to implementation.
- Architecture conventions — PASS: pure AI core does not import SQL/HTTP concrete adapters; no transaction spans inference; analysis/review cannot acquire mutation authority.
- Coding conventions — PASS: strict TypeScript, existing naming/formatting, no new dependency or unchecked production `any`, error propagation retained.
- Testing conventions — PASS: individual TDD/characterization, real SQLite, applicable shared gates and measured coverage; no failing/skipped scenario accepted.
- Documentation conventions — PASS: Git documents are English, planning baseline is reconciled, ADR-09 records the accepted change in radius/reference policy.

## Whole-change review

Cross-slice integration:

1. Shared interfaces — PASS: Research mode, schema 2.0, prompt `research-v1`, validation `analysis-result-v2`, frozen context and result DTO agree through real provider/SQLite/UI tests.
2. Naming consistency — PASS: existing candidate/camera/subject concepts retained; answer references remain distinct from input context.
3. Duplicate logic — PASS: encoding and context execution share their existing paths; the legacy v1 schema is intentionally preserved independently.
4. Missed shared abstractions — PASS: no demonstrated shared component remains unextracted; backend/frontend URL presentation checks belong to separate runtime boundaries.
5. Superseded code — PASS: initial source-lookup restriction is replaced only in Research; required legacy behavior remains in place.

Cross-slice refactoring:

- Clean-code principles — PASS: explicit authority and deadline ownership, immutable results, deterministic pure validation.
- Refactoring patterns — PASS: shared codecs/context workflow, projection after validation, finite configuration at the server boundary.
- Code smells — PASS: no source-audit database, duplicate worker or hidden retry introduced; large-radius text and optional-link visibility findings were fixed with tests.
- Project standards — PASS: ownership, source limits, strict types, TDD, SQLite, dependency and coverage checks verified above.

Implementation principles:

1. Surgical — PASS: all app, test, tooling and documentation changes map to the accepted tasks.
2. Simplicity — PASS: no new library/service/proxy protocol; existing durable lifecycle reused.
3. Think first — PASS: v1 compatibility, exact optional context, short TLS versus long response wait and upstream timeout are explicit decisions.

Artifact gaps: none requiring a new product decision. The shorter upstream timeout and unmeasured live geolocation quality are documented integration limits, not implied passing behavior. Critical/important review findings are resolved. Final assessment: ready to commit after successful cumulative verification.

## Executed verification

Base: `58350f7cd8bc6f43d62b6279bd55d479fe3aca36`. Runtime: Node 22.23.2, Go 1.25.14 and Bun 1.4.2. Checks ran on 20–21 September 2026.

| Gate | Actual result |
|---|---|
| Shared checker self-tests | 72 passed |
| Size/ratchet | Passed; 888 files |
| Dependency boundaries | Passed; 331 frontend files, 607 runtime edges and 487 Go files at the final dependency run |
| Go formatting | Passed; 488 files |
| ESLint | Passed; zero errors, three unchanged inherited warnings |
| Route type generation / TypeScript | Passed after correcting the two new test fixtures' explicit DTO guards |
| Go vet | Passed |
| Full Go race/coverage suite | Passed; root package 320.118 s; all AI packages passed; AI statements 88.13% (4,219/4,787), above 80% |
| Final affected Go race checks | Research integration and provider HTTP passed after the final TLS/SQLite tests; root 7.511 s |
| Frontend tests/coverage | 47 files / 100 tests passed; AI lines 94.64%, branches 86.21%, both above 80% |
| Final formatted Research component/decoder tests | 3 files / 10 tests passed |
| Comparison report CLI | 5 Node tests passed, including actual CLI file input and independent reference uncertainty |
| Go and frontend production builds | Passed through the shared smoke prerequisites |
| Built browser suite | 3 AI-disabled legacy journeys and 12 AI-enabled journeys passed; no retries or skips |
| Container builds | Backend and frontend `linux/amd64` images built successfully from a source export excluding `.env` files; final app behavior included |
| OpenSpec | Strict change validation passed; all 68 scenarios mapped |
| Documents | English, local links and whitespace checks passed |
| Live proxy compatibility | One synthetic request: HTTP 200, 8,085 ms, v2 unknown answer semantically valid; upstream 120-second limit inspected |

Initial failures were resolved rather than skipped: schema/context/DTO behavior during TDD; TypeScript fixture narrowing; new report-script ESLint parsing (now covered by existing tooling rules); and the new browser locator matching both alternative errors (assert both alternatives). The first sandboxed Chromium run failed before tests because macOS denied its launch; the same shared suite passed with permitted local execution. An optional formatter was absent locally, so a temporary pinned CLI applied the existing formatting configuration without changing dependencies.

The three inherited ESLint warnings are in `useDawarich.ts`, `HeaderTitle.tsx` and `useOverviewLayerReconcile.ts`. Existing Next.js middleware deprecation and Node SQLite experimental notices remain. Browser fixtures intentionally deny optional Nominatim requests and assert fallback behavior; they are not unexpected network failures.

### Graph review and limitations

GitNexus 1.6.12 indexed this checkout/branch at the stated base: 10,866 nodes, 32,008 edges and 568 flows. Required upstream impacts were run before symbol edits; shared validator/dispatcher/root integration risk was HIGH/CRITICAL and was reported before editing.

The final MCP change call lost its transport and its concurrent cycle call exhausted the native buffer pool. The working CLI completed change analysis and a complete cycle enumeration with zero cycles. To avoid the CLI's abbreviated display, the same installed `LocalBackend.callTool('detect_changes', ...)` result was captured in full under ignored `out/checks/research-detect-complete.json`, explicitly bound to `immich-places-ai-addon`; it contains no `partial`, `truncated` or error flag. The structured run identified 110 changed files and CRITICAL affected behavior, including production admission/execution, result validation/history/thumbnail reads, provider dispatch/deadlines, launch and review.

Graph symbol/process counts varied across queries and the index reports receiver/cross-language and process-enumeration limits. These counts are discovery evidence, not proof of complete runtime reachability. Source review covered the complete Git diff and new files; the installed dependency checker, Go compiler/race/SQLite tests and built browser suite verify the affected boundaries independently. No zero/UNKNOWN/capped graph result was treated as safety evidence. No application import cycle was found by either the complete graph cycle check or the source dependency gate.

No private matched-photo geolocation benchmark or deployment was performed. Longer live requests are still limited by the existing upstream 120-second configuration. Local verification does not prove live web-search availability or accuracy.

## 24 September rollout and live comparison

Revision `46e8fa8` was pushed and deployed to the isolated NAS preview; both
containers became healthy, SQLite was backed up and the original Places and
proxy containers remained unchanged. See the [rollout record](../../../docs/engineering/nas-research-preview-2026-09-24.md).

The separately authorized [three-photo comparison](../../../docs/engineering/research-live-comparison-2026-09-24.md)
completed six real requests. All Research wire responses failed with commentary
and duplicate JSON, although their isolated documents passed the unchanged v2
validator. This is a confirmed existing proxy aggregation defect, not a successful
application result or an accuracy failure. Keep the change active pending the
shared-proxy follow-up; preserve the passing deterministic gates above without
claiming the live integration gate passed. No private inputs or answers are
committed. Maintained specs are synchronized and strictly valid.
