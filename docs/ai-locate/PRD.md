# Immich Places AI Add-on
## Product Requirements Document

**Version:** 1.0 + checkout reconciliation — proposed baseline for OpenSpec planning  
**Date:** 17 September 2026  
**Repository:** `crazz/immich-places-ai-addon`  
**Product owner:** Alex  
**Status:** Ready for product and engineering review; not an implementation claim.

## 1. Product decision

Extend the existing Immich Places fork with an **AI Locate** capability. Reuse its gallery, album and date filtering, selection, image previews, Leaflet map, authentication, synchronization, and manual review experience. Do not build a second photo-management application or write directly to the Immich database.

The user selects photographs without GPS, sends explicitly authorized image data and optional context to a configurable OpenAI-compatible vision endpoint, reviews proposed **camera coordinates**, **viewing direction**, and **descriptions in selected languages**, and explicitly saves chosen fields to Immich.

**Core invariant: analyze → review → stage → confirm save.** Generating or accepting a proposal must never itself modify Immich. “Processed,” “located,” “staged,” and “saved” are different states.

### 1.1 Evidence and scope of review

The original package reviewed public `main` on 17 September 2026. This repository copy is reconciled against `5e70c6165777949c9d8b50ede3b2768bcaa5df87`; see [RECONCILIATION.md](RECONCILIATION.md). The reconciliation used source inspection and local contract/SQL checks, not an application build or runtime integration test. No authenticated calls were made to the owner's Immich instance.

The observed application uses Next.js/React, Go, SQLite, and Leaflet. Its source already exposes album, GPS, and date filtering. These are integration points, not new features to recreate. [R01, R02, R03, R05]

The checkout also has folder, tag, and hidden-asset filtering. A server-resolved AI selection must preserve the active catalog scope or explicitly reject an unsupported scope; it must never silently broaden it. Current day counts use SQLite `DATE()` while range filters compare timestamp strings, so the consistent source-local date behavior required by FR-01 is work to implement, not an existing guarantee. The current grid is ordered by `fileCreatedAt`, independently of its capture-date filter. [R06, R17]

One important behavior needs special handling: the current location handler writes immediately when called and expands an asset to its stack members. The AI workflow must not inherit that implicit write scope. [R05]

## 2. Problem, audience, and outcomes

Older camera images, imports, and edited photographs may lack usable GPS. Manual map placement is slow, particularly when users recognize the trip but not the exact viewpoint. AI can suggest candidates, but confident landmark recognition can still produce the wrong camera location.

The primary user is an existing Immich Places user restoring a personal photo library. A secondary user is a household member using the same deployment with a separate account. The administrator configures allowed AI hosts, limits, and storage. The extension must preserve the existing multi-user boundary; it must not become a single global queue with shared credentials or results.

The desired outcome is **less work per correctly located photograph**, not the maximum number of generated coordinates. A truthful “location unknown” is preferable to an unsupported exact point.

### 2.1 Success measures

Measure accepted results, review effort, and actual spatial error separately. Proposed beta evaluation uses approximately 200 owner-approved photographs with independently verified GPS, split by trip or sequence to reduce leakage. Benchmark preparation must not alter the original library.

| Measure | Definition / release use |
|---|---|
| Spatial quality | Percentage of returned camera points within 100 m, 1 km, and 25 km of verified ground truth; report the sample and mode. |
| Coverage | Percentage of analyzed photographs for which a camera-location candidate is returned. |
| Review usefulness | Acceptance without moving the point; acceptance after correction; rejection; unresolved. |
| Misleading precision | High-evidence results that are wrong or claim an unjustified point-level location. |
| Direction quality | Circular angular error only on photographs with independently verified viewing direction. |
| Cost efficiency | Provider usage and, where rates are known, estimated cost per accepted result, including failed attempts. |
| Safety | No unconfirmed writes, cross-user disclosure, or writes to assets outside the approved target set in acceptance tests. |

No numerical geolocation-accuracy promise is approved by this document. Quality targets must be set after a baseline; model self-reported confidence is not a measured probability.

## 3. Release scope

### 3.1 First complete release — required

