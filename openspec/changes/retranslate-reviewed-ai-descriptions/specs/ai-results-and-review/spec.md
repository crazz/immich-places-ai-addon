## ADDED Requirements

### Requirement: Explicitly authorize translation of reviewed facts

An owner SHALL be able to request selected description languages from a visible, explicitly reviewed factual text basis bound to the current draft and factual revision. Generation SHALL require a deliberate action identifying the private provider and one to eight distinct valid language tags. Stale location-dependent text SHALL not silently become the basis. The request SHALL transmit only the displayed basis and translation instructions, with no image, hidden location context, neighboring-photo data or automatic link fetching. It SHALL neither rerun geolocation nor mutate Immich.

#### Scenario: T01 Generate from corrected facts
- **GIVEN** an owned draft and a factual text basis explicitly reviewed after a camera correction
- **WHEN** the owner requests English and Ukrainian descriptions
- **THEN** only the approved text basis and language instructions are sent to the chosen provider
- **AND** the camera decision, original analysis and Immich metadata remain unchanged

#### Scenario: T02 Reject absent consent or invalid input
- **GIVEN** an empty or over-limit basis, invalid or duplicate languages, more than eight languages or no explicit generation action
- **WHEN** translation admission is attempted
- **THEN** no provider request is dispatched and the saved draft is preserved

#### Scenario: T03 Retain scene-only and offline review
- **GIVEN** owned retained scene-only facts with no camera point and an unavailable source photo
- **WHEN** the owner explicitly translates the reviewed text
- **THEN** no source image fetch or invented place is required and location remains unknown

### Requirement: Preserve bounded private translation runs

The system SHALL durably bind each translation run to its owner, installation, draft/factual revision, exact approved basis, language set and provider revision before dispatch. The same submission identity and inputs SHALL resolve to the same run across lost acknowledgements and restart; different inputs SHALL not reuse its authority. Only one active run per draft SHALL be admitted, and provider concurrency, queue capacity, deadlines and token/cost limits SHALL also bound translation work. Each language SHALL have at most one provider dispatch per run, without hidden retries or fallback.

#### Scenario: T04 Recover a lost submission acknowledgement
- **GIVEN** a translation run was accepted but its response was lost
- **WHEN** the owner repeats the same identity and inputs after restart or provider disablement
- **THEN** the existing run is returned without another provider dispatch

#### Scenario: T05 Reject substitution and concurrent admission
- **GIVEN** an existing run identity or an active run for a draft
- **WHEN** changed inputs reuse that identity or another active run is submitted for the same draft
- **THEN** no duplicate or substituted dispatch authority is created

#### Scenario: T06 Enforce execution budgets
- **GIVEN** exhausted queue capacity, provider concurrency, tokens, cost or deadline
- **WHEN** translation admission or a language dispatch is considered
- **THEN** the work waits or fails within the declared bounds and no extra call bypasses the limit

### Requirement: Retain independent language outcomes

Every requested language SHALL retain its own complete, unavailable, failed, canceled or interrupted outcome. Successful output SHALL match its requested language, contain bounded valid text and remain tied to the original approved facts; malformed, refused or truncated responses SHALL not create usable text. Failure of one language SHALL not discard another's result or overwrite existing draft text. Inspecting results SHALL not dispatch work.

#### Scenario: T07 Keep partial language success
- **GIVEN** two requested languages and one provider failure
- **WHEN** the run completes
- **THEN** the successful suggestion and failed language are visible independently while both original draft texts remain unchanged

#### Scenario: T08 Reject invalid translation output
- **GIVEN** a response with a wrong language, unexpected fields, malformed structure, invalid or oversized text, refusal or truncation
- **WHEN** it is validated
- **THEN** that language has an explicit unusable outcome without fabricated text, camera geometry or write authority

#### Scenario: T09 Reload retained outcomes
- **GIVEN** a completed run and earlier text versions
- **WHEN** the owner reloads after restart with provider execution disabled
- **THEN** the private outcomes and their original basis remain inspectable with no provider request

### Requirement: Adopt translations only through current draft review

Generation SHALL not adopt text automatically. The owner SHALL explicitly select successful suggestions to apply against the current draft revision and matching factual basis. Adoption SHALL create a new revision, leave geometry, unselected languages and immutable analysis unchanged, and invalidate earlier staged approval. Concurrent edits or changed facts SHALL prevent silent adoption; a possibly acting confirmed write SHALL continue to protect its revision. Suggestions from an obsolete factual basis SHALL be visibly stale.

