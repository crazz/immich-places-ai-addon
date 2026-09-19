# ai-selection-snapshots Specification

## Purpose

Freeze a user's eligible image selection with private scope provenance, explicit exclusions or aggregate query counts, so later AI work can reference an exact bounded set without obtaining authority to transmit or modify those images.

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

### Requirement: Resolve the complete authorized matching set

The system SHALL offer an all-matching selection mode that resolves the complete authorized catalog under the declared normalized scope, independently of loaded pages, display counts or client-provided asset lists. It SHALL preserve album, recursive folder, tag, GPS, visibility and source-local date semantics. The discoverable candidate set SHALL retain catalog suppression of hidden libraries and secondary stack members. Invalid or unsupported scope and mixed explicit/query input SHALL fail without broadening the selection.

#### Scenario: Matching assets span several gallery pages
- **GIVEN** the owner's scope matches eligible photos on multiple pages
- **WHEN** the owner previews all matching assets
- **THEN** every eligible matching photo is included once regardless of the loaded page
- **AND** the snapshot records the normalized scope and all-matching mode

#### Scenario: Combined scope preserves catalog boundaries
- **GIVEN** matching and nonmatching assets across albums, folder siblings, nested folders, tags, GPS states and source-local date boundaries
- **WHEN** all-matching resolution uses those supported scope combinations
- **THEN** membership agrees with the corresponding authorized catalog predicates including inclusive calendar-day bounds and recursive folder component boundaries
- **AND** undated assets participate only when no date bound is requested

#### Scenario: Query discovery does not expose suppressed assets
- **GIVEN** foreign assets, hidden-library assets and secondary stack members would otherwise match the filter
- **WHEN** the user previews all matching assets
- **THEN** those assets contribute neither membership nor counts nor exclusion details

#### Scenario: Ambiguous query input is rejected
- **GIVEN** an all-matching request containing explicit asset IDs, paging inputs, unsupported filters or an invalid scope
- **WHEN** the preview is requested
- **THEN** the request fails without a snapshot or fallback to a wider scope

### Requirement: Report exact query counts and bounded exclusions

The system SHALL report exact matched, eligible and excluded asset counts for one completed resolution, with matched equal to eligible plus excluded. Each discoverable candidate SHALL be counted once and classified by the shared selection eligibility policy. Exclusion reasons SHALL be aggregated without returning an unbounded list of excluded asset identifiers or private metadata. Empty and excluded-only results SHALL return counts and no snapshot.

#### Scenario: Mixed catalog candidates have reconcilable counts
- **GIVEN** four discoverable candidates comprising two eligible images, one video and one individually hidden image in an all-visibility scope
- **WHEN** the all-matching preview completes
- **THEN** matched is four, eligible is two and excluded is two
- **AND** aggregate reasons identify one unsupported type and one hidden-by-policy exclusion
- **AND** only the two eligible IDs enter the snapshot

#### Scenario: Empty and excluded-only queries remain explainable
- **GIVEN** a valid scope matching either no candidates or only ineligible candidates
- **WHEN** the preview completes
- **THEN** exact counts and aggregate exclusion reasons describe the result
- **AND** no empty snapshot is created

#### Scenario: Relational matches do not multiply assets
- **GIVEN** an asset matches multiple underlying catalog relationships
- **WHEN** the query is resolved
- **THEN** that asset contributes at most one to each applicable count and occurs at most once in membership

### Requirement: Reject oversized matching batches without truncation

The system SHALL apply the configured batch limit to eligible all-matching assets and SHALL reject the entire selection when that limit is exceeded. A completed oversized resolution SHALL report exact counts without publishing a snapshot or selecting an arbitrary subset. Numerous excluded candidates SHALL not consume the eligible-asset allowance. Resolution SHALL have bounded resource use; interruption before complete enumeration SHALL return failure without presenting provisional counts as exact.

#### Scenario: Exact-limit and over-limit batches differ explicitly
- **GIVEN** a configured eligible-asset limit of 500
- **WHEN** completed queries contain respectively 500 and 501 eligible assets
- **THEN** the first can produce a 500-asset snapshot
- **AND** the second returns a size-limit error with eligible count 501 and no snapshot

#### Scenario: Exclusions do not exhaust the eligible allowance
- **GIVEN** a scope matching many excluded videos and ten eligible images within the resolution budget
- **WHEN** the configured eligible limit is 500
- **THEN** the preview can create a ten-image snapshot with exact counts for the entire matched set

#### Scenario: Interrupted enumeration has no partial success
- **GIVEN** resolution times out, is cancelled or encounters a catalog/storage failure before completion
- **WHEN** the request finishes
- **THEN** it reports failure without a snapshot or purportedly exact partial counts
- **AND** retry requires a new complete resolution

### Requirement: Freeze query membership from one consistent catalog state

The system SHALL derive counts and retained membership from one consistent catalog state and publish them atomically. Retained membership SHALL have deterministic order. Reading or consuming a snapshot SHALL use its frozen IDs and historical preview counts rather than rerun the all-matching query. Current-authority and eligibility revalidation SHALL invalidate the whole snapshot if a retained target becomes ineligible, without shrinking or extending it.

#### Scenario: Concurrent catalog changes cannot split counts and membership
- **GIVEN** catalog synchronization changes matching rows during preview
- **WHEN** a snapshot is successfully published
- **THEN** its membership and counts describe one consistent state before or after the concurrent change
- **AND** publication failure leaves no usable partial snapshot

#### Scenario: Later matching additions do not join a snapshot
- **GIVEN** a valid retained snapshot and subsequent matching catalog additions or changed browser filters
- **WHEN** its owner reads or consumes it
- **THEN** membership, order, normalized scope and historical counts remain those of the original preview

#### Scenario: A frozen member loses eligibility
- **GIVEN** a retained member becomes inaccessible, hidden or outside its original scope
- **WHEN** the owner reads or consumes the snapshot
- **THEN** the entire snapshot is reported stale and requires a new preview
- **AND** no replacement or smaller usable selection is returned

### Requirement: Preserve selection authority and side-effect boundaries

All-matching preview and retained snapshots SHALL obey the existing selection contract for authentication, owner isolation, mutation origin, AI enablement, installation/policy binding, expiry, quotas, deletion and cleanup. Creating or reading them SHALL perform no provider call, image download, job submission or Immich mutation and SHALL convey no analysis or write consent. Explicit selection and existing manual/GPX/gallery behavior SHALL remain compatible.

#### Scenario: Unauthorized query and retained-resource access fail
- **GIVEN** a missing session, invalid mutation origin, disabled AI or another user's snapshot
- **WHEN** the corresponding preview or read is attempted
- **THEN** the request fails under the shared owner-safe selection error contract without disclosing another user's matching counts or membership

#### Scenario: Shared lifecycle limits apply to both modes
- **GIVEN** an expired or installation-invalid snapshot, an exhausted owner quota, or deletion of the owning account
- **WHEN** all-matching resources are read, created or cleaned respectively
- **THEN** the shared expiry, invalidation, capacity and deletion rules apply equally to explicit and all-matching snapshots

#### Scenario: Query preview preserves existing workflows and sends nothing externally
- **GIVEN** explicit selection, gallery paging, manual GPS and GPX workflows remain available
- **WHEN** all-matching preview and snapshot reads complete
- **THEN** those workflows retain their prior behavior
- **AND** no provider, image-download, job or Immich-write operation occurs
