# CH10 and CH12–CH15 planning batch

Status: CH10, CH12, CH13 and CH14 implemented and archived; CH15 planned. Prepared on 20 September 2026 against `23c6ca6` on `codex/plan-ai-selection-and-validation`. All five changes contain proposal, design, delta specs and vertical-slice tasks; CH10/CH12/CH13/CH14 tasks are complete and CH15 tasks remain unchecked. OpenSpec Plus artifact reviews were performed inline, honoring the explicit instruction to use no subagents.

## Changes and application order

| Order | Change | Outcome | Requirements / scenarios / tasks |
|---|---|---|---|
| 1 | [CH10: consented context](../../openspec/changes/archive/2026-09-20-add-consented-ai-context/proposal.md) | Internal bounded metadata-only Context-assisted analysis | 6 / 18 / 8 |
| 2 | [CH12: durable Visual integration](../../openspec/changes/archive/2026-09-20-connect-visual-analysis-to-durable-jobs/proposal.md) | Authorized real execution with finite resource reservations | 10 / 34 / 9 |
| 3 | [CH13: batch progress and reruns](../../openspec/changes/archive/2026-09-20-add-ai-batch-progress-and-reruns/proposal.md) | Launch both modes, follow/cancel work and create explicit new runs | 8 / 24 / 10 |
| 4 | [CH14: persistent results](../../openspec/changes/archive/2026-09-20-browse-persistent-ai-results/proposal.md) | Private terminal history independent of Missing GPS | 5 / 15 / 7 |
| 5 | [CH15: proposal inspection](../../openspec/changes/review-ai-camera-and-subject-proposals/proposal.md) | Read-only camera/subject, evidence and language presentation | 6 / 17 / 8 |

Totals count each delta occurrence, including deliberately repeated modified requirements: 35 requirements, 108 scenarios and 42 tasks. CH14 technically requires CH12 rather than CH13; the sequence above keeps this batch linear. Finish verification and sync/archive each predecessor before applying a conflicting successor. CH13's modified job requirements preserve CH12's full replacement requirements and inherited scenarios. CH14 creates `ai-results-and-review`; CH15 adds to it afterward. Do not apply CH15 against a main spec where CH14 has not been synchronized.

## Settled integration choices

- CH10 admits four independently consented context classes, at most six metadata-only neighbors, bounded hint/album text and honest capture-time/lineage status. Visual stays context-free.
- CH12 requires explicit provider token-policy readiness; CH03's synthetic image/format observations do not establish token-limit support. Reservations are conservative and durable; costs remain estimates. Accepted jobs are independent of expiring preview snapshots.
- CH13 records the exact frozen context used by an attempt and revalidates completion against it. Changed evidence requires explicit reanalysis. Retry-failed and reanalysis create new linked runs with fresh preview/consent.
- CH14 keeps terminal history readable while execution is disabled. Capture date and **Selected album at launch** filters use retained provenance, with unknown fields left unknown. Current image access remains separately authorized.
- CH15 uses an isolated read-only map composition; existing manual click/drop/save behavior cannot receive AI proposals. Canonical descriptions retain `complete`/`unavailable` status and their actual scene/candidate basis.

No new library is proposed. Existing approved JSON Schema and image libraries remain in place. Draft editing/acceptance belongs to CH16, translation operations to CH17, writes to CH18–CH22 and deletion/retention to CH23–CH24.

## Requirement-to-task and verification trace

Every scenario under a mapped requirement must become an automated behavior/authorization/failure test during apply. The layers below identify where the claim is established; this is a test plan, not executed application-test evidence. Tests for inherited behavior may pass immediately. Real SQLite is required for persistence/concurrency claims. Shared engineering gates, Plus TDD, GitNexus pre-edit impact and pre-commit analysis remain required during implementation.

