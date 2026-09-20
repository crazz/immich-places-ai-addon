# Immich Places AI Add-on
## Technical Design and Integration Specification

**Version:** 1.0 + checkout reconciliation + adopted engineering standards (ADR-07)  
**Date:** 17 September 2026  
**Repository:** `crazz/immich-places-ai-addon`  
**Companion document:** `PRD.md`  
**Status:** Design input for OpenSpec; not production code or a completed implementation.

The adopted [architecture](../engineering/architecture.md), [testing](../engineering/testing.md) and [coding standards](../engineering/coding-standards.md) govern implementation. [ADR-07](../engineering/decisions/ADR-07-ai-internal-packages.md) supersedes the original flat-file AI package placement. The original Word export remains historical and does not contain this amendment.

## 1. Architecture decision

Implement AI Locate inside the existing Go backend and Next.js/React frontend. Use the existing SQLite database for durable jobs, immutable results, review drafts, and write operations. Preserve the existing frontend/backend deployment boundary. Do not add Python, Redis, PostgreSQL, a second photo catalog, or a direct connection to Immich's database.

The AI module is a **proposal producer**. It can read authorized photographs and permitted context; it cannot call the Immich writer. Only a user-confirmed write operation can invoke writeback.

### 1.1 Alternatives considered

| Option | Benefit | Cost / decision |
|---|---|---|
| In-process Go module | Reuses auth, catalog, deployment, and storage; small integration surface. | Shares backend resources; selected for V1 with bounded workers. |
| Separate Go worker service | Better resource and failure isolation without a second language. | Extra deployment and queue coordination; defer until measured load warrants it. |
| Python AI service | Suitable for future local vision/retrieval experiments. | Additional runtime, contracts, and operational burden; no current requirement justifies it. |
| Browser-direct AI calls | Superficially fewer backend changes. | Exposes credentials, lacks durable execution, and complicates privacy; rejected. |
| Rewrite or fork Immich itself | Full native integration. | Discards the selected Places foundation and increases upgrade burden; rejected. |

These are engineering judgments for this project, not measured performance comparisons.

### 1.2 Logical data flow

```text
Existing catalog and selection
    -> AI launch dialog / consent
    -> server-resolved selection snapshot
    -> durable SQLite job items
    -> bounded Go workers
    -> image preparation + permitted context
    -> vision provider -> validation -> immutable analysis
    -> result review -> durable draft
    -> write preview -> explicit confirmation
    -> exact-target Immich API writer -> readback -> audit
```

The geocoder is an optional candidate-resolution dependency, not proof of the camera's position. Research/search tools attach later to the analysis orchestrator, never to the writer.

## 2. Verified baseline and repository integration map

The original package reviewed public `main` on 17 September 2026. This repository copy is reconciled against `5e70c6165777949c9d8b50ede3b2768bcaa5df87`. [RECONCILIATION.md](RECONCILIATION.md) records source evidence and verification limits. The application was not built or run, the Go suite was not executed, and no application migration or live Immich mutation was performed. Proposed paths in later sections do not imply existing files.

| Existing path | Verified integration surface |
|---|---|
| `package.json` | Next.js 16.1.6, React 19.2.3, Leaflet 1.9.4, TypeScript, and build/lint scripts. [R02] |
| `backend/go.mod` | Go 1.25; modernc SQLite driver and Goose migration dependency. [R03] |
| `backend/main.go` | Service construction, protected routes, auth integration, and background services. [R04] |
| `backend/handlers.go` | Date/GPS query parsing, preview proxy, suggestion handler, and immediate location write with stack expansion. [R05] |
| `backend/database.go` | Tenant-scoped assets, existing filter SQL, SQLite WAL, and local synchronized metadata. Description is absent from the inspected asset-column projection. [R06] |
| `backend/migrations.go` | Embedded SQL migrations under `backend/migrations/` using Goose. [R07] |
| `backend/crypto.go` | AES-GCM encryption helpers with a legacy plaintext-compatible read path. [R08] |
| `src/app/page.tsx` and `src/shared/context/AppContext.tsx` | Existing shell, catalog, map, selection, and view-context composition. [R09, R10] |
| `src/shared/context/useAppProviderState.ts` | Selection, suggestions, save callbacks, and map refresh orchestration. [R11] |
| `src/shared/components/PhotoListContainer.tsx` | Album/date controls, catalog selection, and GPX-related presentation. [R12] |
| `src/features/selection/useSelectionController.ts` and `src/shared/types/map.ts` | Selection operations and coordinate-only pending-change types; add an AI source without overloading them with full analyses. [R13, R14] |
| `docker-compose.yml` | Separate frontend and backend services, with backend `/data` persistence. [R15] |
| `src/middleware.ts` and `src/utils/client.ts` | Default `/api/backend` prefix rewritten to `BACKEND_URL`; this checkout has no API route-handler proxy to extend. [R18] |
| `src/shared/context/useCatalogDomain.ts` | Selects album or folder mode and passes tag, GPS, hidden, and date filters into the catalog. [R17] |
| `src/features/selection/useSelectionState.ts` and `locationSave.ts` | In-memory pending coordinates; current save helper repeats failed assets once. Neither is a durable approval protocol. [R19] |
| `backend/interfaces.go` and `backend/immichClient.go` | Existing handler client has image reads and bulk GPS update only; shared HTTP transport is configured with `RetryMax = 3`. [R20] |
| `backend/suggestionService.go` | Uses capture time or a `fileCreatedAt` fallback and includes frequent locations; not a consent-filtered AI evidence source. [R21] |

