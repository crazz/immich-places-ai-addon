# CH18 implementation verification

Status: implemented, verified and synchronized on 25 September 2026. All seven tasks and P01–P18 are complete. Inline reviews; no independent/subagent review claimed.

Base: `5543618ff8825c97012e9f23f5efaebb4721dc1f`, branch `codex/plan-ai-selection-and-validation`, same local checkout. Inline implementation and reviews; no subagents. The unrelated `.brooks-lint-history.json` is excluded.

## Workflow and conventions

- [x] Preflight: all change artifacts, CH16 contract/evidence, engineering standards and relevant product/design/reconciliation/roadmap/batch sections read. OpenSpec reports ready, 0/7 tasks.
- [x] Dependency analysis: sequential slices 1 → 2 → 3. Persistence required by the first usable preview is introduced with slice 1, then hardened in slice 3. Existing route/editor integration files verified in Git; `internal/ai/writepreview`, preview adapters and migration 027 are new.
- [x] Fresh exact comparison and accessible presentation.
- [x] Source/GPS recovery and concurrent authority fencing.
- [x] Durable lifecycle, bounds and private response fencing.
- [x] Whole-change spec, TDD and quality review, full gates and synchronization; authorized commit/handoff closure.

Conventions: English; 500 physical lines; focused pure AI core and root I/O adapters; AI-only frontend state; consumer-owned read/store interfaces; exact owner/installation/revision scope; no provider or mutation interface; immutable original analyses and plans; explicit nullable/zero coordinates; no quality threshold; no manual pending-state changes; bounded requests/reads/storage; short SQLite transactions with no network inside; per-test RED/GREEN/refactor or explicit existing-behavior characterization; real file-backed SQLite, deterministic local HTTP and built browser fixtures; measured AI coverage and all thirteen installed gates against the base. No deployment, push, archive or live calls.

## Graph evidence

GitNexus is bound to `immich-places-ai-addon` at this checkout and base commit; context reports no incomplete index reasons. Bounded process enumeration and unresolved cross-language/dynamic edges still require source/tests. `registerAIDraftRoutes` impact: LOW, one direct caller `newAIResultHandler`, then `main`; existing protected handler/session/no-store path verified. No new writer dependency is introduced.

## Development evidence

Logs are retained locally under ignored `out/checks/ch18/`, using the pinned environment copied from CH16. The test ledger below records actual results as work proceeds.

P01 (`aiWritePreview_test.go`, 0 → 1): route 404 RED, exact GPS response GREEN with all draft tests. Refactor assessment: separate pure fixed-field plan, workflow and request-scoped read adapter; no mutation/provider contract. P02 begins after that pass. New `Create` is not indexed yet (UNKNOWN); source confirms its only production caller is the preview HTTP adapter and its read dependencies are the new session methods. New-symbol gaps remain explicit until index refresh.

P02 (1 → 2 preview tests): unstaged draft returned 200 RED; exact state/revision/body validation GREEN (`p02-*.log`). Strict duplicate/unknown-field decoding is reused at a 4 KiB boundary. Refactor: core validation owns readiness; HTTP owns safe status translation. New route/Build impact is UNKNOWN until reindex; direct local source traces confirm route → Create → Build only.

P07 (new persistence test 0 → 1): missing preview table RED; migration 027, canonical bytes/digest, immutable SQL payload and restart GET GREEN (`p07-*.log`). Refactor assessment: persistence owns the short final revalidation transaction; the per-request adapter retains credential authority privately, never in core plans. `ReadDraft` is newly unindexed; source confirms only the new workflow reads it. Existing draft suites remain green after the additive migration.

P04 (new conflict test 0 → 1): intervening GPS returned a usable plan RED; exact nullable comparison with owned before/current/proposed values and explicit CH16 review/restage recovery GREEN (`p04-*.log`). Refactor assessment: comparison and its safe error payload remain pure; HTTP only translates the result. No baseline or draft revision is silently changed.

