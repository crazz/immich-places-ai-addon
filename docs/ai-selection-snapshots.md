# AI selection snapshots

CH05 freezes explicit local-catalog selections; CH06 freezes all matching candidates under a declared scope. It does not analyze images, call a provider, download images or modify Immich. There is no analysis launch control yet. Later job admission must still check current authority and obtain fresh upstream image permission.

## API

Use the normal authenticated frontend proxy: `POST /api/backend/ai/selection-preview` and `GET /api/backend/ai/selections/{snapshotID}`. Direct backend paths omit `/api/backend`. Both routes require a current session and `AI_ENABLED=true`; all responses use `Cache-Control: no-store`. POST also requires one exact `Origin` equal to `AI_PUBLIC_ORIGIN` and `Content-Type: application/json`.

```json
{
  "mode": "explicit",
  "assetIDs": ["11111111-1111-4111-8111-111111111111"],
  "scope": {
    "view": "all",
    "gpsFilter": "no-gps",
    "hiddenFilter": "visible"
  }
}
```

Scope supports `view: all`, `album` with `albumID`, or `folder` with `folderPath`, plus optional `tagID`, `startDate` and `endDate` (`YYYY-MM-DD`). GPS values are `no-gps`, `with-gps`, `all`; hidden values are `visible`, `hidden`, `all`. Missing filters default to `no-gps` and `visible`. Dates use the recorded source-local calendar day from `dateTimeOriginal`, with no file-creation/upload-date fallback. Gallery ordering remains based on `fileCreatedAt` and ID. Either absent coordinate means missing GPS; zero is valid. Folder matching is recursive, case-sensitive and preserves directory boundaries. Root-only folders, unavailable album/tag references, reversed dates, unsupported fields and conflicting view fields are rejected.

The explicit response includes `snapshotID`, mode, normalized scope, `policyVersion`, ordered `assetIDs`, `exclusions`, creation/expiry and five counts: `requestedCount = uniqueCount + duplicateCount`, `uniqueCount = eligibleCount + excludedCount`. UUID duplicates retain their first occurrence. Each excluded ID has one reason, in this order: `unavailable`, `unsupported_type`, `hidden_by_policy`, `stack_child`, `outside_scope`. Unavailable, foreign and hidden-library IDs disclose no asset metadata. The hidden scope does not override the AI hidden-image policy. Stack children never expand into targets. Zero eligible assets return `snapshotID: null` and create no resource.

All-matching requests use the same scope and omit `assetIDs` entirely:

```json
{"mode":"all-matching","scope":{"view":"all","gpsFilter":"no-gps","hiddenFilter":"visible"}}
```

This mode reads the complete authorized local catalog, independent of loaded gallery pages. It rejects `assetIDs` even when null/empty, paging, cursors, client counts and sorting fields. Discovery suppresses other owners, hidden libraries and stack children before counting. Videos and individually hidden assets permitted by the visibility filter are counted and excluded by the shared AI policy.

Query responses replace per-ID `exclusions` with `matchedCount` and aggregate `exclusionCounts`. `matchedCount = eligibleCount + excludedCount`, `requestedCount = uniqueCount = matchedCount`, and `duplicateCount = 0`. Empty reasons are `{}`; no excluded-ID list is collected or returned. Eligible IDs follow `fileCreatedAt DESC, immichID DESC`. The maximum applies to eligible photos in this mode, so many excluded videos do not consume the allowance. A completed oversized query returns `413 SELECTION_LIMIT_EXCEEDED`, exact matched/eligible/excluded/reason counts and `snapshotID: null`, with no subset IDs. A timeout, cancellation or scan/storage failure reports no purportedly exact partial counts. Retry starts a new complete preview.

GET reloads only the frozen IDs, checks owner, installation, policy, expiry, integrity and relevant original catalog facts, then returns the unchanged manifest, including historical query counts. Query reads do not rediscover matches or recount excluded nonmembers. A removed/hidden target or changed relevant scope fact rejects the whole snapshot. Sync additions, changed browser filters, new stack members and harmless resynchronization cannot alter membership or extend expiry.

