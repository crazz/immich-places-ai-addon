## ADDED Requirements

### Requirement: Protect a draft revision while its GPS write can still act

Editing or rejecting a confirmed draft before its next mutation reservation SHALL invalidate queued or safely retryable approval atomically. Once dispatch is reserved and may act, edits/rejection SHALL wait for a settled outcome; the UI SHALL preserve unsaved client changes and explain the active write. Completing an older approved revision SHALL never mark a newer revision as saved. Original result inspection SHALL remain available throughout.

#### Scenario: R01 Edit before dispatch reservation
- **GIVEN** a confirmed queued or safely retryable draft whose next mutation has not been reserved
- **WHEN** its owner saves an edit or rejection
- **THEN** queued approval is invalidated before the new revision can be written and no stale dispatch follows

#### Scenario: R02 Edit while a write is unresolved
- **GIVEN** a draft revision with a reserved or potentially sent GPS operation
- **WHEN** the owner attempts to save edits or reject it
- **THEN** the saved revision remains fixed, unsaved edits remain available and the active operation must settle before another revision can be saved

### Requirement: Display verified GPS outcomes independently of proposal and review state

AI Results SHALL show explicit confirmation, durable progress, conflict, unresolved, failed, verified no-op and verified-write outcomes separately from analysis and draft review. It SHALL show the exact approved revision and private before/intended/observed values without claiming causal certainty from readback. Confirming with a known overlapping manual pending edit SHALL require explicit resolution. Lost submissions SHALL reconcile using the same identity, including on HTTP origins without secure-context UUID support. Accessible status and recovery SHALL work without map tiles, and account/result changes SHALL fence late private replies.

#### Scenario: R03 Retain results after a verified GPS save
- **GIVEN** an approved GPS operation reaches verified local completion
- **WHEN** the UI refreshes catalog counts and markers
- **THEN** the photo can leave Missing GPS while its immutable result, draft revision and write audit remain available in AI Results

#### Scenario: R04 Confirm accessibly without a secure browser origin
- **GIVEN** keyboard input, failed map tiles and an HTTP origin without secure-context UUID support
- **WHEN** the owner confirms the displayed exact GPS plan and its acknowledgement is uncertain
- **THEN** a stable submission identity permits reconciliation without a duplicate or inaccessible confirmation flow

#### Scenario: R05 Preserve manual work and private view boundaries
- **GIVEN** a same-photo manual pending edit or an in-flight private operation reply
- **WHEN** confirmation is attempted or the user switches account/result
- **THEN** the manual choice must be resolved without silent replacement and obsolete replies cannot populate the new private view

#### Scenario: R06 Keep unresolved and failed outcomes distinct
- **GIVEN** an operation has a conflict, readback failure, known failure or pending local refresh
- **WHEN** its owner inspects AI Results
- **THEN** the actual state and applicable recovery action are visible without a false saved label, discarded draft or automatic provider call