V1 includes provider configuration, vision capability testing, image selection, album/date filters, persistent batch jobs, Visual and Context-assisted modes, structured proposals, processed-results browsing, camera-position review, nullable viewing direction, multilingual descriptions, editable drafts, explicit field-level save, conflict handling, and a write audit trail.

**Direction and descriptions are part of V1.** A GPS-only engineering slice can be developed first, but it is not the complete product requested here.

The application must work when the provider supports image input and ordinary structured JSON but not strict JSON Schema. Strict schema output is preferred where available; server validation is mandatory in both cases. OpenAI documents image inputs and structured-output mechanisms, but third-party compatibility must be measured per endpoint/model. [O01, O02]

### 3.2 Extension scope

A later extension adds web research, additional image-context sharing, sequence-level analysis, provider comparison, and calibrated confidence. Interfaces should permit those additions without changing the approval boundary. Research is not silently simulated by asking a non-browsing model to invent citations. Web-search capability is a separate provider/tool concern. [O03]

### 3.3 Non-goals

V1 excludes automatic writeback, video analysis, face identification, geolocation of a person through identity matching, local model training, a separate geospatial warehouse, direct database writes to Immich, modification of original image files by this add-on, and guaranteed recovery of exact camera position or heading from every photograph. It also excludes a general-purpose browser agent and a promise of universal post-save undo.

## 4. User journey and screens

### 4.1 Configure a provider

In Settings, the user creates a private provider profile with a name, API base URL, secret, model identifier, and approved transport options. V1 implements Chat Completions image input; a Responses adapter is an optional extension, not assumed compatibility.

The connection test uses a bundled, non-private test image and reports image support, JSON support, strict-schema support, and authentication or parameter errors separately. Model listing is optional: a missing `/models` route must not prevent manual model entry. The UI must disclose that a provider test may incur usage.

A configured host and model are not proof of successful geolocation. Changing provider settings does not silently alter already queued jobs.

### 4.2 Select photographs

The user opens the existing catalog with missing GPS selected and filters by album and capture-date range. V1 retains the current single-album selection; selecting across multiple pages is supported through an explicit server-resolved selection snapshot.

Folder, tag, and visibility constraints already active in the catalog remain part of selection preview and submission. An unsupported combination must be explained before submission; a gallery count or the existing “select all” action is not a frozen server selection. [R17, R19, R24]

“Select page” and “Select all matching” must be distinct. The analysis dialog reports the exact number of eligible assets, exclusions, and expanded targets, if any. Newly synced photos do not join an already submitted job.

The processed-results view is independent of the missing-GPS predicate. After a successful GPS write, the image disappears from “Missing GPS” but remains visible in “AI Results.”

### 4.3 Run analysis

The dialog shows selected count, provider, model, mode, languages, data-sharing choices, optional user hint, and limits. A useful hint could describe the trip or suspected place; it is context, not verified evidence.

**Visual mode** sends only the prepared image and task instructions. **Context-assisted mode** may additionally send user-approved capture time, album labels, and bounded nearby-location hints. The default mode may be Context-assisted, but external context disclosure requires an explicit saved preference or consent in the dialog. Additional neighboring images are off by default and outside V1.

Once submitted, each eligible photograph gets a durable queue item. The user can leave the page and return later. Progress distinguishes queued, processing, completed, failed, and canceled items. Cancellation is best effort for an already running external request and does not promise a refund.

### 4.4 Review processed results

The results list has thumbnails, place labels, analysis outcome, evidence status, and review/write state. It supports album/date and AI-state filters. Failed or unknown results remain inspectable and can be retried with a new hint or provider.

Opening a result shows the image alongside the existing map and an AI panel. The panel presents a camera point, uncertainty when justified, a separate subject marker, viewing direction when known, concise evidence, alternatives, warnings, and language tabs. Numeric coordinate editing must remain available if map tiles fail.

The user may move the camera point, choose another candidate, edit direction, correct descriptions, reject the proposal, or run a new analysis. Moving the point invalidates any derived direction and location-dependent text until reviewed or regenerated. The original model result remains immutable.

### 4.5 Stage and save

