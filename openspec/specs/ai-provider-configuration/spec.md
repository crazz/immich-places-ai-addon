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

AI SHALL default to disabled for the installation. Authenticated listing SHALL explain disabled status without revealing stored profiles, and mutations SHALL be unavailable. Enabled mutations SHALL require session authentication, an explicitly configured matching browser origin, bounded valid input and JSON content. Unauthenticated, invalid-origin, malformed, oversized and unsupported-option requests SHALL fail without persistence. Profile operations SHALL make zero provider and Immich calls and SHALL not claim destination approval or capability support.

#### Scenario: Disabled installation preserves existing workflows
- **GIVEN** the default disabled installation
- **WHEN** users browse, place coordinates manually or preview/confirm GPX matches
- **THEN** those workflows remain available
- **AND** provider listing reports disabled and provider mutations cannot change stored data

#### Scenario: Reject unauthorized or invalid-origin mutations
- **WHEN** a settings mutation has no session or a missing, null or nonmatching origin
- **THEN** it fails before persistence and returns a safe structured error

#### Scenario: Validate bounded configuration without external calls
- **WHEN** a user submits malformed, oversized or unsupported configuration
- **THEN** the request fails with an actionable safe error and no saved revision
- **WHEN** valid settings are created, edited or disabled
- **THEN** they are saved without any provider or Immich request, regardless of destination availability

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
