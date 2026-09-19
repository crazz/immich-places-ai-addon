## Purpose

Allow each authenticated user to manage private, durable provider settings and credentials without sending photographs or invoking an external provider. This is CH01's configuration foundation for FR-03 and FR-04.

## Requirements

### Requirement: Keep provider profiles private and revisioned

The system SHALL let an authenticated user create, list and edit only their own profiles, with a name, base URL, manually entered model and enabled state. Creation SHALL start at revision one. Updates SHALL require the current revision and atomically create a new immutable configuration revision. Stale edits SHALL fail without overwriting data. Disablement SHALL apply to the logical profile across all revisions.

#### Scenario: Create and reopen a private profile
- **WHEN** a user saves a valid profile and later reloads it after a backend/database reopen
- **THEN** its configuration, enabled state and revision remain available only to that user
- **AND** responses contain no readable credential

#### Scenario: Edit with a current or stale revision
- **GIVEN** a saved profile
- **WHEN** an edit supplies its current revision
- **THEN** a new revision becomes current and prior configuration remains unchanged
- **WHEN** another edit supplies the old revision
- **THEN** it fails with a conflict and cannot overwrite the newer revision

#### Scenario: Disable a logical profile
- **WHEN** its owner disables a profile using the current revision
- **THEN** the logical profile is disabled across its historical revisions
- **AND** another user's profiles are unchanged

#### Scenario: Another user references a profile
- **WHEN** a user lists profiles or attempts to edit another user's profile ID
- **THEN** no other user's configuration or credential is disclosed or changed
- **AND** the unavailable ID is handled like a missing profile

### Requirement: Encrypt and explicitly remove credentials

Credentials SHALL be stored only as authenticated ciphertext and never returned in readable form, ordinary logs or errors. An omitted edit secret SHALL retain the existing secret only when the destination is unchanged. Explicit replacement SHALL bind the new secret to the new revision. Explicit removal SHALL erase credential material from every revision of the logical profile while retaining non-secret history. Owning-account deletion SHALL delete all profile and credential records. Disablement alone SHALL retain encrypted credentials; removal SHALL not claim to rewrite backups. Plaintext or corrupt stored credentials SHALL fail closed when reused.

#### Scenario: Retain or replace a credential
- **WHEN** an owner edits a profile without changing its destination and omits the secret, or supplies a replacement
- **THEN** the new revision retains or replaces the encrypted secret as requested
- **AND** public data reveals only whether a credential exists

#### Scenario: Change destination without choosing credential treatment
- **GIVEN** a profile with a stored credential
- **WHEN** its destination changes without explicit replacement or removal
- **THEN** the edit is rejected without creating a revision or reusing the credential at a new destination

#### Scenario: Remove credentials and delete the owning account
- **WHEN** an owner explicitly removes a profile credential
- **THEN** all revisions of that profile contain no credential material and unrelated profiles remain unchanged
- **WHEN** that account is deleted
- **THEN** all its provider profiles and versions are removed

#### Scenario: Reuse corrupted or plaintext stored credentials
- **WHEN** an edit would retain a stored credential that is plaintext or fails authentication
- **THEN** the operation fails without leaking the value or creating a revision

### Requirement: Protect configuration and keep saves offline

AI SHALL default to disabled for the installation. Authenticated listing SHALL explain disabled status without revealing stored profiles, and mutations SHALL be unavailable. Enabled mutations SHALL require session authentication, an explicitly configured matching browser origin, bounded valid input and JSON content. Unauthenticated, invalid-origin, malformed, oversized and unsupported-option requests SHALL fail without persistence. Profile creation, listing, editing and disablement SHALL make zero provider and Immich calls and SHALL not themselves claim destination approval or capability support. Only the separately and explicitly invoked capability test SHALL contact a provider under the approved dispatch policy; loading or saving Settings SHALL NOT trigger that test.

#### Scenario: Disabled installation preserves existing workflows
- **GIVEN** the default disabled installation
- **WHEN** users browse, place coordinates manually or preview/confirm GPX matches
- **THEN** those workflows remain available
- **AND** provider listing reports disabled and provider mutations cannot change stored data

