## 1. Contributors can verify authenticated browsing in an isolated application

- [x] 1.1 Provide deterministic browser execution against production frontend/backend artifacts, migrated temporary persistence and synthetic local Immich data, with bounded startup/cleanup, no server reuse, blocked unexpected external traffic and failure evidence.
- [x] 1.2 Characterize registration, Immich key setup, synchronized browsing, session continuity, logout denial and login through the real browser/proxy/API path while preserving application behavior.

## 2. Contributors can detect unintended manual and GPX writes

- [x] 2.1 Characterize manual coordinate preview, cancellation and confirmation, asserting zero pre-confirmation mutations, exact selected-asset payloads and persisted readback with unrelated assets untouched.
- [x] 2.2 Characterize malformed GPX handling and valid matching/preview/confirmation, asserting zero pre-confirmation mutations, exact matched-asset payloads and persisted readback with unmatched assets untouched.

## 3. Local and CI verification report application smoke failures

- [x] 3.1 Include browser smoke execution in full and focused shared verification, rebuilding required artifacts and blocking execution after prerequisite failure, with regression coverage of orchestration and failure propagation.
- [x] 3.2 Make pinned browser installation, failure artifacts, generated-output exclusions and usable commands available in CI and engineering guidance, recording actual evidence and remaining live/AI coverage limits.
