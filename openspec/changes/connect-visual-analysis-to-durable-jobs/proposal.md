## Why

CH09 can analyze one prepared image and CH11 can recover durable work, but no authenticated production path connects them. CH12 makes an explicitly authorized Visual batch executable with bounded usage, fresh access checks and durable results, while preserving the existing no-write boundary.

## What Changes

- Admit frozen selection tokens through authenticated submit, progress and cancel APIs, recording exact provider revision, image consent, languages, membership and limits.
- Connect the existing image preparation and Visual runner to the durable worker, with startup/shutdown lifecycle, current authority checks and explicit failure translation.
- Reserve finite call and token allowances before dispatch, using an explicitly configured provider token policy; report usage and monetary estimates honestly.
- Preserve immutable successes across retries, cancellation and restart, and expose only safe owner-scoped progress.
- Remove the earlier internal-only restrictions from the maintained job contract without enabling automatic retention or Immich writes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-analysis-jobs`: production Visual admission, execution, budgets, progress and cancellation.
- `ai-provider-configuration`: explicit revision-bound production token policy with no inferred compatibility.
- `ai-location-proposals`: allow only the approved output-token limit in the otherwise minimized provider payload.

## Impact

Depends on implemented CH05–CH09 and CH11. Uses the same Go process, SQLite database, approved provider transport and read-only Immich adapters. Adds ordered migrations for admission/usage metadata and backend-only execution settings; no new library is proposed.

CH13 owns the launch/progress UI, Context-assisted durable dispatch and reruns. History browsing, drafts, writeback, automatic retention and production deployment are outside this change. Covers FR-02–05/09 and NFR-01–05/07 in the [PRD](../../../docs/ai-locate/PRD.md), [execution design](../../../docs/ai-locate/TECHNICAL_DESIGN.md#8-durable-execution-states-and-budgets) and [roadmap](../../../docs/ai-locate/OPENSPEC_ROADMAP.md).
