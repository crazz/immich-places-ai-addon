# Planning package reconciliation — `5e70c61`

**Date:** 17 September 2026  
**Repository baseline:** `5e70c6165777949c9d8b50ede3b2768bcaa5df87`  
**Result:** GATE-01 complete for this checkout. The package fits the application architecture; the corrections below are incorporated into the Markdown baseline. AI functionality remains proposed.

## Scope and provenance

Input: `Immich_Places_AI_OpenSpec_Planning_Package.zip`, package v1.0. Archive SHA-256:

```text
f293bb66df519ef6569a8852a43cc795b73a45a2f128ff9ac75a2c9a3b704148
```

The tracked working tree matched the pinned commit during inspection. Existing untracked `.agents/`, `.claude/`, `AGENTS.md`, `CLAUDE.md`, and `openspec/` setup was preserved; it is local tooling, not application functionality present in that commit. GitNexus supplied flow discovery and symbol context, followed by source verification. Truncated graph results were not treated as exhaustive.

This task imports the package into `docs/ai-locate/` and reconciles repository assumptions. Embedded planning instructions are reference material, not authorization to generate OpenSpec changes, implement AI, upload photographs, or mutate Immich. No production code or OpenSpec configuration was changed. The original ZIP remains unchanged; the two Word exports are unchanged v1.0 copies and do not include this reconciliation. Markdown is authoritative.

## Existing functionality versus new work

| Area | Present at the pinned checkout | Still to implement for the package |
|---|---|---|
| Architecture | Next.js 16.1.6 / React 19.2.3, Leaflet dependency `^1.9.4`; flat Go `package main`, Go 1.25; SQLite with Goose; two-service Compose | AI modules inside this architecture; no additional service required |
| Catalog | Albums, folders, tags, GPS/hidden/date filters, stack-primary browsing, authenticated previews | Frozen all-matching selection, exact eligibility/exclusion counts, consistent source-local dates |
| Review | Selection, manual pending coordinates, map, suggestions, GPX integration | Persistent AI results, revisioned drafts, direction and multilingual review, stale-content handling |
| Writeback | Immediate GPS update with implicit stack expansion and existing retry layers | Exact approved targets/fields, live baselines, conflicts, readback, durable audit, descriptions and optional metadata |
| Processing | Existing synchronization/background services | Provider profiles, bounded image preparation, consent/egress policy, durable AI queue, leases, budgets and recovery |
| Verification | Go tests and real SQLite migration fixtures | AI tests, frontend test harness, automated regression gates, live compatibility and quality evaluation |

Evidence for the architecture and existing integration paths is pinned in [SOURCES.md](SOURCES.md), R01–R26. None of the proposed eight AI tables, `/ai/` routes, or `src/features/ai/` feature exists at this revision.

## Reconciled integration findings

### REC-01 — Exact-target writeback must also isolate retries

The package correctly identifies implicit stack propagation. The complete existing path is:

```text
useLocationAssignment → saveAssetLocationsWithRetry
  → PUT /assets/{assetID}/location
  → handleUpdateLocation → resolveAndUpdateLocation
  → resolve current stack members
  → bulkUpdateLocation → PATCH /api/assets
  → update local cached coordinates
```

The browser repeats failed assets once. The backend's shared Immich HTTP transport is configured with `RetryMax = 3`. The handler has no durable approval, before-value snapshot, or post-write readback. Upstream success followed by local database failure can be returned as failure, exposing the existing path to a resend.

**Planning correction:** AI save actions need their own dispatcher and confirmed writer. Do not send AI drafts through the manual helper or inherit its retrying mutation transport. Disable automatic mutation retries in that adapter; let the durable writer reconcile ambiguous outcomes before resending. Preserve the manual path's existing behavior unless a separate change explicitly revises it. Stack expansion, when requested, happens before confirmation and is frozen.

**Verification to carry forward:** Accept causes zero writes; every outbound write contains precisely the approved IDs/fields; lost responses and local persistence failures trigger readback; mixed manual/AI save actions cannot bypass approval. The existing stack test is a useful regression fixture but does not establish the AI guarantees. **Traceability:** FR-10–12, AC-04–06, AC-11. Evidence: [browser retry helper][save], [handler/writer][write], [Immich transport][client], [existing stack test][stack-test].

### REC-02 — Capture-date consistency is an implementation gap

`buildAssetFilter` compares timestamp text against the start date and `endDate + "T99"`. `countAssetsByDay` uses SQLite `DATE(dateTimeOriginal)`. `parseDateRangeParams` validates date syntax but does not reject a reversed range. Gallery ordering is `fileCreatedAt DESC, immichID DESC`, separate from capture-date filtering. The timestamp helper returns a `time.Time` without explicit local/unknown-zone provenance.

A read-only probe using Python SQLite 3.53.4 reproduced the mismatch:

