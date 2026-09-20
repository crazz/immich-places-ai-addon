# Production Visual jobs

CH12 connects the [durable lifecycle](ai-analysis-jobs.md), [frozen selections](ai-selection-snapshots.md), [image preparation](ai-image-preparation.md) and [Visual analysis](ai-visual-analysis.md). The application registers private job APIs and starts a bounded in-process consumer when `AI_ENABLED=true`. AI remains disabled by default. CH13 subsequently adds the [launch and Context-assisted workflow](ai-batch-workflow.md). Automatic retention and an Immich writer remain outside these changes.

## Admission and progress API

All routes use the existing session and `Cache-Control: no-store`. POST requires the exact configured `AI_PUBLIC_ORIGIN`, JSON content type and one unambiguous object of at most 32 KiB. Duplicate fields, including case variants, unknown fields, invalid UTF-8 and trailing objects are rejected. Owner and installation come from server authority; clients cannot supply replacement membership.

`POST /ai/jobs` accepts this shape. Replace the symbolic values with the current private selection/profile/readiness values; the consent configuration must be an exact normalized copy of the submitted configuration.

```json
{
  "idempotencyKey": "one-client-generated-key-for-this-confirmed-request",
  "configuration": {
    "selectionToken": "current-snapshot-uuid",
    "profileId": "saved-profile-id",
    "revision": 1,
    "mode": "visual",
    "format": "strict",
    "allowJson": false,
    "languages": ["en"],
    "primaryLanguage": "en",
    "policyId": "current-readiness-policyID",
    "limits": {
      "maxCalls": 1,
      "maxTokens": 104000,
      "outputTokens": 4000
    }
  },
  "consent": {
    "version": "image-consent-v1",
    "image": true,
    "configuration": "repeat the configuration object here"
  }
}
```

The example illustrates the shape, not a ready-to-send request or a provider recommendation. Use the full object for `consent.configuration`. JSON mode needs both observed JSON support and `allowJson=true`; strict mode needs observed strict support. Languages use the canonical normalization rules and do not require English.

Admission holds one short SQLite write transaction across current snapshot revalidation, active provider/capability/policy checks and job/admission/membership insertion. It performs no inference, image fetch or upstream metadata request. It retains the selection/filter manifest, exact consent and limits, provider model, token policy, and local launch capture/selected-album provenance. Capture text is bounded to 128 bytes and album labels to 4096 bytes; absent, oversized or invalid UTF-8 values remain unavailable. Visual never sends this launch metadata to the provider.

The response is `202` with durable progress including `id`. An identical normalized request with the same key returns the retained job even after snapshot expiry or deletion. Changed input conflicts; a new key needs a current snapshot. Disabled AI rejects POST, including duplicate admission. A failed or uncertain response must be retried with the same key rather than inventing a new run.

- `GET /ai/jobs/{id}` returns immutable configuration, creation time, canceled/blocked flags, state counts, at most 500 item summaries, result IDs and usage status.
- `GET /ai/jobs?limit=20&cursor=...` returns summaries without item arrays. The limit is 1–100. Ordering is descending creation time and ID. The encrypted cursor binds owner, installation and page limit; encode it as a query value and retain the same limit on continuation.
- `POST /ai/jobs/{id}/cancel` accepts `{}` and is idempotent. Queued/retry/blocked work stops immediately; running work is fenced and observes cancellation through its guards/heartbeat.

Reads and cancellation remain available while AI execution is disabled, within the current installation. Foreign and absent IDs use the same unavailable category. The internal CH11 submission API is not production authorization; jobs without a `production-v1` admission cannot be claimed by the production consumer.

## Ordinary launch defaults

Following [ADR-08](engineering/decisions/ADR-08-simple-ai-launch.md), ordinary launch requires no manually authored execution policy. With an enabled provider and current image/format observations, the application supplies a deterministic `execution-default-v1` configuration: 100,000 input tokens reserved per call for scheduling, up to 16,384 requested output tokens (the UI defaults to 4,000), a 15 MiB request envelope and a 10 MiB image envelope. Existing transport and preparation limits may be stricter. The default call count is one per selected image.

These token values are planning estimates, not verified provider limits. Requests use `max_tokens` as an output hint; a compatible proxy may ignore it. Reported usage is retained even above the estimate, without fabricating a policy violation. Calls, payload sizes, concurrency and deadlines remain enforced. Cost remains unknown. Defaults are versioned and bound to the exact owner, installation, provider revision/model and destination fingerprint.

The **Start analysis** action authorizes the displayed image selection and settings; no separate image checkbox is required. The backend still stores the exact `image-consent-v1` configuration for durable scope. Context classes remain explicit selections.

An explicit operator policy takes precedence. A stale restriction for the same owner/profile blocks automatic fallback, including after installation changes; update or remove that restriction deliberately. Invalid policy JSON still fails startup. Automatically generated defaults cannot be supplied as operator attestations.

## Operator-attested token compatibility

`AI_EXECUTION_POLICIES` is an optional JSON array, empty by default, bounded to 128 KiB and 128 entries. Each entry contains:

