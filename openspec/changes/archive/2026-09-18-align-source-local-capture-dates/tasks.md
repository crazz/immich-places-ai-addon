## 1. Catalog dates and calendar counts agree

- [x] 1.1 Align asset filtering and day counts with the recorded source-local date, preserving offsets, offsetless timestamps, inclusive/one-sided ranges and undated membership without changing stored timestamps or gallery ordering.
- [x] 1.2 Apply the shared date policy to album, folder and marker consumers while preserving each owner's existing tenant, membership, GPS, visibility and stack boundaries, with real SQLite regression coverage.

## 2. Catalog endpoints reject invalid date ranges consistently

- [x] 2.1 Reject malformed and reversed ranges across date-taking catalog endpoints while retaining authentication, equal/optional bounds and required calendar-count bounds, with HTTP contract coverage.
- [x] 2.2 Record scenario-to-test evidence and the implemented catalog contract in the maintained specifications and planning guidance, retaining later FR-01 obligations and actual verification limits.