P05 (1 → 2 conflict tests): unrelated timestamp/description changes already passed; changed checksum incorrectly created a plan RED. Separate reviewed-image comparison GREEN (`p05-*.log`). Refactor: material-source conflict precedes GPS comparison, retaining analysis provenance. `Compare` graph gap confirmed against its single workflow caller.

P03 (new GPS test 0 → 1): absent EXIF metadata incorrectly returned 404 RED; absent/null/empty/partial/zero values and signed-zero normalization GREEN (`p03-*.log`). Refactor: normalization copies nullable values, preserving input ownership. Shared `aiDraftStore.source` impact HIGH: observe/acknowledge and baseline routes. Warning issued before changing the absent-EXIF branch; all CH16 draft tests pass.

P10 (1 → 2 GPS tests): already-matching GPS labeled changed RED; shared pure diff computation in creation and reload GREEN (`p10-*.log`). Refactor: one comparison implementation, no mutation/completed-write status. New store `get` graph gap confirmed by its sole HTTP GET caller.

P06 (new source test 0 → 1): malformed/upstream failures incorrectly conflated with private missing references RED; SOURCE_UNAVAILABLE 503 GREEN for wrong shape, duplicates/types/ranges, oversize, redirects, denial, invalid identity and unavailable visibility (`p06-*.log`). Local missing authority remains 404. Refactor: reuse bounded nonredirecting single-attempt transport; classify only at preview boundary. GitNexus refresh passed; ReadMetadata LOW lower-bound (interface dispatch), verified against the concrete workflow/HTTP chain; error mapper HIGH (shared route flows), warned before edit.

P11/P12 (new lifecycle test 0 → 1): expired/edited/rejected/invalidated previews returned usable RED; persisted status checks GREEN without changing bytes, digest or expiry (`p11-p12-*.log`). Fresh replacement performs a new upstream read. Refactor: status is a read projection, separate from immutable plan. Store get impact HIGH with unresolved generic receiver calls; source verifies preview-specific HTTP caller, and lifecycle tests exercise the actual path. Publish impact LOW, exact workflow/route callers.

P13 (1 → 2 lifecycle tests): eleventh active preview accepted RED; ten-preview cap and at-most-100 expired unprotected cleanup GREEN (`p13-*.log`). The retained preview protects its revision, and history survives. Refactor assessment: capacity and cleanup share the existing publication transaction. `protected` is a lifecycle guard for the subsequent confirmed-write/audit consumer, with no executor introduced.

P08 (new race test 0 → 1): characterization of publication fences PASS under Go race instrumentation (`p08-characterization.log`). Barriers hold the real metadata request while draft edit/rejection, account deletion, credential rotation, installation rotation, hidden status or insert failure occurs; zero rows publish. Refactor: existing short transaction and source authority checks suffice; no extra lock or retry.

Canonical encoding (new pure test 0 → 1): 16 KiB bound missing RED → GREEN; stable signed-zero-normalized bytes/digest and unchanged inputs verified. Refactor: fixed struct order and a single payload limit check. P09 (new isolation test 0 → 1): existing private GET/body/origin/session guards characterized PASS; foreign/missing IDs return the same safe code, substitutions cannot alter the saved digest, and responses are no-store. No refactor needed; the owner-qualified lookup is the boundary.

P17/P18 (1 → 2 isolation tests): disabled-execution creation/read, account cascade, unrelated owner preservation, catalog reset, and old/new installation fencing characterized PASS. Initial test compile used a nonexistent fixture constant; corrected to the existing ID helper before the actual characterization run. Refactor: existing account FKs and installation binding suffice; no new retention/deletion mechanism.

Frontend preflight: existing DraftReview/DraftEditor/Baseline/API, result keying and test patterns read. DraftReview impact LOW through ResultDetail and AI Results navigation. Conventions above apply; typed unknown-boundary parsing, accessible numeric values, no manual mutations, abort/key fencing and exact request DTOs are required. UI follows the completed backend dependency before slice review.

P14/UI (new component test 0 → 1): missing preview control RED; exact numeric before/proposed values, zero/absent distinction, single-photo GPS scope, saved revision/status/digest, coarse decision and no mutation control GREEN (`ui-numeric-*.log`). Refactor: typed DTO validation, API and presentation separated; shared draft editor remains independent.

