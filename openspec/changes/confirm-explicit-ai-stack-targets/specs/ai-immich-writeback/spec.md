## ADDED Requirements

### Requirement: Freeze only explicitly reviewed stack GPS targets

The default target SHALL remain the analyzed photo alone. Additional GPS targets SHALL require GPS selected with a valid camera pair on the analyzed photo, deliberate selection from its current stack and independent authorized image/source/baseline review. A plan SHALL contain at most 50 distinct targets including the analyzed photo; duplicate input SHALL not create repeated targets and over-limit input SHALL fail without truncation. Hidden, trashed, inaccessible, non-image, foreign-stack or unreadable selected members SHALL not enter a usable plan. Later additions SHALL never expand the approved target list.

#### Scenario: S01 Select an exact subset
- **GIVEN** an accessible five-photo stack with independently observed GPS baselines
- **WHEN** the owner selects the analyzed photo and two reviewed siblings
- **THEN** only those three IDs and their before/intended values enter the immutable plan

#### Scenario: S02 Retain the single-photo default
- **GIVEN** an analyzed photo in a stack
- **WHEN** the owner does not deliberately add any member
- **THEN** only the analyzed photo can be confirmed and later stack additions remain excluded

#### Scenario: S03 Reject inaccessible or incomplete scope
- **GIVEN** a selected member is hidden, trashed, inaccessible, non-image, outside the stack or cannot be read within bounds
- **WHEN** a target preview is created
- **THEN** no usable partial or expanded plan is substituted and no private inaccessible metadata is disclosed

#### Scenario: S04 Bound and deduplicate target selection
- **GIVEN** repeated IDs or more than 50 distinct selected targets
- **WHEN** the manifest is prepared
- **THEN** repeated IDs cannot cause repeated writes and an oversized selection is rejected without silent truncation

#### Scenario: S20 Reject stack scope without a GPS decision
- **GIVEN** a description-only draft or a missing/invalid camera pair
- **WHEN** additional stack GPS targets are requested
- **THEN** no usable expanded plan or fabricated GPS pair is created

### Requirement: Bind independent target baselines and exact field scope

Each selected target SHALL bind its own reviewed source identity, GPS baseline and exact fields to the same immutable approval. Siblings SHALL receive only the approved GPS pair; description and direction SHALL not propagate. A changed source, selected GPS or lost authority SHALL stop the affected target before dispatch. A selected sibling no longer in the reviewed stack SHALL conflict; unselected membership changes alone SHALL not alter scope or approved values. No stack relation SHALL substitute for access or viewpoint review.

#### Scenario: S05 Preserve sibling descriptions and unrelated fields
- **GIVEN** sibling GPS selections and a description selected only for the analyzed photo
- **WHEN** confirmed target steps execute
- **THEN** siblings receive only the approved GPS pair and their descriptions, direction and other fields remain unchanged

#### Scenario: S06 Observe changed membership
- **GIVEN** a frozen stack target manifest
- **WHEN** another member is added or a selected sibling leaves before its dispatch
- **THEN** the new member is never included and the departed selected sibling conflicts without rewriting the saved manifest

#### Scenario: S07 Stop a changed target independently
- **GIVEN** one selected member's GPS, source or authority changes after confirmation
- **WHEN** that member is checked before dispatch
- **THEN** it is not overwritten and other still-valid target outcomes remain independently actionable

### Requirement: Reserve and recover each exact target without overlap

Confirmation SHALL durably approve the complete target list and obtain all required unresolved-target exclusions atomically or obtain none. Every target SHALL retain its own attempts, generation and sender-completion evidence. The limit SHALL be two mutation attempts per target standard step, with no automatic resend and no reset on restart. Old and new plan versions/accounts SHALL compete for the same target exclusion without private disclosure. Read-only reconciliation and accepted retry-generation replay SHALL never resend completed siblings.