“Accept proposal” persists a local draft and integrates it into the existing pending-change presentation. “Save to Immich” opens the actual write confirmation. A card-level “Add to Immich” shortcut may open this same confirmation, but must not bypass it.

The confirmation lists the exact asset targets and changes by field: GPS pair, selected description, and optional extended metadata. Existing descriptions are preserved by default. The user sees before/after values and must deliberately choose replacement or managed append behavior.

On save, the server rereads current Immich values, detects conflicts, records the approved revision, writes supported fields, and reads them back. The UI can show success, conflict, verification pending, partial success, or failure. A provider response can never authorize this operation.

## 5. Functional requirements

“SHALL” denotes a release requirement. IDs are stable references for downstream OpenSpec requirements and scenarios.

### FR-01 — Reuse the existing catalog

The extension SHALL use the existing authenticated gallery, previews, album and date controls, and map. Missing GPS means either coordinate is absent; latitude or longitude equal to zero is not automatically missing. V1 SHALL accept image assets only and explain exclusions for video, inaccessible, hidden-by-policy, or unsupported assets. Existing GPS and hidden-library semantics are visible in the local database queries. [R06]

Capture dates SHALL use one documented interpretation consistently in the grid, counts, context, and job selection. Capture timestamps without an offset remain local/unknown; they must not be silently treated as the server's timezone. V1 filters the calendar date recorded in `dateTimeOriginal`, preserving the source offset rather than converting it to the server timezone. Missing capture dates appear in an undated group when no date range is set and are excluded from bounded date searches. No upload-date fallback is implicit. The inherited string-based filtering requires boundary tests. [R06]

### FR-02 — Freeze the selection

A job SHALL capture explicit asset IDs and relevant filter metadata. Selection preview and submission SHALL use the same eligibility rules. Changes in album membership, filters, sync state, or stack membership after submission SHALL NOT silently change the job's target set. Duplicate IDs within a batch SHALL collapse to one item.

### FR-03 — Configure private providers

Each provider profile SHALL belong to one application user and use an administrator-approved host. Secrets SHALL be stored server-side encrypted and never returned in readable form. Saving a profile SHALL not automatically send private photographs. Unsupported options SHALL produce actionable errors, not an unexplained provider failure.

### FR-04 — Obtain and record data-sharing consent

The user SHALL control image and context disclosure. Clicking **Start analysis** with the provider and selection displayed SHALL authorize that image analysis without an additional consent checkbox; context classes beyond the images SHALL remain separately selectable. The system SHALL record the authorized context classes, provider revision, and asset selection. The default payload SHALL exclude file paths, people identities, device serials, unrelated albums, and unselected images. Revoking access or disabling the provider SHALL stop future dispatches; already transmitted data cannot be recalled by the add-on.

### FR-05 — Run durable, bounded work

Jobs SHALL survive browser closure and backend restart. The system SHALL limit concurrency, retries, input size, output size, and per-job external calls. A permanent authentication or policy error SHALL not trigger unlimited retries. One bad photo SHALL not prevent unrelated items in the batch from completing. Reanalysis SHALL create a new run, not overwrite a result under review.

### FR-06 — Return an honest location proposal

The output SHALL distinguish the photographer's location from the photographed subject's location. It SHALL support point, street/area, site, city, region, and unknown granularity. A landmark or city centroid SHALL NOT automatically become the camera coordinates.

The model may return candidates without choosing one, or return unknown. Precision radius is nullable and labeled as estimated, not statistically calibrated. A source-free model guess SHALL not be presented as externally verified. The backend SHALL reject invalid coordinates and malformed contracts.

### FR-07 — Represent camera direction separately

Direction SHALL mean horizontal camera azimuth in degrees clockwise from true north, within `[0, 360)`. It SHALL include a method and nullable uncertainty. The bearing toward a visible object is not necessarily the camera's optical-axis direction. Unknown direction SHALL be `null` and SHALL NOT block a GPS-only save. Pitch, roll, and field of view are outside V1.

### FR-08 — Generate multilingual descriptions

