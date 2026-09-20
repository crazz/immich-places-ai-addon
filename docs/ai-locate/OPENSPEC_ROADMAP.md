# Proposed OpenSpec conversion roadmap

**Status:** F01–F03 tooling and the first product batch, CH04 capture-date consistency, CH01 private provider settings, CH02 egress policy and CH03 capability tests, are implemented at this checkout when their shared verification passes. Later product candidates remain proposed.
**Baseline:** [PRD](PRD.md), [technical design](TECHNICAL_DESIGN.md), and [reconciliation](RECONCILIATION.md) against `5e70c6165777949c9d8b50ede3b2768bcaa5df87`.  
**Tooling inspected:** OpenSpec 1.8.0 with the `spec-driven` schema and repository-local OpenSpec Plus rules.

The adopted [architecture](../engineering/architecture.md), [testing](../engineering/testing.md), [coding standards](../engineering/coding-standards.md) and [ADR-07](../engineering/decisions/ADR-07-ai-internal-packages.md) now govern every candidate. They establish implementation rules, not completed tooling or product behavior.

## Correction to the original breakdown

The previous eight items bundled too many independently reviewable outcomes into individual changes. Retain them as capability areas and use the 26 smaller candidate changes below. This is a planning recommendation, not an OpenSpec size limit or a promise that exactly 26 proposals will be needed.

