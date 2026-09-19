## 1. Private profiles survive storage and reopen

- [x] 1.1 Define bounded pure provider configuration validation and redacted data contracts, with positive/negative unit coverage.
- [x] 1.2 Add the provider migration and real SQLite create/list/reopen/upgrade/isolation coverage, including foreign-key and busy-timeout behavior on every pooled connection.

## 2. Revisions and credentials remain safe across edits

- [x] 2.1 Implement atomic revision-checked edits and logical disablement with immutable history, stale/competing edit and tenant-isolation tests.
- [x] 2.2 Enforce encrypted-only credential retention/replacement, destination-change protection, explicit all-revision erasure and account cascades, with corruption and rollback tests.

## 3. Protected configuration is reachable only when enabled

- [x] 3.1 Add default-off configuration and explicit public-origin validation, then integrate authenticated listing/create/update routes with bounded strict JSON, stable safe errors and no external client.
- [x] 3.2 Verify enabled/disabled, session, ownership, origin, malformed/oversized/unsupported input, conflict and zero-outbound-call HTTP contracts using real persistence.

## 4. Users manage profiles through Settings

- [x] 4.1 Add the feature-owned typed API/state and accessible dialog/form through the AI public entry point, with load/disabled/empty/error and create/edit/disable flows.
- [x] 4.2 Cover secret masking/removal/clearing, pending duplicate prevention, cancellation and conflict handling with frontend tests and a real proxy/session browser journey.

## 5. Package, verify and document the capability

- [x] 5.1 Enforce the backend AI 80% statement floor across core and integration code using the race-suite coverage measurement, including untested source; add meaningful checker regressions and retain frontend AI floors.
- [x] 5.2 Update executable/container packaging for internal packages and run the full shared checks and container build.
- [x] 5.3 Complete inline spec/quality/integration reviews and GitNexus change analysis; record scenario evidence, configuration/cleanup/rollback guidance, sync the maintained capability and archive the completed change.
