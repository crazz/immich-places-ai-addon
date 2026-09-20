## ADDED Requirements

### Requirement: Bind context disclosure to explicit classes and current authority

Context-assisted analysis SHALL accept only context classes explicitly authorized for the exact owner, installation, target, provider revision and selection. Capture time, selected-album label, user hint and nearby-location metadata SHALL be separately controlled. Context SHALL not be inferred from image consent, a saved provider, catalog suggestions or a prepared image. Visual mode SHALL remain context-free.

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

### Requirement: Preserve honest capture-time and bounded context semantics

Context SHALL preserve recorded capture-date and offset information, distinguishing missing, invalid and local/unknown-zone timestamps. It SHALL not substitute upload time or server timezone. Temporal neighbor comparisons SHALL require comparable offset-bearing timestamps and an explicit positive window no greater than 24 hours. Context SHALL contain at most six metadata-only neighbors, one selected-album label of at most 256 UTF-8 bytes, a user hint of at most 2,000 UTF-8 bytes and at most 16 KiB of serialized context. Discovery SHALL be finite and need not be exhaustive.

#### Scenario: Offset and missing-time meanings survive
- **GIVEN** targets have an explicit offset, an offsetless capture time or no capture time
- **WHEN** authorized context is produced
- **THEN** each retains its source-local date and explicit time status without upload-date or server-timezone substitution
- **AND** temporal neighbors are omitted when elapsed-time comparison is not supported

#### Scenario: Context limits are deterministic
- **GIVEN** more than six eligible neighbors and valid bounded caller text
- **WHEN** a context bundle is prepared
- **THEN** at most six deterministic metadata-only neighbors are retained within the total limit
- **AND** omission status does not claim complete catalog enumeration

#### Scenario: Invalid bounds do not silently alter user input
- **GIVEN** malformed or oversized hint/label input, an invalid window or an excessive context projection
- **WHEN** preparation is requested
- **THEN** it returns a safe failure without truncating caller text or disclosing a partial bundle

### Requirement: Admit only eligible context sources with honest lineage

Each retained source SHALL be currently accessible to the owner in the same installation and satisfy visibility/library/type and relevance policy. Selected-album context SHALL come only from the explicitly authorized album containing the target. Hidden, inaccessible, duplicate, out-of-window and known AI-origin neighbors SHALL be excluded. Unknown lineage SHALL remain unknown. Unrelated frequent/home locations, unrelated albums, private paths, people/device identifiers and neighbor images SHALL not be disclosed.

#### Scenario: Eligible sources retain their provenance
- **GIVEN** authorized current metadata from an eligible neighbor with unknown lineage
- **WHEN** it is included
- **THEN** its context kind and unknown lineage are retained without a verified or independent-confirmation label

#### Scenario: Ineligible or unrelated sources are excluded
- **GIVEN** candidates include foreign, hidden, inaccessible, known AI-origin, duplicate or out-of-window assets and unrelated frequent locations
- **WHEN** sources are selected
- **THEN** none of those inputs enters the bundle and no neighboring image is fetched

#### Scenario: Album context cannot broaden the authorized scope
- **GIVEN** a supplied album is foreign or unselected
- **WHEN** context is prepared
- **THEN** the invalid caller binding is rejected without exposing any album label

#### Scenario: Former album membership omits optional context
- **GIVEN** the authorized selected album no longer contains the target before the bundle is frozen
- **WHEN** context is prepared
- **THEN** its label is omitted with a safe reason and no other album is substituted

### Requirement: Recheck frozen context before transmission and publication

Prepared context SHALL bind its exact content and source authority independently of caller buffers. Optional unavailable sources may be omitted before the bundle is frozen with safe reasons. An empty bundle SHALL remain an explicit Context-assisted result of preparation. Observed authority, consent, source or installation changes after freezing SHALL invalidate the attempt before further transmission or successful publication; they SHALL not silently replace its evidence.

#### Scenario: Empty authorized context remains explicit
- **GIVEN** consent is valid but no optional source remains eligible
- **WHEN** preparation finishes
- **THEN** it reports an empty Context-assisted bundle with safe omission statuses
- **AND** it does not manufacture evidence or silently change the requested mode

#### Scenario: Frozen-source or authority change discards the attempt
- **GIVEN** a frozen bundle whose source, access, consent or installation changes
- **WHEN** a subsequent dispatch or publication check observes the change
- **THEN** no subsequent transmission or successful result publication occurs

#### Scenario: Caller mutation cannot widen disclosure
- **GIVEN** a prepared bundle
- **WHEN** caller input or a returned copy is changed
- **THEN** the retained content, digest and authorized source set remain unchanged

### Requirement: Validate Context-assisted output against the exact supplied evidence

A Context-assisted attempt SHALL send one prepared target image with controlled instructions and only the approved bounded context. It SHALL retain existing format authority, deadline, size, cancellation and single-reservation bounds. Every returned source reference SHALL resolve to the exact server-authorized supplied bundle; model text SHALL not create verification, radius or heading authority. Canonical camera/subject, outcome, language and uncertainty rules SHALL remain mandatory.

#### Scenario: Valid contextual output retains separate evidence kinds
- **GIVEN** a current authorized bundle and applicable provider revision
- **WHEN** one reserved attempt returns complete valid Context-assisted output
- **THEN** its exact outcome and evidence references are retained without promoting hints to verified facts

#### Scenario: Invented evidence or unsupported geometry fails
- **GIVEN** output invents a source, claims an unsupported radius/heading basis or has invalid canonical geometry/languages
- **WHEN** output is validated
- **THEN** no proposal is returned and no repair or fallback call is made

#### Scenario: Resource failure or cancellation cannot create partial success
- **GIVEN** an attempt exceeds its existing request/response/deadline limits, loses a reservation or is canceled
- **WHEN** the condition is observed
- **THEN** no partial proposal is published and no hidden follow-up request occurs

### Requirement: Return bounded contextual provenance without persistent side effects

The transient handoff SHALL retain an immutable proposal and independently owned mode, context digest, consent-policy version, source records and omission metadata for a later durable consumer. Default diagnostics SHALL redact private context and raw provider input/output. Preparation and analysis SHALL introduce no database writes, external research, Immich mutations or public bypass of durable admission.

#### Scenario: Provenance survives the transient handoff
- **GIVEN** a successful contextual result
- **WHEN** its consumer reads the metadata
- **THEN** the exact supplied source set and context binding are available independently of caller mutations

#### Scenario: Failures and execution remain private and read-only
- **GIVEN** successful, failed, concurrent or canceled synthetic contextual attempts
- **WHEN** they finish
- **THEN** database-write, research and Immich-mutation counters remain zero
- **AND** private hints, source identifiers and credentials are absent from ordinary diagnostics
