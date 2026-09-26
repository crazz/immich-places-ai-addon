## ADDED Requirements

### Requirement: Select one reviewed standard description independently of GPS

An owner SHALL be able to select one complete, nonempty current primary-language description for the analyzed photo independently of GPS. Preserve SHALL be the default description policy and SHALL omit description from writes. Replace and managed append SHALL require deliberate selection. Stale dependent text SHALL require regeneration/adoption or explicit current review before selection. Other languages, direction and unrelated fields SHALL remain unselected. Description-only planning SHALL not require camera coordinates; selecting no writable field SHALL not produce a usable plan.

#### Scenario: C01 Save a description without camera coordinates
- **GIVEN** a current reviewed scene-only description and no camera point
- **WHEN** the owner stages and confirms only that description
- **THEN** the exact description can be written without GPS or invented geometry

#### Scenario: C02 Preserve text during GPS-only saving
- **GIVEN** a GPS decision with stale, failed or unavailable descriptions and existing Immich text
- **WHEN** only GPS is selected
- **THEN** description readiness does not block GPS and no description is sent or replaced

#### Scenario: C03 Reject unusable description selection
- **GIVEN** empty, failed, unavailable or stale unreviewed text, multiple primary languages or no selected writable field
- **WHEN** a write preview is requested
- **THEN** no usable plan is published and current draft values remain intact

### Requirement: Compare exact selected description baselines

Description preview and dispatch SHALL use fresh authorized source and description values, retaining absence separately from unavailable reads. Absent/null description SHALL compare as empty only after a valid complete metadata read; malformed or unavailable data SHALL not imply empty. Comparisons SHALL preserve Unicode content, whitespace and line breaks without trimming or normalization. An intervening selected-description change SHALL conflict before dispatch; unselected GPS changes SHALL not block a description-only plan. The full before/intended text, policy and language SHALL be bound to the displayed immutable plan and its five-minute validity.

#### Scenario: C04 Review empty Unicode and multiline baselines
- **GIVEN** valid metadata with absent, empty or nonempty multilingual text
- **WHEN** a description preview is created and reloaded
- **THEN** the exact before/intended text and presence state remain visible without normalization or a new write

#### Scenario: C05 Detect a changed description
- **GIVEN** a reviewed description baseline
- **WHEN** another client edits it before preview publication or dispatch
- **THEN** the changed field conflicts and cannot be silently overwritten

#### Scenario: C06 Distinguish unrelated changes from unavailable reads
- **GIVEN** a description-only plan
- **WHEN** only GPS changes or the description read is unavailable/malformed
- **THEN** an unrelated GPS change alone does not block it, but unavailable/malformed text never becomes a fabricated empty baseline

### Requirement: Update one owned managed append block without duplication

Managed append SHALL preserve text outside one explicitly owned block and show the complete resulting description before confirmation. Ownership SHALL require retained verified local lineage, not merely a marker in upstream text. Repeated append or a primary-language change SHALL update the same owned block. Missing, edited, duplicate, nested, malformed or unowned marker content SHALL produce a conflict rather than automatic repair or a second block. Observed/final description limits SHALL be 64 KiB of UTF-8; text SHALL not be silently truncated.

#### Scenario: C07 Append while preserving existing text
- **GIVEN** existing user text and no conflicting managed marker
- **WHEN** the owner confirms the displayed first managed append
- **THEN** the user text is preserved exactly and one identifiable block is added and verified

#### Scenario: C08 Repeat append or change primary language
- **GIVEN** a previously verified owned block and unchanged reviewed baseline
- **WHEN** a new approved description or primary language is appended
- **THEN** the same block is replaced once without duplicating it or altering surrounding text

#### Scenario: C09 Reject tampered or unowned blocks
- **GIVEN** an edited, missing, duplicate, nested, malformed or unowned managed marker
- **WHEN** managed append is previewed or dispatched
- **THEN** no repair, takeover or additional block is written and existing text is preserved

#### Scenario: C10 Reject oversized text without truncation
- **GIVEN** an observed or resulting description would exceed the declared bound
- **WHEN** replace or append is previewed
- **THEN** the request fails visibly without truncating the draft or upstream text

### Requirement: Reconcile selected standard fields independently

