# Isolated NAS preview deployment — 20 September 2026

The CH01–CH15 application at commit
`11fca61eddafff998d37c203a77a67df9deef28b` was deployed through the Dockhand API as
the new local stack `immich-places-ai-preview`, environment `1`. Deployment
completed at 19:18 UTC; isolation and health checks completed at 19:19 UTC.
The [deployment configuration and runbook](../../deploy/nas-preview/README.md)
describe its scope and maintenance.

## Artifacts

Both images were built from a clean `git archive` of that commit for
`linux/amd64`, labeled with the full source revision and transferred directly to
the NAS. No registry publication or production image replacement was performed.

| Image | Runtime image configuration digest on NAS |
|---|---|
| `immich-places-ai-preview-frontend:11fca61` | `sha256:c27e9c2ddc7d97cd3ee57dd43c10d84e0af39811d9b5996d38f3df0d7de2ecdb` |
| `immich-places-ai-preview-backend:11fca61` | `sha256:a22877b2c479502f09c5016bbb4aeadcf9790e18e925ff5465e1e44b6b8d45ce` |

The root and preview Compose files passed `docker compose config --quiet` with
dummy key material. The repository size check and `git diff --check` passed.
GitNexus reported no changed application symbols for these configuration and
documentation edits; Compose file impact was `UNKNOWN`, so it was verified by
reading the backend configuration contract and validating the resolved Compose
environment. No application source changed and no new full regression-suite run
is claimed by this deployment record.

## Live checks

- Dockhand created and deployed only the new stack. Its newly generated
  `ENCRYPTION_KEY` was stored as a masked secret and was absent from plaintext
  files in the stack directory. In-memory comparison confirmed that the key
  differs from the existing Places key; neither value was logged.
- Frontend `immich-places-ai-preview` and backend
  `immich-places-ai-preview-backend` both reported `healthy`.
- `/` and `/api/backend/health` returned HTTP 200. The health response reported
  zero synchronized assets and no previous sync time.
- The browser rendered **Sign in** and **Create account** at
  `http://pt-nas.tail5d7615.ts.net:3033`. The registration form was left open for
  the user, with no account submitted.
- The backend bind mount is exclusively
  `/volume2/docker/immich-places-ai-preview/data:/data`. SQLite `quick_check`
  returned `ok`; users, assets, profiles, capabilities, selections, jobs,
  analyses and result history had zero rows. The new installation identity had
  one row.
- Only frontend port `100.110.198.109:3033` is published. The backend has no host
  port. The frontend uses only `immich-places-ai-preview_default`; the backend
  also joins the existing `npm_proxy` network for configured upstream access.
- Before/after comparisons of container IDs, image IDs, start times and hashed
  configurations confirmed that `immich-places`, `immich-places-backend` and
  `codex_proxy` were unchanged. The existing data mount remained independent.

## Scope of the evidence

This is a fresh-install deployment smoke check. No user account, Immich API key,
provider profile or result was copied or created. No private image analysis,
synthetic provider inference or Immich mutation was invoked.

AI settings are enabled with the inspected proxy destination allowlist and one
worker. `AI_EXECUTION_POLICIES` remains `[]`: the new account/profile must be
configured and an exact, verified execution policy installed before production
analysis can run. This deployment does not establish new token compatibility,
geolocation quality, NAS concurrency benchmarks or retention/release gates.

The upstream Immich instance is shared. Separate Places data does not isolate
future user-authorized manual/GPX writes from that real library. Browser sessions
also require distinct hostnames; different ports alone do not separate cookies.

The initial launch restriction described above was superseded later on 20 September by [ADR-08](decisions/ADR-08-simple-ai-launch.md). The [launch correction verification](ai-launch-simplification-verification.md) records the updated preview, ordinary application defaults and successful deployment.