Important distinctions: the existing pending-change user experience is reusable, but it is not an established durable AI draft API. Add server-side AI draft persistence. Never call the existing location handler from an “Accept proposal” action. Do not assume that invoking it for one ID affects only that ID. [R05]

The existing application call is `PUT /assets/{assetID}/location`; it resolves stacks and calls Immich's bulk `PATCH /api/assets` route. It does not perform the fresh metadata read, revision-bound approval, description update, or post-write verification required below. The proposed per-asset adapter is new work. [R05, R20]

## 3. Proposed module boundaries

Put new AI core behavior in small, feature-focused packages under `backend/internal/ai/`, as adopted in [ADR-07](../engineering/decisions/ADR-07-ai-internal-packages.md). Keep the existing backend in place and integrate through narrow adapters and explicit construction in `package main`. Create packages as actual boundaries appear; the following is an ownership map, not a requirement to scaffold every directory in the first change.

| Proposed backend area | Responsibility |
|---|---|
| `backend/ai_*.go` integration files | Thin authenticated handlers, legacy-service adapters and composition; split by resource/responsibility. |
| `backend/internal/ai/analysis/` | Analysis orchestration, validation, consent/context rules, evidence lineage and consumer-owned provider/read interfaces. |
| `backend/internal/ai/jobs/` | Job claiming, leases, bounded attempts, cancellation, budgets and restart recovery. |
| `backend/internal/ai/review/` | Durable draft policy, revisions, edits and stale-content rules. |
| `backend/internal/ai/writeback/` | Confirmed exact-target operations, conflict detection, reconciliation and audit policy; owns its mutation interface. |
| `backend/internal/ai/model/`, when needed | Small shared pure types/contracts used by actual core consumers; no generic catch-all layer. |
| `backend/internal/aiadapters/`, when needed | Concrete provider/Immich transports, image preparation, credential/egress enforcement and SQL repositories, separated from core packages. Adapters needing legacy unexported services may initially live in `package main`. |
| `backend/migrations/<next>_ai_*.sql` | Additive migrations; `018` is next at the pinned checkout, but recheck before creating files. |

Core packages must not import concrete adapters, SQL drivers or the executable package. The change that first introduces subpackages must also update and verify the backend Docker build, which currently copies only root Go sources and migrations. Apply the coding standard's file-size limit to integration files and tests as well as core code.

Frontend additions belong under `src/features/ai/`: provider settings, launch dialog, progress panel, result browser, review panel, map overlays, and an API client. Extend existing selection/context types only with the minimum integration references. Keep full analysis history and server draft state inside the AI feature boundary.

Define small interfaces at actual consumer boundaries, such as `VisionProvider`, `AssetReader`, `ContextSource`, focused persistence operations, and a mutation interface owned by confirmed writeback. Avoid a single all-purpose `AIStore`. The analysis service receives only read-side interfaces and cannot depend on the writer. Defer `SearchProvider` and `CandidateResolver` until an accepted consumer needs them. A provider result cannot supply an authoritative application user ID, Immich asset ID, write target, or approval token.

These interfaces are new. Existing `HandlerImmichAPI` exposes only `bulkUpdateLocation`, `getThumbnail`, and `getPreview`; `SyncImmichAPI` is a separate interface. Add narrow AI read/write adapters and matching test doubles rather than assuming metadata reads or description/custom-metadata writes already exist. The current backend holds one configured Immich URL; a durable installation identity and its reset/invalidation behavior also need implementation. [R04, R20]

## 4. Provider and image transport contract

### 4.1 Capability negotiation

A profile identifies a user, approved base URL, model, credential reference, and immutable revision. A capabilities record contains observed image-input support, JSON mode, strict-schema support, request-size limits if known, supported token-limit parameter, and test timestamp.

The V1 adapter sends a non-streaming request to the configured Chat Completions endpoint, with text and an image data URL. OpenAI supports encoded image input; compatible endpoints must pass a real synthetic-image test rather than inherit that assumption. [O01]

Manual model entry is supported even when discovery is unavailable. Do not attach optional parameters such as temperature, reasoning settings, or a particular maximum-token field unless enabled by the profile's capability policy. An unsupported strict-schema parameter can trigger one explicit downgrade to JSON mode when allowed; authentication failures must not be treated as schema incompatibility.

OpenAI distinguishes schema-constrained output from JSON mode and documents refusal/truncation edge cases. The application therefore validates every result independently of transport mode. A valid JSON object does not prove a correct geolocation. [O02]

