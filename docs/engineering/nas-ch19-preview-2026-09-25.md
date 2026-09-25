# NAS CH19 preview deployment — 25 September 2026

The isolated preview now runs committed revision
`3c2a04280e1e4b7f2c5761d3166794cafff4efd1`, including CH16 drafts, CH18 previews,
CH19 confirmed GPS execution and the recovery fixes in `9cc39d1`.
Live GATE-02 is **pending**; GPS writing remains disabled.

## Deployment and verification

Both images were built successfully for `linux/amd64` from a clean `git archive`
export and labeled with the complete source revision:

- `immich-places-ai-preview-frontend:3c2a042`
- `immich-places-ai-preview-backend:3c2a042`

The images were loaded onto the NAS, and Dockhand updated only the image
references in stack `immich-places-ai-preview`, environment `1`. Deployment used
`pull=false`, `build=false` and `forceRecreate=false`.

Verification at `2026-09-25T15:23:39Z` established:

- Both services are healthy and run the intended images; the proxied health
  endpoint returns HTTP 200 at `http://pt-nas.tail5d7615.ts.net:3033/`.
- The normal startup migration upgraded the existing database from 25 to 28.
  `PRAGMA quick_check` returned `ok`, and there are zero write operations.
- Backend environment, persistent mounts and frontend port bindings were
  preserved. `AI_WRITE_ENABLED` and `AI_WRITE_PROFILE` remain unset, retaining
  their disabled/empty defaults.
- Container identities, start times, images, configuration hashes and mounts
  for `immich-places`, `immich-places-backend` and `codex_proxy` are unchanged.
- The existing encryption key remains in Dockhand secret storage. No credential
  was printed, committed or copied into another installation.

Before deployment, SQLite's backup API created
`/volume2/docker/immich-places-ai-preview/backups/before-3c2a042.db`.
Its `quick_check` returned `ok`, and its file mode is `0600`.

The application source is unchanged from the previously verified CH19 revision.
The [recovery verification](../../openspec/changes/archive/2026-09-25-execute-confirmed-ai-gps-writes/recovery-verification.md)
records all thirteen passing checks; this deployment additionally verifies both
container builds and the actual migration/health/configuration preservation.
It does not claim a new authenticated browser journey or live mutation.

## GATE-02 preparation and remaining work

Read-only API requests verified Immich 3.2.2, current fixture ownership and update
permission. Private target/sibling metadata baselines are retained in ignored
local evidence under `out/checks/ch19-gate02/`; public records omit private IDs,
coordinates and photo payloads. No photo was uploaded or changed.

The target thumbnail was readable, both target and sibling custom metadata were
empty, and target/sibling metadata matched the baseline after deployment. The
thumbnail stayed in process memory and was not transmitted to a provider.

The [pinned v3.2.2 controller](https://raw.githubusercontent.com/immich-app/immich/v3.2.2/server/src/controllers/asset.controller.ts)
implements the single-asset `PATCH` through `updateAssetV3` with `asset.update`
permission. `ApiExcludeEndpoint` deliberately omits it from generated OpenAPI.
The temporary inference that this absence required changing the adapter was
disproved by the controller; all resulting local edits were reverted. The
deployed adapter remains the reviewed CH19 implementation.

Completing the live gate still requires an authenticated preview account session,
an analysis/review draft for the selected fixture, an exact approved GPS plan,
controlled activation with single-request evidence, independent readback and
unrelated-field/sibling comparison, durable audit/local refresh verification,
and final writer disablement. The Immich API credential alone does not establish
a Places account session. No authentication bypass, provider request, write
confirmation or controlled writer activation was performed during deployment.

## Recovery

Preserve the migrated database, encryption key and additive CH16–CH19 tables.
The previous images remain available for diagnosis; an older image cannot expose
the new draft/preview/write history. Prefer a compatible binary rollback and
never restore the pre-deployment backup over newer user data as routine recovery.
No database rollback or photo restoration was performed. Follow the
[GPS operations guide](../ai-gps-writes.md) for retained audit and uncertain-send
recovery rules.
