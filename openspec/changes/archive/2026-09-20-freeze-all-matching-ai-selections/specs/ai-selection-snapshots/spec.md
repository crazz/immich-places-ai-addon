## ADDED Requirements

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