#### Scenario: T10 Adopt selected suggestions
- **GIVEN** current successful suggestions and an unchanged draft/factual revision
- **WHEN** the owner applies one selected language
- **THEN** only that language changes in a new local draft revision and no provider or Immich write occurs

#### Scenario: T11 Reject stale or concurrently edited adoption
- **GIVEN** a manual edit or camera/factual change after generation started
- **WHEN** an older suggestion is applied
- **THEN** the saved draft and unsaved user edits remain intact and renewed comparison is required

#### Scenario: T12 Respect active write authority
- **GIVEN** a draft with an undispatched approval or a reserved possibly acting write
- **WHEN** the owner adopts translated text
- **THEN** a safe edit invalidates undispatched approval atomically, while a possibly acting write blocks the edit until it settles

### Requirement: Retry and cancel translation without hidden regeneration

An explicit retry SHALL create a new bounded run for the chosen unsuccessful languages from a newly confirmed current basis. Successful languages SHALL not be included implicitly. Cancellation, disablement, obsolete installation/provider authority or deletion SHALL prevent new dispatches and late unauthorized publication. Restart SHALL not automatically resend an interrupted reserved language. Completed suggestions SHALL remain separate from the current draft.

#### Scenario: T13 Retry only the chosen unsuccessful language
- **GIVEN** one complete and one failed language
- **WHEN** the owner explicitly retries the failed language with current consent
- **THEN** only that language receives new dispatch authority and the earlier successful suggestion remains intact

#### Scenario: T14 Cancel or restart an active run
- **GIVEN** pending and reserved language items
- **WHEN** the owner cancels or the process restarts during a call
- **THEN** new unauthorized sends stop, interrupted outcomes remain visible and no automatic duplicate call occurs

#### Scenario: T15 Invalidate provider authority before dispatch
- **GIVEN** queued translation work
- **WHEN** provider revision, enablement, destination policy or installation authority changes
- **THEN** obsolete work cannot dispatch using replacement authority and the user receives an explicit outcome

### Requirement: Isolate translation history and lifecycle

Translation reads, admission, cancellation, retry and adoption SHALL enforce owner and installation scope. Draft-linked runs SHALL survive ordinary cleanup and catalog resets while their draft remains retained. Account deletion SHALL remove only the owner's private records and prevent late recreation. Ordinary logs/errors SHALL omit basis text, generated descriptions and secrets.

#### Scenario: T16 Deny foreign translation operations
- **GIVEN** unauthenticated, foreign-owner or obsolete-installation references
- **WHEN** a translation operation or history read is requested
- **THEN** no private content or foreign authority is exposed or used

#### Scenario: T17 Delete during a provider call
- **GIVEN** an in-flight run and another owner's retained history
- **WHEN** the run's owner is deleted
- **THEN** late completion cannot recreate private records and the other owner's data remains intact

#### Scenario: T18 Preserve retained runs without private log output
- **GIVEN** a retained draft and translation history
- **WHEN** cleanup, catalog reset or a provider failure occurs
- **THEN** linked history remains available and ordinary diagnostics contain no private basis, output or credential

### Requirement: Make translation review accessible and explicit

The language editor SHALL expose basis review, language selection, progress, per-language failures, cancellation and explicit adoption through keyboard-operable controls and textual state. It SHALL remain usable without a map, preserve unsaved edits and fence late responses after account/result changes. Navigating or opening a language tab SHALL not initiate generation.

#### Scenario: T19 Complete keyboard translation review
- **GIVEN** a narrow viewport, unavailable map and partial language failure
- **WHEN** a keyboard user generates, inspects and adopts one successful suggestion
- **THEN** the basis, outcomes and saved revision are understandable without color or map interaction

#### Scenario: T20 Fence private late replies
- **GIVEN** an in-flight translation or unsaved basis edits
- **WHEN** the user changes account/result or navigates away
- **THEN** obsolete replies cannot populate the new view and unsaved edits receive explicit handling

#### Scenario: T21 Inspect without dispatch
- **GIVEN** a draft with retained translation outcomes
- **WHEN** its owner opens language tabs or refreshes the review
- **THEN** no generation, adoption or Immich mutation occurs automatically
