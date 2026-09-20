## ADDED Requirements

### Requirement: Inspect separate camera subject and alternative locations

Proposal review SHALL distinguish camera and subject locations through text and marker identity and show available alternatives without approving them. A selected canonical candidate MAY receive initial inspection focus; ambiguous results SHALL not gain a chosen candidate merely by opening them. Unknown or subject-only results SHALL not invent camera geometry. Numeric zero SHALL remain a valid coordinate and null SHALL remain absent.

#### Scenario: Show different camera and subject points
- **GIVEN** a candidate with distinct camera and subject coordinates
- **WHEN** the user opens its proposal
- **THEN** both are explicitly labeled and the subject/landmark is never substituted for the photographer's position

#### Scenario: Inspect an ambiguous alternative
- **GIVEN** an ambiguous proposal with multiple candidates and no selection
- **WHEN** the user focuses one alternative
- **THEN** only local inspection focus changes, with no approved candidate or persisted result change

#### Scenario: Preserve missing and zero coordinates
- **GIVEN** unknown, subject-only or valid zero-coordinate geometry
- **WHEN** the panel and map render
- **THEN** missing camera points stay absent while valid zero values remain visible and are never treated as missing GPS

### Requirement: Present uncertainty without overstating precision

Review SHALL show supplied granularity, uncertainty notes and estimated camera radius with its basis. Radius SHALL be labeled uncalibrated and SHALL not imply statistical confidence or external verification. Null radius SHALL not create a circle or invented accuracy; zero radius SHALL not be treated as absent or proven exactness. Numeric coordinates SHALL remain authoritative when the map projection cannot display them faithfully.

#### Scenario: Show a supplied radius honestly
- **GIVEN** a validated candidate with a supported radius and basis
- **WHEN** uncertainty is displayed
- **THEN** the estimate, units and basis are clear without a calibrated-confidence claim

#### Scenario: Show unknown or zero radius
- **GIVEN** a candidate whose radius is null or zero
- **WHEN** the panel renders precision
- **THEN** null is explicitly unknown without an invented circle and zero is shown as supplied without asserting exact accuracy

#### Scenario: Inspect projection edge cases
- **GIVEN** valid geometry near a pole or across the antimeridian
- **WHEN** the map frames the result
- **THEN** the original numeric coordinates remain intact and usable even if map framing needs a degraded presentation

### Requirement: Keep camera heading distinct from subject bearing

Direction SHALL mean the stored horizontal optical-axis azimuth clockwise from true north, with method and nullable uncertainty. Null direction SHALL remain unknown. Subject bearing SHALL not create a missing heading; review SHALL not invent pitch, roll or field of view. An arrow SHALL require an actual camera point and supplied direction.

#### Scenario: Inspect supplied true-north direction
- **GIVEN** a camera point and valid direction, including zero degrees
- **WHEN** the user reviews heading
- **THEN** the stored degrees, true-north reference, method and uncertainty are shown without converting zero to missing

#### Scenario: Keep unknown direction unknown
- **GIVEN** a camera and subject with no camera direction
- **WHEN** they are displayed together
- **THEN** direction remains unknown and any displayed subject bearing is separately labeled rather than promoted to optical-axis heading

### Requirement: Explain evidence and lineage without fabricated verification

Review SHALL resolve observation references within the canonical result and source references within trusted stored provenance, retaining their distinct meanings. It SHALL show concise support, warnings, limitations and available source lineage. Model-only evidence, user hints and unknown lineage SHALL not become independent verification. Untrusted text SHALL remain inert, and review SHALL not fetch fabricated citations or expose private reasoning.

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

### Requirement: Display exact requested-language descriptions

Review SHALL provide one accessible tab per requested normalized language with primary-language indication, exact canonical `complete` or `unavailable` status and reason. It SHALL preserve scene-only or candidate-specific factual basis and never silently copy a translation, reassign text to an inspected alternative or invent absent text. Changing language tabs SHALL not trigger provider calls or writes.

#### Scenario: Switch between complete and unavailable languages
- **GIVEN** requested languages with complete and unavailable descriptions
- **WHEN** the user switches tabs
- **THEN** exact text or unavailable reason appears with primary-language indication and no fabricated translation or provider call

#### Scenario: Inspect another candidate without rewriting text
- **GIVEN** a description bound to one candidate or only the visible scene
- **WHEN** the user focuses another candidate
- **THEN** the description retains its original labeled basis and is not represented as text for the alternative

#### Scenario: Preserve a non-English language set
- **GIVEN** a valid requested set that excludes English
- **WHEN** review opens or the application locale changes
- **THEN** the same requested language set remains available without adding English or changing stored content

### Requirement: Preserve accessible read-only review under degraded maps

All meaningful spatial/evidence/language information SHALL remain available through keyboard-operable controls and numeric/text presentation when tiles or source images fail. Read-only inspection SHALL never install manual location-edit actions, modify pending coordinates, create a draft or write Immich. Leaving review SHALL preserve the existing manual/GPX state and behavior. Owner/result changes SHALL remove stale private content and map layers.

#### Scenario: Review without map tiles or an image
- **GIVEN** tile failures or unavailable source-image access
- **WHEN** a keyboard user inspects candidates, direction, evidence and language tabs
- **THEN** the complete numeric/text information remains usable with clear labels and predictable focus

#### Scenario: Attempt manual-style map interactions during review
- **GIVEN** an AI proposal displayed while manual pending changes already exist
- **WHEN** the user clicks, drags, drops or opens map context controls
- **THEN** AI inspection creates no pending coordinate, approval or Immich mutation and existing manual state remains unchanged

#### Scenario: Leave or switch private review
- **GIVEN** active proposal layers and detail requests
- **WHEN** the result/account changes or the user returns to manual/GPX mode
- **THEN** stale content and layers are removed, focus is restored and the original manual/GPX workflow still operates
