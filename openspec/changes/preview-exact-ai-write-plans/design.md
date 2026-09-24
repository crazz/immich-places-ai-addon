## Context

See [proposal.md](proposal.md). CH16 is a planned dependency, not implemented code at this planning revision. Its draft identity, optimistic revision, reviewed-image identity and acknowledged GPS baseline form this change's input contract. Apply and synchronize CH16 before starting CH18.

Grounding at `0fb8eaf` used GitNexus history/image/manual-save queries and source verification of `backend/aiResultImage.go`, `aiImageMetadata.go`, `aiImageTransport.go`, `aiResultHTTP.go` and `immichClient.go`. Current image reads are bounded and separately authorized; the general Immich client has automatic retries. The existing analysis source digest includes `updatedAt`, so the CH16 reviewed-image identity is the write-preview comparison, not the old digest. Some Go graph symbols were unresolved; source verification was required. Re-run graph impact before source edits during apply.

Follow the adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md), [coding standards](../../../docs/engineering/coding-standards.md) and ADR-07. No architecture or dependency change is needed.

## Goals / Non-Goals

**Goals:** A persisted, inspectable plan with server-owned scope and an explicit fresh-read boundary, ready for CH19 to consume atomically.

**Non-Goals:** No mutation interface, worker, approval endpoint or dependency from analysis/review to a writer. Do not reuse manual stack expansion, and do not broaden CH16 field selection.

## Decisions

### 1. The preview workflow consumes a narrow draft snapshot

Add `internal/ai/writepreview` for pure plan validation/canonicalization and a workflow consuming narrow draft-reader, metadata-reader and preview-store interfaces. This package has no mutation capability. Root `aiWritePreview*.go` adapters own HTTP, SQL and composition. A small explicit draft handoff exposes only owner/installation, ID/revision/state, analyzed target, camera pair, GPS selection, reviewed source identity and acknowledged GPS baseline. Do not serialize the whole analysis, local translations or provider settings into a plan.

Alternative: calculate a client-only diff. Rejected because it cannot bind authority or prevent revised inputs from being confirmed. Alternative: a single preview-and-write request. Rejected because CH18 must remain independently deployable without mutation.

### 2. Read current values, then revalidate before publication

`POST /ai/write-previews` accepts only `draftId` and `draftRevision`. Authenticate with the current session and existing origin checks. Require a current staged draft with `[gps]`, exactly its analyzed asset, a valid camera pair and an acknowledged baseline. Missing baseline returns `BASELINE_REVIEW_REQUIRED`, not guessed values. A stale heading, translation or wide/unknown radius does not block GPS.

Outside any SQLite transaction, resolve the current credential and local asset authority, then perform a bounded `GET /api/assets/{id}` through a dedicated read adapter. Validate returned ID/type/visibility/not-trashed status and checksum/owner identity. Parse EXIF GPS with duplicate-key/type/range checks and independent nullable latitude/longitude; absent/null EXIF GPS is missing, a wrong-shaped EXIF object is an error. Read success does not prove update permission. The adapter uses the existing configured Immich origin only, forbids redirects, bounds the response to 1 MiB and uses one attempt within a 10-second deadline. No model-supplied destination or private URL is accepted.

Compare reviewed-image identity with CH16's baseline identity. Material image change returns `SOURCE_CHANGED`. Compare nullable GPS pairs exactly after finite-number normalization; a relevant GPS difference returns `IMMICH_CONFLICT` with owned before/current/proposed values. A GPS-only update can change `updatedAt` without changing the image identity. Other metadata changes do not create an irrelevant conflict. Missing/partial GPS remains an exact before-value, never `0,0` by default. If fresh GPS already equals intended GPS, return an explicit `unchanged` diff; CH19 can record a verified no-op without a mutation.

After the read, recheck current installation, credentials, local visibility and exact draft revision/state in a short transaction and insert the plan only if still valid. A competing edit, rejection, account deletion or installation rotation must prevent publication. Remote access can change again after publication; CH19 must repeat it before dispatch. Resolve baseline conflicts through CH16's explicit baseline/current-image review, then stage the new revision and create a new preview; never silently adopt changed GPS.

