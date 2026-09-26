# CH17 → CH20 → CH21 → CH22 planning batch

Status: **CH17, CH20, CH21 and CH22 implemented, verified and synchronized**. OpenSpec Plus Apply follows the requested order. The main agent owns backend work and final verification; the later user instruction permits two Astra Extra High helpers for stable safety review and frontend checkpoints. Original planning evidence follows. Prepared on 25 September 2026 against `0c5bc0c0a4e2fb46b21efa332921ed968dc49209` on `codex/plan-ai-selection-and-validation`. The user selected CH17 plus CH20–CH22 while preparing the Places login for CH19's pending live gate. Planning neither closes that gate nor authorizes another live action.

The installed `openspec-propose` workflow was combined with OpenSpec Plus proposal, design, spec and tasks disciplines (Plus 1.6.1, the latest release observed during preparation). Reviews were performed inline under the user's continuing no-subagent preference; no independent/subagent review is claimed. Product purpose, non-goals and engineering boundaries were resolved from the adopted roadmap and prior decisions. The designs below are reviewable implementation proposals, not claims that newly planned behavior is approved for live operation.

## Order and dependencies

| Order | Change | Testable outcome | Prerequisites |
|---|---|---|---|
| 1 | [CH17 — retranslate reviewed descriptions](../../openspec/changes/retranslate-reviewed-ai-descriptions/proposal.md) | Generate selected languages from reviewed text, then explicitly adopt current suggestions without rerunning geolocation | Implemented CH09, CH11–CH12, CH16 |
| 2 | [CH20 — write approved descriptions](../../openspec/changes/write-approved-ai-descriptions/proposal.md) | Confirm one primary-language description independently of GPS, preserving existing text and safe append behavior | CH19 implemented/synchronized; apply CH17 first in this batch |
| 3 | [CH21 — confirm exact stack targets](../../openspec/changes/confirm-explicit-ai-stack-targets/proposal.md) | Confirm a frozen reviewed stack GPS list and inspect independent target outcomes | CH19; apply/sync CH20 first because this batch evolves its writeback requirements |
| 4 | [CH22 — mirror supported metadata](../../openspec/changes/mirror-supported-ai-metadata/proposal.md) | Optionally mirror reviewed metadata without repeating successful standard-field writes during recovery | CH20; apply/sync CH21 first for this batch's shared target/step contract |

At planning time, each change received a proposal, design, capability deltas, numbered unchecked tasks and a scenario-to-task/test verification plan. The completed tasks and implementation-verification records now contain the execution results. CH17 changes `ai-results-and-review`. CH20–CH22 each evolve that capability and `ai-immich-writeback`; they are separate coherent changes, not one large writeback proposal.

Finish implementation, verification and spec synchronization for each predecessor before applying its successor. Rebase/reconcile later deltas if an earlier implementation changes the contract. The ordered deltas carry complete modified requirement blocks and retained regression scenarios; applying CH21/CH22 out of order would be incorrect. Maintained specs were left unchanged during planning and synchronized after each change was implemented and verified. Migration numbers are allocated sequentially from the actual apply checkout, after existing migration 028, rather than reserved in this plan.

CH17 can be implemented and tested while CH19 live GPS verification is pending. CH20–CH22 can use deterministic fixtures against implemented CH19, but their real write capabilities remain default-off until separately authorized evidence covers the actual version/profile, fields and targets. The [CH19 preview deployment](../engineering/nas-ch19-preview-2026-09-25.md) records the current healthy build with writing disabled.

## Resolved scope and design direction

- Translation starts from a visible reviewed text basis, with no new image/geolocation call. Language failures are independent; generation never overwrites a draft. Explicit adoption checks exact draft/factual revisions and active-write protection.
- A current description can be saved without GPS. Preserve is the default; replace/managed append require deliberate selection. Empty description clearing remains outside this change. Text comparisons preserve exact Unicode, whitespace and line breaks.
- Managed append owns one stable block through verified private lineage. Existing edited, missing, foreign or malformed markers conflict rather than being repaired or duplicated automatically.
- Stack GPS starts with the analyzed photo alone and permits up to 50 explicitly reviewed members. Every target has its own source/baseline/authority. New members never join an approved plan; departed selected siblings conflict. Description and direction are not copied to siblings.
- Optional metadata mirrors only selected current fields under `immich-places-ai-addon` on the analyzed photo, with a reader-visibility disclosure. It follows verified standard fields as a separate step; failures preserve standard success and authoritative local records.
- Every extension preserves durable approval, exact revision, fresh conflict checks, no hidden mutation retry, known-sender-completion rules, target exclusion, private audit and source-consistent readback. Each target/step has a two-attempt ceiling; total possible sends remain bounded by the frozen plan.
- Existing v1 approval bytes/digests and unresolved state remain untouched. New plan versions are explicit and reject unknown content. The existing write switch alone cannot enable new description, stack or metadata capabilities.
- All designs retain Go/Next.js/local SQLite, existing providers/map/test dependencies and one backend process. No new library, architectural departure, proxy change, general retention policy, automatic writeback, upstream deletion or universal undo is included.

## Approaches considered during design

| Change | Selected approach | Other viable approach and accepted trade-off |
|---|---|---|
| CH17 | Focused translation run/item records, existing provider dispatch policies and shared resource limits | Add a translation kind to image-analysis jobs. This could share more persistence, but couples a text-only result and consent to canonical image-analysis contracts; focused records keep each history truthful at the cost of a small new adapter. |
| CH20 | Explicit versioned standard-field plans and additive storage, with shared target exclusion | Rebuild existing plan tables under a new migration and broaden their decoder. This saves tables but increases risk to immutable v1 payloads, ambiguous senders and rollback. Versioned storage needs mixed-history readers. |
| CH21 | One confirmed immutable manifest with per-target state and atomic guard admission | A parent batch of separately confirmed single-target operations. This can isolate retries but makes whole-list approval/overlap handling and partial admission more complex. Per-target state keeps one scope decision and requires carefully qualified SQL. |
| CH22 | An independently reserved metadata step under the same confirmation | A separately confirmed mirror operation linked to standard success. This is viable but adds a second approval workflow; the selected step model matches the optional second-step contract while requiring separate step audit and retries. |

