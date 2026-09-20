# AI results and review

## Purpose

Make private immutable AI analysis history discoverable and understandable independently of the current catalog, without granting draft or Immich-write authority.

## Requirements

### Requirement: Browse private durable terminal history

Authenticated owners SHALL be able to list and inspect retained current-installation succeeded, failed and canceled analysis items, including located, ambiguous and unknown proposals. History SHALL survive restart, source removal and GPS changes without depending on Missing GPS membership. Foreign or old-installation references SHALL not expose private content. AI execution disablement SHALL not prevent authorized history reads or cause provider work.

#### Scenario: Retain results after GPS or catalog changes
- **GIVEN** retained results for a photo that gains GPS or disappears from synchronized assets
- **WHEN** its owner opens AI Results after reload
- **THEN** the retained history remains available independently of current Missing GPS or catalog membership

#### Scenario: Read with execution disabled
- **GIVEN** retained private history and disabled AI execution
- **WHEN** the owner lists or opens results
- **THEN** local history remains readable without provider calls or activation of workers

#### Scenario: Reject foreign history access
- **GIVEN** an unauthenticated, foreign-owner or old-installation reference
- **WHEN** list, detail or item-result access is attempted
- **THEN** no private record or existence information is disclosed

### Requirement: Filter and paginate retained history honestly

History SHALL support bounded stable pagination and filters for terminal execution state, proposal outcome, asset/run, retained capture date and selected album at launch. The album filter SHALL identify its historical selected-scope meaning. Capture dates SHALL preserve source-local recorded dates without upload-time fallback; absent historical facts SHALL remain unknown. Filters and cursors SHALL be validated and scoped to owner, installation and query.

#### Scenario: Page while new work completes
- **GIVEN** multiple pages of history and concurrently completing jobs
- **WHEN** the owner follows the current query's continuation cursor
- **THEN** deterministic pages do not duplicate entries or mix another query/owner, and new completions appear after refresh

#### Scenario: Filter retained album and date facts
- **GIVEN** dated, undated and older provenance-incomplete runs from selected-album and timeline launches
- **WHEN** the owner uses capture-date or selected-album-at-launch filters
- **THEN** only matching retained facts qualify, undated records remain in unbounded/undated views and current catalog membership is not substituted for missing history

#### Scenario: Reject invalid query boundaries
- **GIVEN** a reversed date range, excessive limit, unsupported filter or foreign/mismatched cursor
- **WHEN** history is requested
- **THEN** the request fails safely without unbounded loading or cross-scope data

### Requirement: Keep immutable runs and independent states

Each run SHALL remain separately inspectable with its original model/mode and provenance. Execution state, proposal outcome, review state and write state SHALL be independent. Located, ambiguous and unknown SHALL be successful proposal outcomes; failed/canceled entries SHALL have safe details without fabricated proposals. In the absence of durable draft/write records, review SHALL be unreviewed and write SHALL be not requested.

#### Scenario: Display successful unknown separately from failure
- **GIVEN** a successful unknown proposal and a failed provider attempt
- **WHEN** both appear in history
- **THEN** their execution/outcome labels differ and neither implies approval or a completed write

#### Scenario: Preserve older runs after reanalysis
- **GIVEN** multiple analyses of the same asset
- **WHEN** a newer run completes or provider settings change
- **THEN** every retained run keeps its immutable proposal and original provenance without replacing older review/write state

#### Scenario: Reject corrupt or unsupported detail safely
- **GIVEN** a retained record that cannot be safely decoded under its stored contract version
- **WHEN** the owner opens it
- **THEN** a safe unavailable state appears without rewriting the record, exposing raw data or breaking other history entries

### Requirement: Authorize current images independently of stored results

Retained result ownership SHALL not authorize new upstream image access. Thumbnails and previews SHALL require current owner, installation and asset visibility/access; hidden, missing or inaccessible assets SHALL show an unavailable placeholder while local result history remains readable. Private image references or cached content SHALL never cross account boundaries.

#### Scenario: Display an authorized current thumbnail
- **GIVEN** a retained result whose source is currently accessible
- **WHEN** its owner opens the list or detail
- **THEN** the image is fetched through the authorized read boundary without provider credentials or stored external image URLs

#### Scenario: Lose source access
- **GIVEN** a retained result whose source became hidden, removed or inaccessible
- **WHEN** a thumbnail or preview is requested
- **THEN** no unauthorized image is returned and the UI shows an unavailable placeholder alongside the local history

#### Scenario: Change accounts during image loading
- **GIVEN** an in-flight private image/detail request
- **WHEN** the active account changes
- **THEN** prior account data is cleared and late responses cannot populate the new account's view

### Requirement: Navigate usable history without side effects

AI Results SHALL provide accessible loading, empty, filtered-empty, error, stale and detail states, preserve navigation/filter position and work without successful catalog sync. Model/provider text SHALL remain untrusted display content. Browsing, refresh, filter and detail actions SHALL make no provider call, create no draft or manual pending coordinate, and perform no Immich mutation.

#### Scenario: Return from detail to the same list
- **GIVEN** a filtered paginated history list
- **WHEN** the user opens detail and returns
- **THEN** filters, list position and keyboard focus are restored predictably without resubmitting analysis

#### Scenario: Handle unavailable catalog or history requests
- **GIVEN** failed catalog sync, an empty filter or a failed history request
- **WHEN** the user navigates AI Results
- **THEN** catalog failure does not block local history, and empty/error/stale states provide an understandable bounded retry path

#### Scenario: Render untrusted content without write authority
- **GIVEN** a proposal containing markup-like or instruction-like text
- **WHEN** it is displayed and navigated
- **THEN** it remains inert escaped text and causes no provider call, draft acceptance, pending-coordinate change or Immich write