| Status | Meaning |
|---|---|
| 400 | Invalid or ambiguous selection/scope/JSON |
| 401 / 403 | Missing session / rejected creation Origin |
| 404 | Unknown or foreign snapshot, including one already purged |
| 409 | Known expired, stale or inconsistent snapshot; preview again |
| 413 / 415 | Resource limit / non-JSON content type |
| 429 | Owner has 20 live snapshots |
| 503 | AI disabled, initialization unavailable, global capacity, database busy or deadline; busy/capacity failures are retryable |
| 500 | Sanitized storage failure |

A failed preview returns no usable partial snapshot. A lost successful POST response may be followed by an explicit new preview; existing resources remain subject to quotas and expiry.

## Operator settings and lifecycle

| Backend setting | Default | Allowed |
|---|---|---|
| `AI_SELECTION_MAX_ASSETS` | 500 | 1–5000 unique explicit IDs (including exclusions), or eligible all-matching assets |
| `AI_SELECTION_TTL_SECONDS` | 900 | 60–3600 seconds, absolute and non-sliding |
| `AI_INSTANCE_EPOCH` | `1` | Nonblank value, at most 128 bytes |

The independent ceilings are 10000 raw explicit IDs, a 1 MiB JSON request, a 1 MiB response/retained manifest and item facts, 20 live snapshots per owner shared by both modes, and 1000 retained headers installation-wide (including expired backlog). Local operations have a five-second deadline. Limits reject the entire preview; IDs are never silently truncated.

Expired snapshots become unreadable immediately. Cleanup runs on startup, before creation and every minute, including when AI is disabled. Each pass has a five-second deadline and removes at most 100 headers plus their cascading items. Backlog waits for later passes; no live snapshot is evicted. Cleanup failures log a fixed message without manifests or private paths, and periodic cleanup retries. Account deletion cascades only that owner's snapshots. Backups can retain older metadata and need the existing backup protections.

Migration 021 adds `ai_installation_identity`, `ai_selection_snapshots` and `ai_selection_items`. Items intentionally have no foreign key to synchronized assets, so sync deletion cannot silently shrink a manifest. The installation UUID persists across ordinary restarts. Its connection fingerprint uses the canonical configured Immich origin/base path and epoch, excluding URL credentials and query secrets. A changed endpoint or epoch rotates the UUID and removes older snapshots.

For an Immich replacement at the same URL: disable AI, change `AI_INSTANCE_EPOCH`, restart the backend, and run the existing full catalog sync for connected users before enabling AI again. The epoch is an operator declaration, not remote server identity attestation. CH11 [durable jobs](ai-analysis-jobs.md) reuse this persisted identity: rotation atomically cancels unfinished old work and fences its leases while retaining terminal analyses. Current-installation readers reject old history; explicit owner cleanup can remove it. Rotation does not automatically rebuild the catalog.

Deploy initially with AI disabled. Invalid configured limits fail startup. Selection initialization failure prevents selection API availability; it does not change manual catalog data. For binary rollback, leave the additive tables in place and keep AI disabled. Before starting an older binary that cannot interpret query manifests, expire/remove transient selections through the existing lifecycle; never reinterpret an all-matching manifest as explicit. Do not use destructive down migrations as routine rollback. Keep a consistent database/encryption-key backup under the existing operator procedure.

CH13 must consume the retained token and original normalized scope/counts; it must not rebuild membership from page state or treat a preview as analysis consent. Current authorization and upstream image permission still need checking at job admission.

See [CH05 verification](engineering/ai-selection-verification.md) and [CH06 verification](engineering/ai-matching-selection-verification.md) for automated evidence and measured limits. Local synthetic timing does not establish NAS performance or upstream permission freshness.