UI reload (1 → 2 component tests): retained ID did not fetch a stored comparison RED; private owner/draft/revision-keyed ID-only browser persistence and GET restoration GREEN (`ui-reload-*.log`). Refactor: shared response-scope verification; no coordinates or credentials stored in browser storage. Newly unindexed PreviewSession callers confirmed in DraftReview/WritePreview source.

P15/UI (2 → 3 component tests): a newly overlapping manual edit left a usable comparison visible RED; blocking/retirement aborts pending preview work and clears the reference GREEN (`ui-overlap-*.log`). Unsaved editor changes retire it too. Refactor: read-only overlap input, no shared pending-state writes; explicit new preview is required after resolution.

UI expiry (3 → 4 component tests): visible status stayed usable past five minutes RED; injected browser clock verifies expiry without a request/renewal GREEN (`ui-expiry-*.log`). Refactor: timer is scoped to the displayed immutable plan and cleaned up on replacement/unmount; backend time remains the authority.

UI conflict (4 → 5 component tests): generic failure lost the before/current/proposed conflict RED; validated private conflict DTO and explicit baseline-review/restage guidance GREEN (`ui-conflict-*.log`). Refactor: specialized preview error decoding leaves shared request behavior unchanged; all messages are local safe text.

UI revision recovery (5 → 6 component tests): missing saved-state recovery control RED; comparison delegates to CH16's explicit keep/discard reconciliation GREEN (`ui-revision-*.log`). Refactor: reuses the existing compare workflow; local coordinates remain intact.

P16/UI (6 → 7 component tests): existing abort/key/overlap fences characterized PASS for late create responses after result change, account change and overlapping manual work (`ui-late-characterization.log`). Refactor: no additional state layer required; ID-only storage is scoped and removed on navigation.

Scoped ESLint found invalid object-label property names and a missing fixture return type. Converted presentation rows to tuples and recovery messages to a switch, typed the fixture and retained the same tests. The first build orchestration attempted two `--gate` flags; the runner correctly rejected the unsupported invocation. Separate Go/frontend production build gates both pass. These setup/check failures are not behavioral RED evidence.

Built preview journey RED: real proxy POST `/ai/write-previews` returned 404 (trace confirms missing composition registration). `main` impact UNKNOWN is the Go executable entrypoint; source confirms the protected result-handler mux owns draft routes but had no preview prefixes. Added only the two preview path registrations. No preview executor or manual path is connected.

Migration characterization PASS (`migration-characterization.log`): fresh test databases, CH16 026 upgrade/reopen/repeat, transactional failure at a conflicting table, retained-table legacy analysis reads and owner-qualified FKs across four pooled connections. Readiness characterization PASS (`readiness-characterization.log`): missing/rejected/unstaged/invalid/expanded/unreviewed inputs fail without quality or non-GPS gates. Refactor assessments: existing migration harness and narrow readiness policy suffice.

Built browser GREEN: the real frontend/proxy/backend/SQLite preview journey passes after composition registration (`browser-preview-green.log`). Reload uses GET only with unchanged digest/expiry; a test-only synthetic external GPS edit produces conflict, explicit review/restaging produces an unchanged comparison, and the application performs zero writes/provider calls. The 390px screenshot was inspected: numeric fields, scope, status, expiry and digest are readable without horizontal overflow or map tiles.

Final edge checks: replacement-reference test RED (old ID retained after conflict) → GREEN after clearing browser storage before a new comparison. Response-scope test RED (same revision but substituted coordinates accepted) → GREEN after comparing reviewed image identity, nullable baseline and intended pair against the saved draft. Refactor assessment: reuse the existing private response boundary; no new state/transport abstraction. New-symbol GitNexus results were UNKNOWN; source confirmed the two API helpers call `checkScope` and the keyed wrapper calls `PreviewSession`. Deadline/cancellation characterization PASS against a real stalled HTTP service: one attempt, bounded 10-second read, no plan, unchanged draft. Late stored-GET characterization PASS after private navigation. The protected error-envelope test RED (nonstandard requestId) → GREEN with the existing requestID convention; mapper impact HIGH, warning issued and protected route callers reviewed. A new mock's inferred empty tuple failed TypeScript; corrected the fixture mock declaration, then lint/types pass. No production behavior or type rule was weakened.

