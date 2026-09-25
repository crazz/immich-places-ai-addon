# AI results and review

## Purpose

Make private immutable AI analysis history discoverable independently of the current catalog, preserve explicit local review decisions as revisioned drafts, and support accessible correction and source review without granting Immich-write authority. Inspection alone remains side-effect-free.

## Requirements

### Requirement: Browse private durable terminal history

Authenticated owners SHALL be able to list and inspect retained current-installation succeeded, failed and canceled analysis items, including located, ambiguous and unknown proposals. History SHALL survive restart, source removal and GPS changes without depending on Missing GPS membership. Foreign or old-installation references SHALL not expose private content. AI execution disablement SHALL not prevent authorized history reads or cause provider work.

#### Scenario: Retain results after GPS or catalog changes
- **GIVEN** retained results for a photo that gains GPS or disappears from synchronized assets
- **WHEN** its owner opens AI Results after reload
- **THEN** the retained history remains available independently of current Missing GPS or catalog membership

#### Scenario: Read with execution disabled
- **GIVEN** retained private history and disabled AI execution
- **WHEN** the owner lists or opens results
- **THEN** local history remains readable without provider calls or activation of workers

#### Scenario: Reject foreign history access
- **GIVEN** an unauthenticated, foreign-owner or old-installation reference
- **WHEN** list, detail or item-result access is attempted
- **THEN** no private record or existence information is disclosed

### Requirement: Filter and paginate retained history honestly

History SHALL support bounded stable pagination and filters for terminal execution state, proposal outcome, asset/run, retained capture date and selected album at launch. The album filter SHALL identify its historical selected-scope meaning. Capture dates SHALL preserve source-local recorded dates without upload-time fallback; absent historical facts SHALL remain unknown. Filters and cursors SHALL be validated and scoped to owner, installation and query.

#### Scenario: Page while new work completes
- **GIVEN** multiple pages of history and concurrently completing jobs
- **WHEN** the owner follows the current query's continuation cursor
- **THEN** deterministic pages do not duplicate entries or mix another query/owner, and new completions appear after refresh

#### Scenario: Filter retained album and date facts
- **GIVEN** dated, undated and older provenance-incomplete runs from selected-album and timeline launches
- **WHEN** the owner uses capture-date or selected-album-at-launch filters
- **THEN** only matching retained facts qualify, undated records remain in unbounded/undated views and current catalog membership is not substituted for missing history

#### Scenario: Reject invalid query boundaries
- **GIVEN** a reversed date range, excessive limit, unsupported filter or foreign/mismatched cursor
- **WHEN** history is requested
- **THEN** the request fails safely without unbounded loading or cross-scope data

### Requirement: Keep immutable runs and independent states

Each run SHALL remain separately inspectable with its original model/mode and provenance. Execution state, proposal outcome, review state and write state SHALL be independent. Located, ambiguous and unknown SHALL be successful proposal outcomes; failed/canceled entries SHALL have safe details without fabricated proposals. In the absence of durable draft/write records, review SHALL be unreviewed and write SHALL be not requested.

#### Scenario: Display successful unknown separately from failure
- **GIVEN** a successful unknown proposal and a failed provider attempt
- **WHEN** both appear in history
- **THEN** their execution/outcome labels differ and neither implies approval or a completed write

#### Scenario: Preserve older runs after reanalysis
- **GIVEN** multiple analyses of the same asset
- **WHEN** a newer run completes or provider settings change
- **THEN** every retained run keeps its immutable proposal and original provenance without replacing older review/write state

#### Scenario: Reject corrupt or unsupported detail safely
- **GIVEN** a retained record that cannot be safely decoded under its stored contract version
- **WHEN** the owner opens it
- **THEN** a safe unavailable state appears without rewriting the record, exposing raw data or breaking other history entries

### Requirement: Authorize current images independently of stored results

Retained result ownership SHALL not authorize new upstream image access. Thumbnails and previews SHALL require current owner, installation and asset visibility/access; hidden, missing or inaccessible assets SHALL show an unavailable placeholder while local result history remains readable. Private image references or cached content SHALL never cross account boundaries.

#### Scenario: Display an authorized current thumbnail
- **GIVEN** a retained result whose source is currently accessible
- **WHEN** its owner opens the list or detail
- **THEN** the image is fetched through the authorized read boundary without provider credentials or stored external image URLs

#### Scenario: Lose source access
- **GIVEN** a retained result whose source became hidden, removed or inaccessible
- **WHEN** a thumbnail or preview is requested
- **THEN** no unauthorized image is returned and the UI shows an unavailable placeholder alongside the local history

