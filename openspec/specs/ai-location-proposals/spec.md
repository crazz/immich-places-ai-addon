# AI location proposals

## Purpose

Validate untrusted AI analysis output as a bounded, internally consistent location proposal while preserving uncertainty, server-owned evidence authority and the separate requirement for user review and confirmed writes.

## Requirements

### Requirement: Enforce the canonical bounded result contract

The validator SHALL accept only a complete object conforming to canonical analysis-result schema version 1.0 and the semantic rules below. Server-supplied completion metadata SHALL explicitly identify a complete ordinary result; refused, truncated, tool-response or missing/unknown completion state SHALL fail even if the accompanying bytes look like a complete object. It SHALL reject unsupported versions, missing or unknown properties, wrong types, invalid numeric values, duplicate JSON keys at any depth, trailing content and malformed or truncated JSON. Input bytes, nesting, collection sizes and string lengths SHALL have explicit finite limits; exceeding a limit SHALL fail the complete result rather than truncate or repair it. Schema validation SHALL perform no external resource lookup.

#### Scenario: Canonical fixtures remain valid
- **GIVEN** the canonical synthetic located fixture and unknown-location fixture with their corresponding requested languages and empty Visual evidence bundles
- **WHEN** each complete object is validated
- **THEN** both are accepted without inventing or discarding fields

#### Scenario: Structurally invalid or ambiguous output fails closed
- **GIVEN** output with an unsupported version, missing required nullable field, extra property, wrong type, repeated key, trailing object or truncated JSON
- **WHEN** it is validated
- **THEN** validation fails and returns no validated proposal
- **AND** no missing content is supplied by guessing or defaults

#### Scenario: Numeric and resource bounds are enforced
- **GIVEN** output with overflowing/non-finite numbers or excessive bytes, depth, collection items or text length
- **WHEN** it is validated
- **THEN** the whole result fails with a bounded diagnostic
- **AND** validation does not fetch schemas or follow model-provided links

#### Scenario: Transport failure cannot masquerade as a completed proposal
- **GIVEN** server-supplied completion state reports refusal, truncation, tool response or an absent/unknown state
- **WHEN** even structurally valid result bytes are supplied
- **THEN** validation rejects the attempt without interpreting that state as an unknown-location outcome

### Requirement: Keep outcomes consistent with candidate selection

A `located` result SHALL identify an existing selected candidate with non-null camera coordinates. An `ambiguous` result SHALL contain at least two candidates and no selected candidate. An `unknown` result SHALL have no selected candidate. Validation SHALL preserve these distinctions and SHALL NOT choose, fabricate or promote a camera point for unknown/ambiguous outcomes. A valid unknown result SHALL remain a completed proposal outcome rather than a technical failure.

#### Scenario: Located output selects a supported camera candidate
- **GIVEN** a valid candidate with camera coordinates and a matching selected candidate ID
- **WHEN** the outcome is located
- **THEN** the selected candidate is retained as an unapproved proposal

#### Scenario: Ambiguous and unknown outcomes remain unselected
- **GIVEN** a valid ambiguous result with two candidates or an unknown result with no selection
- **WHEN** validation completes
- **THEN** the original outcome is retained without automatically choosing a candidate or inserting zero coordinates

#### Scenario: Inconsistent outcome selection is rejected
- **GIVEN** a located result without a camera-bearing selected candidate, an ambiguous result with one candidate or a selection, or an unknown result with a selection
- **WHEN** validation runs
- **THEN** the complete result is rejected rather than silently changing its outcome

### Requirement: Resolve identifiers only within authorized evidence

Observation and candidate identifiers SHALL be nonblank and unique within their respective collections. Observation text, candidate place names and support summaries SHALL be nonblank; optional country codes SHALL be null or valid ISO 3166-1 alpha-2 codes. Selected IDs, candidate evidence references and candidate-dependent description references SHALL resolve to the applicable collection. Source references SHALL resolve only within the authoritative input evidence bundle supplied separately by the server; model-written URLs or IDs SHALL NOT create source authority. Provided-context observations SHALL require Context-assisted mode and a nonempty authorized bundle; a candidate citing such observations SHALL also cite at least one authorized source. Visual-mode output SHALL contain no provided-context observations or source references when its authorized bundle is empty. Duplicate references SHALL be rejected rather than inflating apparent support.