Provider settings and capabilities are versioned. Jobs bind to an immutable profile revision. Disabling a profile prevents subsequent dispatches across all revisions. Updating a model or secret must not silently redirect queued private data to another host.

### 4.2 Image preparation

Workers fetch images through the authenticated Immich client on the server, not by sending the provider a private Immich URL. Normalize orientation, decode to a supported raster format, strip unnecessary metadata, and create a bounded analysis copy. Never expose the Immich key in an AI request.

The current preview handler streams the upstream image response; it does not provide these normalization, metadata-stripping, or decode-size guarantees. Reuse authenticated fetching, then implement bounded image preparation in the AI worker. [R05, R20]

Proposed initial limits: one image per item, 2,048-pixel maximum long edge, 10 MiB encoded request-image limit, and a 40-megapixel decode guard. These are configurable operational defaults to benchmark, not model requirements. Prefer an existing suitable preview; unsupported formats fail explicitly. Higher-resolution retry requires an explicit setting and the same limits. Original files are never changed by this module.

Use a temporary file only when needed; clean it on completion and startup recovery. Do not store Base64 payloads in job JSON, logs, or the audit trail.

### 4.3 Credentials and egress

All secrets remain in the backend. Reuse the existing encryption primitives but require encrypted storage for new AI credentials; do not inherit the legacy plaintext read fallback for this new table. [R08]

Administrator policy controls allowed hosts, ports, protocols, and custom headers. HTTPS is required except for explicitly approved local inference endpoints. Credentials are never forwarded on redirects; provider redirects are disabled by default. Request destinations are revalidated after DNS resolution to prevent DNS-rebinding and SSRF attacks. Cloud metadata addresses remain blocked even when private-network providers are allowed.

Research URL fetching uses a different, stricter policy: no access to internal hosts, loopback, link-local addresses, or credential-bearing URLs. All model output and web content is untrusted. The model has no shell, filesystem, arbitrary SQL, or writeback tool.

## 5. Analysis pipeline and evidence model

### 5.1 Pipeline

The worker authorizes the asset, prepares the image, builds the allowed context, invokes the provider, validates the response, and persists a terminal result. Optional candidate resolution can add known POI coordinates, but cannot promote them automatically to camera coordinates.

Context-assisted mode includes at most six metadata-only neighbors within the configured window, the explicitly selected album label when authorized, capture time with timezone status, and a short user hint. Existing suggestions are inputs to evaluate, not ground truth. Exclude hidden or inaccessible assets and omit unrelated frequent/home locations by default.

Do not forward `getSuggestions` output wholesale. At the pinned checkout it falls back to `fileCreatedAt` when capture time is absent and appends frequent locations. Its same-day/neighbor queries exclude hidden libraries but do not filter per-asset `isHidden`. Build a separate consent-filtered context adapter, keep absent capture time absent, and recheck each source's eligibility and provenance. [R06, R21]

Every context item records its source and lineage. A previous AI-generated GPS value must not become independent confirmation of another AI guess. Known AI-origin neighbors are excluded by default; a later opt-in mode may use them with an explicit dependence label. Lack of provenance is not proof that a neighbor is independently verified.

Visual observations, user hints, temporal context, geocoder results, external evidence, and user corrections remain separate categories. No private chain-of-thought is requested or retained; store concise observations, support summaries, contradictions, and source references.

Date selection uses the source-local calendar date in `dateTimeOriginal`, not the server clock. Apply the same helper to asset queries, counts, and frozen selections; preserve offsets and distinguish missing dates. Test midnight, daylight-saving transitions, offsetless timestamps, and start-after-end errors before changing inherited string predicates.

CH04 implements this policy for existing catalog queries through a shared validated calendar-prefix expression and range predicates, with real SQLite and HTTP regression coverage. See the [maintained catalog contract](../../openspec/specs/catalog-capture-dates/spec.md). Frozen selections and the new undated-group presentation remain future work; stored timestamps and independent gallery ordering are unchanged.

The current range predicate compares text (`>= startDate`, `< endDate + "T99"`), but `countAssetsByDay` groups using `DATE(dateTimeOriginal)`. An offset timestamp can therefore match one calendar date and be counted under another. `parseDateRangeParams` validates each date but not range ordering. Capture-date consistency is new work across catalog modes, counts, context, and AI selection; it is not a reason to silently change the existing `fileCreatedAt` sort order. [R05, R06, R17]

Freeze explicit IDs with the active view/filter manifest: album or folder scope, tag, GPS, hidden policy, and date bounds as applicable. Preserve stack-primary browsing semantics while independently validating any explicitly supplied ID. Recompute eligible/excluded counts using the same resolver used for submission; the existing missing-location count endpoint has no folder scope, and current “select all” operates on the supplied page assets. Reject unsupported scopes explicitly. [R17, R19, R24]

### 5.2 Research extension

