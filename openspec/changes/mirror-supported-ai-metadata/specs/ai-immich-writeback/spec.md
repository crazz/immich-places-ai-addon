## ADDED Requirements

### Requirement: Gate optional metadata independently of standard writing

Metadata mirroring SHALL be default-off, separately supported for the current installation/version/permission profile and explicitly selected. Read-only capability checks SHALL never issue a trial mutation or imply write compatibility. Unsupported or unverified metadata SHALL not block local review or supported standard-only plans. A new combined plan SHALL include at least one selected standard field and mirror only the analyzed photo; standard no-op verification is allowed before the selected mirror.

#### Scenario: M01 Keep unsupported mirroring optional
- **GIVEN** an installation without verified metadata support
- **WHEN** the owner reviews a draft or prepares a standard-only write
- **THEN** local direction/translations remain available and no metadata mutation or false support claim occurs

#### Scenario: M02 Require deliberate supported selection
- **GIVEN** verified metadata support and a current draft
- **WHEN** the owner explicitly selects mirroring alongside a standard-field decision
- **THEN** only the analyzed photo's exact mirror can enter the plan

#### Scenario: M03 Reject implicit or mirror-only admission
- **GIVEN** a saved provider profile, an opened review, an unverified profile or a new plan selecting only metadata
- **WHEN** write admission is considered
- **THEN** no implicit probe or new mirror-only operation is authorized

### Requirement: Preview only selected current shareable metadata

The owner SHALL see and approve the exact exported direction, precision, place text, selected translations and minimal provenance, with a disclosure that authorized asset readers may see it. Only current reviewed values SHALL be eligible; unknown heading SHALL not become a fabricated bearing. Export SHALL omit secrets, raw prompts/responses, private hints, neighboring-photo locations and unselected content. It SHALL be bounded to eight selected languages and 64 KiB of structured text without silent truncation. Facts, field choices or language edits SHALL invalidate prior mirror approval.

#### Scenario: M04 Preview minimized reviewed content
- **GIVEN** a draft with current and stale fields, private context and multiple languages
- **WHEN** the owner selects a reviewed subset for mirroring
- **THEN** only that current subset and disclosed minimal provenance appear in the export comparison
- **AND** private context, stale unreviewed values and unselected languages are absent

#### Scenario: M05 Reject stale or oversized export
- **GIVEN** stale selected values, unsupported data, more than eight languages or a payload exceeding 64 KiB
- **WHEN** the combined preview is requested
- **THEN** no usable mirror plan is produced and local records are neither truncated nor discarded

#### Scenario: M06 Invalidate changed metadata selection
- **GIVEN** a current mirror preview
- **WHEN** facts, direction review, language text or selected export fields change
- **THEN** the old preview cannot authorize the changed export

### Requirement: Preserve unrelated and externally changed metadata

A mirror SHALL write only the `immich-places-ai-addon` namespace for the approved analyzed asset. A valid complete metadata read without that key SHALL mean absent; unavailable, malformed, duplicate or truncated data SHALL not. Existing namespace replacement SHALL require recognized verified local ownership and an unchanged fresh before-value. Foreign, incompatible or externally edited values SHALL conflict; unrelated keys SHALL be preserved. Semantic comparison SHALL ignore object key order but preserve array order and string contents.

#### Scenario: M07 Create or replace only the owned namespace
- **GIVEN** an absent namespace or unchanged recognized owned mirror and unrelated custom metadata
- **WHEN** the exact approved mirror step executes
- **THEN** only that namespace is added/replaced and unrelated keys and sibling metadata remain unchanged

#### Scenario: M08 Reject unavailable or conflicting namespace
- **GIVEN** a failed/malformed metadata read or a foreign, edited or incompatible value under the namespace
- **WHEN** preview or dispatch checks run
- **THEN** no empty baseline is fabricated and no automatic takeover or merge occurs

#### Scenario: M09 Compare semantic readback honestly
- **GIVEN** readback with only reordered object keys or with changed array order/text
- **WHEN** the mirror is verified
- **THEN** key order alone does not fail comparison while changed approved content cannot be reported as verified

