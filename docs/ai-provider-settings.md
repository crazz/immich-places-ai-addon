# Private AI provider settings

Provider settings store each user's private configuration. Saving a profile never contacts a provider, sends a photo or changes Immich. An installation destination allowlist gates internal provider dispatch. Owners can run an explicit Settings capability test that uses only bundled synthetic images; that test records observed image/JSON/strict compatibility for the saved revision and does not authorize private photographs or claim geolocation quality.

## Enable the settings

AI is disabled by default. To enable private settings, configure the backend (or the root `.env` used by Docker Compose):

```dotenv
AI_ENABLED=true
AI_PUBLIC_ORIGIN=https://places.example.com
# Default deny: empty allowlist. Approve exact destinations before any dispatch can run.
AI_PROVIDER_EGRESS_POLICY=[]
```

Use the exact browser origin, including a nondefault port when applicable, without a path, query, fragment or trailing slash. For local HTTP development an example is `http://localhost:3032`; the existing session/TLS settings still apply. Restart the backend after changing installation configuration. The backend refuses to start with AI enabled and an invalid/missing public origin, or with a malformed `AI_PROVIDER_EGRESS_POLICY`.

Provider profile saves remain offline: saving a base URL does not approve egress. Internal dispatch requires an exact matching installation rule. Empty or unset `AI_PROVIDER_EGRESS_POLICY` denies every provider request while still allowing offline profile management.

Open **Settings → AI providers** after signing in. Create a profile with its name, HTTP(S) API base URL and manually entered model. The optional key is masked while entered and never loaded back into the form. Profiles are private to the signed-in account. Multiple profiles are supported; this change does not select a provider for an analysis job.

Setting `AI_ENABLED=false` hides stored profiles from listing and rejects profile changes. Existing browsing, manual placement and GPX remain available. Profile-level **Enabled** is a separate setting; dispatch respects both installation and profile controls.

## Existing NAS `codex-proxy` container

Reuse the already-running `codex-proxy` container. Do not create another proxy, publish an extra host port or change its default model for this addon. The Places backend must already share the proxy's Docker network (for example `npm_proxy`); generic Compose in this repository does not attach that operator-specific network.

Before enabling dispatch:

1. Confirm the live proxy endpoint from the backend network, typically `http://codex-proxy:3466/v1`. Host loopback `127.0.0.1:3466` is for host-network clients and is not the backend container's destination.
2. Confirm the Docker network subnet (recheck at deployment; do not treat a past inspection as a portable default).
3. Set an exact local rule for that base URL and CIDR only, for example:

```dotenv
AI_PROVIDER_EGRESS_POLICY=[{"baseURL":"http://codex-proxy:3466/v1","addressClass":"local","allowedCIDRs":["192.168.144.0/20"]}]
```

Do not commit a broad private-network allowance or auto-populate approval from a user profile. After changing the policy, restart the backend so new connections load the updated allowlist.

The proxy's own default model remains an independent operator setting (historically `gpt-5.6-luna` on the inspected image). Enter the addon profile model manually. An authorized [NAS test on 19 September 2026](engineering/nas-provider-capability-2026-09-19.md) observed image, JSON and strict-schema sample support for requested/reported model `gpt-5.6-sol` through the existing proxy. This evidence applies to the recorded runtime and configuration; public documentation is not proof of availability on another installation. Do not change the proxy global default, upgrade its container or silently fall back to another model when Sol is unavailable — create a new profile revision with a chosen model instead.

## Explicit synthetic capability test

Saved enabled profiles expose a labeled **Test provider** action separate from saving. Unsaved edits are not testable; save first. The UI discloses the destination/model, that only bundled synthetic images are sent, that up to three Chat Completions requests may consume provider usage, that the local proxy can forward this synthetic input to its upstream model service, and that cancelling only stops local waiting and cannot reverse upstream usage.

The backend shares one roughly 120-second provider-work deadline across the three probes; the Settings client uses a 130-second complete-operation deadline with cancel/close abort. Each probe uses a 64 KiB response ceiling, tighter than the general provider transport default. Concurrent starts are rejected (busy). Reload reads the last persisted report and never resubmits the POST. Observations are revision-bound and become non-applicable after profile, protocol or destination-policy changes; they are not a TTL guarantee.