## Inline slice reviews

These are self-reviews under the requested inline mode, not independent reviews. The proposal, full design and relevant requirements/scenarios were read for the contract checks; the TDD ledger and standards were then checked before quality approval. Persistence spans the first usable comparison and the lifecycle slice because every preview is durable from its introduction.

| Slice | Spec-compliance verdict and evidence | TDD verdict | Quality verdict |
|---|---|---|---|
| 1: tasks 1.1–1.2 | PASS: staged exact GPS, no precision/heading gate, bounded source validation, numeric comparison and manual-state separation; P01–P03/P06/P10/P14–P16 mapped below | PASS: route, readiness, GPS, diff and component/browser RED→GREEN records; inherited behavior characterized explicitly | PASS: pure consumer-owned workflow, request-local adapter, typed AI-only panel; no writer/provider/manual dependency |
| 2: tasks 2.1–2.2 | PASS: image/GPS conflicts keep the draft and require CH16 review; obsolete revision/key/owner/install cannot publish; P04–P06/P08–P09/P18 | PASS: conflict/source tests RED→GREEN; existing transaction fences pass race characterization | PASS: shared read decoder preserves original image provenance; safe conflict DTO and short final transaction |
| 3: tasks 3.1–3.3 | PASS: same canonical plan after reopen, private scope, expiry/capacity/protected history and disabled reads; P07–P09/P11–P13/P16–P18 | PASS: persistence/lifecycle/capacity RED→GREEN; migrations/lifecycle characterization; UI retirement and boundary regressions individually verified | PASS: immutable payload separated from lifecycle projection; bounded cleanup and ID-only browser storage; late replies abort/fence |

Per-slice checks PASS after fixes: backend draft/preview/migration tests with race detector (`backend-slices.log`, 68.514s); 15 focused frontend tests (`ui-slices.log`); scoped ESLint (`lint-slices*.log`), typegen/types (`types-slices-corrected.log`), Go formatting (`gofmt-slices.log`), production builds and the preview browser journey. No tests skipped. The type-check fixture failure and earlier build-command misuse are recorded check failures, not behavioral RED evidence.

Quality checklist, applied to production and tests across each slice:

| Concern | Relevance and verdict |
|---|---|
| File responsibility | Relevant, PASS: core plan/validation/workflow, HTTP/read/store adapters, DTO/API/panel, and behavior-specific tests |
| Decomposition | Relevant, PASS: interfaces at read/store seams; no SQL/network in pure core |
| File growth | Relevant, PASS: no new file approaches 500 lines; inherited main/test integrations are small |
| Naming | Relevant, PASS: draft revision, baseline, plan, preview and usable/expired/stale match the contract |
| Error handling | Relevant, PASS: protected safe envelope including requestID; malformed/foreign/conflict/unavailable remain distinct |
| Test quality | Relevant, PASS: real file-backed SQLite, real bounded HTTP and built app; injected time/barriers prove races without private/live inputs |
| Clean code principles | Relevant, PASS: single responsibility, explicit scope/nullable values, consumer-owned I/O and one GPS comparison policy |
| Refactoring patterns | Relevant, PASS: shared source reader and exact-diff function reused; DTO validation extracted at its actual boundary |
| Code smells | Relevant, PASS: no speculative services, duplicate workflow, dead earlier implementation or broad mutable state found |
| Project conventions | Relevant, PASS: English, strict types, safe Go errors, current deployment, ordered additive migration, 500-line rule, AI package ownership and no network transaction |
| Surgical changes | Relevant, PASS: every integration change serves preview registration, missing EXIF semantics, migration tip or UI/browser verification |
| Simplicity | Relevant, PASS: existing libraries/clock/auth/transport; no new configuration/dependency or executor |
| Speculative features | Relevant, PASS: protected-preview retention is explicitly required by P13; no confirmation/write control |
| Comments | Relevant, PASS: storage catch comments explain the browser restriction; no narrated tests or commented-out code |
| Testability | Relevant, PASS: clock/client/storage/cancellation supplied at actual seams |
| Readability | Relevant, PASS: bounded adapter methods and separate plan/validation/presentation responsibilities |
| Maintainability | Relevant, PASS: exact durable DTO and migration ownership documented for CH19 |
| TODO/FIXME/skips | Relevant, PASS: none introduced |