#### Scenario: Reject unauthorized or invalid-origin mutations
- **GIVEN** a settings mutation has no session or a missing, null or nonmatching origin
- **WHEN** the backend receives the mutation
- **THEN** it fails before persistence and returns a safe structured error

#### Scenario: Validate bounded configuration without external calls
- **GIVEN** provider settings are available
- **WHEN** a user submits malformed, oversized or unsupported configuration
- **THEN** the request fails with an actionable safe error and no saved revision
- **WHEN** valid settings are created, edited or disabled
- **THEN** they are saved without any provider or Immich request, regardless of destination availability

#### Scenario: Reloading settings does not repeat a capability test
- **GIVEN** a profile has a previous successful or failed test
- **WHEN** Settings is loaded, reloaded or saved
- **THEN** only stored observations are displayed where applicable
- **AND** no provider test or capability fallback is triggered

### Requirement: Provide accessible provider settings

The existing Settings experience SHALL expose private provider management with labeled controls, manual model entry, enabled state and visible loading/empty/error/disabled states. Saved secrets SHALL never be prefilled or persisted in browser storage. Cancel/close SHALL cause no mutation and discard entered secrets. A pending save SHALL prevent duplicate submission; success SHALL display the saved revision. Failure or conflict SHALL remain visible without an automatic mutation retry or silent overwrite.

#### Scenario: Create, edit and disable through Settings
- **WHEN** a user opens Settings and creates, edits or disables a profile
- **THEN** labeled keyboard-accessible controls show the authoritative saved state and revision
- **AND** no connection test or photo disclosure occurs

#### Scenario: Cancel and handle failures safely
- **WHEN** a user cancels or closes with an entered secret
- **THEN** no save occurs and that secret is discarded
- **WHEN** a save is pending or returns a validation/conflict error
- **THEN** duplicate saves are prevented, the error is visible, non-secret edits remain available and entered secret text is cleared

#### Scenario: Explain disabled, empty and unavailable settings
- **WHEN** provider settings are disabled, empty or fail to load
- **THEN** the UI explains that state and offers an appropriate retry or creation action without exposing another account's data

### Requirement: Test only an explicitly selected owned profile with synthetic input

An authenticated owner SHALL be able to explicitly test the current revision of an enabled private profile when AI is enabled. Before submission the UI SHALL identify the destination and model, disclose that up to three synthetic-image requests can consume provider usage, and explain that the local proxy can forward data to its upstream model service. The backend SHALL require the expected current revision, mutation-origin protection, bounded JSON input and CH02's approved dispatch path. It SHALL reject caller-supplied image data, asset IDs, prompts, URLs, model overrides and transport options. Tests SHALL use only server-owned synthetic fixtures and SHALL have no Immich read or write dependency. The selected profile's model SHALL be used without changing the proxy's global default or requiring model discovery.

#### Scenario: Explicitly test the existing proxy with the selected model
- **GIVEN** the owner has saved an enabled profile for the existing approved NAS proxy and manually entered a model
- **WHEN** the owner deliberately starts the disclosed test for its current revision
- **THEN** the existing proxy receives only that test's synthetic input and selected model
- **AND** no model-list request, new container or global default-model change occurs
- **AND** no private photo, Immich URL, library metadata or Immich request is involved

#### Scenario: Reject unauthorized, stale or unapproved tests
- **GIVEN** a missing/expired session, foreign profile, stale revision, disabled installation/profile, invalid origin or unapproved destination
- **WHEN** a capability test is requested
- **THEN** it fails with a safe actionable error before any provider request
- **AND** unavailable and foreign profiles remain indistinguishable

#### Scenario: Reject injected test payloads and options
- **GIVEN** a test request includes unsupported fields, an oversized/malformed body or a non-JSON content type
- **WHEN** the backend validates it
- **THEN** it fails without storing an accepted test or making an external request

### Requirement: Record observed image and output-mode compatibility separately

