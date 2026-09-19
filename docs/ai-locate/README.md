# Immich Places AI Add-on — planning package

English product and technical planning documents for `crazz/immich-places-ai-addon`.

Version 1.0 + checkout reconciliation + adopted engineering standards · 17 September 2026 · Product baseline for OpenSpec planning

Reconciled with [`5e70c61`](https://github.com/crazz/immich-places-ai-addon/tree/5e70c6165777949c9d8b50ede3b2768bcaa5df87). Start with [RECONCILIATION.md](RECONCILIATION.md) for confirmed behavior, required integration corrections, and remaining gates.

Before planning, implementation or review, read the adopted [architecture](../engineering/architecture.md), [testing](../engineering/testing.md) and [coding standards](../engineering/coding-standards.md). These own engineering policy; this package owns AI product requirements and integration planning. [ADR-07](../engineering/decisions/ADR-07-ai-internal-packages.md) records the accepted internal-package amendment. The F01–F03 [tooling prerequisite](OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite) is complete; CH01 adds backend AI coverage enforcement.

## Files

- [PRD.md](PRD.md): product scope, stable requirements, acceptance scenarios, and release gates.
- [TECHNICAL_DESIGN.md](TECHNICAL_DESIGN.md): architecture, integration constraints, persistence, jobs, APIs, writeback, testing, and operations.
- [OPENSPEC_HANDOFF.md](OPENSPEC_HANDOFF.md): capability mapping and a reference planning prompt; not an implementation backlog or an instruction executed during reconciliation.
- [OPENSPEC_ROADMAP.md](OPENSPEC_ROADMAP.md): eight capability areas decomposed into 26 smaller candidate changes, dependencies, requirement ownership, and artifact-generation approach.
- [ENGINEERING_STANDARDS_PROPOSAL.md](ENGINEERING_STANDARDS_PROPOSAL.md): adoption record linking the authoritative engineering standards; the earlier proposal is superseded.
- [SOURCES.md](SOURCES.md): pinned repository references, package provenance, and verification limits.
- [Analysis-result schema](contracts/ai-analysis-result.v1.schema.json): canonical proposed model-output contract.
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

CH04 implements the catalog date-query portion of FR-01: [maintained capture-date contract](../../openspec/specs/catalog-capture-dates/spec.md). It preserves recorded calendar days and raw timestamps, rejects reversed ranges, and keeps undated assets in unbounded browsing. The undated-group UI and AI selection/eligibility remain planned.

CH01 implements private provider configuration: [maintained provider contract](../../openspec/specs/ai-provider-configuration/spec.md) and [operator guide](../ai-provider-settings.md). Profiles are encrypted, revisioned, user-scoped and globally disabled by default. Saving does not contact a provider.

CH02 [provider egress policy](../../openspec/changes/archive/2026-09-19-enforce-ai-provider-egress-policy/proposal.md) and CH03 [provider capability tests](../../openspec/changes/archive/2026-09-19-add-ai-provider-capability-tests/proposal.md) are implemented and archived. CH02 enforces approved destinations and bounded revision-bound dispatch. CH03 provides an explicit synthetic image/JSON/strict test, private persisted observations and Settings results. CI exercises deterministic local fixtures. An authorized [NAS check on 19 September 2026](../engineering/nas-provider-capability-2026-09-19.md) additionally observed image, JSON and strict-schema sample support for `gpt-5.6-sol` through the existing codex-proxy, closing GATE-03 for that recorded configuration. The isolated test did not deploy the addon into production Places. Analysis, jobs and writeback remain planned.
