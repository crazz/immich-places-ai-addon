# Isolated NAS AI preview

This Dockhand-managed local stack runs Research analysis, persistent review drafts,
exact GPS previews and confirmed GPS execution beside the
existing `immich-places` deployment. It is specific to the inspected `pt-nas`
installation; review the addresses and network policy before using it elsewhere.

| Resource | Preview |
|---|---|
| Dockhand environment / stack | `1` / `immich-places-ai-preview` |
| Browser origin | `http://pt-nas.tail5d7615.ts.net:3033` |
| Published listener | Tailscale IP `100.110.198.109:3033` only |
| Compose file on NAS | `/volume2/docker/immich-places-ai-preview/compose.yaml` |
| Persistent SQLite directory | `/volume2/docker/immich-places-ai-preview/data` |
| Private frontend/backend network | `immich-places-ai-preview_default` |
| Shared upstream network | Existing `npm_proxy`, inspected CIDR `192.168.144.0/20` |
| Frontend revision | `3c2a04280e1e4b7f2c5761d3166794cafff4efd1`, `linux/amd64` |
| Backend revision | `3c2a04280e1e4b7f2c5761d3166794cafff4efd1`, `linux/amd64` |

The existing stack, its port `3032`, images and
`/volume2/docker/immich-places/data` remain independent. The preview uses a fresh
database, installation identity, accounts and encryption key. Never mount the
existing data directory into this stack or copy its database/key as preview data.

Use the full browser origin above, distinct from the existing `http://pt-nas:3032`
hostname. Browser cookies are scoped by hostname, not port: using `pt-nas:3033`
would share session cookies with the existing short-hostname deployment. Keep the
two deployments on distinct hostnames when accessing them. The origin must also
match `AI_PUBLIC_ORIGIN` exactly for AI writes. HTTP is enabled only for this
Tailscale-bound deployment; the backend has no published host port.

## Deployment

1. Export each service's committed source revision from the table above with
   `git archive` into a temporary build directory. This excludes local secrets,
   test output and dependency caches.
2. Build the root Dockerfile as `immich-places-ai-preview-frontend:3c2a042` and
   `backend/Dockerfile` with its revision's `backend` build context as
   `immich-places-ai-preview-backend:3c2a042`. Use `--platform linux/amd64 --load`
   and label each image with its full `org.opencontainers.image.revision`.
3. Transfer these two images to the NAS with `docker save` / `docker load`.
   No registry publication is required. `pull_policy: never` prevents replacement
   with unrelated registry images.
4. Create only the new preview directory and its `data` subdirectory. Validate
   `compose.yaml` with `docker compose config --quiet` using a dummy encryption
   key before supplying any real credentials.
5. Create the local stack using Dockhand's authenticated
   `POST /api/stacks?env=1`: `name`, `compose`, `composePath`, `start: false`, and
   `envVars: [{key: "ENCRYPTION_KEY", value: <new random 32-byte hex key>,
   isSecret: true}]`. Generate the value in memory and never print it or commit it.
   Dockhand stores the secret separately; its plaintext `.env` must not contain it.
6. Deploy only this stack with
   `POST /api/stacks/immich-places-ai-preview/deploy?env=1`, with `pull`, `build`
   and `forceRecreate` all false. Retain the SSE completion result.
7. Verify both healthchecks, the first-registration UI, separate mounts/networks,
   secret storage and unchanged IDs/start times for the existing Places and proxy
   containers. Do not create a test account in the user's fresh installation.

For later image updates, build and load new immutable revision tags, update the
changed services' image references through the Compose API, and redeploy only
the preview. The template supports independent `PREVIEW_FRONTEND_REVISION` and
`PREVIEW_BACKEND_REVISION` overrides. Preserve its key and take a consistent
SQLite backup before backend/data changes. Stopping this preview stack in
Dockhand leaves its data in place. Do not use a general Docker prune or touch the
existing stack as part of preview maintenance.

## First use and limits

Register a new local account, then enter an Immich API key through the UI. No
existing Places users, API keys, catalogs, provider profiles or results are copied.
The configured Immich server is still the same upstream library: database
isolation does not create a disposable copy of Immich. Existing manual placement
and GPX writes therefore affect that real library when explicitly used.

AI settings and result/history UI are enabled. Create an enabled private provider
profile for `http://codex-proxy:3466/v1` and the chosen model, then explicitly run
the disclosed synthetic capability test. No provider call runs just from saving
a profile or deploying the stack.

`AI_EXECUTION_POLICIES` stays empty for ordinary analysis, which uses application
defaults after a current capability test. Optional operator restrictions are
explained in the [production guide](../../docs/ai-production-jobs.md). The
[batch workflow](../../docs/ai-batch-workflow.md) launches with one explicit Start
action; token estimates are not a billing guarantee. CH19 supplies confirmed
single-photo GPS execution, but writing remains disabled until the controlled
GATE-02 check is authorized and completed. The template defaults to frontend
`3c2a042` and backend `3c2a042`. When updating the Dockhand local stack through its API,
update the intended image references in the saved Compose content and then
deploy. Frontend-only fixes preserve the backend image and container. Changing only
`PREVIEW_REVISION` through the environment endpoint did not change the local
stack's running images in the verified rollout. Preserve the existing secret and
data mount, and verify actual container image tags after deployment.

The [HTTP launch verification](../../docs/engineering/ai-http-launch-verification.md)
records the request-key correction and its rollout.

The [CH19 rollout verification](../../docs/engineering/nas-ch19-preview-2026-09-25.md) records the current deployment, migration 25→28, checked backup and preserved runtime configuration. The [Research rollout](../../docs/engineering/nas-research-preview-2026-09-24.md) remains the preceding deployment record.