No discarded approach is silently implemented. The individual designs contain the selected architecture, responsibilities, data flow, failure handling, migration/rollout and test boundaries. Standard libraries and installed dependencies are sufficient; no new-library approval is needed.

## Requirement and verification handoff

| Change | Primary product ownership | Verification plan |
|---|---|---|
| CH17 | FR-08/09/10/13; NFR-01/02/03/06; AC-04/08/09/12 | [T01–T21](../../openspec/changes/retranslate-reviewed-ai-descriptions/verification-plan.md) |
| CH20 | FR-08/10–14; NFR-01/02/03/06/08; AC-04/06/08/09/11/12 | [C01–C20 and inherited cases](../../openspec/changes/write-approved-ai-descriptions/verification-plan.md) |
| CH21 | FR-02/10–12/14; NFR-01/02/03/06/08; AC-04/05/06/09/11 | [S01–S20 and inherited cases](../../openspec/changes/confirm-explicit-ai-stack-targets/verification-plan.md) |
| CH22 | FR-07/08/11–14; NFR-01/02/03/06/08; AC-04/08/09/12 | [M01–M21 and inherited cases](../../openspec/changes/mirror-supported-ai-metadata/verification-plan.md) |

Verification plans map every scenario, including retained full-block predecessor scenarios, to tasks and automated boundaries. They identify required fixture extensions rather than claiming those cases already exist. Apply follows Plus TDD with the adopted characterization exception, real SQLite migration/reopen/authorization/race fixtures, local HTTP exact payload/count tests, typed component tests and affected built-browser journeys. All thirteen engineering gates and coverage floors remain applicable before implementation completion; the planning pass runs document/configuration checks only.

## Inline review and resolved findings

Proposal reviews checked intent, cohesion, prerequisites, non-goals and capability coverage. Design reviews checked package ownership, actual source seams, durable state, bounded I/O, security, migration and recovery. Spec reviews checked success/negative/edge Gherkin coverage and ordered-delta preservation. Task reviews checked end-to-end outcomes, requirement-level scope and complete test traceability. Material findings resolved during preparation:

1. The current provider codec requires an image. Translation needs a bounded text-only codec, not a dummy image or a fabricated geolocation result.
2. Old preview/approval storage caps payloads at 16 KiB and binds one GPS point. Description and later target/step plans need explicit new versions and additive storage, preserving old bytes and guards.
3. A description-only save must not require GPS or discard the user's pending manual coordinates. Relevant-field baselines and readiness remain independent.
4. Matching a marker is insufficient ownership. Append requires retained verified lineage and fails closed on edited/duplicated/unowned boundaries.
5. Combined GPS/text readback can be mixed. Preserve field evidence and require a new remaining-field plan once safe; never replay a payload that would resend observed successful fields.
6. The current guard token is unique. Multi-target plans require distinct opaque target tokens and atomic all-target admission, not one reused parent token or operation-wide completion updates.
7. Metadata missing-key errors do not prove absence. Use a valid complete authorized metadata-list read and distinguish absent, unavailable and conflicting owned values.
8. Metadata is a separate mutation. Its attempt identity, completion proof and readback cannot reuse the standard step's state; successful standard writes are never replayed for mirror repair.
9. GPS compatibility must not enable every new capability through the existing write flag. Default-off capability activation is separate and installation/profile-bound.

## Evidence limits

GitNexus was bound to this repository and indexed revision `0c5bc0c`; source checks confirmed relevant draft, preview, writer and migration boundaries. Graph process coverage is capped and some Go receiver relationships are unresolved, so no complete call-graph proof is claimed. Apply repeats impact analysis at its actual checkout.

Context7 resolved the official Immich API documentation, supplemented by exact pinned v3.2.2 controller/DTO/service source linked from the designs. This supports route and schema planning only. No photo, credential or private provider payload was included in documentation queries. Existing CH19 live photo authorization is not broadened by this batch.

No application tests, database migrations, live provider or Immich requests, deployment, spec synchronization or implementation were performed for this planning batch.

## Executed planning checks

- `openspec validate --all --strict --no-interactive`: passed all 15 maintained-spec/change items, with zero failures.
- `openspec status --change <name> --json`: all four changes report `isPlanningComplete: true`; proposal, design, specs and tasks are present. All 35 implementation tasks remain unchecked.
- Planning consistency check: passed ordered delta composition, retention of predecessor scenarios, GIVEN/WHEN/THEN structure, capability declarations and complete scenario/task/test mappings. The batch adds 25 requirements and 82 scenarios; its complete delta blocks contain 214 scenario references including inherited cases.
- Documentation check: passed 26 Markdown files and 141 local links, with no unresolved planning placeholders.
- `bun run check:size --base HEAD`: passed, checking 1,024 files against `0c5bc0c`.
- `git diff --check`: passed. A separate check of all 28 new files passed for trailing whitespace, conflict markers and final newlines.
- GitNexus `detect_changes(scope: all)`: reported low risk and no affected execution flows for the three tracked documentation/configuration files. Untracked planning artifacts are outside that diff; this result is not evidence of runtime coverage or independent review.

Local check evidence is retained under ignored `out/checks/ch17-ch22-planning/`. These results establish planning consistency and readiness for review, not implemented acceptance criteria.
