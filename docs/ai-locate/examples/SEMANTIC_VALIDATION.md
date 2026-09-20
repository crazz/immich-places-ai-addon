# Semantic validation beyond the JSON Schema

> Version boundary: the rules below describe implemented v1 validation. The planned [Research v2 contract](../../../openspec/changes/research-photo-locations-with-web-evidence/specs/ai-location-proposals/spec.md), following [ADR-09](../../engineering/decisions/ADR-09-approximate-ai-location-proposals.md), accepts answer-provided sources and model-estimated city/region radii without an upper precision cutoff. It preserves the v1 fixtures and does not claim that v2 is implemented.

The [canonical embedded JSON Schema](../../../backend/internal/ai/results/ai-analysis-result.v1.schema.json) validates structure and exact numeric bounds. CH07 now implements the accompanying semantic checks in `backend/internal/ai/results/`; see the [caller contract and budgets](../../ai-result-validation.md). A provider-specific transport schema may need simplification for that endpoint's supported subset; the backend still validates against the canonical schema.

The following checks remain mandatory in application code:

1. `located` has a selected candidate with non-null camera coordinates. `ambiguous` has at least two candidates and no automatic selection. `unknown` has no selected candidate.
2. Observation and candidate IDs are unique. Every selected ID and evidence reference exists. Every source reference belongs to the backend's authorized input evidence bundle.
3. Requested language tags are valid, normalized, unique, and covered exactly once. Complete descriptions have nonempty text and no unavailable reason; unavailable descriptions have null text and a nonempty reason.
4. Candidate-based descriptions identify an existing candidate. Scene-only descriptions have null candidate IDs.
5. Null radius and unknown basis are paired in both directions. Numeric radii need referenced visual or server-attested context/source evidence; only point granularity permits zero, and city/region require null/unknown. Accepted values remain uncalibrated.
6. A proposed camera direction requires a selected/reviewable camera viewpoint and supporting visual or known-viewpoint evidence. Merely computing a bearing to the subject is insufficient.
7. Coordinate bounds are checked independently of the model. Consumers remain responsible for current source/capture-date eligibility, consent and authorization. Neither model output nor its IDs authorize an application resource or an Immich write.
8. Model warnings and evidence are untrusted text; render safely and do not execute links or instructions.

The example files pass the implemented canonical and semantic validator with their matching Visual language contexts. Their embedded test copies are byte-identical to these examples at CH07 verification. They have not been sent to a live provider. The synthetic fixture is deliberately invented, not a photo-location finding or a source of real-world geographic facts.

Parsing also enforces bounded bytes/collections/strings/numbers, duplicate-key and Unicode integrity. Transport refusal/truncation/tool responses cannot masquerade as completed unknown outcomes. Server language/source context is separate; errors contain bounded safe paths/codes, and successful proposals expose only independent copies.