Users SHALL choose one or more language tags and one primary language for Immich. English SHALL be available but not mandatory. Descriptions SHALL describe the visible place and scene, not invent historical facts or turn uncertain identification into certainty. The system SHALL retain every requested language's status and text.

Descriptions SHALL share the same approved factual basis. Missing or failed translations SHALL be explicit and retryable without rerunning geolocation. Changing the chosen location SHALL mark affected translations stale. Stale location-dependent text SHALL not be saved without regeneration or explicit user review.

### FR-09 — Keep reviewable result history

Completed, ambiguous, unknown, and failed outcomes SHALL remain accessible. The system SHALL preserve model output separately from user edits. The processed list SHALL show analysis state and write state independently. A newer result SHALL not replace a user's approved draft automatically.

### FR-10 — Persist editable drafts

Acceptance SHALL create or update a durable, per-user draft with a revision number. The draft SHALL survive refresh and navigation. A draft SHALL carry the source result, exact targets, chosen fields, and baseline values. Competing edits from two tabs SHALL be detected through revision checks.

### FR-11 — Confirm every Immich write

The final confirmation SHALL bind to an immutable draft revision, exact target IDs, and exact field values. Any subsequent edit invalidates the confirmation. The default target is only the analyzed asset. Stack propagation SHALL require a separate preview of members and a deliberate choice; stack membership is not evidence of an identical camera viewpoint.

Existing location writes expand to stacks, so the AI writer SHALL use an explicit-target path rather than call that handler unchanged. Descriptions and directions SHALL not be copied to other stack members by default. [R05]

### FR-12 — Preserve existing metadata and handle conflicts

GPS and description choices SHALL be independent. Existing GPS that appears after analysis SHALL produce a conflict, not an overwrite. Existing description text SHALL be preserved unless replacement or a managed append block is selected. Rerunning an append SHALL update the same managed block, not duplicate it.

The server SHALL reread the relevant fields before writing and verify after writing. Unknown network outcomes SHALL trigger reconciliation before resending. The UI SHALL distinguish successful standard-field writes from failed optional metadata writes. Direct SQL access to Immich is prohibited.

### FR-13 — Store extended information without misrepresenting support

The local application database SHALL be the authoritative store for proposals, direction, translations, sources, provenance, and reviews. Standard Immich fields SHALL receive only approved GPS and description values. Extended metadata may be mirrored under a namespaced key when the deployed Immich version and permissions support it.

Immich v3.2.2 source provides coordinate and description updates plus object-valued asset metadata. It does not provide a camera-direction field in the inspected update DTO. Availability on the owner's deployed server remains an integration gate. [I01, I02]

### FR-14 — Record a safe audit trail

For every attempted write, record actor, target, approval revision, selected fields, previous values, intended values, outcomes, and timestamps. Never log secrets, image bytes, or private prompts in ordinary logs. Export and deletion of local AI results SHALL be supported under a retention policy. Deleting local results SHALL not silently remove metadata already written to Immich.

## 6. Non-functional requirements

| ID | Requirement |
|---|---|
| NFR-01 | Authorization is enforced server-side for every provider, job, result, draft, asset, and write request; no cross-user access or cache leakage. |
| NFR-02 | Analysis and save operations are durable, bounded, idempotent at the application boundary, and recoverable after interruption. |
| NFR-03 | AI disabled means existing browsing and manual geotagging remain usable; no AI-provider calls are made. |
| NFR-04 | On the agreed reference NAS, cached gallery/filter requests target p95 under 1 second with 100,000 indexed assets and two active AI jobs. This is a test target, not a measured result. |
| NFR-05 | Queue submission acknowledges within 2 seconds for up to 500 eligible assets on the reference environment; it does not wait for inference. |
| NFR-06 | Review controls are keyboard-operable, state is not communicated by color alone, and point/heading editing has numeric alternatives. |
| NFR-07 | Existing deployment remains a frontend plus backend with persistent local SQLite storage; no Redis, PostgreSQL, or GPU dependency is required for the add-on. [R15] |
| NFR-08 | Upgrade and rollback procedures protect the SQLite database, encryption key, and audit history. Diagnostics expose sanitized error codes and aggregate metrics. |

## 7. Acceptance scenarios

