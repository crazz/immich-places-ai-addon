# Exact AI GPS previews

CH18 implements private, read-only comparisons for a staged local review draft. A preview covers the analyzed photo and its complete GPS pair. Large or unknown estimated error, stale heading and unrelated descriptions do not block it. Previewing never calls a provider, populates manual pending coordinates or mutates Immich.

## Review and recovery

1. Open an accepted draft in AI Results. Review the current image and GPS with the [draft baseline workflow](ai-review-drafts.md), explicitly acknowledge it and stage the saved GPS revision.
2. Choose **Preview exact GPS**. Inspect the photo ID, draft revision, exact before/proposed latitude and longitude, changed/unchanged label, expiry and digest. Absent and partial GPS remain distinct from zero. Numeric controls work with keyboard input and unavailable map tiles.
3. If GPS changed, inspect the reviewed/current/proposed values, explicitly review and acknowledge the new baseline, then stage the new revision and create a fresh preview. A material image change requires current-image review. A stale draft offers **Compare saved revision**; source failure leaves the saved draft editable.
4. Resolve same-photo manual pending work by explicitly saving or discarding that manual choice. Local editor changes or new manual overlap retire the displayed preview. Account/result navigation removes private preview state and fences late replies.

Previews expire five minutes after creation. Reloading retrieves the same comparison without contacting Immich or extending expiry. An already-matching pair is labeled **unchanged**; it does not claim that a write occurred. Read access does not establish update permission. CH18 has no confirmation control or executor.

## HTTP and stored contract

- `POST /ai/write-previews` accepts only `draftId` and `draftRevision`, at most 4 KiB, with current session and origin checks. The server resolves all plan values from the owned staged revision.
- Creation makes one metadata GET to the configured Immich origin, forbids redirects, limits the response to 1 MiB and applies a 10-second deadline. It validates identity/visibility and exact nullable GPS, using CH16's separately reviewed image identity rather than treating an unrelated timestamp change as image replacement.
- After the read, a short SQLite transaction rechecks installation, credentials, local visibility and exact draft revision/state before publication. No database transaction spans the network request.
- `GET /ai/write-previews/{id}` is private, installation-bound and `no-store`. It returns the stored plan/digest and a current `usable`, `expired` or `stale` projection. It makes no upstream request. Catalog reset and disabled provider execution preserve local inspection.
- The fixed-field `gps-preview-v1` payload binds owner, installation, preview ID, draft ID/revision, analysis, reviewed-image identity, target, `[gps]`, nullable before values, finite intended values, observation/creation/expiry and `exact-nullable-gps-v1`. Signed zero is normalized. SHA-256 hashes the same canonical bytes stored and displayed, bounded to 16 KiB. A digest is not authentication.

Malformed requests return 400; foreign/unavailable references return indistinguishable 404 errors. Readiness, stale revision, source/GPS conflict and capacity failures return 409; source/storage failures return 503. An expired private GET remains inspectable with an unusable status. Errors contain safe local text, a code and a request ID; only an authorized GPS conflict includes before/current/proposed values.

## Persistence, deployment and CH19 handoff

Migration 027 adds `ai_write_previews`, with owner/installation-qualified references to immutable draft revisions. Payload, digest and identity/validity fields cannot be updated. Account deletion cascades only owned rows; installation rotation does not transfer old authority. Retained previews protect referenced draft revisions and their history.

Publication permits ten active previews per current draft revision and deletes at most 100 expired, unprotected rows per pass. Stale revisions are not active authority. CH19 must atomically protect a consumed preview with `protected=1` when creating its durable confirmation/audit, and must preserve that reference during cleanup. The flag is only a retention guard; it grants no approval.

CH19 must reload the private stored plan, verify exact ID/digest/scope/revision/validity, durably consume it once, and repeat source/GPS/authority checks before dispatch. The narrow preview workflow exposes only draft reading, metadata reading and plan publication. No current interface permits a mutation. A fresh read can become obsolete immediately afterward; stored `usable` status is not a remote permission lock.

Back up SQLite consistently before rollout. Binary rollback retains the additive tables; do not run the destructive down migration as recovery. Tests cover fresh/026 upgrade/reopen/repeat migrations, transactional failure and pooled foreign keys. Earlier analysis-read queries work with preview tables retained; actual deployment/rollback was not performed.

Keep CH19's planned `AI_WRITE_ENABLED=false` until separately authorized disposable-photo GATE-02 evidence verifies the actual Immich version, mutation route, permissions, exact field preservation and readback. Synthetic preview tests do not close that gate. No live provider, private photo or live Immich mutation was used. See the [CH18 verification record](../openspec/changes/preview-exact-ai-write-plans/implementation-verification.md) for P01–P18 and actual checks.