The test SHALL report independent observations for image input, JSON mode and strict-schema mode, distinguishing supported, explicitly unsupported and unverified outcomes. Image support SHALL require a response matching facts visible in the synthetic fixture that are not supplied as answers in its prompt. A successful output-mode observation SHALL require a complete non-refused response that passes server validation of the synthetic contract and fixture facts. Strict-schema support SHALL mean that the schema request was accepted and the returned sample validated; it SHALL NOT claim proof that the provider always enforces schemas. Image plus JSON-mode support SHALL remain a usable compatibility outcome when strict schema is explicitly unsupported. Missing usage, token-limit support and provider size limits SHALL remain unknown rather than invented.

#### Scenario: Observe full compatibility
- **GIVEN** the selected endpoint/model correctly reads the fixture and returns valid JSON and strict-schema samples
- **WHEN** the explicit test completes
- **THEN** the report marks all three observations supported for that exact revision and test protocol
- **AND** it labels them as observed synthetic compatibility rather than geolocation accuracy or guaranteed schema enforcement

#### Scenario: Preserve useful JSON-only compatibility
- **GIVEN** image and JSON-mode samples pass and the strict-schema request is explicitly rejected as unsupported
- **WHEN** the report is produced
- **THEN** it reports image and JSON support with strict schema unsupported
- **AND** it identifies JSON-only compatibility without changing the configured model, destination or sending a replacement request

#### Scenario: Strict schema can work without JSON mode
- **GIVEN** image input passes, JSON mode is explicitly unsupported and the strict-schema sample passes
- **WHEN** the report is produced
- **THEN** it reports strict-schema compatibility while preserving the separate JSON-mode result

#### Scenario: A successful HTTP status is insufficient
- **GIVEN** a response ignores the image, supplies wrong fixture facts, malformed or schema-invalid JSON, a refusal, a tool call or truncated output
- **WHEN** the test validates that response
- **THEN** the affected observation is not marked supported
- **AND** no invalid sample produces a compatibility pass or an analysis/draft/write artifact

#### Scenario: Keep unknown limits and metadata honest
- **GIVEN** a provider omits usage, size limits or token-limit parameter support
- **WHEN** the test succeeds or fails
- **THEN** those fields remain unknown
- **AND** local request/time ceilings are not presented as discovered provider limits

### Requirement: Bound capability checks without hidden fallback or retries

One explicit test SHALL issue at most three non-streaming provider requests within a single 120-second provider-work deadline, using no private input and CH02's stricter applicable transport bounds. The installation SHALL run at most two tests concurrently and at most one per user; excess requests SHALL fail visibly rather than queue. Concurrent duplicate starts SHALL not multiply calls. Each probe SHALL recheck the current profile revision, session authority and installation/profile enablement. Authentication, unavailable-model, policy, network, rate-limit and server failures SHALL remain distinct from unsupported-mode evidence and SHALL stop remaining probes. Refusal, truncation or invalid output SHALL not trigger repair or another model/mode attempt. Cancellation SHALL stop subsequent probes and attempt to cancel in-flight I/O without promising usage reversal. A later retry SHALL require a new explicit user action.

#### Scenario: Run only the disclosed bounded probes
- **GIVEN** the owner starts one accepted test
- **WHEN** the configured endpoint completes each permitted probe
- **THEN** at most three requests run within the shared deadline
- **AND** no SDK retry, repair call, undisclosed downgrade, tool execution or model substitution occurs

#### Scenario: Reject duplicate or excess starts
- **GIVEN** that user's test is running or both installation test slots are occupied
- **WHEN** another test start arrives
- **THEN** it reports a busy result without another accepted test or provider request
- **AND** a double submission for the same running profile cannot double provider usage

#### Scenario: Distinguish failures from mode incompatibility
- **GIVEN** a probe receives an authentication/model/policy error, rate limit, server error or network failure
- **WHEN** the test processes the failure
- **THEN** it reports that safe failure category and stops remaining probes
- **AND** it does not reinterpret the failure as unsupported strict schema or a successful JSON fallback

#### Scenario: Cancel, time out or revoke between probes
- **GIVEN** a test has started
- **WHEN** its caller cancels, its deadline expires, its session is revoked, its profile revision changes, or the profile/installation is disabled before a later dispatch
- **THEN** no subsequent probe is admitted
- **AND** already transmitted input is described as potentially consumed
- **AND** no automatic restart or retry follows

