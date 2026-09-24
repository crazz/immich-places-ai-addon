# Immich Places AI Add-on — planning package

English product and technical planning documents for `crazz/immich-places-ai-addon`.

Version 1.0 + checkout reconciliation + adopted engineering standards · 17 September 2026 · Product baseline for OpenSpec planning

Reconciled with [`5e70c61`](https://github.com/crazz/immich-places-ai-addon/tree/5e70c6165777949c9d8b50ede3b2768bcaa5df87). Start with [RECONCILIATION.md](RECONCILIATION.md) for confirmed behavior, required integration corrections, and remaining gates.

Before planning, implementation or review, read the adopted [architecture](../engineering/architecture.md), [testing](../engineering/testing.md) and [coding standards](../engineering/coding-standards.md). These own engineering policy; this package owns AI product requirements and integration planning. [ADR-07](../engineering/decisions/ADR-07-ai-internal-packages.md) records the accepted internal-package amendment. The F01–F03 [tooling prerequisite](OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite) is complete; CH01 adds backend AI coverage enforcement.

## Research planning amendment — 20 September 2026

[ADR-09](../engineering/decisions/ADR-09-approximate-ai-location-proposals.md) and the active [Research proposal](../../openspec/changes/research-photo-locations-with-web-evidence/proposal.md) add best-effort coordinates, unrestricted estimated-error magnitude and optional answer-provided links. Search metadata and a proxy extension are not required. Implementation is complete in the active change; see the [Research guide](../ai-research.md) and [verification record](../../openspec/changes/research-photo-locations-with-web-evidence/implementation-verification.md). Historical v1 contracts remain unchanged.

## Files

- [PRD.md](PRD.md): product scope, stable requirements, acceptance scenarios, and release gates.
- [TECHNICAL_DESIGN.md](TECHNICAL_DESIGN.md): architecture, integration constraints, persistence, jobs, APIs, writeback, testing, and operations.
- [OPENSPEC_HANDOFF.md](OPENSPEC_HANDOFF.md): capability mapping and a reference planning prompt; not an implementation backlog or an instruction executed during reconciliation.
- [OPENSPEC_ROADMAP.md](OPENSPEC_ROADMAP.md): eight capability areas decomposed into 26 smaller candidate changes, dependencies, requirement ownership, and artifact-generation approach.
- [CH10 and CH12–CH15 planning batch](CH10_CH12_CH15_PLAN.md): change order, dependencies, requirement/task mapping and current implementation status.
- [CH16 → CH18 → CH19 planning batch](CH16_CH18_CH19_PLAN.md): prepared local drafts, exact GPS previews and confirmed single-photo writes, with ordered prerequisites and verification plans. These changes are not implemented; codex-proxy work is deferred.
- [Consented context](../ai-consented-context.md): CH10's internal Context-assisted contract and verification.
- [Persistent history](../ai-results-history.md): CH14 private terminal history, stable pagination and separately authorized current images.
- [Proposal review](../ai-proposal-review.md): CH15 read-only camera/subject geometry, uncertainty, evidence and exact language descriptions.
- [Batch workflow](../ai-batch-workflow.md): CH13 gallery preview/consent, durable Context execution, progress, cancellation and explicit reruns.
- [ENGINEERING_STANDARDS_PROPOSAL.md](ENGINEERING_STANDARDS_PROPOSAL.md): adoption record linking the authoritative engineering standards; the earlier proposal is superseded.
- [SOURCES.md](SOURCES.md): pinned repository references, package provenance, and verification limits.
- [Analysis-result schema](../../backend/internal/ai/results/ai-analysis-result.v1.schema.json): canonical v1.0 model-output contract embedded by CH07.
- [Unknown-location example](examples/ai-analysis-result.unknown.json) and [synthetic example](examples/ai-analysis-result.synthetic.json): contract fixtures; the synthetic example is not a real geolocation.
- [Semantic validation](examples/SEMANTIC_VALIDATION.md): checks required in addition to JSON Schema.
- [Original PRD export](exports/Immich_Places_AI_PRD.docx) and [original design export](exports/Immich_Places_AI_Technical_Design.docx): unchanged Word reading copies from package v1.0. They do **not** include the checkout reconciliation or adopted engineering amendments.

Markdown is the planning source of truth. Requirement IDs, V1 scope, and the provider contract are preserved. The reconciliation clarifies existing behavior and implementation gaps; it does not mark proposed AI features as implemented. Future OpenSpec changes belong under the repository's existing `openspec/` workspace and should link to this baseline.

## Verification boundary

The original package used browser inspection of public `main`. Local source and GitNexus traces have now been reconciled against `5e70c6165777949c9d8b50ede3b2768bcaa5df87`; GATE-01 is complete for that checkout. Reconcile again if implementation starts from a different revision. Referenced Immich API source remains pinned to v3.2.2; the deployed server and provider capabilities remain unverified.

This reconciliation imports and updates planning documentation only. No application build, Go test suite, application migration, private-photo upload, or Immich write was performed. Go was unavailable in the session environment. The canonical schema and both examples passed local structural validation; a separate SQLite probe reproduced the date-bucketing discrepancy. See the reconciliation report for the exact verification boundary. Nothing was committed or pushed.

Subsequent standards adoption updated agent instructions, OpenSpec configuration and the planning documents. It did not implement application behavior, test harnesses or CI checks. The original reconciliation report remains historical evidence for the pinned source checkout.

Subsequent F01–F03 implementation installed shared engineering checks, frontend unit/component tests and deterministic browser smoke journeys. See the dated [verification baseline](../engineering/verification-baseline.md) for actual results; the historical reconciliation above does not describe current tooling availability.

## Main design decisions

Extend the existing Go/Next.js/SQLite application. Treat AI as a proposal producer. Retain camera/subject separation, nullable heading, selected-language descriptions, durable jobs/results, and explicit field-level approval. Use an exact-target API writer that does not inherit implicit stack expansion. Keep full results locally and mirror optional extended metadata only when supported and approved.

CH04 implements the catalog date-query portion of FR-01: [maintained capture-date contract](../../openspec/specs/catalog-capture-dates/spec.md). It preserves recorded calendar days and raw timestamps, rejects reversed ranges, and keeps undated assets in unbounded browsing. The undated-group UI remains planned; CH05–CH06 implement AI selection/eligibility below.

CH01 implements private provider configuration: [maintained provider contract](../../openspec/specs/ai-provider-configuration/spec.md) and [operator guide](../ai-provider-settings.md). Profiles are encrypted, revisioned, user-scoped and globally disabled by default. Saving does not contact a provider.

CH02 [provider egress policy](../../openspec/changes/archive/2026-09-19-enforce-ai-provider-egress-policy/proposal.md) and CH03 [provider capability tests](../../openspec/changes/archive/2026-09-19-add-ai-provider-capability-tests/proposal.md) are implemented and archived. CH02 enforces approved destinations and bounded revision-bound dispatch. CH03 provides an explicit synthetic image/JSON/strict test, private persisted observations and Settings results. CI exercises deterministic local fixtures. An authorized [NAS check on 19 September 2026](../engineering/nas-provider-capability-2026-09-19.md) additionally observed image, JSON and strict-schema sample support for `gpt-5.6-sol` through the existing codex-proxy, closing GATE-03 for that recorded configuration. The isolated test did not deploy the addon into production Places. Production analysis admission and writeback remain planned; CH11 adds internal durable jobs below.

CH05 implements [explicit selection snapshots](../../openspec/specs/ai-selection-snapshots/spec.md): strict owner-scoped preview/read APIs, scope-aware eligibility, immutable expiring manifests, installation binding and bounded cleanup. See the [operator guide](../ai-selection-snapshots.md) and [verification](../engineering/ai-selection-verification.md). It makes no provider/image-fetch/Immich-write calls and adds no launch UI.

CH06 implements all-matching selection through the same snapshot contract: whole-catalog scope resolution, exact aggregate counts, eligible-asset limits and frozen query membership. See [CH06 verification](../engineering/ai-matching-selection-verification.md).

CH07 implements [result validation](../../openspec/specs/ai-location-proposals/spec.md): one embedded canonical schema, bounded parsing, semantic/evidence/language checks and immutable proposals without write authority. See the [caller contract](../ai-result-validation.md) and [verification](../engineering/ai-result-validation-verification.md). CH09 adds internal provider analysis; CH11 adds durable jobs/results exercised with synthetic execution.

CH08 implements [bounded image preparation](../ai-image-preparation.md): authorized exact-asset previews, orientation normalization, metadata removal and transient JPEG copies with decode/transmission limits. The internal adapter rechecks authority and source freshness; it has no public dispatch route. See [verification](../engineering/ai-image-preparation-verification.md) for synthetic evidence and live-compatibility limits.

CH09 implements [single-image Visual analysis](../ai-visual-analysis.md) through the approved transport, exact private revisions and current image authority. It requires explicit caller authorization and one dispatch reservation, validates the complete canonical result and returns an immutable handoff. See [verification](../engineering/ai-visual-analysis-verification.md). No public route or durable worker is connected; full-schema model compatibility and geographic quality remain unverified.

CH11 implements [recoverable durable jobs](../ai-analysis-jobs.md): private idempotent submissions, bounded leases/calls, cancellation/restart fencing, immutable validated history and controlled installation/retention lifecycle. See [verification](../engineering/ai-analysis-jobs-verification.md). A synthetic executor proves the lifecycle; CH12 still owns production admission and the real Visual connection.


CH10 and CH12–CH15 are now implemented and archived. [The batch record](CH10_CH12_CH15_PLAN.md) records their dependency order and verification. CH12 supersedes the earlier internal-only Visual/job limits; CH13 also connects consented Context-assisted work. CH14 retains private history independently of source availability, and CH15 provides isolated read-only inspection. Draft editing, translation operations and confirmed writing remain planned.
