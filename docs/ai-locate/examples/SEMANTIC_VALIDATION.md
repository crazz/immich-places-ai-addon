# Semantic validation beyond the JSON Schema

The canonical JSON Schema validates structure and basic numeric bounds. A provider-specific transport schema may need simplification for that endpoint's supported subset; the backend still validates against the canonical schema.

The following checks remain mandatory in application code:

1. `located` has a selected candidate with non-null camera coordinates. `ambiguous` has at least two candidates and no automatic selection. `unknown` has no selected candidate.
2. Observation and candidate IDs are unique. Every selected ID and evidence reference exists. Every source reference belongs to the backend's authorized input evidence bundle.
3. Requested language tags are valid, normalized, unique, and covered exactly once. Complete descriptions have nonempty text and no unavailable reason; unavailable descriptions have null text and a nonempty reason.
4. Candidate-based descriptions identify an existing candidate. Scene-only descriptions have null candidate IDs.
5. Unknown radius basis requires null radius; non-null radius requires a supported basis. Reject non-finite numbers and misleading precision/granularity combinations.
6. A proposed camera direction requires a selected/reviewable camera viewpoint and supporting visual or known-viewpoint evidence. Merely computing a bearing to the subject is insufficient.
7. Bounds and source-local dates are checked independently of the model. Neither model output nor its IDs authorize an application resource or an Immich write.
8. Model warnings and evidence are untrusted text; render safely and do not execute links or instructions.

The example files pass the canonical JSON Schema. They have not been sent to a live provider. The synthetic fixture is deliberately invented, not a photo-location finding or a source of real-world geographic facts.