### Requirement: Keep capability evidence private durable and revision-bound

The backend SHALL retain the latest test record per owned profile revision with a unique attempt identity, protocol version, policy identity, timestamps, independent observations and safe failure categories. Test startup SHALL persist before provider dispatch. Completion SHALL update only the matching attempt. Results SHALL survive database reopen, remain owner-scoped, cascade on account deletion and contain no credential, image bytes or raw provider response. Any profile edit SHALL leave the new revision untested; old or late observations SHALL never become the current revision's proof. Disablement, current policy changes and test-protocol changes SHALL make old evidence inapplicable for current use. Interrupted tests SHALL be visible after restart or expiration and SHALL NOT resume automatically. Persistence failure SHALL not be reported as a durable successful test or cause provider replay.

#### Scenario: Persist and reload a private observation
- **GIVEN** a completed test for an owned profile revision
- **WHEN** the backend/database is reopened and the owner reloads Settings
- **THEN** the same revision-bound safe report remains available without another provider call
- **AND** another user cannot read or change it

#### Scenario: Invalidate after edits or policy changes
- **GIVEN** a previous successful report
- **WHEN** the profile receives a new revision, is disabled, or the active destination policy/test protocol changes
- **THEN** that report is not represented as current usable proof
- **AND** saving or loading the changed settings does not retest automatically

#### Scenario: Reject late publication into a newer configuration
- **GIVEN** a test is in flight and the owner edits the profile or starts a later permitted attempt
- **WHEN** an old response arrives
- **THEN** it cannot overwrite the newer attempt or become evidence for the new revision
- **AND** the current view explains that current settings need their own completed test

#### Scenario: Recover interruption without replay
- **GIVEN** an accepted test did not persist completion before process exit or expiration
- **WHEN** the backend restarts or the owner reads the expired test
- **THEN** the record reports interruption rather than remaining permanently running or claiming success
- **AND** no provider request is replayed without a new explicit test action

#### Scenario: Handle storage failure and account deletion
- **GIVEN** test storage fails or the owner is deleted during a test
- **WHEN** startup or completion is attempted
- **THEN** failure before persisted startup causes zero provider requests
- **AND** failure after dispatch reports unavailable durable evidence without replaying requests
- **AND** account deletion removes its test records and a late response cannot recreate them

### Requirement: Expose an accessible explicit test and honest results

Provider Settings SHALL offer a labeled keyboard-accessible test action for saved current enabled profiles, separate from saving. It SHALL show the destination/model and usage disclosure, pending, busy, unsupported, unverified, failed, interrupted, stale and completed states with textual labels. Pending UI SHALL prevent duplicate starts, support cancel/close, and avoid displaying late results for a changed profile or another account. Failure SHALL offer an appropriate explicit retry/reload action without automatically resubmitting. Secrets SHALL remain masked/cleared under CH01 rules and SHALL never enter browser storage or result diagnostics. Synthetic compatibility SHALL NOT be labeled geolocation quality or permission to send private photographs.

#### Scenario: Start and inspect a test using the keyboard
- **GIVEN** an owner opens a saved enabled profile in Settings
- **WHEN** the owner activates its labeled test action after the visible disclosure
- **THEN** the UI shows pending and then the three observed capability states for the tested revision
- **AND** reload shows persisted results without retesting

#### Scenario: Handle cancel and stale responses safely
- **GIVEN** a test is pending in Settings
- **WHEN** the owner closes/cancels, edits the profile or switches accounts before a response arrives
- **THEN** cancellation is requested and no late result is shown as current for the changed context
- **AND** duplicate submission is prevented while the request is pending

#### Scenario: Explain failures without exposing secrets or inventing readiness
- **GIVEN** a test is busy, unsupported, unverified, failed, interrupted or stale
- **WHEN** Settings renders the result
- **THEN** it gives a textual explanation and an appropriate explicit retry, reload or configuration action
- **AND** it exposes neither secret values nor raw provider responses
- **AND** it distinguishes observed compatibility from private-photo authorization and geolocation quality