| Field | Contract |
|---|---|
| `version` | `execution-v1` |
| `binding` | Exact `owner`, canonical UUID `installation`, `profile`, positive `revision`, `model`, and SHA-256 `egressFingerprint` |
| `outputField` | Explicitly attested `max_tokens` or `max_completion_tokens` |
| `maxInputTokens`, `maxOutputTokens` | Positive integers, each at most 1,000,000,000 |
| `maxRequestBytes`, `maxImageBytes` | Positive bounds no greater than 15 MiB and 10 MiB; existing preparation/codec limits can be stricter |
| `evidenceRef` | Private operator evidence identifier, 1–128 UTF-8 bytes without controls |
| Optional tariffs | All of `currency` (three uppercase letters), `inputMicrosPerMillion`, `outputMicrosPerMillion`; rates 0–1,000,000,000 |

When opting into an attested policy, the operator must establish that the attested input allowance covers the whole enforced image/prompt/schema envelope for that destination/model. CH03 image/format observations do not establish token parameter support or pricing. No tokenizer guess, private probe or automatic attestation is made. Obtain identities from the private profile/current installation records and the deployment's canonical destination-policy fingerprint; do not copy evidence from another revision, owner or installation. Editing a profile with an explicit restriction requires a new exact revision policy or removal of that restriction. Invalid/incomplete/duplicate configuration fails startup even when AI is disabled.

Private profile DTOs expose `executionReadiness`: `policy_required`, `capability_required`, `ready`, `policy_violated` or `unavailable`. The projection distinguishes `source: application-defaults` and `source: operator-attested` from observed capability results and supplies safe limits plus `policyID`. It excludes `evidenceRef`, credentials and documents. Reading/saving a profile never probes the production policy.

## Finite allowances and honest accounting

`maxCalls=0` selects one call per asset; an explicit positive limit is at most three times membership. Each item has at most three claims/calls. `maxTokens` is a positive integer at most 1,000,000,000,000 and must cover one full input allowance plus the selected output ceiling. `outputTokens` is positive and cannot exceed the policy maximum. A smaller batch budget may stop pending items with `failure: budget` before their images are prepared.

Immediately before every provider dispatch, the current lease atomically reserves one call, the full input allowance and the selected output ceiling. The body contains only the resolved output-limit field; default requests do not claim provider enforcement of that field. Reservations are retained after timeout, shutdown, unknown delivery or smaller/missing usage; retries require another permitted reservation. There are no transport retries, format fallbacks or repair calls.

Usage reports distinguish `reservedTokens` from nullable `inputReported`, `outputReported` and `totalReported`. `reportedStatus` is `unknown`, `partial` or `complete`; complete means every reserved attempt reported input and output counts, not independently verified billing. Accounting reads the bounded response envelope even when the canonical proposal is invalid. Untrustworthy, malformed or absent usage remains unknown. Production accounting accepts nonnegative integer counts up to 1,000,000,000,000 so values above the largest attested ceiling can still invalidate a policy.

Exceeding an attested allowance durably records a policy violation and blocks further dispatch under that policy. It does not reverse charges. `costStatus` is `unknown` without complete tariffs; `estimatedMicros` is then null. With tariffs, the rounded-up conservative reservation estimate and currency are labeled `estimated`. Optional `maxEstimatedMicros` is a nonnegative integer at most 1,000,000,000,000,000 and is accepted only with complete tariffs. Estimates are not a billing guarantee.

## Worker settings and lifecycle

| Setting | Default | Bounds |
|---|---:|---|
| `AI_JOB_WORKERS` | 2 | 1–16 global leases and fixed consumer goroutines |
| `AI_JOB_PER_OWNER` | 1 | 1–global workers |
| `AI_JOB_LEASE_SECONDS` | 180 | At most 600, strictly greater than 120 + heartbeat seconds |
| `AI_JOB_HEARTBEAT_SECONDS` | 30 | 1–60 |
| `AI_JOB_IDLE_MS` | 1000 | 100–30000 |

Each attempt has a 120-second deadline; image preparation keeps its existing 30-second bound. Workers share the application shutdown context and are joined before database shutdown. Browser closure/session expiry does not cancel an accepted job. Recovery remains bounded to 100 expired leases per pass. Exact active profile, policy/consent, current image/source and lease authority are rechecked at dispatch and publication. No transaction spans a network call.

Timeout/connection/429/5xx failures retry within finite allowances. Invalid images/proposals and inaccessible individual assets fail that item. Authentication, model, consent/profile or policy loss blocks the job. Storage and lease failures preserve recovery semantics; there is no optimistic success on failed persistence.

Migration 023 adds admission, launch provenance, attempt accounting and policy-violation records. Existing immutable analyses are preserved; internal jobs do not gain inferred production consent. Installation rotation fences unfinished work and retains terminal history. Account deletion cascades private records. Existing explicit bounded cleanup cascades job-owned admission/launch/usage records; policy violations remain until account deletion or a genuinely different attestation. No automatic cleanup is installed. Back up SQLite and key material consistently; operational rollback disables workers and restores a compatible backup rather than destructively downgrading active data.

The [verification record](engineering/ai-production-jobs-verification.md) describes synthetic evidence. Deployment-specific token support, model quality, reference-NAS timing/concurrency and GATE-04 retention/operating decisions remain separate release evidence.
