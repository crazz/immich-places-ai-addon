## Context

See the proposal for motivation and FR-01 in the planning PRD for the product policy. The current catalog uses text predicates for capture-date ranges but groups calendar counts with SQLite `DATE()` on the complete timestamp, allowing UTC normalization to move the bucket. Asset/folder queries share a filter builder; album counts and map markers repeat the date predicates. Most endpoints share range parsing, while map markers validate dates separately.

The existing day-count endpoint requires both bounds and accepts album, tag, GPS and visibility filters; it does not accept a folder scope. This change preserves those API boundaries. Current gallery ordering uses file creation time and ID, independently of capture dates. Undated gallery membership is preserved; a new undated-group presentation and frozen-selection UI are outside this bounded query correction.

## Goals / Non-Goals

**Goals:**

- One documented source-local calendar policy across existing catalog query owners and one validation policy across date-taking HTTP endpoints.
- Preserve authentication, tenant joins, hidden-library/per-asset policies, GPS/stack filtering, pagination and ordering.
- Keep the existing schema, driver, deployment and API response shapes, with real database and HTTP regression evidence.

**Non-Goals:**

- Timestamp migration, timezone inference, upload-date fallback, new query scopes, changes to legacy suggestion/GPX semantics or any AI runtime/package.
- Claiming the rest of FR-01 is implemented: undated-group presentation, AI eligibility and selection snapshots remain later work.

## Decisions

### Keep date policy at the existing catalog persistence boundary

A focused backend file owns the source-local date expression and range predicate construction. The date is the canonical leading `YYYY-MM-DD` calendar date of `dateTimeOriginal`; the offset and time remain untouched in stored/returned metadata. Calendar validation operates only on that prefix, never on the offset-bearing timestamp. Missing, empty or invalid calendar prefixes have no dated bucket and cannot enter a bounded search. No server timezone, UTC conversion or file-creation fallback is involved.

Range construction retains raw text lower/upper bounds for canonical source timestamps so existing timestamp indexes remain usable, and excludes values without a valid calendar prefix whenever either bound exists. The established inclusive end-day behavior is retained for date-only, `T`-separated and space-separated timestamp text. Calendar aggregation uses the same source-day expression. The shared helper supplies predicates to asset filters, album count queries and map-marker filters; folder queries already consume the asset filter. SQL identifiers come only from fixed internal aliases, while date values remain bound parameters.

This is query behavior, not AI domain logic, so it remains in the existing backend persistence/composition package. No empty AI package, new database dependency, migration or Docker packaging change is needed. Extracting duplicated date predicates into the focused policy file lowers the inherited database-file size ceiling without growing existing exceptions or moving unrelated query behavior.

### Validate date ranges consistently before queries

All date-taking catalog handlers use one parser for optional canonical date bounds and start/end ordering. The map-marker path adopts that parser while retaining its other validation. Same-day and one-sided bounds remain valid; the day-count endpoint keeps its existing requirement for both dates. Invalid/reversed ranges return the existing bad-request error shape without executing a catalog query. Authentication precedes date handling.

### Verify query results and endpoint contracts on real SQLite

Focused new Go test files use normal migrated, file-backed test databases. Fixtures cover offsets crossing UTC midnight, offsetless capture times, daylight-saving overlap/boundary dates, leap days, inclusive ends, missing/empty dates and invalid calendar prefixes. Compare concrete asset IDs and day buckets, including preserved timestamps and independent sort order. Exercise album/tag/GPS/visibility variants, folder listings/counts and marker counts with separate users and hidden-library/stack exclusions where those owners already apply them.

HTTP tests prove consistent malformed/reversed range rejection across assets, day counts, missing-location counts, albums, folders, folder assets and markers, plus the authenticated day-count response. Existing backend race tests and the full shared gates retain manual/GPX/browser coverage. New behavior uses observed RED/GREEN; unchanged behavior can be characterized directly. There is no frontend implementation change requiring a separate rendering harness change.

### Deployment and recovery remain schema compatible

Existing data is read in place. Deployment changes only query and validation code; rollback restores the previous query behavior without a data conversion. Persisted timestamps and public asset DTOs stay intact. Document the correction separately from the pinned historical reconciliation and retain later FR-01 obligations in the roadmap.

## Risks / Trade-offs

- Shared filters have critical graph impact across multiple endpoints → cover callers with real scoped fixtures and HTTP validation, then run the full suite and GitNexus change review.
- Offsetless data cannot identify a real instant → preserve its recorded calendar day and raw timestamp without assigning a timezone.
- Existing malformed calendar data previously produced inconsistent filtering or count errors → consistently exclude it from dated queries while retaining it in unbounded results; do not rewrite it.
- Repeated source-day validation adds query work → retain indexable raw range predicates and existing indexes; no reference-NAS performance claim is made by functional tests.
- Legacy scopes differ between endpoints → preserve those supported scopes and policies; this change establishes date interpretation rather than silently expanding or redefining them.
