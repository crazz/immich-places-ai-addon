## MODIFIED Requirements

### Requirement: Enforce the canonical bounded result contract

The validator SHALL accept only a complete object conforming to canonical analysis-result schema version 1.0 for Visual/Context-assisted results or version 2.0 for Research results and the corresponding semantic rules below. Server-supplied completion metadata SHALL explicitly identify a complete ordinary result; refused, truncated, tool-response or missing/unknown completion state SHALL fail even if the accompanying bytes look like a complete object. It SHALL reject unsupported versions, missing or unknown properties, wrong types, invalid numeric values, duplicate JSON keys at any depth, trailing content and malformed or truncated JSON. Input bytes, nesting, collection sizes and string lengths SHALL have explicit finite limits; exceeding a limit SHALL fail the complete result rather than truncate or repair it. Schema validation SHALL perform no external resource lookup.

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

#### Scenario: LP01 Accept a complete Research result without changing old results
- **GIVEN** a version 2.0 Research answer and retained version 1.0 Visual/Context-assisted answers
- **WHEN** validation and historical readback run
- **THEN** each uses its version-specific contract without rewriting old results or requiring search metadata
- **AND** unsupported version/mode combinations remain invalid

### Requirement: Resolve identifiers only within authorized evidence

Observation and candidate identifiers SHALL be nonblank and unique within their respective collections. Observation text, candidate place names and support summaries SHALL be nonblank; optional country codes SHALL be null or valid ISO 3166-1 alpha-2 codes. Selected IDs, candidate evidence references and candidate-dependent description references SHALL resolve to the applicable collection. For version 1.0, source references SHALL resolve only within the authoritative input evidence bundle supplied separately by the server; model-written URLs or IDs SHALL NOT create source authority. For version 2.0 Research, references SHALL resolve within displayed input context or the bounded source list in that answer. Answer-provided sources SHALL be accepted as unverified references without preauthorization, tool-event provenance or independent verification. A missing or unsafe Research source link SHALL be omitted from active presentation with a bounded indication, without rejecting otherwise valid coordinates. It SHALL never cause a foreign source lookup. Provided-context observations SHALL require Context-assisted or Research mode and a nonempty authorized input-context bundle; a candidate citing such observations SHALL also cite at least one source from that input-context bundle. Visual-mode output SHALL contain no provided-context observations or source references when its authorized bundle is empty. Duplicate references SHALL be rejected rather than inflating apparent support.

#### Scenario: Local references resolve and preserve their kinds
- **GIVEN** unique candidates and observations with valid evidence references and source IDs already present in the authorized input bundle
- **WHEN** validation runs in the corresponding mode
- **THEN** those references are retained without upgrading visual observations to external evidence

#### Scenario: Dangling and duplicate identities are rejected
- **GIVEN** duplicate or blank IDs, blank required evidence text, an invalid country code, a missing selected/evidence/description candidate ID, or a repeated evidence/source reference
- **WHEN** validation runs
- **THEN** the result fails without manufacturing a replacement ID or reference

#### Scenario: Foreign or invented source claims confer no authority
- **GIVEN** version 1.0 output references an ID from another user's bundle or a new URL absent from this request's authorized bundle
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

#### Scenario: LP02 Accept Research sources from the answer
- **GIVEN** a version 2.0 answer supplies new public source URLs and references them in its explanation
- **WHEN** its result is validated
- **THEN** safe references remain attached without a backend-supplied source list or search metadata
- **AND** source identity confers no factual or write authority

#### Scenario: LP03 Keep coordinates when a source link is unusable
- **GIVEN** a valid Research estimate has an unsafe URL, missing source reference or no sources
- **WHEN** the answer is projected for review
- **THEN** valid coordinates and estimated error remain available, unsafe or unresolved links are inactive or omitted, and no foreign source is fetched

### Requirement: Require consistent and conservative radius claims

For version 1.0 results, an unknown radius basis SHALL require a null estimated radius, and a null radius SHALL use the unknown basis. A numeric radius SHALL have an allowed supporting basis and be finite and nonnegative; non-point granularity SHALL NOT claim zero radius. City and region estimates SHALL use null radius and unknown basis in this initial policy, avoiding unsupported fine numeric precision without inventing calibrated accuracy thresholds. Context-extent and source-reported bases SHALL require a referenced authorized input source that supports that basis. A visual estimate SHALL have referenced visual observations. For version 2.0 Research, a numeric estimated radius SHALL be finite and nonnegative; the validator SHALL accept a model-estimate basis with a concise explanation, without independent source evidence. City and region estimates SHALL be allowed to retain a representative point and numeric radius. There SHALL be no upper geographic-error acceptance threshold or minimum confidence. A missing radius SHALL remain null with unknown basis without discarding usable coordinates. All accepted radii SHALL remain estimates rather than measured errors or guaranteed bounds.

