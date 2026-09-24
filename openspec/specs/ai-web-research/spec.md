# ai-web-research Specification

## Purpose

Help users locate photographs by returning the best available camera-coordinate estimate, estimated error and useful references, leaving the decision to use the coordinates to the user.

## Requirements

### Requirement: Research the best available camera-location estimate

Research SHALL use one selected prepared photograph per item and the displayed user hints, with optional selected-album label and recorded capture time. Its instructions SHALL ask the configured AI to investigate likely locations using its available research tools, compare reference photographs, maps and descriptions, and challenge misleading hints. The answer SHALL prioritize the best available camera-coordinate estimate and its estimated error. The system SHALL NOT impose a maximum acceptable error or minimum confidence threshold. Coarse site, city or region estimates SHALL remain valid proposals when their approximate meaning is explained. The camera and photographed subject SHALL remain distinct, and a representative point SHALL not be described as an exact camera position.

#### Scenario: WR01 Return an approximate camera location
- **GIVEN** a photograph and hints support a likely location with estimated error of 500 meters
- **WHEN** the provider returns that coordinate estimate
- **THEN** the result retains the coordinates and estimated error for review without requiring a more precise answer

#### Scenario: WR02 Keep estimates larger than 500 meters
- **GIVEN** the best available estimate is a city or region with a radius of several or many kilometers
- **WHEN** the answer supplies a representative point, estimated radius and explanation
- **THEN** it remains a valid approximate proposal without an error-size or confidence cutoff

#### Scenario: WR03 Challenge an incorrect hint
- **GIVEN** the user suggests one town but the photograph and reference material suggest another
- **WHEN** the AI proposes the better candidate
- **THEN** the answer explains the discrepancy and retains the estimated coordinates and uncertainty

#### Scenario: WR04 Preserve uncertainty without withholding useful candidates
- **GIVEN** one candidate is preferable but uncertain, several candidates are equally plausible, or no meaningful candidate exists
- **WHEN** Research completes
- **THEN** respectively it retains the preferred estimate with alternatives, multiple coordinate proposals without a forced selection, or an unknown outcome
- **AND** a large radius or low confidence alone does not cause an unknown outcome

### Requirement: Accept useful references from the AI answer without search metadata

The system SHALL accept bounded source URLs, available titles and concise relevance explanations supplied in the AI answer. It SHALL NOT require search logs, tool events, source-access metadata, independent verification, or sources known to the backend before analysis. The instructions SHALL ask for useful real references and prohibit inventing citations, but their absence or lack of independent confirmation SHALL NOT invalidate otherwise valid coordinate estimates. Source lists SHALL be optional, and the system SHALL not label answer-provided references as independently verified. It SHALL not request or store private chain-of-thought.

#### Scenario: WR05 Retain answer-provided links
- **GIVEN** a complete answer contains public reference URLs and short explanations without tool metadata
- **WHEN** the result is accepted and stored
- **THEN** the links and explanations remain available beside the coordinate estimate
- **AND** no search-metadata record or proxy extension is required

#### Scenario: WR06 Keep coordinates when sources are absent
- **GIVEN** a valid coordinate estimate and estimated error with no reference URLs
- **WHEN** the result is accepted
- **THEN** the estimate is retained without inventing links, rejecting it or launching a verification request

#### Scenario: WR07 Keep uncertain source claims reviewable
- **GIVEN** an answer supplies a reference whose current availability and contents were not checked by the application
- **WHEN** the result is displayed
- **THEN** its safe URL is usable as an AI-provided reference without a claim of independent verification

### Requirement: Preserve existing data boundaries without a research audit subsystem

Research SHALL use the existing authorized provider destination and one prepared photograph, with only the contextual inputs shown at Start. It SHALL not add neighboring photographs or unrelated private metadata, forward Immich credentials or private image URLs, create public image uploads, or acquire write authority. Image text, hints and reference descriptions SHALL remain untrusted data. The application SHALL retain only the bounded answer and normal job/input provenance; it SHALL not add a search-event or page-access audit subsystem. Existing finite duration, call and payload limits SHALL remain distinct from geographic-error magnitude, which has no quality cutoff.

#### Scenario: WR08 Send only displayed inputs
- **GIVEN** an entered hint with album and date context left off
- **WHEN** the user starts Research
- **THEN** the approved provider receives one prepared photograph and that hint without omitted context, neighboring photographs or Immich credentials

#### Scenario: WR09 Retain read-only behavior
- **GIVEN** an image or answer includes instructions to read private files or update Immich
- **WHEN** the application handles the answer
- **THEN** no such operation is authorized or performed and the coordinate proposal remains separate from any user-confirmed write

#### Scenario: WR10 Distinguish an incomplete response from a coarse answer
- **GIVEN** one response is truncated or exceeds its payload/deadline limit and another is complete with a large estimated error
- **WHEN** execution and validation complete
- **THEN** only the incomplete or out-of-bounds response fails technically
- **AND** the coarse complete estimate remains eligible for review

### Requirement: Compare usefulness without an accuracy acceptance gate

A repeatable comparison SHALL use the same owner-approved photographs and contextual inputs for the existing and new workflows, identifying intentional differences between modes. It SHALL record proposed coordinates, estimated error, outcome, useful links, concise explanation and latency. Measured error SHALL be calculated only where independent camera-reference coordinates and their uncertainty are available. The report SHALL separate measured error from model-estimated error and SHALL not turn the comparison into a maximum-error or minimum-confidence product gate. Synthetic tests SHALL remain distinct from live quality evidence, and private photographs, hints and locations SHALL not be committed by default.

#### Scenario: WR11 Compare estimates against known camera locations
- **GIVEN** matched inputs and independent camera-reference coordinates for some photographs
- **WHEN** a comparison report is assembled
- **THEN** it shows estimated and measured error separately, including coarse and failed outcomes rather than dropping them
- **AND** no result becomes unavailable solely because its measured or estimated error is large

#### Scenario: WR12 Report unmeasured accuracy honestly
- **GIVEN** only synthetic fixtures or photographs without independent camera coordinates are available
- **WHEN** verification is reported
- **THEN** contract checks and unmeasured geographic accuracy remain distinct without inventing a measured error or a passing quality gate