Defer `SearchProvider`, `CandidateResolver` and search adapters until an accepted feature needs them; do not create unused extension interfaces in V1. A later Research mode can retrieve candidates, resolve POIs, compare geometry, and associate claims with source IDs created by the backend while preserving the approval boundary. OpenAI-compatible inference does not imply built-in search; OpenAI's own web search is an explicit tool capability. [O03]

The inherited geocoder chain must not be invoked blindly for automated workloads. Public Nominatim has a maximum one-request-per-second policy, caching and application-identification requirements, additional bulk restrictions, and a prohibition on autocomplete. Research must use an explicitly approved backend/provider policy and be able to bypass a Nominatim-first chain. Existing manual geocoder use should also be reviewed for compliance. [M01]

### 5.3 Position, heading, and uncertainty

Coordinates use WGS84 decimal latitude/longitude. Store camera and subject coordinates separately. Granularity and radius describe the camera estimate; a city-level candidate must not carry a fabricated 10-meter radius.

Camera azimuth is clockwise from true north. A backend may compute camera-to-subject bearing for a consistency check, but it must display that quantity as **subject bearing**, not camera heading. Off-center composition, panoramas, crops, mirrored images, and uncertain positions invalidate naive heading checks. When the user changes a camera or subject point, invalidate dependent heading rather than quietly retain it.

Model self-assessment is not calibrated confidence. The UI's verification status is derived from recorded evidence provenance: visual-only, context-supported, externally-supported, or user-reviewed. These categories are not probabilities, and user review does not imply surveyed accuracy.

## 6. Structured analysis contract

The [canonical backend schema](../../backend/internal/ai/results/ai-analysis-result.v1.schema.json) defines the provider-facing result, with examples in `examples/`. CH07 embeds its unchanged v1.0 bytes and implements bounded structural/semantic validation; see the [caller contract](../ai-result-validation.md) and [verification](../engineering/ai-result-validation-verification.md). This full analysis contract has not been acceptance-tested with a model endpoint. Provider invocation and the server-owned persistence envelope below remain later work.

The object contains `schema_version`, `outcome`, `selected_candidate_id`, `observations`, `candidates`, `descriptions`, and `warnings`. An outcome is `located`, `ambiguous`, or `unknown`. Candidates contain separate camera and subject information, nullable direction, evidence references, and a support summary.

Descriptions are an **array of language-tagged entries**, not an unrestricted map. This keeps object shapes closed for strict-schema providers. Every object disallows unknown properties; optional values are represented explicitly as nullable values. The backend can derive a simplified transport schema for a provider's supported subset while retaining the canonical server validator. [O02]

### 6.1 Server-owned envelope

The server wraps the validated model object with `analysisID`, `userID`, `instanceKey`, `assetID`, job/item IDs, timestamps, image/context hashes, provider revision, requested languages, prompt/schema versions, normalized source records, usage, validation findings, and verification status. These identity and authorization fields are never accepted from model output.

Each source record has a backend-assigned ID, type, retrieval time, provider, normalized URL or local reference, and a short supporting excerpt where permitted. Models refer to source IDs already supplied in their input. A model-written URL alone is not evidence; Visual mode normally has no external source records.

### 6.2 Semantic validation beyond JSON Schema

Enforce finite numbers, coordinate ranges, azimuth range, nonnegative uncertainty, unique candidate/observation IDs, and valid references. `located` requires a selected candidate with a camera point; `ambiguous` requires multiple candidates and no automatic selection; `unknown` requires no selected candidate. Descriptions must match requested languages exactly once or report an explicit unavailable state.

The radius basis must agree with the presence of a radius. Candidate-dependent descriptions require a real candidate ID; scene-only descriptions do not. The language tag is validated independently of string shape. All source IDs must belong to the authorized evidence bundle. Treat inconsistent output as invalid or downgraded-for-review according to a documented rule, never as ready to save.

Keep the immutable provider result separate from the edited draft. Retranslation creates a new description revision linked to the approved place facts; it does not silently move the GPS point.

## 7. Persistence and migrations

Add tables through the existing Goose migration mechanism. The pinned checkout has migrations `001`–`017`, ending at `017_add_original_path.sql`; `018` is the next available number at this revision only. The existing database is `immich-places.db`; initialization executes WAL, foreign-key, and busy-timeout PRAGMAs. Confirm connection-pool PRAGMA behavior when implementing new relations. Existing migration tests cover fresh databases, legacy bootstrap, and repeated migration runs; extend these fixtures for AI rather than replacing the runner. [R06, R07, R22]

