## Why

The current catalog selection describes browser state, not an immutable, owner-bound set for later AI work. Before durable jobs can use selected photographs, the server must explain which explicitly selected images are eligible and preserve exactly that selection despite later synchronization or filter changes.

## What Changes

- Resolve explicit asset IDs using the authenticated user's catalog access, image eligibility, visibility policy and active catalog scope; collapse duplicates and explain exclusions without disclosing another user's assets.
- Freeze the resulting exact IDs and source/filter metadata in a private, expiring snapshot with authoritative counts. Later catalog or stack changes cannot add targets.
- Provide bounded preview and snapshot access, with current authorization, installation identity, expiry and AI enablement checks; a snapshot never substitutes for later access or data-sharing consent.
- Preserve source-local capture-date behavior and the existing manual/GPX selection workflows. Establish one eligibility contract for CH06 query selection and later job admission.

## Capabilities

### New Capabilities

- `ai-selection-snapshots`: Private, bounded, expiring snapshots of explicitly selected eligible image IDs, with filter provenance, exclusions and an exact immutable target set.

### Modified Capabilities

None. The implemented `catalog-capture-dates` contract is reused unchanged.

## Impact

CH05 depends on implemented CH04 `align-source-local-capture-dates`. It affects the AI selection core, authenticated backend API, catalog read adapters, additive local snapshot persistence and deterministic authorization/persistence fixtures. It covers the explicit-selection portion of FR-01–02, NFR-01–03 and NFR-07–08, and the applicable isolation/freeze parts of AC-01/AC-09. Planning references are the [PRD](../../../docs/ai-locate/PRD.md), [reconciliation REC-03/REC-05](../../../docs/ai-locate/RECONCILIATION.md) and [roadmap](../../../docs/ai-locate/OPENSPEC_ROADMAP.md).

All-matching selection belongs to CH06. This change does not add an analysis launch UI, jobs, image preparation, provider calls, context sharing, result validation, writeback, implicit stack expansion or live Immich compatibility claims. Preview is independently testable through its protected API; the complete launch UI remains CH13. No departure from the adopted engineering standards is proposed.