#### Scenario: Local references resolve and preserve their kinds
- **GIVEN** unique candidates and observations with valid evidence references and source IDs already present in the authorized input bundle
- **WHEN** validation runs in the corresponding mode
- **THEN** those references are retained without upgrading visual observations to external evidence

#### Scenario: Dangling and duplicate identities are rejected
- **GIVEN** duplicate or blank IDs, blank required evidence text, an invalid country code, a missing selected/evidence/description candidate ID, or a repeated evidence/source reference
- **WHEN** validation runs
- **THEN** the result fails without manufacturing a replacement ID or reference

#### Scenario: Foreign or invented source claims confer no authority
- **GIVEN** output references an ID from another user's bundle or a new URL absent from this request's authorized bundle
- **WHEN** validation runs
- **THEN** the result is rejected and neither source is loaded or disclosed

#### Scenario: Visual output cannot claim supplied context
- **GIVEN** Visual mode with an empty authorized source bundle
- **WHEN** output includes a provided-context observation or a nonempty source reference
- **THEN** validation rejects the unsupported provenance claim

#### Scenario: Context observations retain a source connection
- **GIVEN** Context-assisted output contains a provided-context observation
- **WHEN** the authorized bundle is empty or a candidate cites that observation without an authorized source reference
- **THEN** the result is rejected for unsupported context provenance

### Requirement: Preserve separate camera and subject coordinates

Camera and subject coordinates SHALL remain separate finite WGS84 latitude/longitude pairs within their canonical ranges. Zero SHALL remain a valid numeric coordinate and SHALL NOT stand for absence. A subject point alone SHALL NOT satisfy a camera location requirement, and validation SHALL NOT copy or infer one pair from the other. Both pairs SHALL remain explicit nullable values where the canonical schema permits absence.

#### Scenario: A landmark does not become the photographer position
- **GIVEN** an unknown result whose candidate has a subject location and null camera location and direction
- **WHEN** validation succeeds
- **THEN** the subject remains a separate marker proposal and no camera selection is created

#### Scenario: Boundary and zero coordinates remain meaningful
- **GIVEN** otherwise valid coordinates at latitude -90 or 90, longitude -180 or 180, or a pair containing zero
- **WHEN** validation runs
- **THEN** valid boundary/zero values are preserved
- **AND** values beyond those ranges are rejected

### Requirement: Require consistent and conservative radius claims

An unknown radius basis SHALL require a null estimated radius, and a null radius SHALL use the unknown basis. A numeric radius SHALL have an allowed supporting basis and be finite and nonnegative; non-point granularity SHALL NOT claim zero radius. City and region estimates SHALL use null radius and unknown basis in this initial policy, avoiding unsupported fine numeric precision without inventing calibrated accuracy thresholds. Context-extent and source-reported bases SHALL require a referenced authorized input source that supports that basis. A visual estimate SHALL have referenced visual observations. Accepted radii SHALL remain uncalibrated estimates, never verified error guarantees.

#### Scenario: Supported visual estimate or unknown radius is retained
- **GIVEN** a visually supported area estimate with radius 250 meters, or a camera estimate with null radius and unknown basis
- **WHEN** validation runs
- **THEN** the corresponding radius state is retained as an uncalibrated proposal

#### Scenario: Contradictory or misleading precision is rejected
- **GIVEN** unknown basis with numeric radius, null radius with a claimed basis, non-point zero radius, or a city/region estimate with numeric radius
- **WHEN** validation runs
- **THEN** the complete result is rejected without silently nulling or enlarging the radius

#### Scenario: Radius provenance requires the matching evidence kind
- **GIVEN** a context-extent or source-reported radius without a referenced authorized source supporting that kind, or a visual radius without referenced visual observations
- **WHEN** validation runs
- **THEN** validation rejects the unsupported basis claim

### Requirement: Require a reviewable viewpoint and evidence for direction

Direction SHALL be null when not proposed. A non-null camera direction SHALL be attached to a candidate with a non-null reviewable camera viewpoint, true-north azimuth in `[0, 360)`, a supported method and nullable uncertainty within `[0, 180]`. A visual-estimate method SHALL reference visual observations; known-viewpoint alignment SHALL reference server-authorized alignment evidence. Merely providing camera and subject points or a subject bearing SHALL NOT establish direction evidence. Validation SHALL NOT compute a missing heading or imply that evidence-reference checks establish real-world optical-axis accuracy.