| Proposed table | Key content and constraints |
|---|---|
| `ai_provider_profiles` | ID, userID, name, activeRevision, enabled; stable logical profile. |
| `ai_provider_versions` | Profile ID + revision, userID, base URL, model, encrypted secret, capabilities/settings JSON; immutable configuration. |
| `ai_jobs` | ID, userID, instanceKey, provider revision, mode, languages, consent snapshot, selection hash, limits, aggregate status, cancellation flag. |
| `ai_job_items` | ID, jobID, userID, assetID, state, attempt count, nextAttemptAt, lease token/expiry, source fingerprint, latest analysis ID, sanitized error. Unique `(jobID, assetID)`. |
| `ai_analyses` | ID, itemID, userID, assetID, outcome, immutable result JSON, evidence bundle, provenance/usage JSON, validation report, createdAt. |
| `ai_drafts` | ID, userID, instanceKey, source analysis ID, revision, target manifest, selected field values, baseline snapshot, content-stale flags, review state. |
| `ai_write_operations` | ID, userID, idempotency key, confirmed plan hash, draft revision, exact immutable plan, status, approval timestamp. |
| `ai_write_targets` | Operation ID + assetID, userID, before/intended/observed JSON, per-step states, verification attempts, timestamps, sanitized errors. |

Every read and write includes user scope, including joins and uniqueness rules. `instanceKey` identifies the configured Immich installation, not just a display URL. Changing the installation requires a controlled reindex and invalidates pending approvals; IDs must not be reused across servers accidentally.

Useful indexes cover `(userID, createdAt)`, `(state, nextAttemptAt)`, `(userID, assetID, createdAt)`, and job item membership. Preserve immutable analyses even when their source disappears from the synchronized catalog, but suppress image access and purge private data according to deletion policy. Do not make synchronization overwrite AI history.

Descriptions need not be added to the entire local asset index in V1. Fetch live asset metadata when preparing and verifying writes, and store only the relevant baseline in the draft/audit. This avoids expanding every sync operation just to support AI review.

Backup the database and encryption key together, but protect them separately. Use SQLite's consistent backup mechanism or stop the service for a filesystem copy; do not copy only the live `.db` while ignoring WAL state. SQLite WAL is not suitable for a shared network-filesystem database. Use local host storage; a distributed worker design would require a different coordination plan. [M02]

## 8. Durable execution, states, and budgets

### 8.1 Independent state dimensions

| Dimension | States / meaning |
|---|---|
| Item execution | queued, running, retry_wait, blocked, succeeded, failed, canceled. |
| Analysis outcome | located, ambiguous, unknown; only populated after valid completion. |
| Review | unreviewed, draft, staged, rejected; tied to a result/draft revision. |
| Write | not_requested, queued, writing, verifying, succeeded, partial, conflict, failed. |

A job derives its aggregate status from items. “Succeeded + unknown” is valid. “Staged” does not imply an Immich mutation. The AI Results query must not be restricted to currently missing-GPS assets.

### 8.2 Claiming and recovery

Claim one eligible item in a short atomic SQLite transaction, storing a random lease token and expiry. Commit before any network request. Heartbeat while working; finalize only when the lease token still matches. This fencing rule prevents a late worker from overwriting a recovered result.

On restart, reconcile expired leases and move eligible interrupted items to retry_wait. On cancellation, stop undispatched items and mark running work cancel-requested; late results cannot create drafts or writes. Completed items and accepted drafts remain intact.

Use at-least-once dispatch semantics with bounded attempts. An external provider may charge twice if a request was processed but its response was lost; neither SQLite nor an application idempotency key guarantees exactly-once provider billing. Capture provider request IDs where available and avoid opaque SDK retry loops layered over worker retries.

### 8.3 Proposed initial defaults

Defaults are operator-configurable and must be tested: global concurrency 2; per-user concurrency 1; maximum 500 assets per submission; 120-second provider timeout; 180-second renewable lease; 30-second heartbeat; at most 3 provider dispatches per item in total. A schema-repair call counts toward that total. Poll visible progress every 3 seconds and back off when idle.

Retry transient 429/5xx/network failures with jitter and bounded `Retry-After`. Treat authentication, disallowed egress, unsupported image format, refusal, and persistent invalid output as nontransient unless an explicit capability fallback applies. Pause new provider dispatches on repeated systemic failures rather than consuming every item.

Enforce hard call/token limits. Monetary budgets are estimates when tariffs or usage are incomplete. Reserve estimated concurrent cost before dispatch and stop launching new requests at the cap; in-flight provider charges cannot be canceled reliably. Do not label a monetary estimate as an exact billing ceiling.

Deduplicate repeated HTTP submissions by an application idempotency key. A content/context hash supports explicit reuse of unchanged results, but never caches across users or bypasses a requested reanalysis. Detect changed image versions before using a previously staged result.

## 9. Proposed application API

Paths below are relative to the Go backend. The default frontend prefix is `/api/backend`, rewritten by `src/middleware.ts` to `BACKEND_URL`; `src/utils/client.ts` also supports a configured client prefix. Register AI routes on the existing protected mux and use the existing frontend client convention. No `/ai/` routes exist in this checkout. [R04, R18]

