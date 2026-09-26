## ADDED Requirements

### Requirement: Review optional metadata visibility and partial success

AI review SHALL provide keyboard-operable default-off mirror selection, the exact chosen contents and asset-reader visibility disclosure. It SHALL show standard and optional step outcomes independently, preserve all local translations and reviewed direction when the mirror is unsupported/failed, and identify that a custom metadata mirror does not promise native Immich UI display or file/EXIF updates. Retrying an incomplete step SHALL be explicit, and account/result changes SHALL fence private replies.

#### Scenario: M19 Inspect disclosure and exact export accessibly
- **GIVEN** a narrow viewport, keyboard input and unavailable map
- **WHEN** the owner chooses optional mirroring
- **THEN** the exact exported content, visibility disclosure, target and step order are accessible before confirmation

#### Scenario: M20 Review a partial result without losing local data
- **GIVEN** successful standard fields and an unsupported, failed or unresolved optional mirror
- **WHEN** the owner reloads AI Results
- **THEN** both outcomes and the permitted recovery action are visible while local language/direction records remain available

#### Scenario: M21 Fence stale private mirror responses
- **GIVEN** an in-flight mirror status or confirmation reply
- **WHEN** the account/result changes or newer operation history becomes current
- **THEN** the old reply cannot leak private content, overwrite the newer outcome or discard its unresolved identity

## MODIFIED Requirements

### Requirement: Edit camera decisions without inventing precision or writable fields

Drafts SHALL allow explicit candidate selection or user-supplied camera coordinates, optional local heading and per-language description edits. Unknown or subject-only proposals SHALL start without camera geometry. Coordinates SHALL be finite and in range, with zero valid and missing values distinct. Heading SHALL be null or within zero inclusive to 360 exclusive. Radius, confidence and unrelated field availability SHALL not impose a GPS quality threshold. Staging SHALL require at least one supported explicitly selected field and current field-specific readiness. GPS SHALL require a finite camera pair; a description-only decision SHALL not. The default target SHALL remain the analyzed asset; an additional stack GPS target SHALL require explicit independent review under the target-manifest contract. Description and direction SHALL remain per-image and SHALL not propagate to siblings. Primary-language and description-policy choices SHALL be revisioned and no client/model input SHALL expand supported scope. Supported optional mirror selection SHALL be separately revisioned and require at least one selected standard field; stale/unreviewed export content SHALL not gain authority through GPS selection.

#### Scenario: D05 Choose a coarse or ambiguous proposal
- **GIVEN** alternatives with low confidence, a 500-meter or larger estimated radius, or unknown uncertainty
- **WHEN** the owner explicitly chooses a camera point and stages GPS
- **THEN** that point is retained as a local decision without a precision threshold or automatic selection of a different candidate

#### Scenario: D06 Supply an absent camera point
- **GIVEN** an unknown or subject-only proposal
- **WHEN** the owner opens a draft and later enters a valid camera pair
- **THEN** the camera remains absent until that explicit entry and the subject is never silently substituted

#### Scenario: D07 Validate geometry and field scope
- **GIVEN** an editable draft
- **WHEN** inputs contain zero coordinates, invalid or partial coordinates, an out-of-range heading, an unapproved asset or an unsupported write field
- **THEN** valid zero values are preserved and invalid edits or expanded write scope are rejected without changing the saved revision

### Requirement: Protect a draft revision while its GPS write can still act

Editing or rejecting a confirmed draft before its next mutation reservation SHALL invalidate queued or safely retryable approval atomically. Once any selected standard or optional step dispatch on any approved target is reserved and may act, edits/rejection SHALL wait for a settled outcome; the UI SHALL preserve unsaved client changes and explain the active write. Completing an older approved revision SHALL never mark a newer revision as saved. Original result inspection SHALL remain available throughout.

#### Scenario: R01 Edit before dispatch reservation
- **GIVEN** a confirmed queued or safely retryable draft whose next mutation has not been reserved
- **WHEN** its owner saves an edit or rejection
- **THEN** queued approval is invalidated before the new revision can be written and no stale dispatch follows

#### Scenario: R02 Edit while a write is unresolved
- **GIVEN** a draft revision with a reserved or potentially sent GPS operation
- **WHEN** the owner attempts to save edits or reject it
- **THEN** the saved revision remains fixed, unsaved edits remain available and the active operation must settle before another revision can be saved
