## 1. An owner can explicitly test synthetic vision and reload its private result

- [ ] 1.1 Deliver a persisted, owner/current-revision-bound synthetic image check through the implemented CH02 dispatcher, with finite request/time/concurrency bounds, safe observations and no Immich dependency; verify image interpretation, exact selected model and zero private-data calls.
- [ ] 1.2 Expose the protected explicit test action and redacted stored summary through the existing provider API, including session/origin/input/policy checks, stale/foreign/disabled rejection and concurrent-start prevention; cover all scenarios of “Test only an explicitly selected owned profile with synthetic input” and preserve offline configuration operations.
- [ ] 1.3 Make accepted tests and completed observations survive real SQLite migration, upgrade and reopen without exposing credentials or replaying requests; deliver the persisted-start and owner-isolation portion of “Keep capability evidence private durable and revision-bound”.

## 2. Users receive independently validated image, JSON and strict-schema evidence

- [ ] 2.1 Deliver the fixed three-probe protocol and separate supported/unsupported/unverified observations, including JSON-only and strict-only compatibility, strict sample validation and honest unknown usage/limits; cover the successful and unsupported-mode scenarios of “Record observed image and output-mode compatibility separately”.
- [ ] 2.2 Reject wrong visual facts, malformed/duplicate-key/schema-invalid output, refusal, truncation and tool responses without repair or fallback; distinguish authentication, unavailable-model, rate-limit, policy and network/server failures using safe categories and counted receiving-fixture evidence.
- [ ] 2.3 Enforce the shared deadline, per-probe byte limits, single-request dispatches, global/per-user limits and zero hidden retries across the full sequence; cover the bounded execution and duplicate/excess-start scenarios of “Bound capability checks without hidden fallback or retries”.

## 3. Cancellation, edits and restarts cannot publish stale capability proof

- [ ] 3.1 Stop subsequent probes after cancellation, deadline, session revocation, profile edit or disablement, and reject late publication for another attempt/revision; cover the remaining authority/cancellation scenarios with deterministic race tests.
- [ ] 3.2 Preserve safe terminal or interrupted evidence across startup, expired attempts, failed completion writes and account deletion; make revision/policy/protocol changes invalidate current applicability without automatic retesting, covering the remaining durable-evidence scenarios with real SQLite fixtures.

## 4. Settings presents an explicit accessible test and honest persisted results

- [ ] 4.1 Deliver the feature-owned test action, destination/model and usage disclosure, typed result states, complete-operation timeout/cancellation, duplicate prevention and stale/account-change protection; cover “Expose an accessible explicit test and honest results” with keyboard/component tests while preserving CH01 secret handling.
- [ ] 4.2 Deliver the complete synthetic test-and-reload browser journey through the real frontend proxy, backend and SQLite with explicit local provider policy, request recording and no real-service calls; preserve offline saves and AI-disabled legacy journeys, with the design's applicable size/dependency, lint/types, race/coverage, migration and build/container checks passing.
- [ ] 4.3 Make operator guidance accurately describe the existing codex-proxy connection, manually selected stronger-model candidate, usage/cancellation limits, additive migration/rollback and opt-in live evidence; keep Sol availability and GATE-03 unverified until an authorized live test supplies actual observations.
