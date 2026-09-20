## ADDED Requirements

### Requirement: Reuse the configured provider without a search-metadata readiness gate

Research SHALL use the selected owned provider revision, model, credentials and approved destination with the existing image/output compatibility checks. Missing search logs, tool events, page-access evidence or research-specific capability observations SHALL NOT block launch or invalidate an answer. The application SHALL not require a new search service, key, proxy extension or metadata test. Image/output compatibility SHALL not be described as proof of web-search availability. Saving or loading settings SHALL remain free of provider calls, and an explicit unsupported request response SHALL be reported without switching providers or models automatically.

#### Scenario: PC01 Launch through the existing provider
- **GIVEN** an enabled owned profile with current image/output compatibility and no search-metadata capability record
- **WHEN** a valid Research selection is started
- **THEN** the existing configured provider receives the request without a new research test or service requirement

#### Scenario: PC02 Preserve useful answers without telemetry
- **GIVEN** the provider returns coordinates, estimated error and optional source URLs with no tool telemetry
- **WHEN** the application accepts its answer
- **THEN** the absent telemetry does not cause a readiness or result-validation failure

#### Scenario: PC03 Report a concrete transport limitation
- **GIVEN** an endpoint explicitly rejects the request or returns an authentication, network or timeout error
- **WHEN** the attempt ends
- **THEN** the actual safe failure is shown without inventing search capability evidence or silently changing the model, endpoint or credential

### Requirement: Supply finite research execution defaults without an accuracy limit

Research SHALL use finite application defaults for request duration, provider calls, payload and retained-answer sizes. Explicit operator restrictions SHALL remain authoritative, and current owner, revision and destination controls SHALL remain enforced. Those execution bounds SHALL NOT include a maximum geographic error or minimum confidence. Missing usage or prices SHALL remain unknown. Users SHALL not need a manually authored policy or an additional consent checkbox to use ordinary defaults.

#### Scenario: PC04 Use ordinary bounded defaults
- **GIVEN** an otherwise ready profile and no explicit operator restriction
- **WHEN** its owner opens Research launch
- **THEN** usable finite execution defaults are supplied without an accuracy cutoff or manual policy document

#### Scenario: PC05 Enforce actual authority and execution limits
- **GIVEN** a foreign or disabled profile, an unapproved destination or an incompatible explicit operator restriction
- **WHEN** Research admission or dispatch is requested
- **THEN** it is denied before transmission with a safe actionable reason
- **AND** that denial is unrelated to the magnitude of a proposed geographic error
