## MODIFIED Requirements

### Requirement: Explain evidence and lineage without fabricated verification

Review SHALL resolve observation references within the canonical result and version 1.0 source references within trusted stored input provenance, retaining their distinct meanings. Version 2.0 Research references SHALL also resolve within the source list returned in its stored AI answer, without tool metadata or prior backend source authorization. It SHALL show concise support, warnings, limitations and available source lineage. Model-only evidence, user hints and unknown lineage SHALL not become independent verification. Untrusted text SHALL remain inert, and review SHALL not fetch fabricated citations or expose private reasoning.

#### Scenario: Inspect source-free Visual evidence
- **GIVEN** a Visual proposal with observations and no external context
- **WHEN** evidence is expanded
- **THEN** it is labeled model/visual evidence without external-verification claims or invented source links

#### Scenario: Inspect contextual lineage
- **GIVEN** Context-assisted provenance with user hints, allowed sources and unknown lineage
- **WHEN** source details are shown
- **THEN** each kind and uncertainty remain distinct and unknown lineage is not presented as corroboration

#### Scenario: Handle unsafe text or unresolved references
- **GIVEN** markup-like text, instruction-like content or corrupt unresolved evidence references
- **WHEN** the detail is rendered
- **THEN** text stays inert, invalid references fail safely and no automatic research, URL fetch or mutation occurs

#### Scenario: RV09 Review answer-provided Research references
- **GIVEN** a retained Research answer contains source URLs and explanations without a search log
- **WHEN** its owner opens the evidence
- **THEN** safe answer-provided links remain available without an independent-verification claim or a metadata gate


## ADDED Requirements

### Requirement: Present coordinates and estimated error for the user's decision

Research detail SHALL prominently show the preferred camera coordinates and estimated error, with an understandable label such as estimated ±500 m or the equivalent in kilometers. There SHALL be no maximum-error filter, minimum-confidence requirement or precision-based disablement of result inspection. Coarse representative points SHALL retain their approximate site/city/region meaning. Unknown error SHALL display as unknown without hiding usable coordinates. Alternatives, concise explanation and the distinction between camera and subject SHALL remain visible. The user decides whether the estimate is useful; this read-only review SHALL not apply coordinates or create approval/write authority.

#### Scenario: RV01 Review a 500-meter estimate
- **GIVEN** a Research result with proposed coordinates and an estimated radius of 500 meters
- **WHEN** the owner opens its detail
- **THEN** both appear prominently as a proposal with estimated ±500 m uncertainty
- **AND** no quality threshold hides the result or prevents inspection

#### Scenario: RV02 Review a city or region estimate
- **GIVEN** a representative point with an estimated radius of several or many kilometers and low confidence
- **WHEN** the result is opened
- **THEN** its approximate meaning, coordinates, radius and alternatives remain reviewable
- **AND** it is not presented as an exact measured camera position

#### Scenario: RV03 Inspect unknown error and ambiguous alternatives
- **GIVEN** a useful point with unknown error or several equally plausible coordinate candidates
- **WHEN** the owner inspects the result
- **THEN** the point or alternatives remain visible with the unknown error or ambiguity explained

#### Scenario: RV04 Use results without maps or pointer input
- **GIVEN** a narrow viewport, keyboard navigation and unavailable map tiles or current photograph
- **WHEN** the owner opens Research detail
- **THEN** coordinates, estimated error, explanations and available source links remain accessible
- **AND** no manual pending coordinate or Immich state is changed

### Requirement: Show safe answer-provided links without a search audit

Research results SHALL display safe HTTP(S) source links and available titles or relevance explanations returned in the AI answer, without requiring a search log, tool provenance or page-access status. Links SHALL be optional and SHALL not be presented as independently verified. All text SHALL remain inert; unsafe schemes, credential-bearing URLs, clearly local/private destinations and unresolved references SHALL not become active links. An unusable reference SHALL not hide or invalidate otherwise valid coordinates. The application SHALL not automatically fetch pages, thumbnails or source availability during persistence or review. Deliberate external navigation SHALL avoid passing the application referrer or opener access.

#### Scenario: RV05 Follow a source from the answer
- **GIVEN** a Research answer includes a public reference URL and short explanation without search telemetry
- **WHEN** the owner opens the result and activates its labeled link
- **THEN** that reference is available through accessible deliberate navigation without a verification badge or search-metadata prerequisite

#### Scenario: RV06 Preserve coordinates with absent or unsafe links
- **GIVEN** a valid coordinate estimate has no sources, a malformed URL, an unsafe scheme or markup-like source text
- **WHEN** its detail is rendered
- **THEN** the coordinates remain visible, text remains inert and unusable destinations are inactive or omitted
- **AND** no automatic source request, script execution or mutation occurs

#### Scenario: RV07 Retain history when a reference disappears
- **GIVEN** an owned retained result whose referenced page later changes or becomes unavailable
- **WHEN** its detail is reloaded with execution disabled
- **THEN** the stored answer remains readable without a source-site request or a claim of current page availability

#### Scenario: RV08 Preserve legacy display and account isolation
- **GIVEN** a version 1.0 Visual/Context-assisted result or an account switch during a Research read
- **WHEN** review updates
- **THEN** legacy results keep their established presentation and the old account's answer is cleared
- **AND** no fabricated external sources are added to legacy history
