## ADDED Requirements

### Requirement: Confirm one immutable plan durably and idempotently

A GPS mutation SHALL require explicit owner confirmation of the exact current unexpired preview and matching digest. The approval SHALL durably identify its actor, installation, draft revision, source identity, exact asset, fields, before/intended GPS and time before any mutation can begin. The same submission key and plan SHALL resolve to one operation across duplicate requests, lost responses and restart; changed-plan key reuse and reuse of a consumed preview SHALL not create another operation. A draft, result, browser flag or digest alone SHALL not grant write authority.

#### Scenario: W01 Confirm and recover the approved operation
- **GIVEN** a current owned unexpired GPS preview
- **WHEN** the owner confirms its exact plan and reloads after a restart
- **THEN** one durable operation identifies the approved revision and values before any mutation dispatch

#### Scenario: W02 Reconcile a lost confirmation response
- **GIVEN** confirmation was committed but its response was lost
- **WHEN** the client looks up or repeats the same key and plan
- **THEN** the same operation is returned without another approval or mutation reservation

#### Scenario: W03 Reject stale or substituted approval
- **GIVEN** an expired, edited, rejected, foreign or wrong-digest preview, a changed-plan reused key or a consumed preview
- **WHEN** confirmation is attempted
- **THEN** no different or duplicate operation is authorized and no mutation is sent

#### Scenario: W04 Fail before durable approval
- **GIVEN** approval storage fails or the process stops before approval commit
- **WHEN** confirmation cannot complete durably
- **THEN** no Immich mutation occurs

### Requirement: Mutate only the approved single-asset GPS pair

Execution SHALL send only the approved finite camera latitude/longitude to exactly the approved asset. It SHALL never expand stack scope, clear GPS using fabricated zeroes, substitute subject coordinates or include description, direction, timestamps, rating, favorite or custom metadata. Numeric zero SHALL remain valid. Every potentially sent attempt SHALL have a durable reservation; transport/browser retries, redirects and automatic alternate mutation routes SHALL not cause hidden sends. Unsupported adapter/version configurations SHALL fail before mutation.

#### Scenario: W05 Write exact GPS while preserving other fields
- **GIVEN** a confirmed plan for one photo with stack siblings and unrelated existing metadata
- **WHEN** the operation executes
- **THEN** only that photo's approved GPS pair is sent and siblings and unrelated fields remain unchanged

#### Scenario: W06 Preserve valid zero and omit unsupported fields
- **GIVEN** a confirmed finite GPS pair containing zero and a draft with local descriptions or heading
- **WHEN** the mutation payload is sent
- **THEN** the zero value is preserved as a coordinate and only the GPS pair is included

#### Scenario: W07 Prevent hidden retries or fallback
- **GIVEN** a mutation encounters a redirect, dropped connection, timeout, throttling or server failure
- **WHEN** the request returns or loses its response
- **THEN** no transport/browser layer resends it or tries another mutation route without the operation's reconciliation and explicit retry decision

#### Scenario: W08 Reject unsupported compatibility
- **GIVEN** mutation dispatch is disabled or the Immich adapter/version is unsupported
- **WHEN** new confirmation or mutation dispatch is attempted
- **THEN** no request mutates Immich and the local draft/history remains available

### Requirement: Recheck source baseline and current authority before dispatch

Each mutation attempt SHALL recheck current owner/installation/access, reviewed image identity, draft revision, enablement and relevant fresh GPS before-values. Known changes SHALL prevent dispatch and surface conflict or renewed review. Already-matching approved before/intended values SHALL produce a verified no-op without a mutation. A different client's changes between the final read and write cannot be guaranteed atomic; the product SHALL not claim remote compare-and-swap protection.

#### Scenario: W09 Detect fresh GPS or source conflict
- **GIVEN** a confirmed plan whose GPS baseline or source image changes before dispatch
- **WHEN** the writer performs its fresh checks
- **THEN** it records a conflict or renewed-review outcome without overwriting the changed values