#### Scenario: Nullable heading does not invalidate a camera proposal
- **GIVEN** an otherwise valid located result with a camera point and null direction
- **WHEN** validation runs
- **THEN** the camera proposal is accepted with direction still null

#### Scenario: Supported heading preserves method and uncertainty
- **GIVEN** a candidate camera viewpoint and method-appropriate evidence
- **WHEN** direction has azimuth zero or just below 360 and null or valid numeric uncertainty
- **THEN** those values and their provenance are preserved as unverified direction proposals

#### Scenario: Unsupported or out-of-range heading is rejected
- **GIVEN** direction without a camera point, without method-appropriate evidence, based only on subject bearing, with azimuth 360, or with uncertainty outside its bounds
- **WHEN** validation runs
- **THEN** the complete result fails without inventing a viewpoint, rotating the azimuth or calculating a substitute heading

### Requirement: Cover the requested language set exactly

The server's requested language set and primary language SHALL be validated independently of model output. Language tags SHALL be valid normalized BCP 47 tags, unique after normalization, and the primary language SHALL belong to the requested set. Output SHALL cover that set exactly once without requiring English. Complete entries SHALL have nonblank text and null unavailable reason; unavailable entries SHALL have null text and a nonblank reason. Candidate-based entries SHALL identify an existing candidate; scene-only entries SHALL have a null candidate ID. Normalization SHALL NOT translate text, supply missing languages or repair inconsistent statuses.

#### Scenario: Requested non-English languages and unavailable status are valid
- **GIVEN** Ukrainian and Portuguese are requested with Ukrainian primary
- **WHEN** one entry is complete and the other explicitly unavailable with consistent basis fields
- **THEN** validation accepts both statuses without requiring English or fabricating the unavailable text

#### Scenario: Tags normalize without duplicating coverage
- **GIVEN** a requested language whose output tag differs only by valid normalization such as `pt-br` versus `pt-BR`
- **WHEN** exactly one matching entry is present
- **THEN** the normalized language is retained
- **AND** two entries collapsing to the same tag are rejected

#### Scenario: Incomplete or contradictory descriptions are rejected
- **GIVEN** a missing or extra language, invalid tag, blank complete text, non-null unavailable text, missing unavailable reason or inconsistent candidate/scene basis
- **WHEN** validation runs
- **THEN** the complete result fails without losing a requested language or guessing its status

#### Scenario: Invalid server language context fails separately
- **GIVEN** an empty, over-limit or normalized-duplicate requested set, or a primary language outside it
- **WHEN** validation is invoked
- **THEN** the caller receives an invalid-context failure rather than a provider-result success

### Requirement: Return safe immutable validation outcomes without authority

Validation SHALL return either a complete validated proposal or bounded deterministic findings without a partially validated result. Model output SHALL NOT supply application user/installation/asset/job IDs, source authorization, verification status, approval or write-plan fields. All accepted prose SHALL remain untrusted content; validation SHALL neither execute it nor certify its geographic truth. Mutation of caller input after validation SHALL NOT alter the retained proposal. Validation SHALL perform no image/provider/Immich calls, persistence or writes and SHALL convey no authority to perform them.

#### Scenario: Model envelope injection is rejected
- **GIVEN** otherwise valid output includes an application identity, approval, verification or write-plan field
- **WHEN** validation runs
- **THEN** the extra field is rejected and no authorized application result is created

#### Scenario: Failures are bounded and redact private payloads
- **GIVEN** malformed output includes private text, coordinates, injected property names or URLs
- **WHEN** validation reports failure repeatedly
- **THEN** it returns stable bounded codes and safe structural paths without raw values or untrusted property names
- **AND** no partial proposal is returned or logged

#### Scenario: Validated data is independent and side-effect free
- **GIVEN** an accepted proposal contains untrusted warning or evidence text
- **WHEN** the caller later mutates its input buffers or inspects the result
- **THEN** the validated result remains unchanged and the text is neither executed nor promoted to verified evidence
- **AND** provider, image, database and Immich-write counters remain zero