**AC-01 — Filter and freeze.** Given 80 eligible photos in an album/date range, when the user selects all matching and confirms analysis, then those 80 IDs are captured. A later sync adding another photo does not add it to the job. Requirements: FR-01–02.

**AC-02 — Unknown location.** Given an uninformative image, when the model cannot establish a camera location, then the result is completed/unknown with no GPS point, not a fake `0,0` coordinate or a technical error. Requirements: FR-06, FR-09.

**AC-03 — Landmark is not viewpoint.** Given a recognized distant monument, when only its position is supported, then the subject marker is shown separately and camera GPS remains unselected. Requirements: FR-06–07.

**AC-04 — Acceptance does not write.** Given a valid proposal, when the user clicks Accept, then a local draft is saved and no Immich mutation occurs. A subsequent refresh restores the draft. Requirements: FR-10–11.

**AC-05 — Stack scope.** Given a selected image in a five-image stack, when the user approves only that image, then only that ID can be written. A newly added sixth member is never included without a new confirmation. Requirements: FR-02, FR-11.

**AC-06 — Conflict.** Given another client added GPS or changed the description after review, when save begins, then the add-on shows the conflicting fields and does not overwrite them silently. Requirements: FR-12.

**AC-07 — Restart and cancellation.** Given queued and running items, when the backend restarts, then completed results remain and interrupted leases are reconciled. Cancellation prevents new dispatches and late responses cannot resurrect canceled work. Requirements: FR-05.

**AC-08 — Language and partial write.** Given English and Ukrainian descriptions, when English is selected for the standard description and optional extended-metadata write fails, then standard-field success remains recorded and the UI identifies the incomplete mirror. Requirements: FR-08, FR-12–13.

**AC-09 — Tenant isolation.** Given two users, when either supplies the other's provider, result, or asset ID, then the server denies access without returning private content. Requirements: FR-03–04, NFR-01.

**AC-10 — Incomplete provider output.** Given a truncated response, refusal, unsupported schema mode, or out-of-range coordinate, then the system uses the configured bounded handling path and never creates a writable proposal from invalid data. Requirements: FR-03, FR-05–06.

**AC-11 — Successful GPS write.** Given verified writeback, then the image leaves the missing-GPS catalog but remains available in AI Results with the saved revision and audit information. Requirements: FR-09, FR-12–14.

**AC-12 — User correction invalidates dependent content.** Given a staged location and direction, when the point is moved to a different candidate, then the direction and descriptions are marked for review and the old approval cannot be reused. Requirements: FR-07–08, FR-10–11.

## 8. Release gates and unresolved decisions

Release gates are: successful integration against an explicitly recorded Immich version; provider capability fixtures; all safety scenarios passing; migration/restart tests; existing manual workflow regression tests; and a published baseline quality report. Read-only libraries, ownership restrictions, missing preview formats, and missing write capabilities must be represented as explicit compatibility outcomes rather than hidden failures.

The following are implementation-time decisions, not invitations to silently change product scope: the first production model and provider; the exact Immich version under test; approved geocoding/search hosts for later Research mode; the chosen retention period; and the reference hardware/test fixture. Proposed defaults are described in the technical design.

A full post-save rollback to “GPS absent” is not a V1 promise. V3.2.2's inspected coordinate schema accepts numbers rather than null, so restoring absent GPS requires a separately verified API capability. Pre-save discard and an audit of old values remain mandatory. [I04]

## 9. OpenSpec handoff boundary

Use this document as the product source of truth and `TECHNICAL_DESIGN.md` as the architectural baseline. Preserve requirement IDs and acceptance scenarios when deriving capabilities and changes. OpenSpec should generate detailed specifications and tasks after reconciling the actual checkout; this document deliberately does not prescribe a task-by-task backlog.

When a technical choice conflicts with an explicit product requirement, surface the conflict and record a decision. Do not solve implementation difficulty by removing multilingual output, review, persistent processing, or explicit write approval.

## References

Reference IDs resolve in `SOURCES.md`. Repository statements are observations of public source, not runtime verification. All other requirements, targets, defaults, and architectural choices in this document are proposals for this project.
