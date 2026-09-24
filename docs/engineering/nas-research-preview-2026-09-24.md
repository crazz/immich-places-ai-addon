# NAS Research preview rollout — 24 September 2026

## Deployed revision

Both isolated preview services now run committed revision
`46e8fa85fbeeaafd9a792c1653df377751fb3786`, built from a clean `git archive` export
for `linux/amd64` and labeled with that complete source revision. The branch
`codex/plan-ai-selection-and-validation` was pushed before building.

- Frontend: `immich-places-ai-preview-frontend:46e8fa8`.
- Backend: `immich-places-ai-preview-backend:46e8fa8`.
- Browser: `http://pt-nas.tail5d7615.ts.net:3033/`.
- Dockhand environment/stack: `1` / `immich-places-ai-preview`.

The update loaded the two local images onto the NAS, changed only their saved
Compose image references through Dockhand, and deployed only that stack with
pull/build/force-recreation disabled. Credentials remained in existing secret
storage; no secret was logged or committed.

## Data and isolation verification

Before deployment, the SQLite backup API made a consistent live backup at
`/volume2/docker/immich-places-ai-preview/backups/before-46e8fa8.db`.
`PRAGMA quick_check` returned `ok`; the backup has mode `0600`.

Verification at `2026-09-24T20:59:00Z` established:

- Both preview healthchecks were healthy and the proxied health endpoint returned
  HTTP 200.
- Both running image tags matched the intended revision; loaded images were
  `amd64` with the expected source label.
- Backend runtime environment, persistent mounts and published frontend port
  bindings matched their pre-deployment values.
- Container IDs, image IDs, start times, configuration hashes and mounts for
  `immich-places`, `immich-places-backend` and `codex_proxy` were unchanged.
- The browser rendered the preview sign-in page. The browser session was signed
  out; an authenticated live UI journey was not claimed.

The existing backend configuration uses the application's ten-minute Research
budget. This does not extend any shorter limit enforced by the upstream proxy.
The rollout itself made no provider call and no Immich mutation. The separately
authorized three-photo comparison uses temporary private inputs and the actual
analysis prompts, wire codecs and result validator against the existing NAS
proxy; it does not create application jobs or import those photos into Immich.

## Recovery

Preserve the database and encryption key. Prefer a compatible rollback retaining
v2 Research readers, as described in the [Research guide](../ai-research.md).
An older v1-only build cannot display newly stored v2 answers. Do not restore the
backup over newer user data or use a destructive migration as a routine rollback.
