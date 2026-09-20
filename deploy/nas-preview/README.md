# Isolated NAS AI preview

This Dockhand-managed local stack runs the CH01–CH15 application beside the
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
| Image revision | `11fca61eddafff998d37c203a77a67df9deef28b`, `linux/amd64` |

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

1. Export the committed application source with `git archive` into a temporary
   build directory. This excludes local secrets, test output and dependency caches.
2. Build the root Dockerfile as `immich-places-ai-preview-frontend:11fca61` and
   `backend/Dockerfile` with the `backend` build context as
   `immich-places-ai-preview-backend:11fca61`. Use `--platform linux/amd64 --load`
   and label both images with `org.opencontainers.image.revision`.
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

For later image updates, build and load new immutable revision tags, set
`PREVIEW_REVISION` for this stack and redeploy only the preview. Preserve its key
and take a consistent SQLite backup first. Stopping this preview stack in
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

`AI_EXECUTION_POLICIES` starts empty. Production analysis stays unavailable until
an operator adds a verified policy bound to this new owner, installation, profile
revision, model and egress fingerprint. Follow the
[production guide](../../docs/ai-production-jobs.md) and
[batch workflow](../../docs/ai-batch-workflow.md); a historical synthetic proxy
test is not an execution-policy attestation for this fresh installation. Result
history starts empty. CH15 review is read-only and does not implement AI writeback.