### Requirement: Execute and recover the mirror as its own approved step

The analyzed photo's standard step SHALL verify before its optional mirror can dispatch. The mirror SHALL retain its own durable reservation, attempt count, generation, before/intended/observed values and sender-completion evidence under the same exact approval. Each step SHALL allow at most two explicit mutation attempts, without hidden retries. Reconciliation SHALL read before deciding success or a retry; readback failure or unknown sender completion SHALL remain unresolved. Only an incomplete eligible mirror step SHALL be retried, never successful standard fields. Accepted-generation replay SHALL resolve locally without new reads or sends.

#### Scenario: M10 Preserve standard success when mirror fails
- **GIVEN** verified GPS/description followed by a failed optional metadata attempt
- **WHEN** the result is reported
- **THEN** standard success and local refresh remain recorded alongside an incomplete mirror, with no rollback or standard resend

#### Scenario: M11 Recover a lost mirror acknowledgement
- **GIVEN** the mirror was applied but its acknowledgement or local outcome storage failed
- **WHEN** authorized readback observes the intended namespace and unchanged image
- **THEN** the mirror becomes verified without repeating either mutation step

#### Scenario: M12 Retry only a known eligible mirror
- **GIVEN** a known-complete unsuccessful mirror, unchanged baseline and current unexpired authority with remaining budget
- **WHEN** the owner explicitly retries or repeats its accepted generation
- **THEN** only one remaining mirror attempt can be reserved and duplicate recovery returns the stored operation locally

#### Scenario: M13 Keep unknown or expired work read-only
- **GIVEN** an ambiguous sender, unavailable readback, exhausted budget or expired approval
- **WHEN** mirror recovery or retry is requested
- **THEN** no unauthorized new mutation occurs and the actual incomplete state remains visible

#### Scenario: M14 Block mirror after incomplete standard work
- **GIVEN** the analyzed target's standard fields are unresolved, conflicting or only partially observed
- **WHEN** the optional mirror would otherwise run
- **THEN** it remains blocked and cannot misrepresent those fields as fully verified

### Requirement: Retain step audit and authority across lifecycle changes

Old approvals SHALL not gain metadata permission during upgrade. New step audit and local reviewed content SHALL survive restart and ordinary cleanup. Source/owner/installation/credential/enablement changes SHALL be rechecked before every step; deletion SHALL prevent late private recreation while preserving required opaque unresolved exclusion. No migration or local deletion SHALL silently remove an upstream mirror. Ordinary logs SHALL not expose private mirrored content. Live capability claims SHALL require separately authorized metadata preservation/readback evidence.

#### Scenario: M15 Upgrade and reopen without inferred mirror authority
- **GIVEN** retained single-photo and stack approvals with completed or unresolved standard work
- **WHEN** metadata support is installed and reopened
- **THEN** earlier bytes, scopes and recovery remain intact and no metadata step is added to old approval

#### Scenario: M16 Revoke authority between steps
- **GIVEN** standard success with the mirror still pending
- **WHEN** source, access, credentials, installation or dispatch enablement changes
- **THEN** the obsolete mirror cannot send while the standard outcome remains retained

#### Scenario: M17 Delete or clean up during mirror work
- **GIVEN** private mirror work and another owner's history
- **WHEN** deletion or ordinary cleanup occurs during a possible send
- **THEN** ownership and unresolved exclusion remain safe, late private recreation is prevented and no upstream deletion is implied

#### Scenario: M18 Separate source support from live verification
- **GIVEN** documented metadata routes and passing synthetic tests but no authorized live mirror evidence
- **WHEN** release readiness is assessed
- **THEN** metadata compatibility remains unverified and mirror dispatch stays disabled

## MODIFIED Requirements

### Requirement: Preview only a staged exact GPS decision