Each target standard step SHALL send at most one non-retrying request per reserved attempt containing exactly its selected GPS and/or description fields. Readback SHALL verify all selected fields before the step is fully successful. A mixed baseline/intended observation SHALL retain per-field evidence but SHALL not allow replay of the combined payload; a new plan for remaining fields requires known prior-sender completion and renewed review. Unknown completion, third values, no-op detection and explicit bounded retries SHALL retain the CH19 safeguards. Descriptions SHALL compare exactly under the stated absence rule; GPS SHALL retain its documented tolerance.

#### Scenario: C11 Verify a combined exact update
- **GIVEN** one photo with both GPS and description approved
- **WHEN** its reserved standard step executes
- **THEN** one request contains only those fields and success requires matching source-consistent readback for both

#### Scenario: C12 Recover a lost description response
- **GIVEN** approved text was applied but its acknowledgement or local persistence failed
- **WHEN** reconciliation reads the intended text
- **THEN** verified state is retained without sending the description again

#### Scenario: C13 Retain a mixed standard-field outcome
- **GIVEN** a combined request is followed by intended GPS but baseline description, or the reverse
- **WHEN** readback is evaluated
- **THEN** the result remains partial and neither successful field is resent through a combined retry
- **AND** a newly reviewed remaining-field plan cannot proceed while the earlier sender may still act

#### Scenario: C14 Preserve replay identity and attempt limits
- **GIVEN** a lost confirmation/retry acknowledgement, expired approval or ambiguous sender
- **WHEN** the owner repeats an accepted identity or requests another attempt
- **THEN** accepted identity recovery remains local while new attempts require fresh authority and the unchanged two-attempt ceiling

### Requirement: Preserve historical approvals across standard-field upgrades

An upgrade SHALL retain old GPS-only plan bytes, digests, audit, attempt counts and unresolved target exclusions. Old approvals SHALL never acquire description authority. New plans SHALL be explicitly versioned, bounded to 1 MiB and rejected if unknown, malformed or substituted. Overlapping old/new plans SHALL share target exclusion, including across accounts. New description writes SHALL remain disabled without profile-specific authorized compatibility evidence.

#### Scenario: C15 Reopen old approval and unknown sender state
- **GIVEN** retained queued, verified and ambiguous CH19 operations
- **WHEN** the application upgrades and reopens its database
- **THEN** the original GPS scope and recovery state are preserved and no description permission is inferred

#### Scenario: C16 Reject cross-version overlap or altered plans
- **GIVEN** an unresolved old target or an unknown, oversized or altered new plan
- **WHEN** confirmation or execution is attempted
- **THEN** no overlapping or substituted mutation is authorized and private values are not disclosed

#### Scenario: C17 Keep unverified description support disabled
- **GIVEN** only synthetic or live GPS evidence is available
- **WHEN** description dispatch is enabled without description-specific compatibility evidence
- **THEN** rollout is blocked while local review and supported history remain available

## MODIFIED Requirements

### Requirement: Preview only a staged exact GPS decision

An owner SHALL be able to request a before/after preview for the current staged revision of an owned draft with acknowledged source and selected-field baselines. GPS selection SHALL require a finite camera pair; description selection SHALL require the current reviewed primary-language text and explicit description policy. The plan SHALL target exactly the analyzed asset with GPS, description or both. Unauthorized client/model input SHALL not add targets, stack members or unsupported fields. Heading, confidence and estimated error SHALL not add GPS readiness gates. Preview operations SHALL make zero provider calls and zero Immich mutations and SHALL not grant write approval.

#### Scenario: P01 Preview one coarse camera decision
- **GIVEN** an owned staged draft with valid camera coordinates, acknowledged baseline and a large or unknown estimated error
- **WHEN** the owner requests a preview
- **THEN** an exact single-photo GPS comparison is returned without writing, provider work or a precision threshold

#### Scenario: P02 Reject an incomplete or expanded decision
- **GIVEN** an unstaged, rejected or baseline-unavailable draft, a GPS-selected draft with invalid camera coordinates, or a request adding unapproved targets or unsupported fields
- **WHEN** a preview is requested
- **THEN** no usable plan is created and the draft remains available for correction

### Requirement: Use fresh authorized source and GPS before-values

Preview creation SHALL read current authorized metadata and compare the reviewed image identity and acknowledged before-values for each selected field. GPS-selected plans SHALL retain exact nullable GPS comparison; description-selected plans SHALL follow the exact description baseline policy. Changed selected fields SHALL conflict rather than silently becoming approved baselines. Changed source image SHALL require renewed current-image review. Unavailable or malformed selected-field metadata SHALL produce no usable plan. Unrelated metadata changes SHALL not falsely imply image replacement. Fresh read access SHALL not be presented as verified write permission.

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

