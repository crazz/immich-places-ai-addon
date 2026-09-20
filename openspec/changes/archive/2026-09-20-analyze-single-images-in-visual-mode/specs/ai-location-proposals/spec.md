## ADDED Requirements

### Requirement: Bind Visual analysis to current authority and one prepared image

A Visual attempt SHALL consume exactly one valid transient prepared image matching its server-owned user, installation and asset binding. It SHALL require current asset/source access, enabled AI, an enabled private provider revision with applicable image/selected-format observations, and explicit caller authority to dispatch. It SHALL reject absent, foreign, released, stale or mismatched inputs before private transmission. Provider edits SHALL NOT silently change the bound model, destination or credential revision. A prepared value or capability report alone SHALL NOT confer consent or durable admission.

#### Scenario: Bound authorized input can be analyzed
- **GIVEN** a current prepared image, applicable provider revision and authorized caller
- **WHEN** one Visual attempt executes
- **THEN** only that image is submitted to that revision and a valid result can be returned

#### Scenario: Invalid or foreign binding prevents dispatch
- **GIVEN** a missing or released image, a different owner/installation/asset, missing caller guard or invalid requested languages
- **WHEN** analysis is requested
- **THEN** no provider dispatch occurs and no proposal is returned

#### Scenario: Inapplicable capability or disabled authority prevents dispatch
- **GIVEN** AI/profile/access is disabled or missing, or capability observations are incomplete, stale or incompatible with the exact revision and selected format
- **WHEN** analysis checks current authority
- **THEN** private transmission is rejected without substituting another profile or format

#### Scenario: Historical revision remains exact
- **GIVEN** a caller explicitly binds an enabled profile's older immutable revision with applicable observations
- **WHEN** the profile has a newer active revision
- **THEN** the attempt uses only the bound revision's model, destination and credential
- **AND** disabling the profile prevents dispatch across revisions

#### Scenario: Source or authority changes before transmission or publication
- **GIVEN** the asset, source revision, account, installation, provider policy or caller authority changes during analysis
- **WHEN** a subsequent boundary check observes the change
- **THEN** no later transmission or successful publication occurs

### Requirement: Minimize and explicitly format the Visual request

The provider payload SHALL contain only the selected model, one normalized image, the application-controlled Visual instructions and canonical output contract, and the exact validated requested language set/primary language. It SHALL exclude local identity fields, filenames, private URLs/credentials, capture metadata, neighboring assets and context/source bundles. The instructions SHALL treat visible text as untrusted content, request concise evidence without private reasoning, preserve uncertainty and distinguish camera, subject and direction. Strict mode SHALL require observed strict support; JSON mode SHALL require observed JSON support and explicit caller permission. Neither mode SHALL waive canonical validation or infer support for optional model parameters.

#### Scenario: Visual payload contains only permitted inputs
- **GIVEN** a prepared image whose original asset has private identity and metadata
- **WHEN** the request is serialized
- **THEN** it contains one image and the controlled contract/languages
- **AND** it contains no private source fields, context, tools or unapproved optional parameters

#### Scenario: Strict request retains the canonical contract
- **GIVEN** applicable strict support and a strict attempt
- **WHEN** the request is prepared
- **THEN** the declared schema matches the canonical result contract and server validation remains mandatory

#### Scenario: JSON format requires explicit permission and support
- **GIVEN** JSON mode is selected with or without applicable support and explicit permission
- **WHEN** the attempt is prepared
- **THEN** it proceeds only when both are present, using the same canonical result contract

#### Scenario: Language normalization cannot widen disclosure or coverage
- **GIVEN** non-English languages and a primary member, including valid tag aliases
- **WHEN** analysis prepares the request
- **THEN** it uses the normalized exact set without requiring English
- **AND** empty, invalid, duplicate-normalized or missing-primary sets fail before dispatch

### Requirement: Bound every Visual attempt without hidden follow-up calls