#### Scenario: W10 Lose authority after confirmation
- **GIVEN** a queued or retry-eligible operation
- **WHEN** access is revoked, the asset becomes hidden/trashed, credentials change, approval expires or the owner/installation becomes obsolete before dispatch
- **THEN** the obsolete authority cannot send another mutation

#### Scenario: W11 Verify an unchanged plan
- **GIVEN** a confirmed preview whose before and intended GPS pairs are equal
- **WHEN** fresh verification still observes that pair and source
- **THEN** the operation records a verified no-op and sends no mutation

### Requirement: Serialize overlapping writes without treating lease expiry as permission

Only one unresolved AI write SHALL hold a given installation/asset target at a time, including across different addon accounts. Competing requests SHALL not disclose another owner's private plan. Worker expiry, restart or late completion SHALL not grant a second sender authority while an earlier attempt may still act. Unrelated targets SHALL remain eligible within bounded runtime concurrency.

#### Scenario: W12 Compete for the same target
- **GIVEN** two plans, possibly owned by different accounts, target the same photo
- **WHEN** they are confirmed or claimed concurrently
- **THEN** only one active target operation can proceed and the competing response exposes no other owner's private values

#### Scenario: W13 Recover a stale worker
- **GIVEN** an operation may have sent a mutation and its worker expires or the process restarts
- **WHEN** recovery or a late worker runs
- **THEN** recovery verifies the existing operation first and no expired worker can overwrite newer local state or authorize a concurrent resend

### Requirement: Reconcile ambiguous outcomes before any bounded explicit retry

A potentially sent request SHALL be followed by authorized readback before success or resend is decided. Matching intended GPS with unchanged source SHALL establish observed desired state; GPS matching neither intended nor baseline SHALL produce a conflict, and unavailable readback SHALL remain unresolved. An unchanged baseline alone SHALL not prove an earlier timed-out request cannot still act. Retry SHALL require explicit owner action, unchanged fresh baseline, valid current approval/authority, established prior-sender completion and remaining attempt budget. There SHALL be at most two mutation attempts per confirmed operation. Read-only status checks SHALL not reset that budget or create a mutation.

Repeating an already-recorded retry generation SHALL return the current private operation without another upstream read or attempt allocation, including after successful execution, restart, approval expiry or write disablement. Owner and installation scope SHALL remain mandatory. A generation that has not been accepted SHALL still require all fresh retry eligibility checks before allocating an attempt.

#### Scenario: W14 Recover remote success with lost response
- **GIVEN** Immich applied the approved pair but the mutation response was lost
- **WHEN** readback observes the intended pair and unchanged source
- **THEN** the operation records verified desired state without resending

#### Scenario: W15 Keep unresolved outcomes honest
- **GIVEN** readback fails or GPS remains at baseline while an earlier possibly accepted request cannot be proven complete
- **WHEN** reconciliation runs, including an observation of unchanged baseline
- **THEN** the operation remains unresolved rather than claiming success or allowing a blind resend

#### Scenario: W16 Retry a known completed unsuccessful attempt
- **GIVEN** prior sender completion is established, fresh GPS still equals baseline and approval/authority and the second-attempt allowance remain valid
- **WHEN** the owner explicitly retries
- **THEN** only one next attempt of the exact approved GPS step can be reserved

#### Scenario: W17 Reject exhausted stale or concurrent retries
- **GIVEN** exhausted mutation allowance, expired/changed approval, unresolved sender completion or concurrent retry requests
- **WHEN** retry is requested
- **THEN** no extra or unauthorized mutation is sent and the durable attempt count is not reset

#### Scenario: W18 Detect a different readback value
- **GIVEN** a potentially sent request followed by changed source or GPS matching neither approved before nor intended values
- **WHEN** readback completes
- **THEN** conflict is recorded with owned observed values and no automatic overwrite occurs

#### Scenario: W27 Recover a retry after its successful execution
- **GIVEN** a retry for an inspected generation was accepted and its second mutation succeeded
- **WHEN** the owner repeats that accepted generation after losing its acknowledgement
- **THEN** the current operation is returned without an upstream read or an additional mutation attempt

