## Why

The selection and durable execution contracts need a usable launch and progress workflow. CH13 lets users explicitly authorize a frozen batch, follow its durable state and start new runs without losing earlier results, including Context-assisted analysis once CH10 is available.

## What Changes

- Add AI launch controls for selected assets, the current page and all matching eligible assets through the same frozen preview contract.
- Show exact provider/mode/language/limit choices, exclusions and separate image/context consent before submission.
- Connect CH10's Context-assisted attempt to durable jobs with frozen evidence provenance and correct mode-specific completion validation.
- Show recoverable progress, safe per-item failures, cancellation and honest usage status across refresh/navigation.
- Make retry-failed and reanalysis explicit new runs with fresh preview/consent, preserving all earlier results.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-analysis-jobs`: complete launch/progress experience, mode-aware durable admission/completion and explicit reruns.

## Impact

Depends on CH06 and the planned [CH10](../2026-09-20-add-consented-ai-context/proposal.md) and [CH12](../2026-09-20-connect-visual-analysis-to-durable-jobs/proposal.md). Apply and sync those changes first. Affects `src/features/ai/`, narrow gallery integration, jobs admission/execution and additive provenance storage. Uses existing frontend, map, HTTP and Go dependencies.

CH14 owns persistent result browsing and CH15 owns detailed proposal presentation. Draft editing/acceptance, translation-only reruns, writes, Research and automatic context sharing are outside this change. Covers FR-02/04/05/09 and NFR-01–07 in the [PRD](../../../../docs/ai-locate/PRD.md) and the [roadmap](../../../../docs/ai-locate/OPENSPEC_ROADMAP.md).
