## Why

The AI writer deliberately targets one photo, while some reviewed stack members may share an appropriate camera position. CH21 lets the user approve an exact additional GPS target list without inheriting the legacy writer's implicit stack expansion.

## What Changes

- Offer an explicit stack-member review with independent visibility, source and GPS baseline checks for every selected photo.
- Keep the analyzed photo as the default target and require deliberate selection of additional members, with clear viewpoint warnings.
- Freeze the displayed target list and approved GPS pair; later stack additions never enlarge it.
- Show per-target success, conflict and unresolved outcomes, preserving completed work when another member cannot be written.
- Keep description and direction choices restricted to their individually reviewed photo and preserve overlap protection, durable audit and readback for every target.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-results-and-review`: Add exact stack GPS target selection and accessible per-target write outcomes.
- `ai-immich-writeback`: Extend immutable plans and confirmed execution to an explicitly reviewed bounded target list.

## Impact

Prerequisite: CH19. Apply after CH20 in this batch and reconcile the shared writeback deltas before implementation; this ordering prevents a later single-photo delta from undoing the multi-target contract. Traceability: FR-02, FR-10–12, FR-14 and AC-04–06/AC-09/AC-11.

Affected areas are private target review, stack reads, preview/writeback and per-target persistence, catalog refresh and AI results UI. Reuse existing Go/Next.js/SQLite services and dependencies; no architectural departure is proposed. Live stack compatibility remains separately gated.

Non-goals: automatic selection of all members, arbitrary cross-stack batches, changing stack membership, description/direction propagation, cross-target atomicity, GPS clearing, codex-proxy changes and new retention policy. Stack membership is not evidence of a shared viewpoint or authority to overwrite GPS.
