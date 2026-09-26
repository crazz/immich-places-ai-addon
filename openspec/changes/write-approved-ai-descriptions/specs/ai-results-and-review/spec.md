## ADDED Requirements

### Requirement: Review independent standard-field choices and outcomes

The owner SHALL select GPS and one primary-language description independently, inspect exact full text before/after values and choose preserve, replace or managed append through accessible controls. Field and policy edits SHALL create a new revision and invalidate old previews without writing. Outcomes SHALL distinguish each selected field, overall partial state and local refresh; description-only writes SHALL not change Missing GPS membership or consume manual pending GPS. Overlapping selected GPS SHALL still require explicit manual-choice resolution, and late private replies SHALL remain fenced.

#### Scenario: C18 Review and confirm text with keyboard input
- **GIVEN** a narrow viewport, unavailable map and a current language description
- **WHEN** a keyboard user chooses the description policy and opens confirmation
- **THEN** the full text comparison, selected fields, language, revision and expiry are accessible before any mutation

#### Scenario: C19 Preserve manual GPS during a text-only operation
- **GIVEN** unsaved manual GPS for the same photo
- **WHEN** a description-only operation completes
- **THEN** the pending GPS and gallery membership are unchanged and no manual save occurs

#### Scenario: C20 Keep field outcomes and newer review separate
- **GIVEN** a partial standard-field outcome, an uncertain identity or a newer edited draft
- **WHEN** history reloads or an older reply arrives
- **THEN** exact field outcomes remain visible without marking the newer revision saved or discarding unresolved authority

## MODIFIED Requirements

### Requirement: Edit camera decisions without inventing precision or writable fields

Drafts SHALL allow explicit candidate selection or user-supplied camera coordinates, optional local heading and per-language description edits. Unknown or subject-only proposals SHALL start without camera geometry. Coordinates SHALL be finite and in range, with zero valid and missing values distinct. Heading SHALL be null or within zero inclusive to 360 exclusive. Radius, confidence and unrelated field availability SHALL not impose a GPS quality threshold. Staging SHALL require at least one supported explicitly selected field and current field-specific readiness. GPS SHALL require a finite camera pair; a description-only decision SHALL not. Targets SHALL remain the analyzed asset only. Primary-language and description-policy choices SHALL be revisioned and no client/model input SHALL expand supported scope.

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

Editing or rejecting a confirmed draft before its next mutation reservation SHALL invalidate queued or safely retryable approval atomically. Once any selected-field dispatch is reserved and may act, edits/rejection SHALL wait for a settled outcome; the UI SHALL preserve unsaved client changes and explain the active write. Completing an older approved revision SHALL never mark a newer revision as saved. Original result inspection SHALL remain available throughout.

#### Scenario: R01 Edit before dispatch reservation
- **GIVEN** a confirmed queued or safely retryable draft whose next mutation has not been reserved
- **WHEN** its owner saves an edit or rejection
- **THEN** queued approval is invalidated before the new revision can be written and no stale dispatch follows

#### Scenario: R02 Edit while a write is unresolved
- **GIVEN** a draft revision with a reserved or potentially sent GPS operation
- **WHEN** the owner attempts to save edits or reject it
- **THEN** the saved revision remains fixed, unsaved edits remain available and the active operation must settle before another revision can be saved