A capability is a lasting behavior contract. Several changes can introduce or evolve its requirements over time. The first change introduces the implemented subset; later changes add new requirements or modify existing ones through deltas. Do not create a new capability for every task, and do not mark the full planned capability implemented after its first slice. See [OpenSpec concepts](https://github.com/Fission-AI/OpenSpec/blob/main/docs/concepts.md).

Keep the reference package authoritative for product scope. Visual and Context-assisted analysis, camera/subject separation, nullable direction, multilingual descriptions, persistent jobs/results/drafts, confirmed writes, and lifecycle controls all remain required V1 scope.

## Change sizing rules

- One clearly stated behavior or contract, with a focused acceptance story and failure cases.
- Independently reviewable and verifiable against completed prerequisites; roughly a coherent PR when practical, without a fixed file/task/time quota.
- Include the UI, API, persistence, migrations and tests needed for that outcome. Do not split only along technical layers.
- Split independently useful additions, such as description writes, stack propagation, and metadata mirroring, from the first safe single-asset GPS write.
- Keep inseparable safety properties together. Every exposed endpoint is authorized from its first introduction. Every production provider dispatch is bounded, consented and recoverable. The first mutation includes confirmation, conflict detection, durable audit and readback.
- An internal contract can be validated with fixtures before its UI exists. Keep incomplete user workflows unavailable; fixtures do not establish live compatibility.

At formal proposal time, split a candidate again if its design reveals independently releasable behavior or multiple substantial workflows. Conversely, retain closely coupled steps together when separating them would leave an unsafe or unverifiable intermediate state.

## Engineering tooling prerequisite

Complete the bounded tooling setup before feature implementation. It is separate from the 26 product candidates and introduces no AI product capability. The work is split into three independently verifiable changes:

| ID | Outcome and current status |
|---|---|
| F01 `enforce-engineering-checks` | [Archived artifacts](../../openspec/changes/archive/2026-09-18-enforce-engineering-checks/proposal.md); all ten tasks and final review are complete. Installed: exact file-size baseline/ratchet, source dependency checks, checker regressions, shared format/lint/types/Go race-test/build command and read-only CI with pinned toolchains/frozen dependencies. Recorded baseline lint errors and gofmt differences are corrected; inherited warnings remain visible. |
| F02 `add-frontend-unit-tests` | [Archived artifacts](../../openspec/changes/archive/2026-09-18-add-frontend-unit-tests/proposal.md); all five tasks and inline final review complete. Installed: offline Vitest/React Testing Library harness, seven pagination logic/component tests, full-source V8 coverage and the native 80% line/branch floor for future AI TypeScript. All twelve shared local checks pass; the same command is wired into CI. AI code remains absent/unmeasured. |
| F03 `add-application-smoke-tests` | [Archived artifacts](../../openspec/changes/archive/2026-09-18-add-application-smoke-tests/proposal.md); all six tasks and inline review complete. Installed: three Chromium journeys for authentication/browsing, manual confirmation and GPX preview/confirmation against real production builds and temporary SQLite, with synthetic Immich/tiles and blocked external egress. All thirteen shared local checks pass; CI uses the same command, installs Chromium and retains failure evidence. Live compatibility and future AI acceptance remain separate. |

Use the [testing guide](../engineering/testing.md#available-commands-and-remaining-setup) for installed commands, comparison-base selection, formatting and enforcement limits. The [verification baseline](../engineering/verification-baseline.md) retains pre-enforcement evidence separately from completion evidence. Future AI scope must meet the adopted coverage floors; absent AI packages are unmeasured.

The setup does not need to invent empty AI packages to enforce architecture. The first change introducing an internal package must include the matching import rules and backend Docker build update/verification. Split the tooling setup if its implementation reveals independently verifiable work; do not fold a broad cleanup into it.

The initial UI harness is complete here and is not additional CH01 setup. CH01 expands provider behavior coverage, enforces backend AI statement coverage and verifies the backend container with internal packages. Expand the harnesses as product capabilities arrive; unexecuted live and quality checks must still be reported explicitly.

## Capability areas

| Area | Stable capability | Candidate changes |
|---|---|---|
| Providers | `ai-provider-configuration` | CH01–CH03 |
| Selection | `catalog-capture-dates` (CH04 foundation), `ai-selection-snapshots` (CH05–CH06) | CH04–CH06 |
| Analysis | `ai-location-proposals` | CH07–CH10 |
| Jobs | `ai-analysis-jobs` | CH11–CH13 |
| Results/review | `ai-results-and-review` | CH14–CH17 |
| Writeback | `ai-immich-writeback` | CH18–CH22 |
| Lifecycle | `ai-result-lifecycle` | CH23–CH24 |
| Operations | `ai-operations` | CH25; CH26 may be documentation/verification only |

The groupings are not eight large parent OpenSpec changes. They organize the roadmap. A candidate may touch an upstream capability when its observable contract changes; declare that delta explicitly rather than duplicate requirements.

## Candidate changes and prerequisites

Dependencies below are product planning constraints, not claimed CLI-enforced scheduling. All feature implementation also requires the engineering tooling setup above; a dash means no preceding product candidate. IDs are stable roadmap references; ordering is a partial order rather than a requirement to implement every row serially.

| ID / proposed change name | Bounded outcome and acceptance boundary | Depends on |
|---|---|---|
| CH01 `add-private-ai-provider-profiles` | Implemented: private encrypted revisioned profiles, settings UI, default-off installation control, session/ownership/origin protection, credential cleanup and backend AI coverage. See the [maintained contract](../../openspec/specs/ai-provider-configuration/spec.md) and [operator guide](../ai-provider-settings.md). Saving sends no image or provider request. | — |
| CH02 `enforce-ai-provider-egress-policy` | Implemented: approved destinations, redirect/credential rules, DNS/address validation and bounded internal dispatch. Reuse the existing NAS codex-proxy through explicit endpoint/network approval. | CH01 |
| CH03 `add-ai-provider-capability-tests` | Implemented: explicit synthetic-image tests, separate image/JSON/strict observations, revision-bound private evidence and Settings UI. Select the model in the addon profile; no private photo, implicit fallback or proxy redeployment. The [19 September NAS check](../engineering/nas-provider-capability-2026-09-19.md) observed all three capabilities for `gpt-5.6-sol`, completing GATE-03 for the recorded configuration. | CH02 |
| CH04 `align-source-local-capture-dates` | Implemented: shared source-local catalog date filtering/counts, invalid/reversed range handling and unchanged gallery ordering. See the [maintained contract](../../openspec/specs/catalog-capture-dates/spec.md). Undated-group presentation remains later FR-01 work; CH05–CH06 implement selection/eligibility. | — |
| CH05 `freeze-explicit-ai-selections` | Implemented: [maintained snapshot contract](../../openspec/specs/ai-selection-snapshots/spec.md) and [operator guide](../ai-selection-snapshots.md); expiring user-bound explicit-ID snapshots, deduplication, eligibility/exclusions, frozen scope and minimal persistent installation binding. Later catalog changes cannot expand them. | CH04 |
| CH06 `freeze-all-matching-ai-selections` | Implemented: complete-query selection through shared eligibility, exact aggregate counts, eligible-asset limits and durable frozen membership. [Verification](../engineering/ai-matching-selection-verification.md) includes whole-catalog and failure-path evidence; shared selection specs are maintained after CH05. | CH05 |
| CH07 `validate-ai-analysis-results` | Implemented: [maintained proposal contract](../../openspec/specs/ai-location-proposals/spec.md), canonical and semantic validation for outcomes, references, camera/subject coordinates, radius/direction and language statuses; no write authority. [Verification](../engineering/ai-result-validation-verification.md) covers bounded parsing, race/fuzz, offline packaging and the pinned v6.0.3 dependency audit. CH09 and CH11 implement internal analysis and durable synthetic-execution consumers. | — |
| CH08 `prepare-bounded-ai-images` | Implemented: [bounded transient image copies](../ai-image-preparation.md), exact authorized reads, orientation/metadata normalization, decode/transmission budgets and authority/source freshness checks. [Verification](../engineering/ai-image-preparation-verification.md) covers synthetic HTTP/SQLite/raster tests and offline packaging; deployed image compatibility remains unverified. | — |
| CH09 `analyze-single-images-in-visual-mode` | Implemented: [internal Visual analysis](../ai-visual-analysis.md) through the approved transport with current authority, exact revisions, explicit reservation and complete canonical validation. [Verification](../engineering/ai-visual-analysis-verification.md) covers synthetic framing, cancellation, freshness and zero-write behavior. No public admission or worker is connected. | CH03, CH07, CH08 |
| [CH10 `add-consented-ai-context`](../../openspec/changes/archive/2026-09-20-add-consented-ai-context/proposal.md) | Implemented and archived: Internal bounded metadata-only context in the shared analysis path with explicit consent, source eligibility, capture-time status and provenance. Exclude hidden/inaccessible or disallowed AI-origin sources and unrelated frequent/home hints. | CH04, CH09 |
| CH11 `add-recoverable-ai-job-execution` | Implemented: [durable private jobs/results](../ai-analysis-jobs.md), idempotent submission, bounded claims/reservations, leases/fencing, cancellation and restart recovery. Reuses CH05 installation rotation and provides bounded private cleanup. [Verification](../engineering/ai-analysis-jobs-verification.md) covers real SQLite and synthetic execution; CH12 must connect production admission/dispatch. | CH01, CH05, CH07 |
| [CH12 `connect-visual-analysis-to-durable-jobs`](../../openspec/changes/archive/2026-09-20-connect-visual-analysis-to-durable-jobs/proposal.md) | Implemented and archived: Execute Visual analysis from frozen selections through CH11, binding consent/profile revision and enforcing access checks, enablement, deadlines, call/token limits and bounded transport behavior. Add submission/cancellation access for the usable workflow. | CH09, CH11 |
| [CH13 `add-ai-batch-progress-and-reruns`](../../openspec/changes/archive/2026-09-20-add-ai-batch-progress-and-reruns/proposal.md) | Implemented and archived: Job launch/progress UI for supported modes, page/all-matching submission, cancellation, retry-failed and reanalysis as explicit new runs. Context-assisted dispatch is enabled only after CH10's contract is present; reruns do not replace reviewed results. | CH06, CH10, CH12 |
| [CH14 `browse-persistent-ai-results`](../../openspec/changes/archive/2026-09-20-browse-persistent-ai-results/proposal.md) | Implemented and archived: Authorized, filtered history for completed/unknown/ambiguous/failed outcomes, independent of Missing GPS; preserve earlier results and show analysis state separately from review/write state. | CH12 |
| [CH15 `review-ai-camera-and-subject-proposals`](../../openspec/changes/archive/2026-09-20-review-ai-camera-and-subject-proposals/proposal.md) | Implemented and archived: Read-only map/panel presentation of separate camera/subject points, alternatives, uncertainty, nullable heading, evidence and language tabs. Include keyboard access and numeric presentation when maps fail. | CH14 |
| CH16 `persist-revisioned-ai-review-drafts` | Accept/edit/reject durable drafts with numeric/map editing, explicit targets and field choices, baseline metadata, concurrent-edit detection and stale-content invalidation. Accept performs zero writes; AI drafts cannot enter the existing manual save path. | CH15 |
| CH17 `retranslate-reviewed-ai-descriptions` | Recreate selected languages from the reviewed facts without rerunning geolocation; retain per-language failures, factual revision and stale-state rules. | CH09, CH16 |
| CH18 `preview-exact-ai-write-plans` | Fetch fresh authorized Immich values and produce an expiring, revision-bound, single-target GPS plan with a digest and before/after values. This change exposes no mutation executor. | CH16 |
| CH19 `execute-confirmed-ai-gps-writes` | First actual mutation: single-asset approved GPS only, with durable confirmation/audit, fresh conflict checks, exact targets, serialized overlapping writes, disabled hidden transport retries, post-write readback and restart/ambiguous-outcome reconciliation. Refresh local state only after verification. These guarantees ship together. | CH18, CH11 |
| CH20 `write-approved-ai-descriptions` | Extend the existing preview/confirmation/executor contract with independent primary-language description selection and preserve/replace/managed-append semantics, including safe repeated append and readback. | CH19 |
| CH21 `confirm-explicit-ai-stack-targets` | Extend GPS write scope through a separately displayed and confirmed member list with per-member authorization/baselines. Later membership changes cannot expand the plan; descriptions/direction are not propagated by default. | CH19 |
| CH22 `mirror-supported-ai-metadata` | Optional supported namespaced metadata step for approved direction/translations/provenance, with per-step audit, readback and partial-success handling. Required local records work without it. | CH20 |
| CH23 `export-and-delete-ai-results` | User-controlled export and deletion, active-work conflicts and retained-audit redaction rules. Local deletion never silently changes Immich; new resources already have isolation and account-deletion policies. | CH16, CH19 |
| CH24 `apply-ai-retention-policy` | Execute approved retention rules, cleanup/recovery scheduling and account-deletion integration across provider, job, draft/result and audit data. Prove retained drafts/audits and in-flight work are handled consistently. | CH23 |
| CH25 `expose-safe-ai-operational-diagnostics` | Sanitized aggregate queue/provider/write metrics and operator configuration guidance. No images, prompts, secrets or precise private locations leak into ordinary diagnostics. | CH12, CH19 |
| CH26 `verify-ai-upgrade-and-recovery` | Backup/restore and upgrade/rollback runbook with executable verification for the final data model, encryption key and existing two-service/local-SQLite deployment. If this is only docs/test tooling, use the supported no-spec-change workflow rather than inventing behavior requirements. | CH20, CH21, CH22, CH24, CH25 |

CH26's final verification also includes the tables and lifecycle behavior introduced by transitive prerequisites. Real corrective behavior discovered during verification should be a separately scoped fix or an explicit change to the relevant owner, not an unlimited “hardening” backlog hidden inside CH26.

## Safety boundaries and useful milestones

CH01 never tests a provider merely because a profile is saved. CH03 may make a deliberately requested synthetic-image call under CH02's egress policy. CH08–CH10 establish internal image/analysis contracts; private production analysis only becomes reachable through CH12's durable, authorized path. CH13 adds the complete V1 launch/progress experience, including Context-assisted mode.

CH11 keeps restart recovery and cancellation with the initial queue state machine. CH19 keeps approval, durable audit, conflicts, exact scope and readback with the initial mutation executor. These are larger than a UI enhancement but remain cohesive safety contracts. Do not split them into releases that temporarily permit unbounded analysis, unconfirmed writes, unlogged mutations or blind retries.

After CH12/CH14, durable Visual analysis and persistent results are available. CH13 adds the full batch/mode workflow. CH15–CH17 complete review and multilingual correction. CH19 provides a safe GPS-write milestone; CH20–CH22 extend field/scope support. The full V1 also requires lifecycle controls, operational verification and the release gate below. Intermediate milestones do not redefine V1 as GPS-only.

Independent work is possible on provider setup, capture-date correction, validation and image preparation. Serialize shared migration allocation and overlapping edits. Finish/sync an earlier delta before applying another change to the same requirement, or explicitly reconcile the combined delta. The [OpenSpec archive workflow](https://github.com/Fission-AI/OpenSpec/blob/main/skills/openspec-bulk-archive-change/SKILL.md) identifies overlapping capability deltas; this roadmap does not assume automatic dependency management.

## Requirement and acceptance ownership

These are completion owners, not permission to defer a safeguard until the last listed change. Extend each row to exact requirement/scenario names, task IDs and tests as its actual change is drafted.

| Requirement | Candidate owners |
|---|---|
| FR-01 — Catalog/capture dates | CH04–CH06; CH14–CH16 retain the existing browsing/map experience |
| FR-02 — Frozen selection | CH05–CH06; CH11–CH13 consume snapshots; CH21 freezes additional write targets |
| FR-03 — Private providers | CH01–CH03 |
| FR-04 — Consent | CH02 controls egress; CH08–CH10 minimize payloads; CH12–CH13 bind/enforce dispatch consent |
| FR-05 — Durable bounded work | CH11–CH13 |
| FR-06 — Honest proposals | CH07, CH09–CH10, CH15 |
| FR-07 — Direction | CH07, CH09, CH15–CH16; optional mirror in CH22 |
| FR-08 — Multilingual descriptions | CH07, CH09, CH15–CH17, CH20 |
| FR-09 — History | CH11, CH13–CH14 |
| FR-10 — Durable drafts | CH16 |
| FR-11 — Explicit writes | CH18–CH22; each extension preserves the initial confirmation boundary |
| FR-12 — Preservation/conflicts | CH19–CH22 |
| FR-13 — Local extended record/mirror | CH11, CH14–CH17, CH22 |
| FR-14 — Audit/export/deletion | CH19–CH22 create audit with every write; CH23–CH24 complete lifecycle controls |
| NFR-01 — Authorization/isolation | Every change introducing a resource/operation; integrated tenant tests at release |
| NFR-02 — Durability/idempotency | CH11–CH13, CH19–CH22 |
| NFR-03 — AI disabled | CH01 and every dispatch/mutation consumer; existing workflow regression at release |
| NFR-04 — Gallery performance | CH04–CH06 and CH11–CH13; measured reference-workload acceptance at release |
| NFR-05 — Submission latency | CH11–CH13; measured 500-asset submission acceptance at release |
| NFR-06 — Accessible review | CH15–CH17; accessible controls also accompany all earlier/later UI changes |
| NFR-07 — Deployment topology | Every change preserves it; CH26 verifies the final deployment |
| NFR-08 — Recovery/diagnostics | Migration checks with every owning change; CH24–CH26 complete operational behavior/evidence |
| AC-01 — Filter/freeze | CH04–CH06, CH11–CH13 |
| AC-02 — Unknown location | CH07, CH09, CH11, CH14–CH15 |
| AC-03 — Landmark versus viewpoint | CH07, CH09, CH15 |
| AC-04 — Accept does not write | CH16; regression in CH19–CH22 |
| AC-05 — Stack scope | CH19 and CH21 |
| AC-06 — Conflict | CH19–CH22 |
| AC-07 — Restart/cancellation | CH11–CH13 |
| AC-08 — Languages/partial write | CH17, CH20, CH22 |
| AC-09 — Tenant isolation | Every applicable change, repeated in the complete acceptance run |
| AC-10 — Incomplete provider output | CH03, CH07, CH09, CH11–CH13 |
| AC-11 — Saved GPS/retained history | CH14 and CH19 |
| AC-12 — Correction invalidation | CH16–CH20 |

## Converting candidates to actual OpenSpec changes

Do not scaffold 26 complete changes up front. Keep this inventory and traceability map, then fully develop a small ready batch. Plan the tooling prerequisite first; the first product batch should be CH01 and CH04: private profile management and capture-date consistency. CH07 and CH08 are independent subsequent candidates if useful capacity is available. This is a recommendation for later work, not an instruction to start implementation now.

For each candidate, use the installed schema and the appropriate OpenSpec Plus proposal/spec/design/tasks workflow. The supplied documents already answer the general product purpose, V1 scope, non-goals and architecture constraints; ask only about unresolved decisions material to that candidate.

| Input | Artifact |
|---|---|
| One bounded outcome plus relevant PRD scope | `proposal.md`: Why / What Changes / Capabilities / Impact |
| Relevant FR/NFR obligations and AC examples | `specs/<capability>/spec.md`: testable normative requirements and concrete success/failure scenarios |
| Relevant technical design, ADRs and REC findings | `design.md`: decisions, affected contracts, migration/failure strategy and rationale |
| Resolved design and scenarios | `tasks.md`: numbered test-first implementation and verification work |

Specs and design follow proposal; tasks depend on both in the installed schema. Initially there are no main AI specs: the first verified slice introduces its actual requirements. Later slices use ADDED for new behavior and full MODIFIED blocks only for requirements that already exist. Keep existing catalog behavior and newly promised consistency guarantees distinguishable. Do not create placeholder requirements for future slices.

Reference the canonical JSON schema/examples rather than cloning contracts into each change. Shared mechanisms belong in the owning design and implemented contracts; every consumer still gets authorization and failure-path scenarios. Keep implementation details out of behavior specs.

Validate each actual change with the installed CLI, including `openspec validate <change-name> --strict --no-interactive`, and check traceability separately. The candidate inventory itself is not an implementation-ready set of OpenSpec artifacts. Use the required Plus artifact reviews when actual artifacts are generated.

Implementation uses Plus apply/TDD and the relevant GitNexus checks. Every behavior spec scenario becomes at least one automated test; live compatibility, performance and quality claims additionally need the specified external/measured evidence. Use the prerequisite frontend harness and extend existing Go fixtures throughout. Tests, tenant controls and migrations are not a final change.

After implementation and verification, reconcile/sync the change's deltas and archive it before advancing conflicting successors. Keep normative requirements in one main capability spec; several small changes can evolve it. Preserve existing OpenSpec/Plus configuration and its adopted engineering/reference-document context. Allocate migration numbers from the actual implementation checkout, not from this roadmap.

## Compatibility gates and release acceptance

| Gate | When it matters |
|---|---|
| GATE-01 — Repository baseline | Complete at `5e70c61`; refresh affected evidence at each actual implementation base |
| GATE-02 — Immich compatibility | Verify image access for CH08 and metadata reads for CH16/CH18. Verify mutation methods/rights/readback on authorized disposable fixtures before claiming CH19–CH22 live compatibility. Package references to v3.2.2 are not evidence of the deployed version |
| GATE-03 — Provider capability | Complete for the [recorded 19 September NAS configuration](../engineering/nas-provider-capability-2026-09-19.md): `gpt-5.6-sol` via existing `codex-proxy`, synthetic image/JSON/strict samples supported. Revalidate changed configurations. CH09 validates internal analysis integration with synthetic fixtures; full-schema live compatibility remains unverified |
| GATE-04 — Operating decisions | Allowed destinations before live provider use; cleanup/retention policies before sensitive persistence is enabled; owner-approved hardware/benchmark data before performance and quality evaluation |

Record these decisions before marking an affected design implementation-ready. Unaffected work can proceed against explicit contracts and fixtures. No planning action authorizes private-photo upload, live write tests or production rollout by itself.

Treat final acceptance as a release gate, not an oversized “finish everything” change. The gate aggregates evidence from the owning changes: AC-01–12, manual/GPX/AI-disabled regression, tenant isolation, restart/migration recovery, deployed Immich and provider compatibility, NFR-04/NFR-05 measurements, and an owner-approved quality baseline for Visual and Context-assisted modes. Quality evaluation support belongs with CH09/CH10; performance verification belongs with selection/job changes. Report measured spatial quality, coverage, direction quality and cost separately; do not invent an accuracy threshold before a baseline.

If the gate finds missing behavior, open a bounded fix against its owner. Do not quietly enlarge CH26 or create a catch-all safety backlog. The release gate does not replace tests within each change, and passing document validation does not satisfy it.

Research/web search, neighboring-image sharing, sequence analysis, provider comparison, automatic writeback and universal post-save undo remain outside V1. Direction and multilingual descriptions remain inside. Optional metadata mirroring is capability-gated and user-selected; required local direction/translation records remain usable without it.
