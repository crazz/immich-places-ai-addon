# CH16 → CH18 → CH19 planning batch

Status: prepared, not implemented. Prepared on 24 September 2026 against `0fb8eaf` on `codex/plan-ai-selection-and-validation`. All artifacts are English. OpenSpec Plus proposal, design, specification and task reviews were performed inline under the user's no-subagent instruction. No independent/subagent review is claimed.

## Application order

| Order | Change | Outcome | Requirements / scenarios / tasks |
|---|---|---|---|
| 1 | [CH16: persistent review drafts](../../openspec/changes/persist-revisioned-ai-review-drafts/proposal.md) | Accept, edit, stage and reject durable local decisions without Immich writes | 7 / 20 / 10 |
| 2 | [CH18: exact GPS previews](../../openspec/changes/preview-exact-ai-write-plans/proposal.md) | Inspect fresh single-photo GPS before/after values for an exact revision | 6 / 18 / 7 |
| 3 | [CH19: confirmed GPS writes](../../openspec/changes/execute-confirmed-ai-gps-writes/proposal.md) | Confirm, execute and reconcile one exact GPS operation with durable audit | 9 / 32 / 16 |

Total: 22 requirements, 70 behavioral scenarios and 33 unchecked implementation tasks. Each change includes `proposal.md`, `design.md`, capability deltas, `tasks.md` and `verification-plan.md`.

Apply in the requested order. Finish, verify, synchronize/archive and commit each change before applying its successor. CH16 depends on implemented CH15. CH18 depends on CH16. CH19 depends on CH18 and implemented CH11; CH17 is not required for GPS-only writes. Allocate migrations serially after inspecting the actual tip; 026/027/028 are planning expectations, not reserved applied versions.

CH16 adds requirements to maintained `ai-results-and-review`. CH18 introduces `ai-immich-writeback`; CH19 then adds distinct requirements to it and to `ai-results-and-review`. Do not apply CH19 before CH18 has been synchronized. The three deltas use unique added requirement names and preserve existing read-only inspection/preview requirements. Main specs stay unchanged until implementation and synchronization.

## Settled boundaries

- Local acceptance and editing never call a provider or mutate Immich. Original analyses remain immutable; draft revision and write outcome are separate.
- Camera coordinates can be coarse, including ±500 meters or much more; confidence, direction, description availability and estimated-error magnitude do not impose a GPS threshold. The user decides whether the point is useful.
- Local drafts can be saved offline. Fresh baseline acknowledgement is explicit and binds the displayed current GPS/image; missing metadata is distinct from missing GPS. A replaced image can be reviewed again without rerunning AI, with old provenance preserved.
- GPS plans target only the analyzed photo and the latitude/longitude pair. No stack expansion, descriptions, direction, metadata mirroring or GPS clearing is included.
- Preview expiry is five minutes. Changed source, baseline, revision or authority prevents using obsolete approval. A stored digest is not a credential or client-supplied authority.
- CH19 uses a separate exact writer, default-off operator dispatch switch, durable per-target exclusion, no hidden retry and readback before local refresh. An ambiguous timeout does not authorize a second send. A completed unsuccessful attempt can be retried explicitly only within the same valid approval and a two-attempt ceiling.
- Draft edits cancel undispatched or safely retryable approval; while a reserved send can still act, its approved revision remains fixed. Existing manual pending state is preserved and known overlapping choices must be resolved explicitly.
- No new library, database, worker service or architecture departure is proposed. Existing Go/Next.js/SQLite and map dependencies are sufficient. Account/installation isolation, cleanup protection, bounded I/O and rollback are included with each new resource.

codex-proxy work is deferred by the user's explicit instruction. The [live comparison](../engineering/research-live-comparison-2026-09-24.md) remains evidence of a Research response-framing limitation, not a dependency of these changes. Any retained valid Visual, Context-assisted or Research result can be reviewed. The three supplied photographs were authorized for analysis, not for mutation testing.

## Verification handoffs

| Change | Scenario groups | Verification plan |
|---|---|---|
| CH16 | D01–D20: durability, editing, stale content, baselines, isolation and accessible manual-state separation | [CH16 verification](../../openspec/changes/persist-revisioned-ai-review-drafts/verification-plan.md) |
| CH18 | P01–P18: exact read-only comparison, fresh conflicts, immutable scope, expiry and isolation | [CH18 verification](../../openspec/changes/preview-exact-ai-write-plans/verification-plan.md) |
| CH19 | W01–W26 and R01–R06: approval, exact writes, concurrency, recovery, verified refresh, lifecycle and review UI | [CH19 verification](../../openspec/changes/execute-confirmed-ai-gps-writes/verification-plan.md) |

Every scenario maps to tasks and planned automated layers. Apply must replace planned evidence with exact executed test names and outcomes in each change's implementation verification record. Real SQLite is required for persistence/race claims; deterministic HTTP fixtures inspect exact fields/targets/request counts. Full installed engineering gates and relevant built-browser journeys remain required. No application tests, migrations, provider requests, Immich mutations or deployments are performed by this planning batch.

GATE-02 remains separate: real write rollout needs an explicitly authorized disposable Immich photo and actual version/route/readback evidence. Until then, finish deterministic implementation checks and keep `AI_WRITE_ENABLED=false`. Passing mocked adapters is not a live compatibility claim. No private photo is to be silently repurposed as a write fixture.

## Inline artifact review

Reviewed proposal scope/non-goals, design alternatives/package ownership, source grounding, authority boundaries, failure/recovery/rollout, Gherkin coverage and task slicing. Concrete findings resolved during preparation:

1. Existing image digests include metadata timestamps. Preserve analysis semantics and introduce a separately reviewed image identity so a GPS write does not invalidate its own image identity.
2. Draft acceptance must survive upstream outages. Keep baseline unavailable explicit and make fresh acknowledgement a later read-only action.
3. Baseline observations require server persistence, binding, expiry and bounds; they cannot be trusted client-supplied metadata.
4. The manual frontend/backend save path adds retries and stack scope. CH19 must use its own exact non-retrying transport.
5. An expired lease cannot fence an already-sent upstream request. Keep target exclusion and reconcile before any possible retry; unknown completion stays unresolved.
6. Reconciliation must distinguish unchanged baseline from a conflicting third value; neither implies a successful write.
7. Confirmed draft editing needs an atomic boundary around dispatch reservation; safely retryable approval must also be invalidated by a new edit.
8. A delayed catalog sync can overwrite newly verified GPS. Reuse the existing sync pause/drain boundary for final readback and publication, with deadline/failure coverage.
9. CH19 modifies two capabilities and depends on CH18's not-yet-created main spec. Record ordered synchronization explicitly instead of pretending the capability already exists.

Planning verification consists of strict OpenSpec validation, ordered delta composition, scenario/task/verification coverage, English/placeholder checks, local-link checks and pre-commit GitNexus change analysis. Implementation and live compatibility remain unverified until their own execution records exist.

Executed document checks: `openspec validate --all --strict --no-interactive` passed all 11 current specs/changes. All three change statuses report planning complete. Ordered composition found no duplicate added requirement names; all 70 scenarios have GIVEN/WHEN/THEN and are covered by task and verification references. The 20 relevant Markdown files passed local-link and English checks, all new artifacts passed placeholder checks, and `git diff --check` passed. No application suite was run for this documentation-only batch.