#### Scenario: S08 Reject an overlapping target atomically
- **GIVEN** one selected photo is held by another unresolved operation, possibly another account or old plan version
- **WHEN** the multi-target plan is confirmed
- **THEN** no partial approval or subset of new target reservations survives and foreign private values remain hidden

#### Scenario: S09 Recover independent target outcomes
- **GIVEN** one successful target, one ambiguous sender and one undispatched target at restart
- **WHEN** recovery runs
- **THEN** the successful target is not resent, ambiguous work is reconciled read-only and only valid unexpired undispatched authority can proceed

#### Scenario: S10 Retry only an eligible exact target
- **GIVEN** one known-complete unsuccessful target and successful siblings
- **WHEN** its owner explicitly retries the eligible target or repeats the accepted generation
- **THEN** only that target can reserve its remaining attempt and accepted replay returns stored state without another send

### Requirement: Preserve partial completion through expiry and lifecycle changes

Multi-target writes SHALL expose per-target verified, no-op, conflict, failed, expired and unresolved states without claiming batch atomicity. Expiry SHALL block unstarted attempts but not discard earlier outcomes or prevent read-only reconciliation. A possibly acting sender on any member SHALL protect the draft revision. Safe edits, disablement, shutdown, deletion and installation changes SHALL prevent new obsolete sends while preserving needed unresolved exclusion and retained authorized audit. Only verified GPS targets SHALL refresh the local catalog.

#### Scenario: S11 Finish partially at approval expiry
- **GIVEN** some targets completed while other targets have not reserved a send
- **WHEN** the five-minute approval expires
- **THEN** remaining new sends stop and completed or unresolved outcomes remain visible with no implicit renewed approval

#### Scenario: S12 Edit or disable while a member can act
- **GIVEN** one possibly acting target and other pending targets
- **WHEN** an edit, rejection, disablement or shutdown occurs
- **THEN** no obsolete pending send is introduced, active revision protection remains and prior successes are not described as rolled back

#### Scenario: S13 Delete or rotate with unresolved targets
- **GIVEN** multi-target work and another owner's retained records
- **WHEN** the owner is deleted or the installation rotates
- **THEN** late work cannot recreate private records, required opaque unresolved exclusions persist and unrelated owners remain intact

#### Scenario: S14 Refresh only verified targets
- **GIVEN** successful, conflicting and unavailable catalog targets
- **WHEN** results are published or an older sync completes
- **THEN** only source-consistent verified GPS is refreshed, missing rows are not fabricated and retained result history survives

### Requirement: Preserve earlier approvals when adding stack support

Upgrade SHALL retain single-target plan bytes, attempts, audit and unresolved guards, and SHALL never expand an older approval into a stack. New multi-target plans SHALL be explicitly versioned and bounded; unknown or substituted scope SHALL fail closed. Stack live compatibility SHALL require exact-target authorized evidence beyond the single-photo GPS gate.

#### Scenario: S15 Upgrade while an old write is unresolved
- **GIVEN** retained single-photo approval and sender state
- **WHEN** stack support is installed and the database reopened
- **THEN** that approval remains single-photo and excludes overlapping new targets until safe settlement

#### Scenario: S16 Keep unverified stack writes unavailable
- **GIVEN** synthetic stack checks or only single-photo live evidence
- **WHEN** multi-target production rollout is considered
- **THEN** stack compatibility remains unverified until separately authorized member-specific mutation/readback evidence exists

## MODIFIED Requirements

### Requirement: Preview only a staged exact GPS decision

An owner SHALL be able to request a before/after preview for the current staged revision of an owned draft with acknowledged source and selected-field baselines. GPS selection SHALL require a finite camera pair; description selection SHALL require the current reviewed primary-language text and explicit description policy. The default plan SHALL target exactly the analyzed asset with GPS, description or both. Additional stack GPS targets SHALL require the explicit independently reviewed target-manifest contract; siblings SHALL carry GPS only. Unauthorized client/model input SHALL not add targets or unsupported fields. Heading, confidence and estimated error SHALL not add GPS readiness gates. Preview operations SHALL make zero provider calls and zero Immich mutations and SHALL not grant write approval.

