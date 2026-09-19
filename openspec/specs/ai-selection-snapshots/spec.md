# ai-selection-snapshots Specification

## Purpose

Freeze a user's eligible image selection with private scope provenance and explicit exclusions, so later AI work can reference an exact bounded set without obtaining authority to transmit or modify those images.

## Requirements

### Requirement: Protect selection resources by current authority

Selection preview and snapshot access SHALL require an authenticated current user, enabled AI and the active installation binding. Creating a snapshot SHALL require the configured application origin and a valid JSON request. Resource access SHALL verify ownership independently of knowledge of the snapshot identifier. Foreign and nonexistent resource identifiers SHALL be indistinguishable and SHALL disclose no selection data. Responses SHALL prohibit shared caching.

#### Scenario: Owner creates and reads a selection
- **GIVEN** AI is enabled and the authenticated owner supplies a valid explicit selection from the configured origin
- **WHEN** the owner previews it and reads the resulting snapshot
- **THEN** the response contains only that owner's selection and is not cacheable

#### Scenario: Missing session or foreign snapshot
- **GIVEN** a snapshot belongs to one user
- **WHEN** an unauthenticated caller or a different user requests it
- **THEN** the unauthenticated caller is rejected and the other user receives the same unavailable result as for a nonexistent snapshot
- **AND** neither response contains IDs, counts, scope metadata or exclusions from the snapshot

#### Scenario: Disabled AI or invalid creation origin
- **GIVEN** otherwise valid selection input
- **WHEN** AI is disabled or the creation origin is absent, duplicated or different from the configured origin
- **THEN** creation is rejected before selection data is persisted
- **AND** disabling AI also prevents snapshot reads

### Requirement: Bound and deduplicate explicit input

An explicit preview SHALL accept only a nonempty bounded list of valid asset identifiers and a supported scope. It SHALL collapse equivalent duplicate identifiers once, preserve first-occurrence order, and report requested, unique, duplicate, eligible and excluded counts with `requested = unique + duplicate` and `unique = eligible + excluded`. Limits SHALL be checked before unbounded work. Malformed, ambiguous, unknown-field, oversized or over-limit input SHALL fail as a whole; the system SHALL NOT truncate it into a smaller successful selection.

#### Scenario: Duplicates are one target
- **GIVEN** an input list contains A, B and a repeated equivalent representation of A
- **WHEN** both distinct assets are eligible
- **THEN** the target order is A then B, requested count is three, unique and eligible counts are two, and duplicate count is one

#### Scenario: Invalid or ambiguous input
- **GIVEN** input contains an invalid identifier, an empty list, unsupported mode, unknown field, duplicate object key or trailing JSON
- **WHEN** preview is requested
- **THEN** the request is rejected without creating a snapshot or returning a partial preview

#### Scenario: Size and batch bounds
- **GIVEN** a request exceeds the raw-entry, byte, response or configured unique-asset limit
- **WHEN** preview is requested
- **THEN** an explicit limit error is returned and no partial snapshot exists
- **AND** duplicates cannot bypass raw-entry bounds and excluded IDs cannot bypass the explicit unique-asset limit

### Requirement: Preserve the declared catalog scope

The selection SHALL retain and enforce the normalized active all-catalog, single-album or recursive-folder view, applicable tag, GPS, visibility and source-local capture-date filters. Unsupported or conflicting scope fields SHALL be rejected rather than ignored or broadened. Folder matching SHALL preserve case and directory-component boundaries. Date interpretation SHALL reuse `catalog-capture-dates`; missing GPS SHALL mean either coordinate is absent, with zero treated as a valid coordinate.

#### Scenario: Explicit IDs outside the active scope
- **GIVEN** supplied owner-visible image IDs span matching and nonmatching album, tag and GPS membership
- **WHEN** a valid scoped preview is resolved
- **THEN** only matching eligible IDs become targets and nonmatching IDs are reported as outside scope
- **AND** the effective filter defaults and supplied scope are retained

#### Scenario: Recursive folder boundaries
- **GIVEN** assets exist directly under `/Trip`, under `/Trip/Day1`, under `/Trips` and under `/trip`
- **WHEN** the selection scope is `/Trip/`
- **THEN** matching normalizes the trailing slash and includes only the first two paths subject to other eligibility rules
- **AND** empty or root-only folder scopes are rejected without falling back to all assets

#### Scenario: Date and GPS edge cases
- **GIVEN** assets include offset-bearing and offsetless dates, undated assets, a partial GPS pair and a complete pair containing zero
- **WHEN** source-local date bounds and a GPS filter are applied
- **THEN** recorded calendar days follow the maintained date contract, undated assets are excluded by either date bound, and only the partial pair counts as missing GPS

#### Scenario: Invalid or unsupported scope
- **GIVEN** a scope combines album and folder fields, uses an unknown filter, supplies an unavailable album or tag, or has a reversed date range
- **WHEN** preview is requested
- **THEN** the scope is rejected with an explanatory owner-safe error and no broadened result

### Requirement: Explain eligibility without leaking inaccessible assets

Only accessible catalog images that are not in hidden libraries, not individually hidden, not stack children and within the declared scope SHALL be eligible. Each unique excluded explicit ID SHALL receive exactly one deterministic reason. Absent, foreign and hidden-library IDs SHALL share an unavailable reason and SHALL disclose no metadata. Owner-visible unsupported types, individually hidden images, stack children and outside-scope assets SHALL have distinct reasons, in that precedence after unavailable. No file extension SHALL be represented as proof that later image decoding is supported.