Cross-task refactor assessment: no further extraction is needed. The same nullable comparison drives creation/reload diff, shared source parsing owns GPS validation, and API validation is centralized. Browser reference retirement fixed a real lifecycle overlap; no redundant plan or manual-state store remains.

## Scenario-to-test map

| Scenario | Concrete automated evidence |
|---|---|
| P01 | `TestAIWritePreviewExactGPSWithoutQualityGateOrMutation`; `shows an exact numeric GPS comparison for a coarse staged decision without a mutation control` |
| P02 | `TestAIWritePreviewRejectsIncompleteExpandedAndObsoleteInput`; `TestReadinessRequiresOnlyExactStagedGPSAndAcknowledgedBaseline` |
| P03 | `TestAIWritePreviewPreservesAbsentPartialAndZeroGPS` |
| P04 | `TestAIWritePreviewGPSConflictRequiresExplicitBaselineReview`; built preview/conflict journey; conflict component test |
| P05 | `TestAIWritePreviewDistinguishesMaterialImageFromMetadataChanges` |
| P06 | `TestAIWritePreviewRejectsMalformedBoundedAndRedirectedSource`; `TestAIWritePreviewBoundsSourceDeadlineAndCancellation` |
| P07 | `TestAIWritePreviewPersistsCanonicalPlanAndReloadsWithoutRead`; stored-ID restoration component test; browser reload |
| P08 | `TestAIWritePreviewFencesConcurrentDraftAuthorityAndStorageChanges` |
| P09 | `TestAIWritePreviewPrivateReferencesCannotSubstituteScope`; `rejects malformed or substituted comparisons at the private response boundary` |
| P10 | `TestAIWritePreviewAlreadyMatchingGPSIsUnchangedAcrossReload`; built preview/conflict journey |
| P11 | `TestAIWritePreviewExpiresAndInvalidatesWithoutChangingSavedPlan`; five-minute UI expiry test |
| P12 | Same backend lifecycle test; editor-dirty retirement component test |
| P13 | `TestAIWritePreviewBoundsCapacityAndCleanupProtectsRetainedHistory` |
| P14 | Built `previews exact GPS across reload and explicit conflict review without AI writes`; keyboard/numeric component test |
| P15 | `retires a comparison when manual work overlaps or the editor becomes dirty, preserving both decisions`; existing draft/manual browser journey |
| P16 | `fences late private replies across result changes, account changes and a new manual overlap`; `discards an in-flight stored preview read when the private view changes` |
| P17 | `TestAIWritePreviewLifecycleKeepsOtherOwnersAndDisabledReads` |
| P18 | Same lifecycle test; concurrent authority/deletion test; private references test |

## Whole-change review

PASS after local fixes, before the cumulative gate. Shared Go/TypeScript interfaces and error envelopes agree; canonical bytes displayed and stored identify the same plan. Naming is consistent. Nullable comparison and read decoding are reused; there is no superseded workflow, missing shared abstraction or executor. Clean code/decomposition, ownership, privacy, TDD and migration conventions above apply across all three slices. All code traces to tasks, with no emergent configuration or dependency. Composition registration and browser reference retirement were made explicit and verified. No proposal/spec/design/task gap requires changing the approved contract. Live compatibility remains outside synthetic verification.

## Final graph and synchronization evidence