#### Scenario: Change accounts during image loading
- **GIVEN** an in-flight private image/detail request
- **WHEN** the active account changes
- **THEN** prior account data is cleared and late responses cannot populate the new account's view

### Requirement: Navigate usable history without side effects

AI Results SHALL provide accessible loading, empty, filtered-empty, error, stale and detail states, preserve navigation/filter position and work without successful catalog sync. Model/provider text SHALL remain untrusted display content. Browsing, refresh, filter and detail actions SHALL make no provider call, create no draft or manual pending coordinate, and perform no Immich mutation.

#### Scenario: Return from detail to the same list
- **GIVEN** a filtered paginated history list
- **WHEN** the user opens detail and returns
- **THEN** filters, list position and keyboard focus are restored predictably without resubmitting analysis

#### Scenario: Handle unavailable catalog or history requests
- **GIVEN** failed catalog sync, an empty filter or a failed history request
- **WHEN** the user navigates AI Results
- **THEN** catalog failure does not block local history, and empty/error/stale states provide an understandable bounded retry path

#### Scenario: Render untrusted content without write authority
- **GIVEN** a proposal containing markup-like or instruction-like text
- **WHEN** it is displayed and navigated
- **THEN** it remains inert escaped text and causes no provider call, draft acceptance, pending-coordinate change or Immich write

### Requirement: Inspect separate camera subject and alternative locations

Proposal review SHALL distinguish camera and subject locations through text and marker identity and show available alternatives without approving them. A selected canonical candidate MAY receive initial inspection focus; ambiguous results SHALL not gain a chosen candidate merely by opening them. Unknown or subject-only results SHALL not invent camera geometry. Numeric zero SHALL remain a valid coordinate and null SHALL remain absent.

#### Scenario: Show different camera and subject points
- **GIVEN** a candidate with distinct camera and subject coordinates
- **WHEN** the user opens its proposal
- **THEN** both are explicitly labeled and the subject/landmark is never substituted for the photographer's position

#### Scenario: Inspect an ambiguous alternative
- **GIVEN** an ambiguous proposal with multiple candidates and no selection
- **WHEN** the user focuses one alternative
- **THEN** only local inspection focus changes, with no approved candidate or persisted result change

#### Scenario: Preserve missing and zero coordinates
- **GIVEN** unknown, subject-only or valid zero-coordinate geometry
- **WHEN** the panel and map render
- **THEN** missing camera points stay absent while valid zero values remain visible and are never treated as missing GPS

### Requirement: Present uncertainty without overstating precision

Review SHALL show supplied granularity, uncertainty notes and estimated camera radius with its basis. Radius SHALL be labeled uncalibrated and SHALL not imply statistical confidence or external verification. Null radius SHALL not create a circle or invented accuracy; zero radius SHALL not be treated as absent or proven exactness. Numeric coordinates SHALL remain authoritative when the map projection cannot display them faithfully.

#### Scenario: Show a supplied radius honestly
- **GIVEN** a validated candidate with a supported radius and basis
- **WHEN** uncertainty is displayed
- **THEN** the estimate, units and basis are clear without a calibrated-confidence claim

#### Scenario: Show unknown or zero radius
- **GIVEN** a candidate whose radius is null or zero
- **WHEN** the panel renders precision
- **THEN** null is explicitly unknown without an invented circle and zero is shown as supplied without asserting exact accuracy

#### Scenario: Inspect projection edge cases
- **GIVEN** valid geometry near a pole or across the antimeridian
- **WHEN** the map frames the result
- **THEN** the original numeric coordinates remain intact and usable even if map framing needs a degraded presentation

### Requirement: Keep camera heading distinct from subject bearing

Direction SHALL mean the stored horizontal optical-axis azimuth clockwise from true north, with method and nullable uncertainty. Null direction SHALL remain unknown. Subject bearing SHALL not create a missing heading; review SHALL not invent pitch, roll or field of view. An arrow SHALL require an actual camera point and supplied direction.

#### Scenario: Inspect supplied true-north direction
- **GIVEN** a camera point and valid direction, including zero degrees
- **WHEN** the user reviews heading
- **THEN** the stored degrees, true-north reference, method and uncertainty are shown without converting zero to missing

#### Scenario: Keep unknown direction unknown
- **GIVEN** a camera and subject with no camera direction
- **WHEN** they are displayed together
- **THEN** direction remains unknown and any displayed subject bearing is separately labeled rather than promoted to optical-axis heading

### Requirement: Explain evidence and lineage without fabricated verification