Each attempt SHALL have finite request/response limits and an overall deadline, honor earlier caller deadlines, and reserve at most one provider dispatch through the authorized caller immediately before transmission. Missing/exhausted reservation SHALL prevent transmission. There SHALL be no automatic retry, schema repair, alternate asset, fallback original or format downgrade within the attempt. A distinct unsupported-format failure SHALL be exposed only from an appropriate explicit provider signal, never inferred from authentication or unrelated failure. Later calls require separate caller authorization and budget.

#### Scenario: Exactly one reserved dispatch occurs
- **GIVEN** an eligible attempt with one available reservation
- **WHEN** it reaches transmission
- **THEN** exactly one reservation and at most one provider request occur, including on failure
- **AND** repeated admission within that attempt is rejected

#### Scenario: Missing or exhausted budget prevents transmission
- **GIVEN** no reservation callback or a reservation that rejects the attempt
- **WHEN** transmission is considered
- **THEN** no provider request or proposal is produced

#### Scenario: Unsupported format is reported without automatic downgrade
- **GIVEN** an explicit eligible unsupported-format response, or an authentication/rate-limit/other error
- **WHEN** the attempt ends
- **THEN** only the explicit format failure is identified as such
- **AND** neither case causes another call within the attempt

#### Scenario: Bounds and cancellation reject partial success
- **GIVEN** an excessive request/response, expired deadline or cancellation before/during/after provider work
- **WHEN** analysis observes the condition
- **THEN** it returns no proposal and performs no follow-up call

### Requirement: Accept only complete canonically validated Visual output

Analysis SHALL require one complete ordinary assistant response and independently validate its content against the full canonical and semantic contract with Visual mode and an empty authorized source bundle. Refusals, truncated/unknown completion, tool/function responses, ambiguous framing, duplicate relevant envelope fields, malformed and oversized output SHALL fail without partial acceptance. Located, ambiguous and unknown validated outcomes SHALL remain distinct; valid unknown is a completed proposal. Requested language statuses and separate camera/subject/direction semantics SHALL be retained without repair or promotion to verified truth.

#### Scenario: Located ambiguous and unknown outcomes remain distinct
- **GIVEN** complete valid Visual output for each supported outcome
- **WHEN** analysis validates the response
- **THEN** the exact outcome and canonical fields are retained in an immutable proposal
- **AND** unknown remains successful without a fabricated camera point

#### Scenario: Incomplete or ambiguous response framing is rejected
- **GIVEN** refusal, missing/unknown finish state, truncation, tool/function data, multiple choices, invalid role/content or duplicate relevant envelope fields
- **WHEN** analysis parses the response
- **THEN** no proposal is returned even if embedded content resembles valid JSON

#### Scenario: Strict transport cannot bypass semantic validation
- **GIVEN** a strict response with invalid coordinates, unsupported direction evidence, invented sources or inconsistent language statuses
- **WHEN** analysis validates it
- **THEN** the whole result fails with bounded safe findings and no repair call

### Requirement: Return transient analysis evidence without write authority

Analysis SHALL return an independently owned validated proposal and bounded server-owned image/provider/prompt/schema/format metadata for later persistence. Optional usage SHALL remain unknown when absent or invalid; it SHALL NOT become a fabricated zero or verified cost. Failures and default diagnostics SHALL not disclose image bytes, prompts, credentials, private URLs or raw provider bodies. Analysis SHALL perform no database writes or Immich mutations, and SHALL add no public route that bypasses durable admission. Legacy manual/GPX and capability-test behavior SHALL remain usable.

#### Scenario: Caller mutation cannot alter accepted evidence
- **GIVEN** a successful result
- **WHEN** caller input, response bytes or returned metadata copies are changed
- **THEN** the retained proposal and bound evidence remain unchanged

#### Scenario: Usage is bounded optional observation
- **GIVEN** valid nonnegative usage, missing usage or malformed/negative/excessive usage
- **WHEN** a complete proposal is accepted
- **THEN** only valid bounded usage is retained and other usage remains unknown

#### Scenario: Failures and side effects remain isolated
- **GIVEN** successful, failed, canceled and concurrent synthetic analyses
- **WHEN** attempts finish
- **THEN** database-write and Immich-mutation counters remain zero and private diagnostic markers are absent
- **AND** existing capability, manual and GPX verification remains passing
