# Source register and verification limits

**Reviewed:** 17 September 2026

The original package used browser retrieval of public `main`. This repository copy has been reconciled with local checkout `5e70c6165777949c9d8b50ede3b2768bcaa5df87`; all repository references below are pinned to that commit. See [RECONCILIATION.md](RECONCILIATION.md) for findings and validation limits.

The source archive is `Immich_Places_AI_OpenSpec_Planning_Package.zip` (v1.0), SHA-256 `f293bb66df519ef6569a8852a43cc795b73a45a2f128ff9ac75a2c9a3b704148`. Markdown was imported and reconciled; the schema, fixtures, semantic-validation notes, and original Word exports were preserved byte-for-byte. The Word files under `exports/` do not include reconciliation changes.

No application build, Go test run, application migration, authenticated Immich call, or live writeback was completed. This task writes planning documentation only. Local schema/fixture validation and an isolated SQLite date-expression probe are recorded in the report. Proposed architecture, limits, schemas, requirements and acceptance targets remain design decisions, not implemented behavior.

External references I01–I04, O01–O03, M01–M02, and S01 retain the original package's source descriptions; they were not reverified during local checkout reconciliation. Immich references are pinned to the v3.2.2 tag. The owner's deployed server version and provider capabilities remain unverified.

## R01 — Fork README and repository overview

[Fork README and repository overview](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/README.md)

Documented features, auth/deployment overview and frontend proxy defaults. README is not a runtime test.

## R02 — Frontend dependency manifest

[Frontend dependency manifest](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/package.json)

Declared Next.js, React, Leaflet, TypeScript and build/lint scripts.

## R03 — Go module manifest

[Go module manifest](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/go.mod)

Go toolchain declaration, SQLite driver and Goose dependency.

## R04 — Go service composition and route registration

[Go service composition and route registration](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/main.go)

Protected application routes and startup wiring.

## R05 — Asset handlers and current location writer

[Asset handlers and current location writer](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlers.go)

Date/GPS parsing; preview reads; handleUpdateLocation; resolveAndUpdateLocation expands stack IDs and writes immediately.

## R06 — Local database and asset queries

[Local database and asset queries](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/database.go)

WAL setup, tenant scope, missing-GPS and source timestamp filtering, stack-primary filtering and synchronized asset projection.

## R07 — Embedded database migration runner

[Embedded database migration runner](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/migrations.go)