#### Scenario: Mixed eligible and excluded selection
- **GIVEN** explicit input contains an eligible image, a video, an individually hidden image, a stack child and an out-of-scope visible image
- **WHEN** preview is resolved
- **THEN** only the eligible image is frozen and every other unique ID has its corresponding exclusion reason
- **AND** no stack member is added automatically

#### Scenario: Foreign and unavailable IDs are indistinguishable
- **GIVEN** submitted IDs include a foreign asset, an absent asset and an asset in a hidden library
- **WHEN** preview is resolved
- **THEN** each receives the same unavailable reason without returned filename, path, type or other asset metadata

#### Scenario: Hidden scope does not override AI policy
- **GIVEN** the catalog scope includes hidden images and a submitted image is individually hidden
- **WHEN** AI selection is previewed
- **THEN** the original visibility scope is retained but that image is excluded as hidden by policy
- **AND** a zero-eligible result has exact counts and no usable snapshot identifier

### Requirement: Publish an exact atomic snapshot

A successful nonempty preview SHALL publish an immutable owner-bound snapshot containing exact eligible membership, normalized scope, selection-policy version, relevant source provenance, authoritative counts and absolute creation/expiry times from one consistent catalog observation. If persistence, cancellation or resource bounds prevent completion, no usable partial snapshot SHALL be published. Later sync additions, browser filter changes or new stack members SHALL NOT add targets. No snapshot SHALL store image bytes, credentials or unrelated asset metadata.

#### Scenario: Catalog changes after preview
- **GIVEN** a snapshot contains eligible A and B
- **WHEN** sync adds C, a stack gains another member or the browser changes filters
- **THEN** the snapshot still identifies only A and B under its original scope
- **AND** unrelated changes or harmless resynchronization do not alter its counts or expiry

#### Scenario: Concurrent sync and atomic failure
- **GIVEN** sync changes relevant catalog state while preview resolves, or persistence fails before publication
- **WHEN** the preview operation finishes
- **THEN** it either publishes one internally consistent manifest with matching counts or fails without a usable snapshot
- **AND** it never combines membership from one observation with counts from another

### Requirement: Expire and revalidate retained membership

Snapshots SHALL remain readable after an ordinary backend restart until their absolute expiry, subject to current session, ownership, AI enablement, installation binding, policy and retained-target eligibility. Revalidation SHALL inspect only frozen targets and their relevant original scope facts. Loss of access, a changed relevant fact, a changed installation binding or expiry SHALL prevent use of the complete snapshot; it SHALL NOT silently shrink, refresh or extend it. Missing cleanup records SHALL be treated as unavailable. A snapshot SHALL NOT authorize provider dispatch or stand in for fresh upstream image permission checks.

#### Scenario: Reopen before expiry
- **GIVEN** a valid snapshot has been persisted and its relevant catalog facts remain unchanged
- **WHEN** the backend restarts and the owner reads it before expiry
- **THEN** the same exact membership, scope, counts and expiry are returned

#### Scenario: A frozen target becomes ineligible
- **GIVEN** a valid snapshot contains A and B
- **WHEN** B is removed, hidden, loses accessible membership or no longer matches a relevant frozen filter
- **THEN** snapshot use is rejected as stale without returning a smaller usable selection or replacing B

#### Scenario: Expiry, policy or installation change
- **GIVEN** an existing snapshot
- **WHEN** its expiry is reached, its policy changes or its installation binding rotates
- **THEN** it cannot be used even if the caller still possesses its identifier
- **AND** rereading it cannot renew its lifetime

### Requirement: Bound snapshot retention and ownership cleanup

The installation SHALL enforce positive bounded selection capacity and lifetime settings, finite per-owner and installation-wide live snapshot limits, and bounded cleanup. Expired metadata SHALL become inaccessible immediately and be physically removed on startup and during periodic cleanup while the backend is running, including with AI disabled. Account deletion SHALL remove that user's snapshots. Quota rejection SHALL NOT evict or expose another valid snapshot. Cleanup failures SHALL remain observable without logging private manifests.

#### Scenario: Quota and cleanup
- **GIVEN** live snapshot capacity is reached and some other snapshots have expired
- **WHEN** an owner creates another preview
- **THEN** expired entries are removed before capacity is evaluated and remaining exhaustion produces an explicit error
- **AND** valid snapshots are preserved

#### Scenario: Expiry while AI is disabled
- **GIVEN** persisted snapshots expire while AI is disabled
- **WHEN** the bounded cleanup cycle or next startup runs
- **THEN** expired snapshot metadata is removed without enabling AI or making provider requests

#### Scenario: Account deletion and cleanup failure
- **GIVEN** one user owns snapshots and another user's snapshots also exist
- **WHEN** the first account is deleted
- **THEN** only its snapshots are removed
- **AND** a separate expired-record cleanup failure leaves expired data inaccessible, reports sanitized diagnostics and is retried on a later bounded cleanup pass

### Requirement: Keep selection separate from analysis and writes

Previewing, loading, invalidating and cleaning snapshots SHALL cause zero provider calls, image downloads and Immich mutations. This capability SHALL preserve existing catalog, manual placement and GPX behavior when AI is disabled or selection input fails. It SHALL NOT expose a launch workflow before the durable analysis prerequisites exist.

#### Scenario: Selection is locally verifiable
- **GIVEN** deterministic catalog fixtures and external-call counters
- **WHEN** valid, excluded, stale and failed selection operations run
- **THEN** the selection results follow the contract and all provider, image-fetch and Immich-write counters remain zero

#### Scenario: Existing workflows remain available
- **GIVEN** AI is disabled or an AI selection request was rejected
- **WHEN** an authenticated user browses, manually places an asset or uses GPX confirmation
- **THEN** existing workflows retain their behavior and no incomplete AI launch control appears
