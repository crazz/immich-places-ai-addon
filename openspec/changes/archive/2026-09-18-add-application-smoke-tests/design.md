## Context

F01 supplies shared verification and F02 supplies frontend unit/component tests. Neither executes the real browser/proxy/API/persistence path. The adopted standard already chooses Playwright against production builds and local external-service fakes. This tooling change preserves application behavior and intentionally skips product delta specifications.

The existing application registers users, stores their Immich key, synchronizes assets into SQLite and offers manual and GPX location confirmation. Numeric place-search input sets coordinates locally. GPX preview matches timestamps in the real backend without writing upstream; the existing confirmed location endpoint performs the Immich update and local persistence. These are the integration boundaries the harness must retain.

## Goals / Non-Goals

**Goals:**

- Reproducible, small browser coverage of authenticated browsing, manual placement and GPX import through the built Next.js frontend, actual proxy, Go server and migrated SQLite database.
- Synthetic fixtures, observable upstream mutations, isolated state and no ordinary-run dependence on internet services or private credentials.
- Visible failure propagation, bounded process lifecycle and one shared local/CI command, with useful failure artifacts.

**Non-Goals:**

- Application behavior changes, AI implementation, comprehensive UI coverage, cross-browser certification, container/deployment changes, live Immich compatibility or performance measurement.
- An AI flag assertion before the application implements that flag, or claims that these legacy journeys establish future AI safety invariants.

## Decisions

### Execute production artifacts with native Playwright lifecycle management

Pin Playwright alongside the existing development tools and run Chromium with one worker, no retries and focused-only tests forbidden. Use Playwright's native web-server readiness and shutdown handling for the frontend, backend and fake Immich service. Dedicated loopback ports cannot reuse an existing server. Startup and shutdown have finite deadlines; a small backend launcher owns temporary database storage, launches the compiled binary from an isolated working directory and removes its data on exit. Explicit synthetic configuration prevents local dotenv files or inherited service credentials from selecting a real library.

The shared runner builds both applications before browser execution, including when the smoke gate is selected alone. A failed prerequisite blocks smoke execution, preventing stale artifacts from producing a misleading pass. Other independent checks retain their existing reporting behavior. CI installs the pinned Chromium runtime and uploads failure evidence without changing release publishing.

### Preserve real application boundaries and fake only external services

Registration, login, key setup, synchronization, browsing, preview, confirmation and persisted readback use real application endpoints through the frontend proxy. The application creates users and runs normal migrations; the harness does not seed private database internals or mock its own API.

A narrow local Immich HTTP fake validates synthetic keys, serves a small catalog and image bytes, records exact mutation bodies and updates its catalog state. Each test owns a distinct account/key; assets start without GPS and without stacks, keeping the legacy write scope explicit. Unknown fake routes are visible errors. The browser receives synthetic map tiles, blocks and reports unexpected external destinations, and blocks service workers. Backend outbound HTTP proxies reject external traffic while loopback Immich calls remain direct. Optional Nominatim label refresh after save/reload is explicitly denied and recorded to exercise the existing offline fallback; unknown destinations fail the suite. The fixture state API is local test infrastructure and never becomes an application route.

### Characterize three existing workflows

Authentication/browsing coverage establishes registration and Immich setup, synchronized photo rendering, session continuity, logout denial and successful login. Manual coverage selects a specific photo, enters numeric coordinates, checks the no-write preview/cancel boundary and confirms only that photo's coordinates. GPX coverage uploads malformed and valid synthetic input, observes failure and preview without upstream writes, then confirms the matched photo. A second unmatched photo makes unintended expansion observable. Persisted coordinates are read through the actual application after reload.

Locators prefer accessible names and observed stable asset identifiers where legacy photo cards lack accessible names. Polling assertions wait for real synchronization and UI state; arbitrary sleeps and test-order dependence are excluded. Known behavior tests may pass immediately against unchanged application code under the user-approved characterization rule. New shared-runner behavior follows normal test-first development.

### Keep maintenance and evidence bounded

Browser helpers and external fixtures live outside application feature code, split by their actual responsibilities and subject to the 500-line limit. Plain Node fixture scripts use the existing tooling lint rules; application lint rules remain intact. Generated Playwright reports and results have exact documented producer/path exclusions. Traces and screenshots are retained on failure; normal runs report tests and durations. Documentation distinguishes installed gates, executed local evidence and unexecuted remote/live checks.

## Risks / Trade-offs

- Synthetic Immich responses cannot prove live version compatibility; retain separate opt-in evidence requirements.
- Chromium alone trades broad platform coverage for a fast, deterministic regression gate. The adopted Node/Go/Bun environment and Linux CI remain the supported harness baseline.
- Fixed isolated ports can conflict with another run; fail clearly instead of attaching to that run or a user's server.
- Legacy photo cards expose few accessible locators; identify their real thumbnail asset URLs without changing product UI solely for tests.
- Forced process termination may leave temporary data if graceful cleanup cannot run; temporary storage contains synthetic data only, and the runner still reports failure. Ordinary successful and failed test teardown must release processes and storage.
- Future AI integration must add real disabled/enabled and authorization scenarios. The current suite establishes only existing non-AI behavior.
