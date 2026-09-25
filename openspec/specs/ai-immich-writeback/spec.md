# ai-immich-writeback Specification

## Purpose

Let owners inspect exact revision-bound GPS changes and, in a subsequent change, authorize durable single-asset execution with explicit conflict handling and verified outcomes.

## Requirements

### Requirement: Preview only a staged exact GPS decision

An owner SHALL be able to request a GPS before/after preview for the current staged revision of an owned draft with an acknowledged baseline and a finite camera pair. The preview SHALL target exactly the analyzed asset and GPS pair. Client/model input SHALL not add targets, stack members or other writable fields. Heading, descriptions, confidence and estimated error SHALL not introduce additional GPS readiness gates. Preview operations SHALL make zero provider calls and zero Immich mutations and SHALL not grant write approval.

#### Scenario: P01 Preview one coarse camera decision
- **GIVEN** an owned staged draft with valid camera coordinates, acknowledged baseline and a large or unknown estimated error
- **WHEN** the owner requests a preview
- **THEN** an exact single-photo GPS comparison is returned without writing, provider work or a precision threshold

#### Scenario: P02 Reject an incomplete or expanded decision
- **GIVEN** an unstaged, rejected, baseline-unavailable or invalid-camera draft, or a request adding targets or fields
- **WHEN** a preview is requested
- **THEN** no usable plan is created and the draft remains available for correction

### Requirement: Use fresh authorized source and GPS before-values

Preview creation SHALL read current authorized metadata and compare the reviewed image identity and acknowledged GPS baseline. Changed GPS SHALL produce a conflict rather than silently becoming the approved baseline. Changed source image SHALL require renewed current-image review. Unavailable or malformed metadata SHALL produce no usable plan. Unrelated metadata changes SHALL not falsely imply image replacement. Fresh read access SHALL not be presented as verified write permission.

#### Scenario: P03 Observe missing partial and zero GPS
- **GIVEN** an accessible asset whose current GPS is absent, partial or zero-valued and matches the acknowledged baseline
- **WHEN** a preview is created
- **THEN** the exact before-values remain distinguishable and no missing value is fabricated as zero

#### Scenario: P04 Detect an intervening GPS change
- **GIVEN** GPS changed since the draft's baseline was acknowledged
- **WHEN** preview creation reads current GPS
- **THEN** the owner sees the before/current/proposed conflict and must explicitly review the changed baseline before a new usable preview

#### Scenario: P05 Distinguish image change from unrelated metadata
- **GIVEN** a reviewed source image
- **WHEN** a preview observes changed image content or only unrelated metadata/timestamps
- **THEN** changed image content requires renewed review while unrelated changes alone do not replace or invalidate the image identity

#### Scenario: P06 Fail without a fabricated baseline
- **GIVEN** upstream timeout, access denial, malformed metadata or an unavailable source
- **WHEN** preview creation is attempted
- **THEN** no usable plan is published and saved draft values remain unchanged

### Requirement: Bind an immutable preview to its exact authority and revision

Every preview SHALL bind owner, installation, draft revision, reviewed source identity, exact target, selected fields, before/intended GPS and validity period. Its displayed comparison and digest SHALL identify the same immutable plan. Client modifications or cross-scope references SHALL not substitute a different plan. A draft or authority change during creation SHALL prevent publication of a usable preview. An already-matching GPS pair SHALL be shown as unchanged, not as evidence of a new write.

#### Scenario: P07 Retain the same plan across reload
- **GIVEN** a persisted usable preview
- **WHEN** its owner reloads it after an application restart
- **THEN** the same values, scope, revision, digest and validity period are displayed without a new upstream request

#### Scenario: P08 Race preview creation with an edit
- **GIVEN** preview creation is reading current source metadata
- **WHEN** the draft is edited/rejected, the account is deleted, credentials change or the installation rotates before publication
- **THEN** no usable plan is published under the obsolete revision or authority

#### Scenario: P09 Reject plan substitution
- **GIVEN** another owner's preview or a request attempting to substitute target, coordinates, fields or revision
- **WHEN** the reference is read or submitted as a new preview input
- **THEN** no private plan is disclosed or altered and no new scope is approved

#### Scenario: P10 Display an unchanged GPS pair
- **GIVEN** the current authorized GPS already equals the intended pair
- **WHEN** a preview is created
- **THEN** the comparison is labeled unchanged and no mutation or completed-write claim is made

### Requirement: Expire and invalidate previews without changing drafts

Previews SHALL expire five minutes after creation. Reading or refreshing a stored preview SHALL not extend that validity. Draft revision/state changes and installation changes SHALL invalidate old preview authority. Creating a replacement SHALL require a fresh authorized read. Expiration, invalidation and bounded temporary-preview cleanup SHALL preserve the draft and retained review history.

#### Scenario: P11 Inspect an expired preview
- **GIVEN** a preview whose five-minute validity elapsed
- **WHEN** the owner reloads it or attempts to continue using it
- **THEN** it is visibly unusable and can only be replaced by a fresh preview without discarding the draft

#### Scenario: P12 Edit after preview
- **GIVEN** an unexpired preview of a staged draft
- **WHEN** the owner changes or rejects that draft
- **THEN** the old comparison remains an old revision and cannot authorize the new decision

#### Scenario: P13 Bound temporary preview growth
- **GIVEN** repeated preview creation or expired previews awaiting cleanup
- **WHEN** the configured preview capacity is reached or cleanup runs
- **THEN** creation/cleanup remains bounded and does not delete drafts, referenced review history or protected confirmed-write records

### Requirement: Show a usable accessible comparison without manual-save side effects

Preview UI SHALL present photo identity, exact GPS before/after values, single-photo scope, revision, expiry and changed/unchanged status through keyboard-accessible text/numeric controls. Failed maps SHALL not hide the comparison. Conflicts and source failures SHALL offer review/retry paths without silently changing a baseline. AI preview data SHALL not enter manual pending coordinates. Overlapping manual pending edits SHALL require explicit resolution before continuing, and account/result changes SHALL remove private stale content.

#### Scenario: P14 Review without a map
- **GIVEN** a narrow viewport, keyboard input and unavailable map tiles
- **WHEN** the owner opens the GPS preview
- **THEN** the complete before/after comparison and validity state remain usable with no write side effect

#### Scenario: P15 Preserve manual pending coordinates
- **GIVEN** manual edits exist, including an edit to the same photo
- **WHEN** the owner opens a preview or a manual edit appears while it is open
- **THEN** manual edits are preserved and the overlapping choice must be resolved before continuing with the AI plan

#### Scenario: P16 Handle late private preview replies
- **GIVEN** preview creation or reading is in flight
- **WHEN** the account or selected result changes
- **THEN** the old reply cannot populate the new private view or become usable authority there

### Requirement: Preserve preview isolation through lifecycle changes

Preview reads and creation SHALL enforce current owner and installation scope. Account deletion SHALL remove only owned preview data and prevent its recreation by late requests. Disabling provider execution SHALL preserve authorized local preview inspection and read-only preview generation without launching AI work. Installation rotation SHALL never transfer an old preview's authority to a new instance.

#### Scenario: P17 Access previews with execution disabled
- **GIVEN** retained staged drafts and disabled provider execution
- **WHEN** the owner reads or generates a preview using current source access
- **THEN** the read-only workflow remains available without provider execution or Immich mutation

#### Scenario: P18 Isolate deletion and installation rotation
- **GIVEN** private previews for different owners and an in-flight request
- **WHEN** one owner is deleted or the installation changes
- **THEN** unrelated private data is preserved and obsolete references/late requests cannot expose or recreate usable preview authority
