## Why

The confirmed writer currently saves GPS only, leaving reviewed multilingual descriptions local. CH20 lets the user save one chosen language independently of GPS while protecting existing text and retaining the same confirmation and recovery guarantees.

## What Changes

- Add independent GPS and primary-language description choices to single-photo review and exact write confirmation.
- Preserve descriptions by default; offer deliberate replacement or a managed append block whose repeated use does not duplicate text.
- Show the complete before/after description and reject stale, failed or unreviewed dependent text.
- Detect external edits, verify the selected fields after writing, and retain field-specific outcomes when only some intended values are observed.
- Keep ambiguous outcomes and local refresh failures recoverable without blindly resending a write.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-results-and-review`: Extend draft field choices and retained outcomes to reviewed primary-language descriptions.
- `ai-immich-writeback`: Extend exact single-photo previews, confirmed writes and reconciliation to approved standard descriptions.

## Impact

Prerequisite: CH19 implementation and synchronized contracts. The proposed batch applies CH17 first so regenerated descriptions use the same review rules; manually reviewed current text remains usable without a successful regeneration. Traceability: FR-08, FR-10–14 and AC-04/AC-06/AC-08/AC-12.

Affected areas are draft review, exact preview/writeback, the versioned Immich adapter, private approval/audit persistence and AI confirmation UI. Reuse the existing Go/Next.js/SQLite deployment and dependencies; no architectural departure is proposed. Live description compatibility needs its own authorized GATE-02 evidence; a GPS pass alone does not establish it.

Non-goals: stack propagation, direction as a standard Immich field, optional metadata mirroring, description clearing, automatic replacement, universal undo, codex-proxy work and retention-policy changes. All other languages remain local, and local acceptance still makes no Immich mutation.