#### Scenario: W28 Recover an accepted retry after restart and disablement
- **GIVEN** an accepted retry is retained across restart, approval expiry and write disablement
- **WHEN** its owner repeats the accepted generation while Immich is unavailable
- **THEN** the current private operation is returned locally with its original attempt count
- **AND** foreign owners, obsolete installations and unaccepted generations receive no duplicate-recovery authority

### Requirement: Verify GPS before local success publication

Success SHALL require source-consistent GPS readback using a documented numerical tolerance, distinct from geographic accuracy. A successful HTTP response alone SHALL not count. Verified GPS SHALL be persisted before refreshing local markers and Missing GPS membership. Local persistence or refresh failure after remote success SHALL trigger reconciliation without resending. A delayed catalog synchronization SHALL not overwrite the newly verified local pair with older data. Missing local assets SHALL not be fabricated to make refresh look successful.

#### Scenario: W19 Observe rounded GPS or mismatched success
- **GIVEN** an upstream success response with GPS readback rounded within the documented tolerance or outside it
- **WHEN** verification evaluates that readback
- **THEN** only a matching pair qualifies as observed desired state, with no claim of measured geographic accuracy or sidecar completion

#### Scenario: W20 Recover local failure after remote success
- **GIVEN** intended GPS was observed but local outcome/catalog persistence fails
- **WHEN** the operation is reconciled after recovery
- **THEN** fresh readback precedes local publication and no GPS mutation is resent to repair local storage

#### Scenario: W21 Preserve verified GPS against an older sync
- **GIVEN** catalog synchronization started before the GPS mutation
- **WHEN** its older response arrives after verified GPS publication
- **THEN** it cannot immediately replace the verified local pair with the stale pair

#### Scenario: W22 Handle an unavailable local catalog row
- **GIVEN** intended upstream GPS is verified but the local photo row is missing
- **WHEN** the writer publishes its outcome
- **THEN** upstream verification and unavailable/pending local refresh are reported separately without inventing a local asset or resending

### Requirement: Retain private audit and respect write lifecycle controls

Approval, before/intended/observed GPS, exact revision/target, attempts, outcomes and timestamps SHALL remain privately inspectable and survive restart and ordinary cleanup. Ordinary logs SHALL omit secrets and private payloads. Disablement/shutdown SHALL stop new sends while preserving ambiguous operations for read-only reconciliation. Account deletion SHALL remove only owned private records and prevent their late recreation; it SHALL not claim to undo an already-received Immich request. Installation rotation SHALL never reuse new-instance authority for an old operation. New writes SHALL remain disabled until the configured live adapter passes explicitly authorized disposable-fixture verification.

#### Scenario: W23 Disable while work is active
- **GIVEN** queued and potentially sent operations
- **WHEN** global/write enablement is disabled or shutdown begins
- **THEN** no new mutation is dispatched and potentially sent work remains available for read-only reconciliation rather than false cancellation/success

#### Scenario: W24 Preserve audit through restart and cleanup
- **GIVEN** confirmed or completed operations with referenced previews and drafts
- **WHEN** the backend restarts, catalog data resets or ordinary cleanup runs
- **THEN** exact private approval/outcome history remains available without new mutations or private log output

#### Scenario: W25 Delete or rotate during an operation
- **GIVEN** owned operations, another account's data and an in-flight sender
- **WHEN** the owner is deleted or the installation changes
- **THEN** obsolete authority cannot dispatch or recreate private data, unrelated owners remain intact and already-sent requests are not falsely described as rolled back

#### Scenario: W26 Keep unverified live writes disabled
- **GIVEN** deterministic tests pass but the configured live adapter lacks authorized disposable-fixture evidence
- **WHEN** the release is prepared
- **THEN** live compatibility remains unverified and real mutation dispatch stays disabled
- **AND** analysis-only photo authorization is not reused as write-test authorization
