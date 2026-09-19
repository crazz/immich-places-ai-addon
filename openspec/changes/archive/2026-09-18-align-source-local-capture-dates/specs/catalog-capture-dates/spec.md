## Purpose

Keep existing catalog capture-date filtering and calendar counts consistent with the date recorded by the source, independently of server timezone, while retaining existing access and browsing boundaries. This capability implements the date-query portion of PRD FR-01.

## ADDED Requirements

### Requirement: Preserve source-local calendar dates

The catalog SHALL interpret a capture date as the valid leading `YYYY-MM-DD` calendar date recorded in `dateTimeOriginal`. Offset-bearing timestamps SHALL keep their recorded day rather than move to a UTC/server day. Offsetless timestamps SHALL remain local/unknown, without inferred timezone or upload-date fallback. Stored and returned capture timestamp text SHALL remain unchanged.

#### Scenario: Offset timestamps cross UTC midnight
- **GIVEN** one eligible asset captured at `2024-01-01T00:30:00+14:00` and another at `2024-01-01T23:30:00-12:00`
- **WHEN** the user filters and requests calendar counts for January 1
- **THEN** both assets belong to January 1 and its count is two
- **AND** neither is assigned to December 31 or January 2

#### Scenario: Offsetless and daylight-saving capture times
- **GIVEN** eligible offsetless timestamps and timestamps on a daylight-saving transition, including repeated local clock times with different offsets
- **WHEN** their recorded calendar day is selected
- **THEN** filtering and counting use that recorded day for every asset
- **AND** no timezone is added to the offsetless timestamp

### Requirement: Apply inclusive ranges and undated membership consistently

Date bounds SHALL include the whole recorded start and end calendar days. A single bound SHALL constrain only its corresponding side. Null, empty or invalid calendar prefixes SHALL have no dated bucket, SHALL remain eligible under an otherwise matching unbounded query, and SHALL be excluded whenever either capture-date bound is set. Calendar counts SHALL report only matching dated assets.

#### Scenario: Inclusive end and one-sided bounds
- **GIVEN** eligible assets at the beginning and end of a leap day, before it and after it
- **WHEN** an exact-day, start-only or end-only range is applied
- **THEN** the exact day includes both boundary assets, and one-sided ranges retain the matching side
- **AND** timestamps with a date-only, space-separated or `T`-separated representation follow the same calendar rule

#### Scenario: Missing and invalid calendar dates
- **GIVEN** otherwise eligible assets with null, empty and invalid calendar dates and independent valid file-creation dates
- **WHEN** browsing is unbounded
- **THEN** those assets remain present without a substituted capture date
- **WHEN** either capture-date bound is supplied
- **THEN** those assets are excluded and produce no dated calendar count

### Requirement: Preserve catalog scope and ordering

Every existing date-aware catalog owner SHALL use the same date interpretation, including asset lists/totals, supported calendar counts, album counts, folder trees/assets and map marker lists/totals. Each owner SHALL retain its existing supported scope, tenant isolation, visibility/GPS/stack policies and ordering. This change SHALL NOT add an unsupported scope to an endpoint or change gallery ordering from its independent file-creation/ID ordering.

#### Scenario: Scoped assets and counts agree
- **GIVEN** eligible and excluded assets across users, albums, tags, GPS states, visibility states and stacks
- **WHEN** matching supported filters and date bounds are applied to asset lists and calendar counts
- **THEN** the date bucket count equals the matching listed assets for that day
- **AND** another user's or otherwise excluded assets do not contribute

#### Scenario: Folder, album and marker consumers retain date boundaries
- **GIVEN** folder and album memberships containing timestamps around source-local date boundaries
- **WHEN** date bounds are applied to folder lists/totals, album counts and map marker lists/totals
- **THEN** each owner uses the recorded day consistently while preserving its existing membership/GPS/visibility rules

#### Scenario: Gallery sort remains independent
- **GIVEN** assets whose file-creation order differs from their capture-time order
- **WHEN** capture-date filtering is applied
- **THEN** eligible assets retain the existing descending file-creation and asset-ID order

### Requirement: Reject invalid catalog date ranges

Authenticated catalog endpoints accepting date ranges SHALL reject malformed dates and start-after-end ranges with HTTP 400 using the existing error response shape, before querying data. Same-day ranges SHALL be valid. Optional one-sided bounds SHALL remain valid where already supported; calendar counts SHALL continue to require both bounds. Unauthenticated requests SHALL remain unauthorized.

#### Scenario: Invalid ranges across catalog endpoints
- **GIVEN** an authenticated request to a date-taking asset, count, album, folder or marker endpoint
- **WHEN** a bound is malformed or the start date is after the end date
- **THEN** the response is HTTP 400 with an explanatory error and no catalog query result

#### Scenario: Valid and missing required bounds
- **GIVEN** an authenticated catalog request
- **WHEN** valid equal or optional one-sided bounds are supplied
- **THEN** the existing supported query succeeds
- **WHEN** a calendar-count request omits either required bound
- **THEN** it is rejected with HTTP 400

#### Scenario: Authentication is preserved
- **WHEN** an unauthenticated caller requests a date-filtered catalog result
- **THEN** the response remains HTTP 401 and contains no catalog data