The current protocol is `capability-v2`. It rejects repaired or trailing JSON and rechecks current authority after DNS and before publishing results. Earlier `capability-v1` reports remain historical and require an explicit retest. Migration 020 preserves those records and adds optional numeric usage and the consumption disclosure. Attempt totals include only fields reported for every dispatched probe; missing, malformed or overflowing values remain unknown, including after reload. These observations do not establish billing totals.

An unapproved destination, missing session, foreign or disabled profile, stale revision or busy slot fails as a structured error before a test is accepted. Once a test is accepted and its outcome is persisted, the request succeeds and returns that report, including failed lifecycles and unsuccessful observations: HTTP success describes the test operation, not provider compatibility.

A successful fixture or live result is observed synthetic compatibility only. It is not permission to send private photographs and not a geolocation quality score. GATE-03 is complete for the [recorded NAS configuration](engineering/nas-provider-capability-2026-09-19.md); other endpoint/model/runtime combinations require their own authorized live evidence.

### Opt-in live evidence (not ordinary CI)

Ordinary regression and smoke suites use a deterministic local provider fixture and never call NAS, OpenAI or the real Immich library. When an operator separately authorizes a live check against the installed `codex-proxy`:

1. Approve the exact proxy base URL and network CIDR as above.
2. Save an enabled profile with the manually entered candidate model (for example `gpt-5.6-sol`) and a usable key.
3. Run **Test provider** once from Settings.
4. Record proxy/runtime image, endpoint, requested and any reported model, protocol/revision, timestamp and the three observation states. Label usage unknown when the provider omits usage fields.
5. If Sol is unavailable, report that fact and choose another model in a new revision; do not modify the proxy default.

Fixture success in CI does not close GATE-03 for the NAS.

## Editing and credentials

Every save creates a revision. A stale edit fails with a conflict; reload the profiles to edit the latest version. The UI does not automatically retry writes. Non-secret fields stay available after a save error, and entered key text is cleared. If a response is lost, reload before resubmitting to check whether the save succeeded.

- Leave the key blank while editing to keep the current key at the same base URL.
- Enter a replacement to associate a new encrypted key with the new revision. Earlier encrypted keys remain in history.
- Change the base URL only with an explicit replacement or removal when a key is stored.
- Choose **Remove all stored keys for this profile** and save to clear credential material from every revision of that profile. Non-secret history remains.
- Disabling a profile retains its encrypted keys. It does not remove them.

Credentials use the existing backend `ENCRYPTION_KEY`. Preserve that key together with a consistent database backup; replacing it without migrating encrypted data prevents reuse of stored keys. Responses disclose only `hasSecret`, never the plaintext or ciphertext. Cancel/close discards an entered key without saving it. The application does not persist provider settings or keys in browser storage. Capability reports never include secrets or raw provider responses.

## Stopping dispatch and rollback

To stop subsequent provider dispatch without removing saved profiles:

1. Set `AI_ENABLED=false`, or set `AI_PROVIDER_EGRESS_POLICY=[]`, then restart the backend.
2. In-flight admitted requests cannot be recalled; the restart closes old connections and prevents new dispatches.

Take a consistent database backup before upgrading. Migrations 018 (profiles) and 019 (capability checks) are additive. Normal binary rollback retains the upgraded SQLite database and safe capability observations; do not run a destructive down migration:

1. Stop the current backend.
2. Keep the upgraded SQLite database and `ENCRYPTION_KEY`.
3. Run the previous backend executable with AI disabled (or with an empty egress policy).
4. Do not run migration 018/019 destructive down migrations as routine rollback.

Key removal and account deletion affect the active database. They do not rewrite backups or promise forensic erasure of old SQLite pages/WAL files. Protect backups, database files and the encryption key with the installation's existing access controls and retention policy. No Immich AI write needs reconciliation for configuration or egress-policy rollback.

## Migration and cleanup

Migration 018 adds user-scoped profile/version tables. Migration 019 adds `ai_provider_capability_checks` keyed by owner, profile and revision. Fresh installation, upgrade, reopen, concurrent revision conflicts and transaction rollback are tested with real SQLite. Account deletion cascades through provider profiles, versions and capability rows; this change does not add an account-deletion UI or endpoint.
