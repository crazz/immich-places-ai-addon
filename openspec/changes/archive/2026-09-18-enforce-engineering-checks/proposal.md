## Why

The adopted engineering standards currently depend on manual review, and the baseline has five lint errors plus Go formatting drift. Contributors need reproducible checks that reject violations before new AI functionality is built.

## What Changes

- Enforce the adopted 500-physical-line limit, explicit exclusions and inherited-file no-growth policy, with actionable failure reports.
- Detect prohibited dependencies in new AI code and frontend dependency cycles while distinguishing inherited violations from new ones.
- Give contributors and CI the same documented check entry point for source rules, lint, types, existing backend tests, formatting and builds, using pinned toolchains and the frozen dependency lockfile.
- Correct the identified baseline lint errors, Go formatting differences and observed sync-test cleanup race without changing application behavior; retain an honest record of remaining warnings and unavailable checks.
- Keep scope to engineering verification. Frontend unit and application smoke harnesses belong to the next two changes. AI features, new services, dependency upgrades, live-provider tests, library mutations and broad legacy refactoring are non-goals.

## Capabilities

### New Capabilities

None in the application's behavior specifications. This tooling-only change explicitly declares `skip_specs: true`; automated checks will verify its developer-facing outcomes.

### Modified Capabilities

None. Existing browsing, manual location placement and GPX behavior remain unchanged.

## Impact

Repository check tooling, dependency-boundary validation, local developer commands, CI configuration, ignore rules and engineering documentation. Existing frontend lint findings affect three source files; five Go files need formatting. Five existing sync tests also need reliable cleanup within the adopted file-size policy. No application API, schema, stored user data or deployment topology changes. This is the first of the three approved tooling changes and precedes product implementation.