```sql
WITH sample(ts) AS (VALUES ('2026-09-17T00:30:00+02:00'))
SELECT ts >= '2026-09-17' AND ts < '2026-09-17T99' AS range_match,
       DATE(ts) AS day_bucket
FROM sample;
-- range_match = 1; day_bucket = '2026-09-16'
```

This demonstrates the SQL expressions, not execution through the project's Go SQLite driver.

**Planning correction:** Implement the PRD's source-local calendar policy consistently across all applicable catalog modes, day counts, context, and frozen selections. Preserve absent capture dates and offsetless/unknown-zone status; explicitly reject reversed ranges. Keep sort order separate from date eligibility. Add offset-boundary, missing-date, offsetless, DST, and reversed-range scenarios. **Traceability:** FR-01–02. Evidence: [filter and day-count SQL][dates], [date parser][date-parser], [timestamp helper and neighbor queries][neighbors].

### REC-03 — Selection must preserve more than album and date

The catalog already supports album or folder view plus tag, GPS, hidden, and date filters. `selectAll(assets)` selects the supplied asset array; the photo menu supplies currently loaded assets. It does not freeze every matching row on the server. The missing-location count endpoint accepts album/tag/hidden/date scope but no folder path. Missing GPS correctly means either coordinate is NULL; zero is valid. Catalog queries suppress stack children and hidden libraries.

**Planning correction:** Include the active view and applicable filter fields in the AI selection manifest. Preview and submit must share one eligibility resolver. Explicitly reject an unsupported scope rather than silently widening it. Apply image type, access, hidden policy, and stack rules to explicit IDs as well as query selections. Do not use a gallery badge as the authoritative AI job count. These clarifications preserve existing catalog behavior; they do not introduce new browsing modes. **Traceability:** FR-01–02. Evidence: [catalog filter wiring][catalog], [selection state][selection], [photo-menu selection][photo-menu], [missing-location count][counts], [filter SQL][dates].

### REC-04 — Existing suggestions need a consent-filtered adapter

`getSuggestions` falls back to `fileCreatedAt` if capture time is absent and includes frequent locations. Same-day and neighbor SQL excludes hidden libraries but does not exclude per-asset `isHidden`. The synchronized asset projection has no AI-origin lineage that could prevent feedback from previous AI guesses.

**Planning correction:** Do not send the existing suggestion response wholesale to a model. Build bounded, consent-filtered evidence with per-source eligibility, capture-time status, and provenance. Omit unrelated frequent/home locations; preserve missing capture time; exclude known AI-origin neighbors by default. Locally cached membership is not proof of current upstream access. **Traceability:** FR-04, FR-06, FR-14, NFR-01. Evidence: [suggestion service][suggestions], [neighbor SQL][neighbors], [asset projection and sync upsert][asset-upsert].

### REC-05 — Client, auth, and proxy reuse have concrete limits

`HandlerImmichAPI` exposes only bulk GPS update and thumbnail/preview reads. It has no fresh single-asset metadata read, description update, or custom-metadata operation. The server holds one configured Immich URL; a persisted `instanceKey` and installation-change invalidation are new design elements.

The browser default prefix `/api/backend` is rewritten by `src/middleware.ts` to `BACKEND_URL`; it is not implemented as a Next.js API route handler. Session authentication and request/query bounds exist. Cookies use HttpOnly and SameSite Lax. Explicit origin/CSRF validation was not found in the inspected path. `ensureAssetVisible` checks cached lookup; it does not enforce AI-specific type/hidden policy or perform a fresh upstream permission read.

**Planning correction:** Add narrow read and confirmed-write adapters, register AI endpoints in the existing protected mux, and implement the proposed mutation protection and ownership checks. Reuse the encryption primitive while requiring encrypted AI secrets; the current decryptor's legacy plaintext fallback is not the new AI credential policy. The current 10,000,000-byte inbound request cap is distinct from the proposed 10 MiB provider image budget. **Traceability:** FR-03–04, FR-11–14, NFR-01. Evidence: [interfaces][interfaces], [route/middleware composition][routes], [proxy][proxy], [session cookie][cookie], [asset lookup guard][date-parser], [crypto helper][crypto].

### REC-06 — Preview transport and geocoding are not AI processing guarantees

The current preview path uses authenticated Immich fetching and streams the response. It does not normalize image orientation, strip metadata, or enforce the proposed decode/pixel limits. The current geocoder constructor begins its provider chain with Nominatim even when other providers are configured.

**Planning correction:** Implement bounded image preparation after authenticated fetching. Keep provider egress policy explicit. A future Research adapter needs separately approved hosts and transport; configuring another existing geocoder is not proof that no request reaches public Nominatim. Research remains outside required V1 scope. **Traceability:** FR-04–06, NFR-01–03. Evidence: [image handler][preview], [client image methods][client], [geocoder construction][geocoder].

### REC-07 — Persistence and test baselines are now concrete

