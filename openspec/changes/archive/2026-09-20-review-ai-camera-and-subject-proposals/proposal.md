## Why

A stored proposal is useful only when the user can distinguish the photographer's possible position from the subject, uncertainty and unsupported guesses. CH15 adds an accessible read-only map and evidence panel over CH14's immutable detail so a plausible landmark is never silently presented as the camera location.

## What Changes

- Show distinct camera and subject markers, explicitly inspectable alternatives and estimated uncertainty only when supplied and supported.
- Present nullable true-north camera heading separately from a bearing toward the subject, with method and uncertainty.
- Display concise evidence, source lineage, limitations and warnings without turning model guesses into external verification.
- Add requested-language tabs with explicit text/status and primary-language indication, without fabricated translations.
- Preserve keyboard access and complete numeric/text presentation when map tiles or source images are unavailable.
- Keep all inspection read-only and isolated from manual pending coordinates, drafts and Immich writes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-results-and-review`: read-only spatial, evidence and multilingual proposal inspection. CH14 creates this capability; apply and sync CH14 before this dependent delta.

## Impact

Depends on planned [CH14](../2026-09-20-browse-persistent-ai-results/proposal.md), consuming CH07's canonical proposal and CH10/CH13 provenance when present. Adds focused view models/panels under `src/features/ai/` and a narrow read-only overlay integration with the existing Leaflet map. Reuses installed map/UI libraries; no new dependency or schema change is planned.

CH16 owns choosing/editing/accepting a durable draft, and CH17 owns translation correction/retry. This change has no approval or write actions. Covers FR-06–09/13, AC-02/03 and NFR-01/06 in the [PRD](../../../../docs/ai-locate/PRD.md), [technical design](../../../../docs/ai-locate/TECHNICAL_DESIGN.md) and [roadmap](../../../../docs/ai-locate/OPENSPEC_ROADMAP.md).
