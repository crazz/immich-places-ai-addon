## ADDED Requirements

### Requirement: Preserve explicit review decisions as durable local drafts

An owner SHALL be able to accept a valid retained analysis into one durable draft, edit it, stage a GPS decision, reject it and reopen a rejected draft. Acceptance, editing, staging and rejection SHALL perform zero provider calls and zero Immich mutations. The original analysis and older runs SHALL remain immutable. Repeated acceptance SHALL return the existing draft without overwriting edits. Failed, canceled or corrupt analysis records SHALL remain inspectable without being fabricated into proposals.

#### Scenario: D01 Accept and reload a draft
- **GIVEN** an owned valid retained analysis
- **WHEN** the owner explicitly accepts it and later reloads after a restart
- **THEN** the saved draft and its revision remain available beside the unchanged original result
- **AND** neither acceptance nor reload has sent a provider request or changed Immich

#### Scenario: D02 Reconcile repeated acceptance
- **GIVEN** acceptance succeeded but its response was lost, or the draft was subsequently edited
- **WHEN** acceptance of the same analysis is repeated
- **THEN** the existing draft is returned without a duplicate or replacement of its latest edits

#### Scenario: D03 Reject and reopen without affecting another run
- **GIVEN** a saved draft and a newer analysis of the same photo
- **WHEN** the owner rejects and later reopens the draft
- **THEN** the original result, newer run and prior draft decisions remain distinct and no Immich write occurs

#### Scenario: D04 Keep invalid execution separate from review
- **GIVEN** a failed, canceled or corrupt result without a valid proposal
- **WHEN** draft creation is attempted
- **THEN** no proposed camera position or accepted draft is fabricated

### Requirement: Edit camera decisions without inventing precision or writable fields

Drafts SHALL allow explicit candidate selection or user-supplied camera coordinates, optional local heading and per-language description edits. Unknown or subject-only proposals SHALL start without camera geometry. Coordinates SHALL be finite and in range, with zero valid and missing values distinct. Heading SHALL be null or within zero inclusive to 360 exclusive. GPS staging SHALL require an explicit valid camera pair and GPS selection for the analyzed asset only. Radius, confidence, heading and description availability SHALL not impose a GPS quality threshold. Targets and supported write fields SHALL not be expanded by client input or model text.

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
- **WHEN** inputs contain zero coordinates, invalid or partial coordinates, an out-of-range heading, another asset or an unsupported write field
- **THEN** valid zero values are preserved and invalid edits or expanded write scope are rejected without changing the saved revision

### Requirement: Detect stale revisions without losing user edits

Every persisted draft change SHALL require the current revision and create a new revision. Concurrent stale changes SHALL fail without overwriting either the durable draft or the client's unsaved edits. Content changes to a staged draft SHALL return it to draft state and invalidate any preview bound to the previous revision. Lost acknowledgements SHALL be reconciled against current saved state before replay; reanalysis SHALL never replace an existing draft automatically.

#### Scenario: D08 Edit from two tabs
- **GIVEN** two tabs opened the same draft revision
- **WHEN** one saves and the second submits its older revision
- **THEN** the second receives a visible conflict and can compare its unsaved edits with the saved version without overwriting it

#### Scenario: D09 Edit after staging or lose a save response
- **GIVEN** a staged draft or a save with an uncertain acknowledgement
- **WHEN** content changes or the owner reloads saved state to reconcile
- **THEN** an accepted edit produces a new unstaged revision and old previews cannot represent it
- **AND** reconciliation does not blindly resubmit an uncertain edit

### Requirement: Keep dependent draft content honest after location changes

Camera movement or candidate changes SHALL mark inherited camera direction, radius and location-dependent descriptions for renewed review. Scene-only content SHALL retain its independent basis. Each language SHALL retain its own availability and review status. A deliberate correction or review SHALL bind only the affected content to current place facts. Missing or stale non-GPS fields SHALL not block GPS-only staging or become implicitly approved by it.

#### Scenario: D10 Move the camera
- **GIVEN** a draft with inherited heading, radius, candidate-dependent text and scene-only text
- **WHEN** the camera point or selected candidate changes
- **THEN** the dependent values are visibly stale while independent scene-only text keeps its basis
- **AND** their old estimates are not presented as support for the corrected point