An owner SHALL be able to request a before/after preview for the current staged revision of an owned draft with acknowledged source and selected-field baselines. GPS selection SHALL require a finite camera pair; description selection SHALL require the current reviewed primary-language text and explicit description policy. The default plan SHALL target exactly the analyzed asset with GPS, description or both. Additional stack GPS targets SHALL require the explicit independently reviewed target-manifest contract; siblings SHALL carry GPS only. Unauthorized client/model input SHALL not add targets or unsupported fields. Heading, confidence and estimated error SHALL not add GPS readiness gates. Preview operations SHALL make zero provider calls and zero Immich mutations and SHALL not grant write approval. A separately supported and selected metadata step may accompany at least one standard field on the analyzed photo under the optional-mirror contract; it SHALL not propagate to siblings.

#### Scenario: P01 Preview one coarse camera decision
- **GIVEN** an owned staged draft with valid camera coordinates, acknowledged baseline and a large or unknown estimated error
- **WHEN** the owner requests a preview
- **THEN** an exact single-photo GPS comparison is returned without writing, provider work or a precision threshold

#### Scenario: P02 Reject an incomplete or expanded decision
- **GIVEN** an unstaged, rejected or baseline-unavailable draft, a GPS-selected draft with invalid camera coordinates, or a request adding unapproved targets or unsupported fields
- **WHEN** a preview is requested
- **THEN** no usable plan is created and the draft remains available for correction

### Requirement: Use fresh authorized source and GPS before-values

Preview creation SHALL read current authorized metadata and compare the independently reviewed image identity and acknowledged before-values for each selected field of every target. GPS-selected plans SHALL retain exact nullable GPS comparison; description-selected plans SHALL follow the exact description baseline policy. Changed selected fields SHALL conflict rather than silently becoming approved baselines. Changed source image SHALL require renewed current-image review. Unavailable or malformed selected-field metadata SHALL produce no usable plan. Unrelated metadata changes SHALL not falsely imply image replacement. Fresh read access SHALL not be presented as verified write permission. Selected optional metadata SHALL additionally require a valid namespace-specific before-value under the mirror contract.

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

Every preview SHALL bind owner, installation, draft revision, each reviewed source identity, the frozen exact target list, per-target selected fields and before/intended values, description language/policy when selected and validity period. Its displayed comparison and digest SHALL identify the same immutable plan. Client modifications or cross-scope references SHALL not substitute a different plan. A draft or authority change during creation SHALL prevent publication of a usable preview. Already-matching selected values SHALL be shown as unchanged, not as evidence of a new write. Selected optional metadata SHALL bind exact namespace, export contents, disclosure acknowledgement, step ordering and capability profile to the same digest.

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

Preview UI SHALL present every selected photo identity, exact per-target before/after values and the complete frozen target scope, description language/policy where selected, revision, expiry and changed/unchanged status through keyboard-accessible text/numeric controls. Failed maps SHALL not hide comparisons. Conflicts and source failures SHALL offer review paths without silently changing a baseline. AI data SHALL not enter manual pending coordinates. Selected GPS overlapping manual pending edits SHALL require explicit resolution, and account/result changes SHALL remove private stale content. When metadata is selected, the exact exported content, reader-visibility disclosure and independent standard/mirror step outcomes SHALL also be accessible.

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

Every selected-field mutation SHALL require explicit owner confirmation of the exact current unexpired preview and matching digest. Approval SHALL durably identify actor, installation, draft revision, each source identity, exact target list, per-target selected fields and before/intended values, description policy/language and approval time before any mutation can begin. The same submission key and plan SHALL resolve to one operation across duplicate requests, lost responses and restart; changed-plan key reuse and reuse of a consumed preview SHALL not create another operation. A draft, result, browser flag or digest alone SHALL not grant write authority. A selected metadata step SHALL be explicitly included in this durable approval; neither an older plan nor standard-field success SHALL imply mirror permission.

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