| Change | Requirement | Tasks | Primary verification layers |
|---|---|---|---|
| CH10 | Bind context disclosure to explicit classes and current authority | 1.1 | Policy, authenticated adapter |
| CH10 | Preserve honest capture-time and bounded context semantics | 1.2 | Policy, serialization |
| CH10 | Admit only eligible context sources with honest lineage | 1.3 | Policy, SQLite, read-only HTTP fixtures |
| CH10 | Recheck frozen context before transmission and publication | 2.1 | Adapter, concurrency/cancellation |
| CH10 | Validate Context-assisted output against the exact supplied evidence | 2.2, 2.3 | Canonical validator, synthetic provider |
| CH10 | Return bounded contextual provenance without persistent side effects | 3.1, 3.2 | Ownership/copy semantics, no-write boundary |
| CH12 | Admit exact consented production selections | 1.2 | SQLite atomicity/idempotency, protected API |
| CH12 | Protect bounded job progress and cancellation | 1.3, 3.3 | Protected API, pagination, disabled execution |
| CH12 | Execute real Visual work under current authority | 2.1, 2.3 | Composed synthetic provider/Immich, SQLite |
| CH12 | Reserve finite usage before every production call | 2.2 | SQLite concurrency, timeout/restart fixtures |
| CH12 | Start and stop a bounded recoverable production consumer | 3.1, 3.2 | Lifecycle, restart, upgrade, snapshot cleanup |
| CH12 | Freeze bounded private submissions atomically | 1.2 | Real SQLite rollback/reopen |
| CH12 | Recover and cancel without late publication | 3.1, 3.3 | Concurrent lease/cancellation fixtures |
| CH12 | Reuse installation identity and provide controlled lifecycle cleanup | 3.2, 3.3 | SQLite migration/account/installation lifecycle |
| CH12 | Require explicit production token compatibility policy | 1.1, 2.2 | Configuration, private readiness API, serialization |
| CH12 | Minimize and explicitly format the Visual request | 2.1, 2.2 | Exact provider payload, canonical validator |
| CH13 | Preview the user's exact batch intention | 1.1 | Pure transitions, RTL, browser, protected API |
| CH13 | Confirm explicit mode and data disclosure | 1.2, 1.3 | RTL/browser, server admission |
| CH13 | Persist and validate the exact contextual attempt | 2.1, 2.2 | SQLite/restart, canonical validation, provider fixture |
| CH13 | Follow durable progress across navigation | 1.3, 3.1 | Hook/DTO, RTL, browser reload/offline |
| CH13 | Cancel and rerun without replacing history | 3.2, 4.1, 4.2 | Protected API, SQLite, browser |
| CH13 | Keep launch and progress accessible without write authority | 3.3 | Keyboard RTL/browser, no-write regression |
| CH13 | Freeze bounded private submissions atomically | 1.3 | SQLite mode/consent transaction |
| CH13 | Admit exact consented production selections | 1.3 | Protected API, SQLite idempotency |
| CH14 | Browse private durable terminal history | 1.1 | SQLite reopen/isolation, protected API |
| CH14 | Filter and paginate retained history honestly | 1.2 | SQLite keyset/watermark, date/provenance fixtures |
| CH14 | Keep immutable runs and independent states | 2.1, 2.2 | SQLite/detail DTO, RTL |
| CH14 | Authorize current images independently of stored results | 2.3 | Authenticated image fixture, account-switch browser |
| CH14 | Navigate usable history without side effects | 3.1, 3.2 | RTL/browser, inert text, network assertions |
| CH15 | Inspect separate camera subject and alternative locations | 1.1 | Pure view models, RTL, map browser |
| CH15 | Present uncertainty without overstating precision | 1.2 | Geometry view models, numeric/map fixtures |
| CH15 | Keep camera heading distinct from subject bearing | 1.3 | View models, RTL/map browser |
| CH15 | Explain evidence and lineage without fabricated verification | 2.1 | Reference mapping, inert text, RTL |
| CH15 | Display exact requested-language descriptions | 2.2 | Canonical DTO/view models, keyboard tabs |
| CH15 | Preserve accessible read-only review under degraded maps | 3.1, 3.2, 3.3 | Keyboard/degraded browser, no-write/manual regression |

## Inline artifact review and document validation

Reviewed proposal scope/non-goals, design ownership/failure/persistence/rollout boundaries, spec consistency and task coverage/slicing for all five changes. No subagent or independent review is claimed. Findings corrected during review:

1. Removed CH11's superseded internal-only/public-executor exclusions through full CH12 modified requirement blocks while retaining inherited scenarios.
2. Made CH12 token compatibility explicit rather than inferring optional parameter support from CH03 observations; separated attested policy, reservations, reported usage and estimated cost.
3. Preserved CH12 admission scenarios in CH13 and documented mode-specific durable revalidation instead of using Visual-only completion for Context output.
4. Separated invalid album bindings from optional lost album membership in CH10 scenarios.
5. Added accepted-job execution after preview cleanup and a monotonic terminal-history watermark for stable CH14 pagination under timestamp ties/clock changes.
6. Corrected CH15 reference namespaces (`evidence_refs` versus `source_refs`), canonical language statuses and description-to-candidate basis; isolated manual map actions.

Validation: every change passes `openspec validate <change> --strict --no-interactive`; `openspec validate --all --strict --no-interactive` passes all 10 current specs/changes. Ordered delta composition preserves all modified targets and inherited scenarios. All 108 scenarios have Given/When/Then, every requirement maps above to tasks and verification layers, local link targets resolve, and planning documents contain no non-English text or unresolved placeholders. Local CLI instruction/status snapshots and document-check evidence are in ignored `out/checks/ch10-ch15-planning/`.

Application code, main capability specs and dependencies were not modified by this planning batch. No application suite, migration, provider/private-photo request, Immich write, deployment or performance/quality benchmark was run. Implementation proceeds against explicit contracts and synthetic fixtures; live activation still requires the applicable GATE-02/03 compatibility evidence and GATE-04 destination/retention/reference-data decisions. Proposed retention durations remain unapproved and are not activated here.