#### Scenario: D11 Review one language independently
- **GIVEN** stale descriptions in two languages and an unavailable third language
- **WHEN** the owner corrects or explicitly reviews one description
- **THEN** only that description becomes current and no translation or provider request is triggered
- **AND** GPS staging remains possible without approving the other descriptions or heading

### Requirement: Record fresh baseline availability without blocking local review

Drafts SHALL distinguish unavailable metadata from observed absent, partial or present GPS. Local acceptance and editing SHALL remain possible offline. Fresh source/baseline review SHALL require current access, show exact observed GPS and source status, and bind acknowledgement to the exact observation and draft revision. Expired, changed or unauthorized observations SHALL not replace a baseline. Renewed current-image review SHALL be possible without rerunning AI and SHALL preserve the original analysis provenance.

#### Scenario: D12 Save while metadata is unavailable
- **GIVEN** retained owned history whose source is offline, hidden, removed or inaccessible
- **WHEN** the owner accepts or edits a local draft
- **THEN** the local decision remains durable with explicit unavailable baseline/source status and no unauthorized source read

#### Scenario: D13 Review a fresh baseline
- **GIVEN** an accessible current photo with absent, partial, zero-valued or present GPS
- **WHEN** the owner reviews and acknowledges its fresh baseline for the current draft
- **THEN** the exact observed values and reviewed source identity are retained in a new revision without changing Immich or rewriting the original analysis

#### Scenario: D14 Reject stale baseline acknowledgement
- **GIVEN** a prepared baseline observation
- **WHEN** its source, owner access, installation, draft revision or validity period changes before acknowledgement
- **THEN** acknowledgement cannot replace the saved baseline and the owner receives a renewed-review path

### Requirement: Preserve private draft lifecycle independently of execution

Drafts and revisions SHALL be isolated by owner and installation and remain available across catalog resets, source removal and provider execution disablement. Draft references SHALL preserve their source history against ordinary job cleanup. Account deletion SHALL remove only that account's private drafts; late requests SHALL not recreate them. Old-installation references and foreign IDs SHALL disclose no private content or existence information.

#### Scenario: D15 Preserve a retained decision through cleanup
- **GIVEN** a durable draft with referenced history
- **WHEN** catalog data is reset, source membership changes, execution is disabled or ordinary history cleanup runs
- **THEN** the current owner's local draft and referenced original remain available without starting provider work

#### Scenario: D16 Deny foreign and obsolete references
- **GIVEN** another owner, an unauthenticated request or an obsolete installation reference
- **WHEN** draft or baseline operations are requested
- **THEN** no private data is disclosed or modified and no upstream request uses the other scope's authority

#### Scenario: D17 Delete an account during work
- **GIVEN** private drafts for two accounts and an in-flight request for one account
- **WHEN** that account is deleted
- **THEN** only its private drafts are removed and late completion cannot recreate them or affect the other account

### Requirement: Keep draft editing accessible and separate from manual saving

Explicit draft editing SHALL provide keyboard-operable numeric camera/heading controls, visible save/conflict/stale states and usable behavior with unavailable tiles or images. Inspection alone SHALL remain read-only. AI decisions SHALL never enter manual pending coordinates or the legacy save path. Same-asset manual pending edits SHALL be shown as a conflict before AI staging. Account/result changes SHALL clear private editor state and fence late responses; unsaved edits SHALL not be silently persisted or discarded by ordinary navigation.

#### Scenario: D18 Edit with keyboard and failed tiles
- **GIVEN** a narrow viewport and unavailable map tiles
- **WHEN** a keyboard user corrects camera coordinates and saves a draft
- **THEN** the saved values, revision and stale-content feedback are accessible without using the map

#### Scenario: D19 Preserve manual work
- **GIVEN** existing manual pending coordinates, including a pending change for the draft's photo
- **WHEN** an AI draft is inspected or edited and staging is attempted
- **THEN** manual data is unchanged, AI data never enters manual saving and the overlapping choice must be explicitly resolved before staging

#### Scenario: D20 Leave private unsaved editing
- **GIVEN** unsaved edits or an in-flight private draft request
- **WHEN** the owner navigates away, switches result or changes account
- **THEN** navigation handles unsaved edits explicitly and obsolete private responses cannot populate another result or account