Review SHALL resolve observation references within the canonical result and version 1.0 source references within trusted stored input provenance, retaining their distinct meanings. Version 2.0 Research references SHALL also resolve within the source list returned in its stored AI answer, without tool metadata or prior backend source authorization. It SHALL show concise support, warnings, limitations and available source lineage. Model-only evidence, user hints and unknown lineage SHALL not become independent verification. Untrusted text SHALL remain inert, and review SHALL not fetch fabricated citations or expose private reasoning.

#### Scenario: Inspect source-free Visual evidence
- **GIVEN** a Visual proposal with observations and no external context
- **WHEN** evidence is expanded
- **THEN** it is labeled model/visual evidence without external-verification claims or invented source links

#### Scenario: Inspect contextual lineage
- **GIVEN** Context-assisted provenance with user hints, allowed sources and unknown lineage
- **WHEN** source details are shown
- **THEN** each kind and uncertainty remain distinct and unknown lineage is not presented as corroboration

#### Scenario: Handle unsafe text or unresolved references
- **GIVEN** markup-like text, instruction-like content or corrupt unresolved evidence references
- **WHEN** the detail is rendered
- **THEN** text stays inert, invalid references fail safely and no automatic research, URL fetch or mutation occurs

#### Scenario: RV09 Review answer-provided Research references
- **GIVEN** a retained Research answer contains source URLs and explanations without a search log
- **WHEN** its owner opens the evidence
- **THEN** safe answer-provided links remain available without an independent-verification claim or a metadata gate

### Requirement: Display exact requested-language descriptions

Review SHALL provide one accessible tab per requested normalized language with primary-language indication, exact canonical `complete` or `unavailable` status and reason. It SHALL preserve scene-only or candidate-specific factual basis and never silently copy a translation, reassign text to an inspected alternative or invent absent text. Changing language tabs SHALL not trigger provider calls or writes.

#### Scenario: Switch between complete and unavailable languages
- **GIVEN** requested languages with complete and unavailable descriptions
- **WHEN** the user switches tabs
- **THEN** exact text or unavailable reason appears with primary-language indication and no fabricated translation or provider call

#### Scenario: Inspect another candidate without rewriting text
- **GIVEN** a description bound to one candidate or only the visible scene
- **WHEN** the user focuses another candidate
- **THEN** the description retains its original labeled basis and is not represented as text for the alternative

#### Scenario: Preserve a non-English language set
- **GIVEN** a valid requested set that excludes English
- **WHEN** review opens or the application locale changes
- **THEN** the same requested language set remains available without adding English or changing stored content

### Requirement: Preserve accessible read-only review under degraded maps

All meaningful spatial/evidence/language information SHALL remain available through keyboard-operable controls and numeric/text presentation when tiles or source images fail. Read-only inspection SHALL never install manual location-edit actions, modify pending coordinates, create a draft or write Immich. Leaving review SHALL preserve the existing manual/GPX state and behavior. Owner/result changes SHALL remove stale private content and map layers.

#### Scenario: Review without map tiles or an image
- **GIVEN** tile failures or unavailable source-image access
- **WHEN** a keyboard user inspects candidates, direction, evidence and language tabs
- **THEN** the complete numeric/text information remains usable with clear labels and predictable focus

#### Scenario: Attempt manual-style map interactions during review
- **GIVEN** an AI proposal displayed while manual pending changes already exist
- **WHEN** the user clicks, drags, drops or opens map context controls
- **THEN** AI inspection creates no pending coordinate, approval or Immich mutation and existing manual state remains unchanged

#### Scenario: Leave or switch private review
- **GIVEN** active proposal layers and detail requests
- **WHEN** the result/account changes or the user returns to manual/GPX mode
- **THEN** stale content and layers are removed, focus is restored and the original manual/GPX workflow still operates

### Requirement: Present coordinates and estimated error for the user's decision

Research detail SHALL prominently show the preferred camera coordinates and estimated error, with an understandable label such as estimated ±500 m or the equivalent in kilometers. There SHALL be no maximum-error filter, minimum-confidence requirement or precision-based disablement of result inspection. Coarse representative points SHALL retain their approximate site/city/region meaning. Unknown error SHALL display as unknown without hiding usable coordinates. Alternatives, concise explanation and the distinction between camera and subject SHALL remain visible. The user decides whether the estimate is useful; this read-only review SHALL not apply coordinates or create approval/write authority.

#### Scenario: RV01 Review a 500-meter estimate
- **GIVEN** a Research result with proposed coordinates and an estimated radius of 500 meters
- **WHEN** the owner opens its detail
- **THEN** both appear prominently as a proposal with estimated ±500 m uncertainty
- **AND** no quality threshold hides the result or prevents inspection

