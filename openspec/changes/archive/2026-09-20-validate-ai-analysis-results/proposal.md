## Why

A provider returning valid-looking JSON has not yet produced a trustworthy application proposal. CH07 establishes a deterministic validation boundary so malformed contracts, invented references and inconsistent location/language claims cannot enter later analysis and review flows as validated results.

## What Changes

- Validate model output against the existing canonical analysis-result v1 schema, with bounded parsing and explicit rejection of unsupported versions, ambiguous JSON and unknown fields; retain the distinction between a valid unknown location and a refused, truncated or tool response reported by the server-side transport.
- Enforce outcome/candidate consistency, unique IDs, referential integrity and source references restricted to the server-authorized input evidence bundle.
- Preserve camera/subject separation, finite geographic bounds, honest radius provenance and nullable supported camera direction; never repair uncertainty by inventing coordinates or promoting a subject point.
- Enforce normalized requested-language coverage and complete/unavailable status consistency, independently of provider claims.
- Return deterministic validated proposals or sanitized validation failures to internal callers while keeping application identity, evidence authority and future approval state outside model output.
- Keep provider invocation, image preparation, persistence, queue state, result UI and Immich writes out of scope. Validation establishes contract consistency, not factual geolocation accuracy or consent.

## Capabilities

### New Capabilities

- `ai-location-proposals`: internal canonical and semantic validation of bounded AI analysis results for later authorized analysis consumers.

### Modified Capabilities

None.

## Impact

Introduces a focused Go AI result-validation package and deterministic fixtures, using the [canonical schema](../../../../backend/internal/ai/results/ai-analysis-result.v1.schema.json) and [semantic baseline](../../../../docs/ai-locate/examples/SEMANTIC_VALIDATION.md). During implementation, relocate the single canonical schema into that package for embedding and update documentation links; preserve its version and contents rather than create a second handwritten schema. The design resolves the schema-validator dependency; no transport or database adapter is introduced. No preceding product change is required; CH09 and CH11 consume this contract later.

Covers FR-06–08, the internal trust boundary of NFR-01, NFR-07–08, and AC-02/AC-03/AC-10 from the [PRD](../../../../docs/ai-locate/PRD.md). Later consumers retain authorization, persistence, review and write responsibilities. No departure from the adopted architecture, testing or coding standards is proposed.

The owner approved `santhosh-tekuri/jsonschema/v6` v6.0.3 following [library-evaluation.md](library-evaluation.md). Its integration and verification boundaries are captured in the design and tasks; structural document validation does not establish an implemented or tested runtime validator.