| Method and path | Contract intent |
|---|---|
| `GET /ai/providers` | List this user's redacted profiles and capability summaries. |
| `POST /ai/providers` | Create a validated private profile. |
| `PUT /ai/providers/{id}` | Create a new profile revision; optimistic concurrency required. |
| `POST /ai/providers/{id}/test` | Bounded synthetic-image capability test; explicit user action. |
| `POST /ai/selection-preview` | Resolve explicit IDs or current filters to a frozen selection token, counts, exclusions, and expiry. |
| `POST /ai/jobs` | Submit token, provider revision, mode, languages, consent, hint, and limits; return 202 plus job ID. |
| `GET /ai/jobs` / `GET /ai/jobs/{id}` | Paginated jobs and item progress. |
| `POST /ai/jobs/{id}/cancel` | Stop future dispatch and request cancellation of active items. |
| `POST /ai/jobs/{id}/retry-failed` | Create a new bounded retry run; retain the original history. |
| `GET /ai/results` / `GET /ai/results/{id}` | Tenant-scoped history, outcomes, provenance, and source records. |
| `POST /ai/results/{id}/draft` | Accept/edit locally; never call Immich mutations. |
| `PATCH /ai/drafts/{id}` | Change selected fields, candidate, or scope with `If-Match` revision. |
| `POST /ai/drafts/{id}/translate` | Recreate selected language entries from the reviewed fact revision. |
| `POST /ai/write-previews` | Read current Immich values; create a time-limited exact write plan and confirmation digest. |
| `POST /ai/write-operations` | Confirm the plan; enqueue the immutable approved operation. |
| `GET /ai/write-operations/{id}` | Per-asset, per-step outcomes and verified values. |
| `POST /ai/write-operations/{id}/retry` | Reconcile and retry only unfinished steps under the same valid approval. |
| `DELETE /ai/results/{id}` / `GET /ai/results/{id}/export` | Apply retention/deletion policy or export authorized local information. |

Provider enable/disable can use the profile update contract. Result deletion returns a conflict while related work is active or a retained write audit prevents full removal; allow redaction of nonessential result content without falsifying history.

New APIs use typed camelCase DTOs to fit the frontend. The provider-facing schema uses snake_case; conversion is explicit. Error responses contain a stable code, safe message, retryability, and request ID. Suggested codes include `ASSET_UNAVAILABLE`, `PROVIDER_AUTH_FAILED`, `VISION_UNSUPPORTED`, `INVALID_AI_OUTPUT`, `BUDGET_EXCEEDED`, `DRAFT_CONFLICT`, `IMMICH_CONFLICT`, and `WRITE_PARTIAL`.

Enforce session authentication, origin/CSRF protection for mutations, bounded bodies, and ownership on every referenced ID. Never accept a result ID from one user and a draft or provider from another. A selection token is time-limited and user-bound, not an authorization bypass.

Session authentication and request/query bounds exist. Session cookies use `HttpOnly` and `SameSite=Lax`; explicit origin/CSRF validation was not found in the inspected request path and must be added for the proposed mutation contract. The global body cap is 10,000,000 bytes; this is distinct from the proposed 10 MiB server-to-provider image limit and does not establish image decode safety. Existing `ensureAssetVisible` checks local lookup, not the AI type/hidden-policy rules or a fresh upstream permission read. [R04, R05, R23]

## 10. Review and frontend integration

Add “AI Locate” to existing selection actions. Preserve current keyboard selection and manual/GPX workflows. Add an AI-state filter and a dedicated results view within the existing shell, not a replacement browsing implementation.

Add a review panel to the existing image/map experience. Render candidate camera and subject markers differently, use a circle only when a radius exists, and show a heading sector only when direction is supported. The map overlay is draft state; it must not overwrite the synchronized asset coordinates until writeback is verified.

Introduce `source: 'ai'` for staged map coordinates and reference a server draft ID/revision. Do not stuff descriptions or the complete evidence bundle into `TPendingLocation`. A small adapter routes AI-origin save actions through the approved AI write operation. Existing manual changes retain their established path until separately refactored. [R14]

Do not allow a mixed manual/AI save bar to send AI drafts through the old coordinate-only handler. Show separate groups if necessary, or unify the presentation behind a typed dispatcher. When an unsaved manual location and an AI proposal target the same asset, present the conflict; neither silently replaces the other.

Refresh catalog counts, map markers, and AI status after verified writes. Preserve failed drafts. Render generated text as escaped text or sanitized Markdown without executable HTML. Source links use safe URL schemes and do not auto-fetch arbitrary model-provided URLs.

## 11. Immich writeback contract

### 11.1 Versioned adapter

The inspected Immich v3.2.2 DTO accepts GPS pairs and a description string, plus object-valued custom metadata. Camera direction is not a standard update field. Its controller retains deprecated PUT updates and implements PATCH alternatives. The adapter must use a tested version-specific route, not guess from generated documentation alone. [I01, I02]

Read standard values with `GET /api/assets/{id}`. For the v3.2.2 adapter, the proposed standard-field call is `PATCH /api/assets/{id}` with only approved fields. A documented compatible PUT profile may be selected for another tested deployment. Never try a second mutation route automatically after an ambiguous timeout.

A capability check may use read-only requests and declared version support. Any write verification uses disposable fixtures with explicit authorization; do not “test” permissions by rewriting an arbitrary user's real photo.