#### Scenario: RV02 Review a city or region estimate
- **GIVEN** a representative point with an estimated radius of several or many kilometers and low confidence
- **WHEN** the result is opened
- **THEN** its approximate meaning, coordinates, radius and alternatives remain reviewable
- **AND** it is not presented as an exact measured camera position

#### Scenario: RV03 Inspect unknown error and ambiguous alternatives
- **GIVEN** a useful point with unknown error or several equally plausible coordinate candidates
- **WHEN** the owner inspects the result
- **THEN** the point or alternatives remain visible with the unknown error or ambiguity explained

#### Scenario: RV04 Use results without maps or pointer input
- **GIVEN** a narrow viewport, keyboard navigation and unavailable map tiles or current photograph
- **WHEN** the owner opens Research detail
- **THEN** coordinates, estimated error, explanations and available source links remain accessible
- **AND** no manual pending coordinate or Immich state is changed

### Requirement: Show safe answer-provided links without a search audit

Research results SHALL display safe HTTP(S) source links and available titles or relevance explanations returned in the AI answer, without requiring a search log, tool provenance or page-access status. Links SHALL be optional and SHALL not be presented as independently verified. All text SHALL remain inert; unsafe schemes, credential-bearing URLs, clearly local/private destinations and unresolved references SHALL not become active links. An unusable reference SHALL not hide or invalidate otherwise valid coordinates. The application SHALL not automatically fetch pages, thumbnails or source availability during persistence or review. Deliberate external navigation SHALL avoid passing the application referrer or opener access.

#### Scenario: RV05 Follow a source from the answer
- **GIVEN** a Research answer includes a public reference URL and short explanation without search telemetry
- **WHEN** the owner opens the result and activates its labeled link
- **THEN** that reference is available through accessible deliberate navigation without a verification badge or search-metadata prerequisite

#### Scenario: RV06 Preserve coordinates with absent or unsafe links
- **GIVEN** a valid coordinate estimate has no sources, a malformed URL, an unsafe scheme or markup-like source text
- **WHEN** its detail is rendered
- **THEN** the coordinates remain visible, text remains inert and unusable destinations are inactive or omitted
- **AND** no automatic source request, script execution or mutation occurs

#### Scenario: RV07 Retain history when a reference disappears
- **GIVEN** an owned retained result whose referenced page later changes or becomes unavailable
- **WHEN** its detail is reloaded with execution disabled
- **THEN** the stored answer remains readable without a source-site request or a claim of current page availability

#### Scenario: RV08 Preserve legacy display and account isolation
- **GIVEN** a version 1.0 Visual/Context-assisted result or an account switch during a Research read
- **WHEN** review updates
- **THEN** legacy results keep their established presentation and the old account's answer is cleared
- **AND** no fabricated external sources are added to legacy history

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

AI Results SHALL show explicit confirmation, durable progress, conflict, unresolved, failed, verified no-op and verified-write outcomes separately from analysis and draft review. It SHALL show the exact approved revision and private before/intended/observed values without claiming causal certainty from readback. Confirming with a known overlapping manual pending edit SHALL require explicit resolution. Lost submissions SHALL reconcile using the same identity, including on HTTP origins without secure-context UUID support. Late initial history SHALL NOT replace a newer confirmation outcome or discard its unresolved identity. A definitive rejection of a repeated confirmation SHALL release that rejected submission for a fresh comparison; missing lookup results or ambiguous failures alone SHALL NOT discard unresolved authority. Accessible status and recovery SHALL work without map tiles, and account/result changes SHALL fence late private replies.

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

#### Scenario: R07 Preserve a newer confirmation when initial history arrives late
- **GIVEN** initial saved-history loading overlaps a new confirmation in the same draft view
- **WHEN** an older history response arrives after the new submission or its acknowledgement
- **THEN** the newer confirmed operation or unresolved submission identity remains current and recoverable
- **AND** older history cannot enable another approval by clearing the pending submission

#### Scenario: R08 Finish recovery after a definitive repeat rejection
- **GIVEN** an uncertain confirmation never committed and its preview has expired
- **WHEN** repeating that same identity receives a definitive expired rejection
- **THEN** the rejected identity no longer blocks a fresh comparison and explicit confirmation
- **AND** no replacement approval or GPS mutation is requested automatically

#### Scenario: R09 Retain an identity while repeated confirmation is still ambiguous
- **GIVEN** a confirmation remains uncertain and lookup has no saved result
- **WHEN** repeating the same identity fails without a definitive rejection
- **THEN** the same submission stays available for reconciliation and new approval remains blocked
