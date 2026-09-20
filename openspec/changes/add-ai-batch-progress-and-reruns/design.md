## Context

See [proposal.md](proposal.md), [CH10's context contract](../archive/2026-09-20-add-consented-ai-context/design.md), [CH12's production contract](../archive/2026-09-20-connect-visual-analysis-to-durable-jobs/design.md) and the adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md) and [coding standards](../../../docs/engineering/coding-standards.md). Grounded at `23c6ca6`: GitNexus resolves `PhotoList` through the existing gallery/container/selection flow. Source confirms separate page assets and selected IDs, an existing trailing-action integration point and a validated private provider API. Jobs and selection UI clients do not yet exist. Current completion is Visual-only, so Context cannot be enabled by a frontend mode toggle alone.

## Goals / Non-Goals

**Goals:** Users can preview, authorize, launch, monitor, cancel and explicitly repeat bounded Visual or Context-assisted work across refresh/navigation.

**Non-Goals:** No new library, result-history implementation, editable draft, write action, translation-only retry or automatic reanalysis. Research and neighbor-image sharing remain unavailable.

## Decisions

### Launch flow and selection authority

Keep launch components, hooks, DTO validation and pure transitions under `src/features/ai/`, exported through its public index. Integrate narrowly with the existing gallery container and trailing action. Reuse the backend proxy/client conventions, but explicitly disable automatic retries for submit/cancel and ambiguous mutations. Do not place job data in manual pending-coordinate state. No browser credential or provider SDK is introduced.

Offer three explicitly labeled selection intentions: selected IDs, current page IDs and all matching. Explicit/page requests use CH05 snapshots; matching requests use CH06 filters/scope and never simulate all-matching from loaded IDs. Show exact matching/eligible/exclusion counts, scope and expiry from the server. Reject empty/over-limit snapshots and provide safe exclusion reasons. Unsupported active filters must block launch with an explanation rather than silently disappear from the preview request. A filter/page/selection change invalidates the local preview and consent; the server remains authoritative against races. Later catalog/stack changes cannot expand an admitted job.

The launch sheet shows provider/model/revision and readiness, Visual or Context-assisted mode, strict or explicitly allowed JSON output, requested language tags/primary language, finite calls/tokens and optional estimated-cost cap. Default to Visual, strict output and the existing language preference when valid; otherwise choose English with an editable primary language. English is never mandatory. Explain partial completion when the chosen total allowance cannot cover the full batch. Do not invent model capability, price or token usage. Disable submit until selection, provider policy and all input bounds are valid.

Image consent starts unchecked. Context-assisted exposes four independently unchecked classes from CH10: `capture_time`, `selected_album`, `user_hint`, `nearby_locations`; selected album is offered only in a concrete selected-album scope. A hint is at most 2,000 UTF-8 bytes and is submitted only with its class selected. Show the bounded metadata-only neighbor disclosure, source-lineage uncertainty and possibility of no usable context. An empty context choice/bundle is valid and stays Context-assisted. Visual has no context payload. Changing selection, provider revision, mode, context classes/text, languages, format or limits resets confirmation and the submission key. A confirmed ambiguous POST keeps its key for explicit reconciliation; a deliberate new run uses a new key. UI double-click protection complements server idempotency.

### Durable Context-assisted execution

Extend the CH12 admission envelope with versioned mode-specific consent, selected context classes, hint and selected-album binding. Image consent remains required in both modes. Reject context fields in Visual instead of silently keeping hidden data. Enforce bounds server-side and include the exact context choices in the idempotency fingerprint. CH12 token policy must cover the additional bounded context envelope before Context admission is ready.

At the first Context attempt, prepare the CH10 bundle under current read authority and persist its minimal bounded provider projection, private source bindings, digest, evidence-support flags, policy/consent versions and omissions before provider reservation. This allows restart to use the exact evidence that was authorized. On a retry, reuse that frozen bundle only after CH10 freshness checks; changed sources fail that item and require explicit reanalysis rather than silently substituting evidence. Temporary target/neighbor images are never persisted; neighbors remain metadata-only. The owner hint is private persisted input, absent from ordinary logs.

Extend the completion envelope and transaction to revalidate Context proposals against the trusted stored source set, requested languages and exact mode, with a versioned Context prompt identity. Do not accept a model-supplied source set or use the current hard-coded Visual validation. Atomically retain context provenance with the immutable result; old Visual records remain readable. This is an additive migration allocated from the application checkout. CH12's exact provider/access/lease guards, call/token ledger, cancellation and failure policy remain in force for both modes.

### Progress, cancellation and explicit new runs

Use CH12 list/detail APIs as durable truth, with a job ID in navigable UI state. Poll active jobs every two seconds while visible; pause while hidden/offline, resume immediately on return and back off failures up to 30 seconds. Stop polling terminal/blocked jobs, with explicit refresh available. Abort superseded requests and ignore late responses after owner, job or filter changes. Clear private cached state on logout/owner change. Browser closure stops polling only, not the server job.

Show queued, running, retry-wait, blocked, succeeded, failed and canceled counts plus completed proposal outcomes when available. Unknown/ambiguous are successful analyses, not failed transport. Use text labels and accessible progress announcements without stealing focus on every poll. Display stale/offline status, safe failures, consumed allowances and reported/unknown usage separately. Cancellation is explicit, repeat-safe and acknowledges that already transmitted data or charges cannot be recalled; completed results remain reachable by result ID, with detailed browsing supplied by CH14.

Retry-failed takes only terminal `failed` members from an owned parent job. Reanalysis takes explicitly chosen members, including successful unknown/ambiguous items; canceled/blocked members can use explicit reanalysis rather than masquerading as failed. Both re-preview current eligibility and require a new consent confirmation and idempotency key. Extend POST admission with an optional parent job and rerun kind, validating the relationship and allowed membership server-side in the admission transaction. Persist this lineage; never reset parent states or overwrite analyses. If some requested assets are no longer eligible, show the changed preview and require confirmation of the remaining exact set. No automatic retry on polling, mounting, changed hints or selecting another provider.

### Verification and rollout

Map scenarios to pure transition/DTO tests, RTL user-intent and accessibility tests, authenticated API tests and real SQLite provenance/reopen tests. Browser journeys cover explicit/page/all-matching launch, consent reset, uncertain POST reconciliation, mixed progress, cancel/reload and reruns preserving the original run; intercept network calls to prove zero Immich writes. Synthetic provider fixtures cover Context evidence rejection, changed source after restart and Visual payload isolation. Preserve provider Settings, manual, GPX and AI-disabled journeys; run the shared apply gates and backend Docker build. No live photo disclosure follows from the plan.

Apply/sync CH10 then CH12 before CH13. Its modified admission requirements include CH12's full replacement text and scenarios, preventing delta synchronization from dropping production admission. Keep Context launch unavailable until its complete durable backend consumer and policy readiness are installed. Rollback disables new execution and restores a compatible backed-up database if needed; never reinterpret Context records as Visual or mutate old results.

## Risks / Trade-offs

- Gallery filters and snapshot filters may differ; unsupported filters must fail visibly instead of widening the batch.
- Context evidence can become stale while queued; explicit reruns are safer than silently changing the evidence of a retried run.
- Polling can show stale progress during outages; label it stale and retain server state as authority.
- More consent choices increase form complexity; progressive disclosure by mode keeps Visual simple without preselecting context sharing.
