## Why

The catalog can filter an offset-bearing capture timestamp into one calendar day while counting it under another, and currently accepts reversed date ranges. Fixing this confirmed discrepancy gives existing browsing a consistent foundation before AI selection snapshots depend on it.

## What Changes

- Align capture-date filtering and day counts across the applicable catalog scopes with the calendar date recorded in `dateTimeOriginal`, preserving source offsets and offsetless local/unknown meaning.
- Preserve undated membership in unbounded browsing and exclude it from bounded capture-date searches, without substituting upload or file creation dates.
- Reject start-after-end ranges consistently and preserve independent gallery ordering and existing scope/visibility behavior.
- Add focused date-boundary, offsetless, missing-date and range-validation regression coverage using the installed test harnesses.
- Exclude AI snapshots/jobs/context, provider management, timestamp rewriting, gallery sort changes and unrelated suggestion/GPX behavior from this change.

## Capabilities

### New Capabilities

- `catalog-capture-dates`: The shared source-local calendar-date contract for current catalog filtering and day counts, including range validation and undated membership. This is the catalog foundation for the roadmap's later AI selection capability.

### Modified Capabilities

None. No maintained product specification exists yet; this change establishes the bounded catalog contract.

## Impact

Backend catalog queries and date-range handling, applicable catalog count consumers and regression tests, plus requirement traceability and integration documentation. No new service, database engine or external call is required. This implements the capture-date portion of FR-01; future context and frozen selections must reuse its policy in CH05–CH06/CH10.