Goose loads migrations/*.sql. The pinned checkout contains 001–017; 018 is next at this revision only.

## R08 — Credential encryption helpers

[Credential encryption helpers](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/crypto.go)

AES-GCM helpers and legacy plaintext-compatible decryption branch.

## R09 — Application page composition

[Application page composition](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/app/page.tsx)

Existing authenticated shell, photo list, and map integration.

## R10 — Frontend application contexts

[Frontend application contexts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/shared/context/AppContext.tsx)

Catalog, view, selection, and map contexts.

## R11 — Frontend orchestration

[Frontend orchestration](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/shared/context/useAppProviderState.ts)

Selection callbacks, suggestions, save state, and map refresh integration.

## R12 — Photo list integration

[Photo list integration](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/shared/components/PhotoListContainer.tsx)

Album/date controls, catalog view and selection integration.

## R13 — Selection controller

[Selection controller](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/selection/useSelectionController.ts)

Composition of location assignment and save callbacks.

## R14 — Map and pending-location types

[Map and pending-location types](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/shared/types/map.ts)

Coordinate-only pending types and existing source discriminator values.

## R15 — Compose deployment

[Compose deployment](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/docker-compose.yml)

Frontend/backend service split and backend /data volume.

## R16 — Repository license

[Repository license](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/LICENSE)

MIT license with upstream copyright notice. Preserve attribution; no broader legal analysis is provided.

## R17 — Catalog scope and filtering

[src/shared/context/useCatalogDomain.ts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/shared/context/useCatalogDomain.ts)

[src/features/photoGrid/useAssets.ts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/photoGrid/useAssets.ts)

[backend/databaseFolders.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/databaseFolders.go)

Album/folder mode and tag/GPS/hidden/date filters. Preserve the active scope in an AI selection manifest.

## R18 — Frontend backend proxy

[src/middleware.ts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/middleware.ts)

[src/utils/client.ts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/utils/client.ts)

Default /api/backend rewrite to BACKEND_URL and configurable client base prefix.

## R19 — Selection state and manual save retries

[src/features/selection/useSelectionState.ts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/selection/useSelectionState.ts)

[src/features/selection/useLocationAssignment.ts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/selection/useLocationAssignment.ts)

[src/features/selection/locationSave.ts](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/selection/locationSave.ts)

In-memory pending state, coordinate save dispatch, and one browser retry pass for failed assets.

## R20 — Existing client interfaces and transport

[backend/interfaces.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/interfaces.go)

[backend/immichClient.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/immichClient.go)

HandlerImmichAPI exposes bulk location writes and image reads only; the shared transport has RetryMax = 3, and the existing upstream write is bulk PATCH /api/assets.

## R21 — Suggestion context assembly

[backend/suggestionService.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/suggestionService.go)

Capture-time fallback to fileCreatedAt, temporal/album neighbors, and frequent-location inclusion. Not an AI consent boundary.

## R22 — Migrations and Go test fixtures

[backend/migrations](https://github.com/crazz/immich-places-ai-addon/tree/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/migrations)

[backend/migrations_test.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/migrations_test.go)

[backend/database_test.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/database_test.go)

[backend/handlers_test.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlers_test.go)

Migration 017 is the latest; temporary real-database fixtures, fresh/legacy/idempotent migration tests, and the existing stack-write test. Inventoried, not executed.

## R23 — Session and credential handling

[backend/handlersAuth.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlersAuth.go)

[backend/databaseAuth.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/databaseAuth.go)

HttpOnly/SameSite Lax cookies, encrypted user credentials, and session-user lookup. Explicit origin/CSRF enforcement is new work for the proposed AI contract.

## R24 — Selection action and count scope

[src/features/photoGrid/PhotoCardMenu.tsx](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/src/features/photoGrid/PhotoCardMenu.tsx)

[backend/handlersCounts.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/handlersCounts.go)

Select-all receives loaded assets; missing-location counts have album/tag/hidden/date inputs but no folder-path input.

## R25 — Build and release automation

[.github/workflows/release.yml](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/.github/workflows/release.yml)

[Dockerfile](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/Dockerfile)

[backend/Dockerfile](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/Dockerfile)

Existing image builds run frontend and backend builds, without explicit Go tests or lint steps. No frontend test script is declared in R02.

## R26 — Geocoder chain and public endpoint

[backend/geocoder.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/geocoder.go)

[backend/nominatimClient.go](https://github.com/crazz/immich-places-ai-addon/blob/5e70c6165777949c9d8b50ede3b2768bcaa5df87/backend/nominatimClient.go)

Current geocoder chain starts with Nominatim, whose client uses the public endpoint. This is not the proposed Research egress policy.

## I01 — Immich v3.2.2 asset DTOs

[Immich v3.2.2 asset DTOs](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/dtos/asset.dto.ts)

GPS-pair validation, description string, custom metadata object schema and no standard heading field in the inspected update schema.

## I02 — Immich v3.2.2 asset controller

[Immich v3.2.2 asset controller](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/controllers/asset.controller.ts)

Read/update routes, deprecated PUT and PATCH alternatives, metadata routes and permissions.

## I03 — Immich v3.2.2 asset service

[Immich v3.2.2 asset service](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/services/asset.service.ts)

Separate standard-field and custom metadata operations; access checks; sidecar job enqueue.

## I04 — Immich v3.2.2 validation schemas

[Immich v3.2.2 validation schemas](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/validation.ts)

Coordinate schemas accept numeric latitude/longitude with bounds, not null.

## O01 — OpenAI: Images and vision

[OpenAI: Images and vision](https://developers.openai.com/api/docs/guides/images-vision)

Image input formats and vision constraints. Third-party endpoint support is not implied.

## O02 — OpenAI: Structured model outputs

[OpenAI: Structured model outputs](https://developers.openai.com/api/docs/guides/structured-outputs)

Schema-constrained output versus JSON mode, supported schema constraints and edge-case handling.

## O03 — OpenAI: Web search

[OpenAI: Web search](https://developers.openai.com/api/docs/guides/tools-web-search)

Web search as a separately enabled tool/integration capability.

## M01 — OpenStreetMap Foundation: Nominatim Usage Policy

[OpenStreetMap Foundation: Nominatim Usage Policy](https://operations.osmfoundation.org/policies/nominatim/)

Public endpoint rate limits, caching, identification, bulk restrictions and autocomplete prohibition; does not govern separately hosted services.

## M02 — SQLite: Write-Ahead Logging

[SQLite: Write-Ahead Logging](https://www.sqlite.org/wal.html)

WAL behavior, persistent state, backup considerations and same-host/network-filesystem constraints.

## S01 — OpenSpec official repository

[OpenSpec official repository](https://github.com/Fission-AI/OpenSpec)

Requirements/scenarios, design and tasks as distinct planning artifacts. The original review did not inspect a local installation. Local OpenSpec/OpenSpec Plus setup now exists from prior work; this reconciliation did not run artifact-generation commands or change that setup.