### 11.2 Approval protocol

The write-preview step fetches fresh values and builds an immutable plan containing exact target IDs, intended field values, relevant before-values, draft revision, source-image fingerprint, optional mirror choice, user, instance identity, and expiry. Return a digest of canonical serialized plan data.

The confirmation request includes the preview ID, matching digest, and application idempotency key. The backend validates the current draft revision and records approval durably before dispatch. A matching browser flag alone is not approval for a different plan.

Immediately before each target write, reread that target. If relevant values differ from the baseline, stop that target with `IMMICH_CONFLICT`. If the source image changed materially, require a new review. Serialize conflicting add-on writes per target; reject overlapping active plans. External clients can still race between read and write because no atomic conditional-update guarantee has been established. Document this residual limitation rather than claim complete race prevention.

### 11.3 Exact targets and field rules

The AI path writes only the explicitly approved list. Default scope is the analyzed asset; optional stack expansion is resolved before confirmation and frozen. Do not reuse implicit stack expansion. Every extra target needs independent access and baseline checks. GPS propagation can be approved for a stack; descriptions and direction remain per-image unless separately reviewed.

Build one standard update payload per target to combine selected GPS and description values where supported. GPS is an all-or-nothing pair. Unsupported or absent fields are omitted, not set to zero or fabricated values. Updating GPS does not imply replacing descriptions, timestamps, ratings, favorite flags, or other metadata.

Description policy is `preserve`, `replace`, or `managed_append`. Managed append uses a deterministic block marker and hash stored in the audit. Replace a previously owned block only if its current content still matches the recorded version; otherwise raise a conflict. Default standard description is the chosen primary language. All translations remain local.

### 11.4 Optional extended metadata

When enabled and supported, mirror a compact reviewed payload through `PUT /api/assets/{id}/metadata` under key `immich-places-ai-addon`. Include schema version, camera precision, reviewed direction, chosen place, translations, and minimal provenance. Do not mirror API secrets, private user hints, neighboring photographs' coordinates, or raw prompts. Explain that asset readers may be able to read mirrored metadata.

This is an optional second step, not a transaction with the standard update. The inspected service upserts custom metadata separately and checks asset-update access. Its standard update path can queue sidecar work, so successful API readback does not certify completion of all filesystem-side processing. [I03]

Do not assume custom JSON will be visible in Immich's standard details UI or exported into a camera-direction EXIF field. Local storage is authoritative for the add-on's complete record.

### 11.5 Readback, retries, and rollback limits

After each mutation, read the fields back. Compare coordinates with a documented numerical tolerance, not string equality; compare description content and namespaced metadata semantically. Update the local catalog only after confirmation. If Immich succeeded but the local update failed, retain the write operation and reconcile instead of resending blindly.

A lost response moves the target to verifying. Read current state: intended values mean success; unchanged baseline permits a bounded retry; a different value means conflict. For multi-step work, retry only the incomplete step. A batch may have successful, conflicting, and failed targets without pretending to be atomic.

The existing frontend `saveAssetLocationsWithRetry` performs a second pass for failures, and the backend Immich factory supplies a shared retrying transport configured with `RetryMax = 3`. AI writes must bypass the manual save helper and use a mutation transport with automatic retries disabled, so the durable writer can reconcile an ambiguous outcome before any resend. Reads and provider calls need separately explicit, bounded retry policies; do not multiply hidden transport retries by worker attempt budgets. [R19, R20]

Expose pre-save discard and a complete before-value audit. Full restoration to missing GPS is not implemented unless an explicitly tested API supports clearing coordinates; the inspected numeric schemas do not accept null. Do not encode missing GPS as `0,0`. [I04]

## 12. Security, privacy, and lifecycle

Recheck access before image fetch, context assembly, review, and write. A local cached asset is not permanent authorization after sharing or API-key permissions change. Provider tests never use a private library image. Egress consent is versioned and does not expand when a user changes filters later.

Treat image text, album names, hints, web pages, and model content as untrusted data. System instructions and application policies are not editable by those inputs. The structured response can propose facts but cannot alter limits, select another user, call arbitrary tools, or authorize a write.

Persist the minimum context needed for reproducibility: hashes, selected context classes, and bounded evidence records. Proposed retention defaults are 30 days for unreviewed results, 7 days for sanitized terminal-error details, immediate deletion of temporary image copies, and retention of confirmed write history until explicit user/admin action. Accepted drafts remain until saved, rejected, or deleted. These defaults require owner approval before release.

Delete user-associated secrets, queued work, and private results on account deletion under a defined cascade/redaction policy. Local result deletion does not undo Immich metadata; offer a separately confirmed supported deletion of the namespaced mirror only when requested. No default telemetry sends images, locations, or prompts outside the configured processing services.

## 13. Deployment and observability