GitNexus force rebuild PASS after an incremental general query exposed malformed symbol paths. The final source-bound lookup resolves the shared source reader to CH16 observe/acknowledge plus preview ReadMetadata, and resolves the preview validator through the shared callback-based parseJSON. `detect_changes(scope=all)` sees all 43 staged files and returns all 262 changed symbols, 66 affected process records, with neither partial nor truncated flags. Aggregate risk is CRITICAL and was reported; affected baseline, handler/composition and draft UI paths were inspected and covered by focused/full regressions. Unexpected legacy frontend process names arise from the shared validator callback: source confirms only the preview API supplies isWritePreview; no legacy caller/import was changed. The complete returned graph report is retained locally as `gitnexus-detect-changes.json`.

The graph's overall process extraction remains bounded (602 reported flows, 725 unranked entry candidates, 67 budget-cut walks, 3 depth caps and 1503 dropped callees); 417 cross-language property sites cannot be linked. These are explicit analysis limits, not proof of no additional impact. Interface/generic receiver gaps were verified against source and actual protected-handler tests. `route_map`, `api_impact` and `shape_check` cannot resolve the Go preview routes, so typed response-boundary tests and the real proxy/browser journey supply route-contract evidence. The import cycle check is complete and clean (zero cycles/components); source-based dependency checks remain an independent required gate.

OpenSpec synchronization PASS: the approved Purpose plus all six added requirements and eighteen P01–P18 scenarios are now in `openspec/specs/ai-immich-writeback/spec.md` under a single Requirements heading. No other capability was changed and CH19 remains planned. Strict spec validation and strict all-item validation pass (12 items). Changed Markdown local links, exact delta/main equivalence, named backend evidence and diff whitespace checks pass. The change remains active; no archive performed.

## Final executed gates

`bun run check --base 5543618ff8825c97012e9f23f5efaebb4721dc1f` completed with exit 0 (`out/checks/ch18/full-check.log`). All thirteen installed gates PASS: checker regressions, size, dependencies, Go format, ESLint, route type generation, TypeScript, Go vet, race-enabled Go tests/AI coverage, Go build, frontend tests/AI coverage, frontend build and browser smoke.

| Verification | Actual result |
|---|---|
| Check harness | 72 regression tests passed; no skipped tests |
| Size/dependency/format/type gates | PASS; 955 handwritten files inspected for size and 528 Go files checked for formatting |
| Backend race/coverage | PASS for root and all twelve internal packages; root suite 396.327s. AI statements 88.10% (4716/5353), above 80% |
| Frontend unit/component coverage | 53 test files and 124 tests passed. AI lines 93.81% (1047/1116) and branches 84.59% (1620/1915), both above 80% |
| Production builds | Go and Next.js build gates PASS |
| Built browser smoke | 3 legacy auth/manual/GPX journeys passed (6.8s), then all 15 AI journeys passed (1.2m), including CH18's reload/conflict/narrow-screen comparison |
| Migration/rollback contract | Fresh migration, 026 upgrade/reopen/repeat, failed upgrade transaction, retained-table legacy reads and four pooled foreign-key checks PASS on disposable file-backed SQLite |
| Backend container | `docker build -f backend/Dockerfile -t immich-places-ch18-check backend` PASS; local image only, no push/deployment |
| OpenSpec/documents | Strict main-spec and all-item validation PASS (12 items); exact 6-requirement/18-scenario sync, changed local links and whitespace checks PASS |
| Graph | Complete uncommitted-change listing inspected; complete zero-cycle check. Bounded process/unsupported-route limitations are recorded above |

Nonblocking inherited warnings remain in Dawarich, HeaderTitle and the manual overview hook; none is a CH18 ESLint error. The Next.js middleware convention and synthetic fixture runtime print existing deprecation/experimental notices. No check was skipped or weakened. Earlier failures are recorded in the development ledger and were corrected before this successful run.

The operator handoff is `docs/ai-write-previews.md`. Source/tests were unchanged after the final suite; only verification/status documentation was finalized. The unrelated `.brooks-lint-history.json` stays untracked and untouched. No archive, push, deployment, live provider call, private photo or live Immich mutation occurred. GATE-02 remains unverified; CH19 must keep ordinary dispatch default-off until a separately authorized disposable-photo compatibility run.