### 3. Immutable canonical plans and bounded expiry

Add `ai_write_previews`, owned by account and installation, referencing the exact draft revision. Store ID, plan version, canonical payload, digest, created/expiry timestamps and invalidation state. The canonical payload contains owner, installation, draft ID/revision, source analysis ID, reviewed-image identity version/value, target ID, selected `[gps]`, nullable before pair, finite intended pair, observation time, preview ID, expiry and versioned comparison policy. Store no secrets or raw upstream response. Limit plan payloads to 16 KiB and request bodies to 4 KiB.

Use a typed fixed-field encoding with normalized signed zero and stable field order, then SHA-256 over the versioned canonical bytes. Persist the same bytes that were hashed; do not rely on map iteration order. The digest detects plan mismatch; possession of it is not authentication. Later confirmation must reload the private stored row rather than trust a client-supplied plan.

Preview validity is five minutes from creation, using injected server time. `GET /ai/write-previews/{id}` returns the stored before/after with current usable/expired/stale status and never refreshes expiry or contacts Immich. A new preview is a deliberate new read; expired plans cannot be renewed in place. Limit active previews to ten per draft and clean at most 100 expired unreferenced rows per pass; CH19 must protect consumed previews/audit from cleanup. These bounds manage temporary previews, not a new draft/audit retention policy.

Alternative: a signed plan carried solely by the browser. Rejected because durable consumption, restart recovery and owner deletion are simpler with the existing SQLite authority model. No signing library or additional secret is needed.

### 4. Show the exact diff through the AI workflow

Add typed API helpers, preview state and an accessible comparison panel under `src/features/ai/`. Show the photo identity, saved draft revision, before latitude/longitude including missing/partial values, proposed pair, single-photo scope, changed/unchanged result and validity status. Show uncertainty as contextual user guidance, with no threshold. Describe affected data as GPS only; other metadata is preserved. Maps are optional, and numeric comparison remains usable with keyboard input or failed tiles.

Keep draft editing available when preview reads fail. Distinguish stale revision, changed source, GPS conflict, expired preview and temporary source failure using concise recovery actions. Do not silently acknowledge new baselines. If the current editor changes or same-asset manual pending edits appear, retire the displayed usable preview and require resolution before continuing. Account/result changes clear private preview state and fence late responses. CH18 displays comparison only; CH19 later adds the explicit confirmation control.

Protected errors use 400 for malformed scope, 404 for foreign/unavailable IDs, 409 for stale draft/source/GPS conflict or missing baseline, 410 for an expired reference where an action needs validity, and 503 for upstream/storage failure. Private GET can return a stored expired plan with status, so expiration is understandable rather than indistinguishable from deletion. Retrying a read never creates a write operation.

### 5. Migration, lifecycle and verification

Allocate migration 027 after verifying the CH16 migration tip. Use owner-qualified foreign keys; account deletion cascades previews, while ordinary catalog reset leaves local drafts/history intact. Installation rotation invalidates all old preview authority. Referenced draft revisions cannot be pruned while a preview is retained. AI/provider execution disablement does not prevent authorized local reads or read-only preview generation; CH19 separately gates mutation dispatch.

Use short transactions without network I/O. Prove fresh/upgrade/repeated migrations, expiry, capacity and reopen behavior on real file-backed SQLite. Test failures between upstream read and insert. Back up consistently before migration; binary rollback keeps the additive tables, and rolling back does not introduce an executor. The [verification plan](verification-plan.md) maps scenarios and installed gates. Update the operator handoff to distinguish synthetic read evidence from unverified real mutation compatibility. All source/test files obey the 500-line rule.

## Risks / Trade-offs

- Upstream values can change immediately after preview → CH19 repeats authority/source/baseline checks; a preview alone never promises conflict-free execution.
- Exact GPS baseline comparison can flag tiny external edits → deliberately require renewed review instead of silently approving different before-values; readback tolerance is a separate CH19 concern.
- Local and upstream authority checks are not an atomic remote permission lock → fail closed on known changes and make the remaining boundary explicit.
- Multiple tabs can retain old previews → private stored revision/expiry checks remain authoritative, irrespective of a visible button or stale client cache.