#### Scenario: P01 Preview one coarse camera decision
- **GIVEN** an owned staged draft with valid camera coordinates, acknowledged baseline and a large or unknown estimated error
- **WHEN** the owner requests a preview
- **THEN** an exact single-photo GPS comparison is returned without writing, provider work or a precision threshold

#### Scenario: P02 Reject an incomplete or expanded decision
- **GIVEN** an unstaged, rejected or baseline-unavailable draft, a GPS-selected draft with invalid camera coordinates, or a request adding unapproved targets or unsupported fields
- **WHEN** a preview is requested
- **THEN** no usable plan is created and the draft remains available for correction

### Requirement: Use fresh authorized source and GPS before-values

Preview creation SHALL read current authorized metadata and compare the independently reviewed image identity and acknowledged before-values for each selected field of every target. GPS-selected plans SHALL retain exact nullable GPS comparison; description-selected plans SHALL follow the exact description baseline policy. Changed selected fields SHALL conflict rather than silently becoming approved baselines. Changed source image SHALL require renewed current-image review. Unavailable or malformed selected-field metadata SHALL produce no usable plan. Unrelated metadata changes SHALL not falsely imply image replacement. Fresh read access SHALL not be presented as verified write permission.

#### Scenario: P03 Observe missing partial and zero GPS
- **GIVEN** an accessible asset whose current GPS is absent, partial or zero-valued and matches the acknowledged baseline
- **WHEN** a preview is created
- **THEN** the exact before-values remain distinguishable and no missing value is fabricated as zero

#### Scenario: P04 Detect an intervening GPS change
- **GIVEN** selected GPS changed since the draft's baseline was acknowledged
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

Every preview SHALL bind owner, installation, draft revision, each reviewed source identity, the frozen exact target list, per-target selected fields and before/intended values, description language/policy when selected and validity period. Its displayed comparison and digest SHALL identify the same immutable plan. Client modifications or cross-scope references SHALL not substitute a different plan. A draft or authority change during creation SHALL prevent publication of a usable preview. Already-matching selected values SHALL be shown as unchanged, not as evidence of a new write.

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

### Requirement: Show a usable accessible comparison without manual-save side effects

Preview UI SHALL present every selected photo identity, exact per-target before/after values and the complete frozen target scope, description language/policy where selected, revision, expiry and changed/unchanged status through keyboard-accessible text/numeric controls. Failed maps SHALL not hide comparisons. Conflicts and source failures SHALL offer review paths without silently changing a baseline. AI data SHALL not enter manual pending coordinates. Selected GPS overlapping manual pending edits SHALL require explicit resolution, and account/result changes SHALL remove private stale content.

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

### Requirement: Confirm one immutable plan durably and idempotently

Every selected-field mutation SHALL require explicit owner confirmation of the exact current unexpired preview and matching digest. Approval SHALL durably identify actor, installation, draft revision, each source identity, exact target list, per-target selected fields and before/intended values, description policy/language and approval time before any mutation can begin. The same submission key and plan SHALL resolve to one operation across duplicate requests, lost responses and restart; changed-plan key reuse and reuse of a consumed preview SHALL not create another operation. A draft, result, browser flag or digest alone SHALL not grant write authority.

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

Execution SHALL send exactly the approved GPS pair and/or description fields for each immutable target. The analyzed photo alone may include its selected description; independently reviewed stack siblings SHALL receive GPS only. It SHALL never discover additional targets during execution, clear GPS using fabricated zeroes, substitute subject coordinates or include unselected description, direction, timestamps, rating, favorite or custom metadata. Numeric zero SHALL remain valid. Each potentially sent target attempt SHALL have its own durable reservation; transport/browser retries, redirects and automatic alternate mutation routes SHALL not cause hidden sends. Unsupported adapter/version/field configurations SHALL fail before mutation. Retained single-asset approvals SHALL remain single-asset.