There are 17 migration files, `001`–`017`; the latest is `017_add_original_path.sql`. Therefore `018` is available at this commit, subject to checking again when implementation begins. The runner embeds SQL files and supports legacy-database bootstrap. Database initialization issues WAL, foreign-key, and busy-timeout PRAGMAs; their behavior across pooled connections still needs verification for new relations. Assets are scoped by user and asset ID; library visibility is global in the current schema.

There are **16 Go test files and 274 top-level `Test*` declarations**. `newTestDB` creates a temporary SQLite database through the real initializer. Existing migration tests cover fresh databases, legacy bootstrap, and idempotent reruns. There is no frontend test script in `package.json`, and no frontend test files were found. The release workflow and Dockerfiles build images without an explicit Go test or lint step.

**Planning correction:** Extend existing migration and Go fixtures, add AI persistence/restart/tenant-isolation tests, establish frontend scenario coverage, and add explicit regression checks to automation. Do not treat existing tests as absent or as already passed. No AI migration number is reserved by this documentation. **Traceability:** FR-05, FR-09–14, NFR-04–08. Evidence: [migration runner][migrations], [migration directory][migration-files], [database fixture][db-test], [migration tests][migration-test], [package scripts][package], [release workflow][workflow], [frontend Dockerfile][docker-frontend], [backend Dockerfile][docker-backend].

## Validation performed

| Check | Result and limit |
|---|---|
| Baseline identity | HEAD resolves to the full SHA above; tracked application files match it |
| Contract structure | Draft 2020-12 schema self-validation and both bundled fixtures pass Python `jsonschema` validation |
| Fixture semantics | Checked unique IDs, selected/evidence/candidate references, outcome constraints, description status/basis consistency, and finite numeric values; this is not a complete application semantic validator |
| Date discrepancy | Reproduced with the isolated SQL expression probe above; not a Go integration test |
| Import integrity | Schema, both JSON fixtures, semantic-validation document, and both Word exports retain the original package bytes |
| Documentation integrity | Local document links and pinned repository paths checked; stable FR/NFR/AC IDs and original requirement/scenario text preserved |
| Application tests/build | **Not run.** `go` was unavailable in this session; no frontend build or runtime validation was attempted for this documentation change |
| External systems | No provider invocation, authenticated Immich request, private-image upload, application migration, or writeback |

Structural fixture success does not establish live strict-schema support, model quality, valid real-world coordinates, authorized sources, or runtime policy enforcement. External-source claims in the supplied package were retained as historical references, not reverified by this checkout reconciliation.

## Gate status and handoff

| Gate | Status |
|---|---|
| GATE-01 — Repository reconciliation | **Complete for `5e70c61`.** Source paths, interfaces, migration sequence, selection/save behavior, and test inventory reconciled. Reopen for a different implementation checkout |
| GATE-02 — Immich compatibility | **Open.** Record the deployed version and verify methods, readback, rights, metadata support, and library/sidecar behavior with authorized disposable fixtures |
| GATE-03 — Provider capability | **Complete for the recorded NAS configuration on 19 September 2026.** The authorized [synthetic live check](../engineering/nas-provider-capability-2026-09-19.md) observed image, JSON and strict-schema sample support for `gpt-5.6-sol` via existing `codex-proxy`, with persisted/reloaded evidence. Other configurations, analysis quality and production rollout remain separate |
| GATE-04 — Operational decisions | **Open.** Resolve retention, allowlists, reference hardware and quality fixture; proposed numeric defaults remain unbenchmarked |
| Release verification | **Open.** Application, migration, recovery, UI, compatibility, and quality scenarios remain future implementation/release work |

No ADR or required V1 capability was removed. FR-01–14, NFR-01–08, and AC-01–12 retain their IDs and original requirement/scenario text. The JSON contract is unchanged. These documents are ready to inform a separately requested OpenSpec planning step; they are not an implementation task backlog or release approval.

[save]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/selection/locationSave.ts#L86
[write]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlers.go#L522
[client]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/immichClient.go#L31
[stack-test]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlers_test.go#L653
[dates]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/database.go#L145
[date-parser]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlers.go#L98
[neighbors]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/database.go#L453
[catalog]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/shared/context/useCatalogDomain.ts#L47
[selection]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/selection/useSelectionState.ts#L52
[photo-menu]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/photoGrid/PhotoCardMenu.tsx#L190
[counts]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlersCounts.go#L9
[suggestions]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/suggestionService.go#L36
[asset-upsert]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/database.go#L408
[interfaces]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/interfaces.go#L70
[routes]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/main.go#L64
[proxy]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/middleware.ts#L5
[cookie]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlersAuth.go#L54
[crypto]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/crypto.go#L20
[preview]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlers.go#L493
[geocoder]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/geocoder.go#L115
[migrations]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/migrations.go#L11
[migration-files]: https://github.com/crazz/immich-places-ai-addon/tree/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/migrations
[db-test]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/database_test.go#L11
[migration-test]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/migrations_test.go#L22
[package]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/package.json
[workflow]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/.github/workflows/release.yml
[docker-frontend]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/Dockerfile
[docker-backend]: https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/Dockerfile