Keep the current two-service Compose topology. Add backend-only settings such as `AI_ENABLED`, global concurrency, request/image limits, egress allowlists, and retention controls. User-specific endpoints/models/secrets live in private provider profiles; no secret uses a `NEXT_PUBLIC_*` environment variable.

Start AI disabled until migrations and configuration checks succeed. Shut down workers gracefully, stop claiming new items, and release or expire unfinished leases safely. Run one backend writer instance against a local SQLite volume in V1. A second process sharing that volume across hosts is unsupported.

Expose queue depth, outcome counts, retry/error rates, latency, token usage, and verified/partial writes as aggregate metrics. Use request/job/item IDs in logs, not filenames, coordinate values, image content, or full provider responses. Disabling AI must leave the rest of Places functional.

Preserve upstream notices and document the fork's changes. The inspected repository carries an MIT license and copyright notice; retain them rather than replace the upstream attribution. [R16]

## 14. Verification and release strategy

Follow the adopted [testing standard](../engineering/testing.md) for test layers, TDD, coverage and verification gates. Unit tests cover schema and semantic validation, consent filtering, coordinate boundaries, language coverage, revision checks, result-state transitions, managed descriptions, and target manifests. Repository/migration tests use real temporary file-backed SQLite with actual migrations, including reopen/upgrade behavior; in-memory fixtures alone do not establish durability.

At the pinned checkout, 16 Go test files contain 274 top-level `Test*` declarations. `newTestDB` uses a temporary database and real migrations; `TestHandleUpdateLocationWithStack` already exercises local stack propagation. These are source inventory findings, not passing test results. No frontend test script or frontend test files were found. The release workflow and Dockerfiles build artifacts but do not run the Go suite or lint explicitly. The [engineering tooling prerequisite](OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite) owns the initial harnesses and automated gates, before feature implementation. Go was unavailable during reconciliation, so the existing suite remains unexecuted. [R02, R22, R25]

Concurrency/recovery tests cover lease expiry, late responses, restart during requests, duplicate submit, cancellation, budget reservations, overlapping drafts, and local persistence failure after Immich success. Provider fixtures cover strict JSON, JSON-only, malformed output, refusal, truncation, unsupported options, 401/429/5xx, and unknown usage.

Immich integration tests use a separately configured disposable library: one GPS-less image, one geotagged image, existing descriptions, a stack with mixed coordinates, inaccessible assets, and relevant external-library modes. Record the Immich version and verified methods. Validate standard fields, optional metadata, partial write, sidecar caveats, and inability to clear GPS where applicable.

End-to-end tests implement PRD AC-01–12 and confirm that Accept causes zero Immich mutations. Assert the exact IDs of every outgoing write, not merely the number of API requests. Add UI regression checks for existing manual placement, GPX import, selection, map refresh, and AI-disabled operation.

For model evaluation, hide ground-truth GPS and disallowed filename/album clues without modifying originals. Separate Visual from Context-assisted evaluation, split by trip, and report coverage alongside accuracy. Use independently verified camera coordinates rather than POI centroids. Store no benchmark photographs in the public fork without permission.

## 15. Decisions and implementation gates

| ID | Decision / remaining gate |
|---|---|
| ADR-01 | Reuse Go/Next.js/SQLite; no new service in V1. |
| ADR-02 | AI creates proposals only; a separate confirmed writer owns mutations. |
| ADR-03 | Immutable results plus durable, versioned drafts; no direct model-output write. |
| ADR-04 | Explicit asset targets; no implicit stack propagation in the AI path. |
| ADR-05 | Standard GPS/description write required; extended metadata mirror is optional. |
| ADR-06 | V1 includes direction and multilingual descriptions; advanced research is separately scoped. |
| [ADR-07](../engineering/decisions/ADR-07-ai-internal-packages.md) | Accepted: new AI core uses focused internal Go packages; keep legacy services behind adapters and defer unused Research interfaces. Update the Docker build with the first subpackage implementation. |
| GATE-01 | Complete for `5e70c6165777949c9d8b50ede3b2768bcaa5df87`: source paths, interfaces, migrations, selection/save behavior, and test inventory reconciled in `RECONCILIATION.md`. Reopen if the implementation checkout changes; this is not a runtime-test pass. |
| GATE-02 | Record deployed Immich version and validate API methods, rights, and sidecar/library behavior. |
| GATE-03 | Complete for the [19 September 2026 NAS configuration](../engineering/nas-provider-capability-2026-09-19.md): existing codex-proxy with `gpt-5.6-sol` passed synthetic image/JSON/strict samples. Other configurations and full analysis-contract compatibility require separate validation. |
| GATE-04 | Approve retention defaults, network allowlists, and quality benchmark fixture. |

OpenSpec should derive capability specifications, scenarios, and implementation tasks from this baseline, not treat proposed filenames, numeric limits, or unresolved gates as already implemented facts. See `OPENSPEC_HANDOFF.md` for the planning boundary.

## References

Reference IDs resolve in `SOURCES.md`. External facts and inspected code are cited inline; module names, APIs, schemas, limits, and decisions labeled proposed are the design for this fork.
