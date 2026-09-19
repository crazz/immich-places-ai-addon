## Why

Selecting every photo matching a catalog filter must not mean selecting only the loaded page. CH06 extends the frozen-selection contract so the user can obtain the complete eligible set and honest counts without an oversized query silently becoming a smaller batch.

## What Changes

- Resolve all matching assets across the complete authorized catalog using the same normalized album, recursive folder, tag, GPS, visibility and source-local date scope as CH05.
- Apply the shared eligibility policy, expose exact matched/eligible/excluded counts and aggregate exclusion reasons, and freeze the eligible membership in one consistent operation.
- Reject a selection exceeding the configured eligible-asset limit without truncation or creating a snapshot; explain empty and excluded-only selections without creating an empty resource.
- Preserve snapshot ownership, installation binding, expiry, retention, current-access revalidation and the separation from analysis consent and Immich writes.
- Keep loaded-page selection and the existing gallery behavior unchanged. Job submission, launch UI, provider calls, stack expansion and writes remain outside this change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-selection-snapshots`: add complete-query resolution, exact bounded counting and immutable all-matching membership. This capability is introduced by CH05, not yet a main spec at planning time; CH05 must be implemented, verified and synced before applying this additive delta.

## Impact

Extends the selection domain and protected preview API introduced by CH05, together with its SQLite adapter. Reuses CH04 catalog date predicates and the current owner-qualified gallery scope; it introduces no new service, dependency, provider integration or public launch workflow.

Prerequisite: [CH05](../archive/2026-09-20-freeze-explicit-ai-selections/proposal.md). Covers FR-01–02, NFR-01–05/NFR-07–08 as applicable to selection, and AC-01/AC-09 from the [PRD](../../../docs/ai-locate/PRD.md). Performance evidence here is a bounded resolver measurement; submission and end-to-end gallery latency acceptance remain with their roadmap owners. No departure from the adopted engineering standards is proposed.