#### Scenario: W05 Write exact GPS while preserving other fields
- **GIVEN** a confirmed plan for one photo with stack siblings and unrelated existing metadata
- **WHEN** the operation executes
- **THEN** only that photo's approved GPS pair is sent and siblings and unrelated fields remain unchanged

#### Scenario: W06 Preserve valid zero and omit unsupported fields
- **GIVEN** a confirmed GPS-only finite pair containing zero and a draft with unselected local descriptions or heading
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

Each target mutation attempt SHALL recheck current owner/installation/access, the analyzed source and target-specific reviewed image identity, draft revision, enablement and relevant fresh selected-field before-values. Known changes SHALL prevent dispatch and surface conflict or renewed review. Already-matching approved before/intended values for every selected field SHALL produce a verified no-op without a mutation. Unselected field changes alone SHALL not confer new authority or block an otherwise valid plan. A different client's changes between final read and write cannot be guaranteed atomic; the product SHALL not claim remote compare-and-swap protection. Selected siblings SHALL still belong to the reviewed stack before dispatch; later unselected membership changes SHALL not expand the manifest.

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

### Requirement: Reconcile ambiguous outcomes before any bounded explicit retry

A potentially sent standard request SHALL be followed by authorized selected-field readback before success or resend is decided. Matching all intended fields with unchanged source SHALL establish observed desired state; third values SHALL conflict and unavailable readback SHALL remain unresolved. Mixed intended/baseline fields SHALL retain partial evidence and SHALL not permit replay of the combined payload. An unchanged baseline alone SHALL not prove an earlier timed-out sender cannot still act. Retry SHALL require explicit owner action, unchanged complete fresh baseline, valid current approval/authority, established prior-sender completion and remaining attempt budget. There SHALL be at most two standard mutation attempts per approved target standard step. Read-only status checks SHALL not reset that budget or create a mutation.

Repeating an already-recorded retry generation SHALL return the current private target operation without another upstream read or attempt allocation, including after successful execution, restart, approval expiry or write disablement. Owner and installation scope SHALL remain mandatory. A generation that has not been accepted SHALL still require all fresh retry eligibility checks before allocating an attempt. Retry/reconcile SHALL identify the exact target and generation. Completed targets SHALL not be resent to recover another member.

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

Success for a standard step SHALL require source-consistent readback of every selected field. GPS SHALL use the documented numerical tolerance, distinct from geographic accuracy; description SHALL use the exact text comparison policy. A successful HTTP response alone SHALL not count. Verified GPS SHALL be persisted before refreshing local markers and Missing GPS membership; description-only operations SHALL not change that membership. Mixed outcomes SHALL preserve independently verified fields without claiming complete success. Local persistence or refresh failure after remote success SHALL trigger reconciliation without resending. A delayed catalog sync SHALL not overwrite newly verified GPS with older data. Missing local assets SHALL not be fabricated. Each target SHALL retain an independent verified/refresh outcome; overall partial completion SHALL not fabricate all-target success.

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

Approval, before/intended/observed selected-field values, description policy/language, exact revision/target manifest, per-target attempts, outcomes and timestamps SHALL remain privately inspectable and survive restart and ordinary cleanup. Ordinary logs SHALL omit secrets and private payloads. Disablement/shutdown SHALL stop new sends while preserving ambiguous operations for read-only reconciliation. Account deletion SHALL remove only owned private records and prevent their late recreation; it SHALL not claim to undo an already-received Immich request. Installation rotation SHALL never reuse new-instance authority for an old operation. Each new field/scope capability SHALL remain disabled until its configured live adapter passes explicitly authorized fixture verification.

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