#### Scenario: Supported visual estimate or unknown radius is retained
- **GIVEN** a visually supported area estimate with radius 250 meters, or a camera estimate with null radius and unknown basis
- **WHEN** validation runs
- **THEN** the corresponding radius state is retained as an uncalibrated proposal

#### Scenario: Contradictory or misleading precision is rejected
- **GIVEN** version 1.0 output has unknown basis with numeric radius, null radius with a claimed basis, non-point zero radius, or a city/region estimate with numeric radius
- **WHEN** validation runs
- **THEN** the complete result is rejected without silently nulling or enlarging the radius

#### Scenario: Radius provenance requires the matching evidence kind
- **GIVEN** a context-extent or source-reported radius without a referenced authorized source supporting that kind, or a visual radius without referenced visual observations
- **WHEN** validation runs
- **THEN** validation rejects the unsupported basis claim

#### Scenario: LP04 Retain a 500-meter estimate
- **GIVEN** Research returns valid camera coordinates and a model-estimated radius of 500 meters with an explanation
- **WHEN** the result is validated
- **THEN** the coordinates and radius are retained without independent evidence or a precision threshold

#### Scenario: LP05 Retain coarse city and region estimates
- **GIVEN** Research returns a representative city or region point with a finite nonnegative radius of several or many kilometers
- **WHEN** the result is validated
- **THEN** the approximate point, granularity and radius remain available for the user's decision
- **AND** the result is not changed to unknown because of its radius or low confidence

#### Scenario: LP06 Distinguish unknown error from invalid numeric data
- **GIVEN** one Research answer has usable coordinates but cannot estimate a radius and another contains a negative or non-finite radius
- **WHEN** validation runs
- **THEN** the first retains its coordinates with unknown estimated error and the second fails numeric validation

### Requirement: Bind context disclosure to explicit classes and current authority

Context-assisted and Research analysis SHALL accept only context classes explicitly authorized for the exact owner, installation, target, provider revision and selection. Capture time, selected-album label and user hint SHALL be separately controlled. Nearby-location metadata SHALL remain an optional Context-assisted class and SHALL not be included in this Research workflow. Context SHALL not be inferred from image consent, a saved provider, catalog suggestions or a prepared image. Visual mode SHALL remain context-free.

#### Scenario: Only authorized classes are disclosed
- **GIVEN** a current Context-assisted request authorizes capture time and a user hint only
- **WHEN** its context is prepared and transmitted
- **THEN** no album label, neighbor metadata or other context class is disclosed

#### Scenario: Missing or mismatched consent prevents transmission
- **GIVEN** context consent is absent, revoked or bound to another owner, installation, target, selection or provider revision
- **WHEN** an attempt is requested
- **THEN** no context or target image is transmitted to the provider

#### Scenario: Visual does not inherit a previous context choice
- **GIVEN** a caller previously used Context-assisted mode
- **WHEN** a Visual request contains a hint or context bundle
- **THEN** the mismatched request is rejected rather than transmitting that context
- **AND** an ordinary Visual request remains usable with no context disclosure

#### Scenario: LP07 Use only displayed Research context
- **GIVEN** a Research launch shows a user hint and an optional album-label choice
- **WHEN** the user starts that configuration
- **THEN** only the selected context is transmitted and Visual remains context-free

## ADDED Requirements

### Requirement: Keep the best available estimate eligible for user review

A version 2.0 Research result SHALL retain the AI's preferred coordinate estimate when one is proposed, including a coarse representative point or a low-confidence estimate with alternatives. The result SHALL explain its approximate meaning and SHALL NOT use large estimated error, missing references or absence of search metadata as a reason to withhold it. Equally plausible candidates SHALL remain reviewable without a forced choice. Unknown SHALL remain available when the model cannot form any meaningful estimate. The backend SHALL not invent a point, copy the subject position into the camera field automatically or grant approval/write authority.

#### Scenario: LP08 Keep a weak but useful preferred estimate
- **GIVEN** a Research answer prefers one candidate with low confidence, broad uncertainty and other alternatives
- **WHEN** the answer is validated
- **THEN** its preferred coordinates and alternatives remain reviewable without a confidence cutoff

#### Scenario: LP09 Preserve ambiguous candidates or genuine unknown
- **GIVEN** an answer supplies equally plausible coordinate candidates or cannot identify any meaningful location
- **WHEN** the answer is accepted
- **THEN** the candidates remain available without an automatic choice or the genuinely unknown answer remains unknown
- **AND** the backend does not manufacture a camera point
