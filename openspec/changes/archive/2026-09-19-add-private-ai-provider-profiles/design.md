## Context

The existing backend composes SQLite, authentication and routes in `package main`; AES-GCM helpers exist, but legacy decryption accepts plaintext. The frontend has a settings panel, shared dialog shell and `/api/backend` proxy. No AI tables, API or UI exist. The adopted architecture requires focused internal AI packages and the first package requires Docker packaging and backend coverage enforcement.

## Goals / Non-Goals

**Goals:** Private, durable, encrypted provider configuration; immutable revisions with optimistic concurrency; default-off installation control; protected settings mutations; accessible creation/edit/disable UI; explicit credential cleanup and migration/reopen evidence.

**Non-Goals:** Destination approval, provider HTTP transport, capability tests, model discovery, custom headers/options, photos, jobs, analysis, results or writeback. CH02 must approve destinations before any dispatch; CH03 tests capabilities. Saving here conveys neither approval nor compatibility.

## Decisions

### Own validation in a small core package and persistence in integration adapters

`backend/internal/ai/providers/` owns pure configuration types, validation and stable error identities. Focused `backend/aiProvider*.go` files adapt SQL, encrypted secrets and authenticated HTTP using existing unexported services. No all-purpose AI store, provider client or unused job abstraction is introduced. The frontend exposes one settings component through `src/features/ai/index.ts`; its API, hook/state and form stay in the AI feature.

Profile input contains name, base URL, manually entered model, enabled state and optional secret. Bound names to 80 characters, models to 200, URLs to 2048 and secrets to 4096 bytes. Require an absolute HTTP(S) base URL without userinfo, query or fragment; reject unknown transport options instead of silently accepting them. URL syntax validation is not egress authorization. Empty secrets support unauthenticated endpoints. Normalize surrounding name/model/URL whitespace and trailing URL slash; never trim secret bytes.

### Persist revisions atomically and keep credentials out of public data

Migration `018` adds logical `ai_provider_profiles` and immutable `ai_provider_versions`, both scoped by user. Composite keys/foreign keys prevent cross-user version association; user deletion cascades through profiles and versions. Every pooled SQLite connection enables foreign keys and the existing busy timeout through driver DSN options, verified on simultaneously held connections. Existing catalog data is untouched.

Create starts at revision 1. Every successful update, including enable/disable, requires `expectedRevision`, advances the logical active revision and inserts a complete version in one transaction. Compare-and-swap acquires the write before reading the previous version, avoiding a read-to-write upgrade race. Stale edits return 409; another owner's ID is indistinguishable from a missing profile. Logical disablement applies to the profile across all revisions. Future dispatch must check this logical state; there is no dispatch consumer in CH01.

Reuse AES-GCM encryption, but reject a stored AI credential without the encryption prefix or with failed authentication; never inherit the legacy plaintext fallback. Responses contain only a `hasSecret` flag. Omitted secret on edit retains the current secret only when the base URL is unchanged; changing the destination requires explicit replacement or removal. A replacement creates a new encrypted revision and preserves old revision credentials for revision-bound work.

Explicit secret removal is a security erasure operation: purge encrypted secret material from every revision of that profile while retaining revision metadata. This is the only mutation of old version content, besides account deletion. Disablement retains encrypted credentials for re-enable; the UI offers removal separately. Backups are not rewritten by erasure. No account-deletion endpoint exists yet: enforce and test the database cascade without inventing an unrelated account UI. Later retention/jobs must preserve these boundaries and cannot restore a removed credential from another revision.

### Protect settings at the authenticated boundary

`AI_ENABLED` defaults to false. When enabled, require `AI_PUBLIC_ORIGIN`, a canonical frontend HTTP(S) origin without path/query/fragment/userinfo. This explicit origin avoids trusting arbitrary forwarded-host headers through the existing proxy. Mutations require a matching non-null `Origin`, JSON content type, a bounded 16 KiB body, strict known fields and exactly one JSON object. Missing/cross-origin requests fail before persistence. Global disablement returns a redacted disabled state from authenticated listing and rejects all mutations; legacy routes remain unchanged.

Add `GET /ai/providers`, `POST /ai/providers` and `PUT /ai/providers/{id}` under session middleware. The list returns `{enabled, items}`. Create/update return a redacted profile with ID, revision, name, baseURL, model, enabled and hasSecret. Errors use stable code, safe message, retryability and request ID. Do not log credentials or request bodies. Profile operations have no external client, so creation/edit/disable cannot send any provider or Immich request.

### Make settings explicit and recoverable

The existing settings panel opens an AI provider dialog. It explains disabled installations, empty lists and load failures; enabled installations can create or edit a profile and toggle enabled state. Inputs have labels, secret input is a password, model entry is always manual, and a saved secret is never prefilled. Blank edit secret means retain; an explicit removal control clears all retained profile credentials. A destination change explains the replacement/removal requirement.

Show revision/conflict errors without silently overwriting newer settings or automatically retrying a mutation. Preserve non-secret edits after errors; clear entered secrets after each submission and on cancel/close. Cancel causes no mutation. Reopening reloads authoritative profiles. No global cache or local storage holds provider settings/secrets. Use the shared dialog keyboard/focus behavior and a pending state that prevents duplicate saves. No connection-test button appears before CH03.

### Verify the first AI package and deployment

Use real migrated SQLite for fresh/v17 upgrade/reopen, encryption, version immutability, stale and competing edits, isolation, erasure and account cascade. HTTP tests cover auth/origin/body limits/errors/default-off and prove zero outbound calls against a local synthetic destination. Frontend tests cover load/disabled/error/create/edit/disable/cancel/conflict and secret handling. Add a deterministic browser journey for real proxy/session/persistence integration with AI enabled, while retaining separate evidence for disabled legacy journeys.

The shared backend test gate runs race tests once with a complete coverage profile and enforces at least 80% statements across new AI core and integration/adapters, including untested AI source. The existing frontend AI floor measures all AI files at 80% lines/branches. Checker tests prove below-floor and missing/invalid measurements fail. Update the Go build target for the executable with subpackages and copy `internal/` into the backend Docker build. Run the full shared checks and container build.

Deployment is additive and defaults off. Document `AI_ENABLED`, `AI_PUBLIC_ORIGIN`, key/DB backup and credential retention. Rollback runs the previous executable with AI disabled while retaining the new tables; do not use destructive down migration as a routine rollback. No architectural departure or new dependency is required.

## Risks / Trade-offs

- Explicit origin configuration adds setup but avoids ambiguous proxy trust; verify the real browser proxy path.
- Old revision credentials remain encrypted after rotation/disable until explicit erasure/account deletion; disclose this and retain a separate erasure action.
- Per-connection foreign-key enforcement can reveal inherited invalid operations; run the complete backend suite, preserve migration semantics and fix only demonstrated integration regressions.
- Coverage floors apply to actual new AI source; retain meaningful failure-path tests instead of assertions manufactured for a percentage.
- Saved endpoints remain unverified until CH02/CH03; no outbound transport is reachable in this change.
