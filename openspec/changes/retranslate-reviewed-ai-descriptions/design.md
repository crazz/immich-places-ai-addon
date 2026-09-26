## Context

See [proposal](proposal.md). Planning base is `0c5bc0c`; CH16 and CH19 are implemented, while CH19 live GATE-02 is pending. `drafts.Draft` already separates `Revision` and `FactsRevision`, per-language text/status/staleness and immutable analysis. `drafts.Apply` permits local corrections but no provider work. Current Visual/Research codecs require an image, so they cannot be reused as a translation request by supplying a dummy image.

GitNexus queries and source checks covered `drafts.Apply`, `aiDraftStore.edit`, `aiProductionRuntime`, provider request codecs and draft persistence. The index matches the planning base. Graph findings are navigation aids; receiver gaps and capped process coverage do not establish complete impact. Apply must refresh the graph and run impact before symbol edits.

The [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md), [coding standards](../../../docs/engineering/coding-standards.md), ADR-07 and ADR-09 remain binding. FR-08, FR-10, NFR-01–03/06 and AC-08/AC-12 provide the product contract.

## Goals / Non-Goals

**Goals:** add a small text-only translation workflow with durable per-language outcomes, exact egress consent and explicit revision-checked adoption. Reuse provider authority, policy and resource limits without representing translation as a geolocation result.

**Non-Goals:** a generic workflow engine, a second worker service, provider repairs, new dependencies, automatic language detection or factual verification by the model. No source image or Immich writer capability enters the translation core.

## Decisions

### D1. Separate translation operations from immutable analysis

Place pure admission, factual-basis validation and outcome rules in `backend/internal/ai/translations/`. Root `aiTranslation*.go` adapters own SQLite, protected routes and provider composition. Existing `internal/aiadapters/providerhttp` owns a dedicated text-only request/response codec. The consumer owns narrow store and provider interfaces; it cannot import concrete transports or writeback. UI/state stays in `src/features/ai/`, connected through the existing API and public feature boundary.

Add private translation run and language-item records referencing the owned installation, draft revision, factual revision, source result, profile revision, policy fingerprint, normalized language set and idempotency key. Each item stores its bounded outcome and sanitized failure; original analysis and current draft are unchanged by generation. This avoids adding translation-only states to immutable analysis history or falsifying an image consent record.

### D2. Freeze and disclose a reviewed text basis

The user confirms a visible factual text basis for this run. Prefill only explicitly current reviewed text; stale candidate text is not silently reused after moving the camera. The user can correct the basis before submission; coordinates alone do not create place names or factual assertions. The frozen basis identifies scene-only versus location-dependent claims and preserves uncertainty. It is at most 16 KiB of valid UTF-8 and is bound to the current factual revision and its own digest.

The explicit Generate action authorizes this text, selected provider/model and one to eight distinct validated language tags. It sends no image, coordinates as hidden context, original prompt, links for automatic fetching, neighbor information or other draft languages. It creates no source-read requirement merely to translate privately retained user-reviewed text. Existing account/installation ownership and provider authority are checked at admission, immediately before dispatch and before publication.

### D3. Bound durable execution without hidden retries

Persist one item per language, with at most one provider dispatch reservation per item per run. Calls use the existing approved destination, DNS/redirect policy, exact profile revision, deadline, response cap and token/cost accounting. Use a translation-specific JSON schema with an exact language echo and complete/unavailable result; validate UTF-8, a 16 KiB text cap and no extra fields. Do not parse translation as `ai-analysis-result.v1` or let output alter draft geometry.

Run translation items in the existing backend process and share the installation-wide provider concurrency limiter with analysis. Permit one active translation run per draft and at most eight language items per run; admission fails visibly at existing queue capacity. Per-item dispatch is bounded by the configured provider deadline with a maximum of 120 seconds, request JSON by 64 KiB and response JSON by 64 KiB. Reserve a conservative text-token budget before each call and reconcile usage afterward; explicit operator limits remain authoritative. No automatic format fallback, transport retry or post-restart resend is introduced. An interrupted reserved item becomes an explicit unsuccessful/unknown-completion outcome; the user may start a new bounded run after inspecting it. Cancellation prevents new reservations and fences late publication.

### D4. Adopt through the existing draft revision boundary

Successful items are suggestions alongside current text. Applying selected suggestions uses a single owner-qualified compare-and-set against the exact current draft revision and matching factual basis. It creates a new draft revision, preserves unselected languages and geometry, returns staged drafts to editing, and invalidates previews/undispatched approvals through CH19's existing guard. A possibly acting write blocks adoption as it blocks a manual draft edit. Any intervening edit requires reload and renewed comparison; no background merge overwrites manual text. Changed facts make old suggestions stale; canceled or failed items cannot be adopted.

Retry selected unsuccessful languages creates a new run linked to its parent and a fresh explicit consent/budget; successful languages are not included by default. Stable submission identity resolves lost acknowledgements locally, including after disablement. Read-only history never transmits again. UI handles keyboard use, per-language states, cancellation, unsaved edits and account/result response fencing.

### D5. Persistence, rollout and verification

Allocate the next additive Goose migration at apply time. Keep immutable run inputs separate from mutable item state and returned suggestions; composite foreign keys enforce owner/installation scope. Draft-linked records protect their source history from ordinary cleanup, and account deletion cancels/removes private runs without late recreation. Installation rotation invalidates unfinished work. No new timed retention policy is enabled.

Verify fresh databases and upgrades from migration 028, reopen, two-owner isolation and cancellation/revision races with real SQLite. Use local provider fixtures to inspect exact text-only payloads and dispatch counts, and Vitest/RTL plus a built Playwright journey for explicit generation, partial language failure and adoption. The [verification plan](verification-plan.md) maps every scenario to tasks and test layers; run the installed thirteen gates at implementation completion with the actual base and coverage floors. Container builds are required when build inputs change. Ordinary tests remain offline. Text-only live provider compatibility is separate opt-in evidence, not inherited from the image capability sample.

Deployment preserves the two-service topology and encryption key. Disable new translation admissions for rollback, settle active reservations, and use a compatible binary retaining new tables; never delete translation or analysis history to make an older binary start. Old clients retain ordinary draft editing. No implementation or deployment occurs during this planning phase.

## Risks / Trade-offs

- A provider may invent details despite constrained input → keep the source basis visible, require explicit adoption, and state that structural validation does not certify factual or linguistic accuracy.
- A text correction races generation → immutable outputs plus exact revision adoption prevent stale overwrites, at the cost of an explicit re-review.
- Translation could monopolize analysis capacity → shared provider concurrency and bounded per-draft admission; no independent unbounded worker pool.
- Provider completion can be unknown after interruption → retain an honest failed/unknown outcome and require a new explicitly budgeted run rather than silently repeating a billed request.