Every preview SHALL bind owner, installation, draft revision, reviewed source identity, exact target, selected fields, each before/intended value, description language/policy when selected and validity period. Its displayed comparison and digest SHALL identify the same immutable plan. Client modifications or cross-scope references SHALL not substitute a different plan. A draft or authority change during creation SHALL prevent publication of a usable preview. Already-matching selected values SHALL be shown as unchanged, not as evidence of a new write.

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

Preview UI SHALL present photo identity, exact before/after values for every selected field, single-photo scope, description language/policy where selected, revision, expiry and changed/unchanged status through keyboard-accessible text/numeric controls. Failed maps SHALL not hide comparisons. Conflicts and source failures SHALL offer review paths without silently changing a baseline. AI data SHALL not enter manual pending coordinates. Selected GPS overlapping manual pending edits SHALL require explicit resolution, and account/result changes SHALL remove private stale content.

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

Every selected-field mutation SHALL require explicit owner confirmation of the exact current unexpired preview and matching digest. Approval SHALL durably identify actor, installation, draft revision, source identity, exact asset, selected fields and before/intended values, description policy/language and approval time before any mutation can begin. The same submission key and plan SHALL resolve to one operation across duplicate requests, lost responses and restart; changed-plan key reuse and reuse of a consumed preview SHALL not create another operation. A draft, result, browser flag or digest alone SHALL not grant write authority.

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

Execution SHALL send exactly the selected finite camera latitude/longitude pair and/or approved description to the analyzed asset. It SHALL never expand stack scope, clear GPS using fabricated zeroes, substitute subject coordinates or include unselected description, direction, timestamps, rating, favorite or custom metadata. Numeric zero SHALL remain valid. Every potentially sent attempt SHALL have a durable reservation; transport/browser retries, redirects and automatic alternate mutation routes SHALL not cause hidden sends. Unsupported adapter/version/field configurations SHALL fail before mutation. Retained GPS-only approvals SHALL remain GPS-only.

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

Each mutation attempt SHALL recheck current owner/installation/access, reviewed image identity, draft revision, enablement and relevant fresh selected-field before-values. Known changes SHALL prevent dispatch and surface conflict or renewed review. Already-matching approved before/intended values for every selected field SHALL produce a verified no-op without a mutation. Unselected field changes alone SHALL not confer new authority or block an otherwise valid plan. A different client's changes between final read and write cannot be guaranteed atomic; the product SHALL not claim remote compare-and-swap protection.

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

A potentially sent standard request SHALL be followed by authorized selected-field readback before success or resend is decided. Matching all intended fields with unchanged source SHALL establish observed desired state; third values SHALL conflict and unavailable readback SHALL remain unresolved. Mixed intended/baseline fields SHALL retain partial evidence and SHALL not permit replay of the combined payload. An unchanged baseline alone SHALL not prove an earlier timed-out sender cannot still act. Retry SHALL require explicit owner action, unchanged complete fresh baseline, valid current approval/authority, established prior-sender completion and remaining attempt budget. There SHALL be at most two standard mutation attempts per confirmed single-target operation. Read-only status checks SHALL not reset that budget or create a mutation.

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

Success for a standard step SHALL require source-consistent readback of every selected field. GPS SHALL use the documented numerical tolerance, distinct from geographic accuracy; description SHALL use the exact text comparison policy. A successful HTTP response alone SHALL not count. Verified GPS SHALL be persisted before refreshing local markers and Missing GPS membership; description-only operations SHALL not change that membership. Mixed outcomes SHALL preserve independently verified fields without claiming complete success. Local persistence or refresh failure after remote success SHALL trigger reconciliation without resending. A delayed catalog sync SHALL not overwrite newly verified GPS with older data. Missing local assets SHALL not be fabricated.

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

Approval, before/intended/observed selected-field values, description policy/language, exact revision/target, attempts, outcomes and timestamps SHALL remain privately inspectable and survive restart and ordinary cleanup. Ordinary logs SHALL omit secrets and private payloads. Disablement/shutdown SHALL stop new sends while preserving ambiguous operations for read-only reconciliation. Account deletion SHALL remove only owned private records and prevent their late recreation; it SHALL not claim to undo an already-received Immich request. Installation rotation SHALL never reuse new-instance authority for an old operation. Each new field/scope capability SHALL remain disabled until its configured live adapter passes explicitly authorized fixture verification.

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
