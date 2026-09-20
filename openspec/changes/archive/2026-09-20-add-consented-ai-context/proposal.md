## Why

Visual analysis cannot use helpful capture-time, selected-album or nearby-location clues, while existing catalog suggestions are not safe to disclose wholesale. CH10 gives later Context-assisted jobs an explicitly consented, bounded evidence input without turning private metadata or earlier AI guesses into independent confirmation.

## What Changes

- Add an internal Context-assisted attempt alongside Visual, retaining one prepared target image and the canonical result contract.
- Admit only separately authorized context classes: capture time with source-local/unknown-zone status, the explicitly selected album label, a short user hint and at most six metadata-only nearby sources.
- Filter each source for current ownership/access, visibility, temporal relevance and lineage; omit unrelated frequent/home locations and known AI-origin neighbors.
- Preserve backend-assigned source identities and provenance through validation and the transient result handoff, without presenting missing provenance as verification.
- Keep Visual payloads context-free and retain bounded calls, safe failures, cancellation and zero Immich mutations.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-location-proposals`: add consent-filtered context preparation, Context-assisted dispatch and authoritative evidence validation while preserving the existing Visual contract.

## Impact

Depends on implemented CH04 and CH09, with CH07/CH08 validation and image boundaries retained. Affects internal analysis, read-only context acquisition, provider serialization and evidence handoff. CH13 will connect this contract to durable mode admission and the launch UI after CH12's Visual integration.

No new dependency or architectural departure is proposed. Public context/job routes, startup execution, neighboring image sharing, Research/web/geocoder calls, AI-origin context opt-in, drafts, writes and production rollout are outside this change. Covers FR-01/04/06/07 and NFR-01/03/07 in the [PRD](../../../../docs/ai-locate/PRD.md), [evidence design](../../../../docs/ai-locate/TECHNICAL_DESIGN.md#5-analysis-pipeline-and-evidence-model) and REC-04 in the [reconciliation](../../../../docs/ai-locate/RECONCILIATION.md#rec-04--existing-suggestions-need-a-consent-filtered-adapter).
