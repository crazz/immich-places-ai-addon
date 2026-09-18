## Why

Unit tests and builds cannot establish that the browser, frontend proxy, backend and database work together. Before AI integration, contributors need a repeatable check that existing authentication, browsing, manual placement and GPX workflows still work.

## What Changes

- Establish a small deterministic application smoke suite using the Playwright harness adopted in the testing standard.
- Exercise known existing UI behavior against the built frontend, real Go backend and isolated local persistence, with synthetic data and local external-service fakes.
- Verify representative authenticated browsing, manual location confirmation and GPX preview/application boundaries, including the absence of writes before confirmation.
- Make the suite usable locally and in CI, with bounded startup/cleanup and visible failures, and document actual evidence and remaining limits.
- Preserve application behavior. AI feature implementation, broad UI coverage, live services, private libraries, performance testing and deployment changes are non-goals.

## Capabilities

### New Capabilities

None in application behavior. This tooling-only change declares `skip_specs: true` and characterizes existing workflows.

### Modified Capabilities

None. Existing product contracts remain unchanged.

## Impact

Development dependencies and lockfile, browser-test configuration and synthetic fixtures, shared verification/CI integration and engineering documentation. F01 and F02 are completed prerequisites. Ordinary runs require no real Immich/provider credentials or private photos, and do not establish live compatibility or AI quality.
