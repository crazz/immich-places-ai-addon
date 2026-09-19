# Private AI provider settings

Provider settings are the first AI foundation. They store each user's private configuration; saving never contacts a provider, sends a photo or changes Immich. Destination approval, connection/capability tests and analysis arrive in later changes. A saved profile is unverified.

## Enable the settings

AI is disabled by default. To enable private settings, configure the backend (or the root `.env` used by Docker Compose):

```dotenv
AI_ENABLED=true
AI_PUBLIC_ORIGIN=https://places.example.com
```

Use the exact browser origin, including a nondefault port when applicable, without a path, query, fragment or trailing slash. For local HTTP development an example is `http://localhost:3032`; the existing session/TLS settings still apply. Restart the backend after changing installation configuration. The backend refuses to start with AI enabled and an invalid/missing public origin.

Open **Settings → AI providers** after signing in. Create a profile with its name, HTTP(S) API base URL and manually entered model. The optional key is masked while entered and never loaded back into the form. Profiles are private to the signed-in account. Multiple profiles are supported; this change does not select a provider for an analysis job.

Setting `AI_ENABLED=false` hides stored profiles from listing and rejects profile changes. Existing browsing, manual placement and GPX remain available. Profile-level **Enabled** is a separate setting; later dispatch code must respect both controls.

## Editing and credentials

Every save creates a revision. A stale edit fails with a conflict; reload the profiles to edit the latest version. The UI does not automatically retry writes. Non-secret fields stay available after a save error, and entered key text is cleared. If a response is lost, reload before resubmitting to check whether the save succeeded.

- Leave the key blank while editing to keep the current key at the same base URL.
- Enter a replacement to associate a new encrypted key with the new revision. Earlier encrypted keys remain in history.
- Change the base URL only with an explicit replacement or removal when a key is stored.
- Choose **Remove all stored keys for this profile** and save to clear credential material from every revision of that profile. Non-secret history remains.
- Disabling a profile retains its encrypted keys. It does not remove them.

Credentials use the existing backend `ENCRYPTION_KEY`. Preserve that key together with a consistent database backup; replacing it without migrating encrypted data prevents reuse of stored keys. Responses disclose only `hasSecret`, never the plaintext or ciphertext. Cancel/close discards an entered key without saving it. The application does not persist provider settings or keys in browser storage.

## Migration, cleanup and rollback

Migration 018 adds user-scoped profile/version tables to the existing SQLite database. Fresh installation, upgrade from version 17, reopen, concurrent revision conflicts and transaction rollback are tested with real SQLite. Account deletion cascades through its provider profiles and versions; this change does not add an account-deletion UI or endpoint.

Key removal and account deletion affect the active database. They do not rewrite backups or promise forensic erasure of old SQLite pages/WAL files. Protect backups, database files and the encryption key with the installation's existing access controls and retention policy.

Take a consistent database backup before upgrading. To roll back the application, stop it, retain the upgraded database and run the previous executable with AI disabled. Migration 018 is additive; do not run its destructive down migration as routine rollback. No external provider request, photo analysis or Immich AI write needs reconciliation for this configuration-only change.