Execution SHALL send exactly the approved GPS pair and/or description fields for each immutable target. The analyzed photo alone may include its selected description; independently reviewed stack siblings SHALL receive GPS only. It SHALL never discover additional targets during execution, clear GPS using fabricated zeroes, substitute subject coordinates or include unselected description, direction, timestamps, rating, favorite or unselected custom metadata. Numeric zero SHALL remain valid. Each potentially sent target attempt SHALL have its own durable reservation; transport/browser retries, redirects and automatic alternate mutation routes SHALL not cause hidden sends. Unsupported adapter/version/field configurations SHALL fail before mutation. Retained single-asset approvals SHALL remain single-asset. An explicitly approved optional metadata step SHALL send only the designated namespace for the analyzed photo, independently of the standard request. Old plan versions SHALL never gain that permission.

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

Each target mutation attempt SHALL recheck current owner/installation/access, the analyzed source and target-specific reviewed image identity, draft revision, enablement and relevant fresh selected-field before-values. Known changes SHALL prevent dispatch and surface conflict or renewed review. Already-matching approved before/intended values for every selected field SHALL produce a verified no-op without a mutation. Unselected field changes alone SHALL not confer new authority or block an otherwise valid plan. A different client's changes between final read and write cannot be guaranteed atomic; the product SHALL not claim remote compare-and-swap protection. Selected siblings SHALL still belong to the reviewed stack before dispatch; later unselected membership changes SHALL not expand the manifest. The optional metadata step SHALL recheck its own capability, source and namespace baseline after the analyzed target standard step has verified.

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

A potentially sent standard request SHALL be followed by authorized selected-field readback before success or resend is decided. Matching all intended fields with unchanged source SHALL establish observed desired state; third values SHALL conflict and unavailable readback SHALL remain unresolved. Mixed intended/baseline fields SHALL retain partial evidence and SHALL not permit replay of the combined payload. An unchanged baseline alone SHALL not prove an earlier timed-out sender cannot still act. Retry SHALL require explicit owner action, unchanged complete fresh baseline, valid current approval/authority, established prior-sender completion and remaining attempt budget. There SHALL be at most two standard mutation attempts per approved target and selected step. Read-only status checks SHALL not reset that budget or create a mutation.

Repeating an already-recorded retry generation SHALL return the current private target operation without another upstream read or attempt allocation, including after successful execution, restart, approval expiry or write disablement. Owner and installation scope SHALL remain mandatory. A generation that has not been accepted SHALL still require all fresh retry eligibility checks before allocating an attempt. Retry/reconcile SHALL identify the exact target, step and generation. Completed targets SHALL not be resent to recover another member. The mirror contract SHALL apply the same readback, sender-completion and identity principles to optional metadata; a mirror retry SHALL never resend completed standard fields.

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

Success for a standard step SHALL require source-consistent readback of every selected field. GPS SHALL use the documented numerical tolerance, distinct from geographic accuracy; description SHALL use the exact text comparison policy. A successful HTTP response alone SHALL not count. Verified GPS SHALL be persisted before refreshing local markers and Missing GPS membership; description-only operations SHALL not change that membership. Mixed outcomes SHALL preserve independently verified fields without claiming complete success. Local persistence or refresh failure after remote success SHALL trigger reconciliation without resending. A delayed catalog sync SHALL not overwrite newly verified GPS with older data. Missing local assets SHALL not be fabricated. Each target SHALL retain an independent verified/refresh outcome; overall partial completion SHALL not fabricate all-target success. A combined operation SHALL be fully successful only when every selected optional step is independently verified. Mirror failure SHALL not retract verified standard-field evidence or its completed local refresh.

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

Approval, before/intended/observed selected-field values, description policy/language, exact revision/target manifest, per-target/per-step attempts, outcomes and timestamps SHALL remain privately inspectable and survive restart and ordinary cleanup. Ordinary logs SHALL omit secrets and private payloads. Disablement/shutdown SHALL stop new sends while preserving ambiguous operations for read-only reconciliation. Account deletion SHALL remove only owned private records and prevent their late recreation; it SHALL not claim to undo an already-received Immich request. Installation rotation SHALL never reuse new-instance authority for an old operation. Each new field/scope capability SHALL remain disabled until its configured live adapter passes explicitly authorized fixture verification. Optional metadata history SHALL retain namespace before/intended/observed values without authorizing upstream deletion or assuming native UI/file support.

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
