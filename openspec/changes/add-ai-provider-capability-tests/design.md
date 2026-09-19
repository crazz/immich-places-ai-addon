## Context

See the [proposal](proposal.md) and [delta spec](specs/ai-provider-configuration/spec.md). CH01 is implemented at `dbc2a52dfae82b111b107dff702eb94c6d67cb31`; CH02's [guarded dispatch design](../enforce-ai-provider-egress-policy/design.md) is a prerequisite plan, not implemented functionality at this checkout. Apply and verify CH02 before this change; OpenSpec does not enforce that cross-change dependency automatically.

GitNexus's matching index identified `newAIProviderHandler`, its `main`/test callers and `backendFetch` with its existing AI consumers. Source confirms a protected `/ai/` mux, private immutable versions, feature-owned `providerApi.ts`/`ProviderSettings.tsx`, real SQLite and an AI-enabled browser journey. There is no current capability storage or inference path. The shared browser request timeout is 15 seconds; the Go server's write timeout is 150 seconds. A test must fit that actual end-to-end path.

Reuse the installed NAS `codex_proxy` at `http://codex-proxy:3466/v1` on `npm_proxy`. The observed image includes the vision forwarding patch and Codex 0.154.0. Historical synthetic-image success used Luna, not Sol. Its `/v1/models` response was a fixed older list that even omitted its configured `gpt-5.6-luna` default; discovery is not an authority for model availability.

The first intended stronger-model candidate is manually entered `gpt-5.6-sol`, which OpenAI documents with image input and structured output ([model reference](https://developers.openai.com/api/docs/models/gpt-5.6-sol)). This does not establish access through this NAS account/runtime. No model is hardcoded into the addon, no automatic fallback to Luna occurs, and the shared proxy's default remains an independent operator setting. Planning does not run inference or alter the container.

## Goals / Non-Goals

**Goals:** One explicit Settings-to-backend-to-provider workflow that produces safe, private, revision-bound observations. Reuse CH02's network/credential boundary and the installed test infrastructure; add only a small capability workflow, protocol adapter and latest-observation persistence.

**Non-Goals:** No analysis queue, background resumption, image library reader, geolocation validator/benchmark, comparison runner, private-photo consent, writer, model discovery, Responses adapter, custom provider options, new dependency or NAS deployment. The two-field synthetic contract is separate from CH07's full analysis-result schema.

Follow the adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md), [coding standards](../../../docs/engineering/coding-standards.md) and [ADR-07](../../../docs/engineering/decisions/ADR-07-ai-internal-packages.md).

## Decisions

### 1. A bounded synchronous operation through the existing session route

Add `POST /ai/providers/{id}/test` to the existing AI mux through a focused handler. The only request field is a positive `expectedRevision`; session identity comes from the server. Reuse the existing exact-Origin rule, strict JSON/content-type validation and safe error envelope, with a 1 KiB body ceiling. The handler passes the authenticated session authority and current owned revision to the use case. It accepts no image, asset, prompt, model or target URL from the browser.

The operation acquires nonblocking in-process limits of two tests globally and one per user, then atomically persists its running record while checking the expected active revision and enabled state. Busy/duplicate starts are rejected, not queued. This single-backend deployment needs no worker or broker. The session and exact current revision are rechecked at each probe's CH02 dispatch-admission point, including after DNS, and before exposing a current result. A changed revision stops remaining probes; CH02's general historical-revision support does not make an old revision eligible for this user test.

The provider work shares one 120-second deadline, including all probes and admission checks. A bounded completion write gets at most five further seconds, including after client cancellation; it cannot launch more network work. No headers are flushed while probes run. The feature client uses a 130-second request deadline and cancellation through response handling, rather than increasing the shared 15-second default for unrelated APIs. The deployed browser/proxy route must support this bounded response time; deterministic delayed-response tests cover the real local proxy path. An outer proxy cutting the request short is reported as interruption, not retried.

Pre-admission errors keep existing structured error conventions: unauthenticated, unavailable profile, disabled AI/profile, rejected origin/input, revision conflict, busy and denied egress. An accepted test whose outcome is successfully persisted returns its report, including unsuccessful compatibility observations; HTTP success for the test operation is not provider compatibility. A persistence failure returns a safe storage error and cannot claim durable success. The existing provider listing gains a redacted capability summary per active revision; it performs no provider calls.

### 2. Keep protocol I/O separate from capability decisions

| Owner | Responsibility |
|---|---|
| `backend/internal/ai/capabilities/` | Probe sequence, observation states, compatibility decision, shared budget and small consumer-owned persistence/authority/probe interfaces |
| `backend/internal/aiadapters/providerhttp/` | Chat Completions serialization, bounded response-envelope parsing and strict/JSON request modes through CH02's dispatcher |
| Focused `backend/aiProviderCapability*.go` | Session/API integration, owner-scoped SQLite admission/completion/listing, startup interruption recovery and composition |
| `src/features/ai/` | Typed report boundary, test request state, cancellation/stale-response handling and accessible result rendering through the public AI entry point |

The capability core cannot read Immich, call writer code, execute tools or import concrete SQL/HTTP adapters. Existing profile DTOs remain free of secrets. No new SDK or general-purpose JSON Schema library is introduced: Go's existing JSON decoder and explicit validation suffice for the closed two-field synthetic sample; the full analysis schema validator remains CH07. Reuse current Vitest/RTL, Playwright, Go and SQLite infrastructure.

### 3. Three explicit probes, with no repair or fallback call

Bundle a small set of metadata-free JPEG fixtures below 64 KiB, each showing one clearly colored geometric shape with known labels. Select one fixture per attempt using an injected fixture selector; deterministic tests select known fixtures. The prompt asks for the visible color/shape without supplying their answer, and filenames/fixture IDs are not sent. Static observations can only provide bounded evidence, not prove that a model could never guess.

All probes send the profile's exact model, `stream: false`, one text instruction and one `image_url` data URL to the approved Chat Completions operation. The fixed sequence is disclosed before invocation:

| Probe | Request mode | Success observation |
|---|---|---|
| Image | No `response_format`; request the two visible labels in a fixed plain-text form | Both parsed labels match the server-owned fixture facts |
| JSON | `response_format.type = json_object`; explicitly request JSON | One closed object with string `color` and `shape` matches the fixture |
| Strict schema | `response_format.type = json_schema`, a named two-field closed schema and `strict: true` | Provider accepts the schema request and the returned object passes the same local validation |

The schema requires both fields, forbids additional properties and allows all fixture colors/shapes rather than exposing the selected answer. Parse one complete JSON object only; reject extra text, duplicate keys, unknown fields, nulls and wrong types/values. Do not repair Markdown-wrapped or invalid output. Require a single usable assistant text response with successful completion; refusal, absent content, tool calls or non-success finish reasons do not pass. No tool/function definitions are sent or executed.

A passing image probe enables output-mode probes. An explicit JSON-mode unsupported error permits the independent strict probe; a strict-mode unsupported result can leave earlier JSON evidence intact. Authentication, unavailable-model, policy, rate-limit, server/network errors, refusal, truncation and invalid output stop remaining probes. Unsupported classification requires an allowlisted structured provider code/parameter identifying the requested mode; a generic 400/500 or free-form text is insufficient. Treat unknown failures as unverified, never as permission to downgrade. A rejected mode is recorded, not resent in a replacement mode.

Each probe uses CH02 with tighter ceilings: 256 KiB serialized request and 64 KiB response; its timeout is the remaining shared provider-work deadline. Keep CH02's header, TLS, redirect, address and single-request protections. Do not attach optional temperature, reasoning effort or either maximum-token parameter without separately established support; CH03 does not probe or claim that support. Byte/time ceilings are local safeguards, not a precise token or billing cap.

### 4. Report observations without inventing compatibility

Each capability has `supported`, `unsupported` or `unverified` plus a safe reason category. Unattempted probes remain unverified. Persist explicit numeric usage only when safely parsed; absent usage, provider request-size limits and token-limit parameter support remain null/unknown. Store requested model and any bounded provider-reported model identifier separately; never silently replace the configured identifier or claim that an alias confirms an unavailable model.

Evaluate the following rows in order; an explicit unsupported result does not become merely incomplete because later probes were correctly skipped.

| Completed evidence | Compatibility summary |
|---|---|
| A systemic failure | Failed; not currently proven compatible |
| Image explicitly unsupported or both output modes explicitly unsupported | Unsupported for this protocol |
| Image supported and strict sample supported, JSON either supported or explicitly unsupported | Strict-schema sample compatible |
| Image and JSON supported, strict explicitly unsupported | JSON-only compatible |
| Any other unverified observation | Incomplete; not currently proven compatible |

Keep attempt lifecycle (`running`, `completed`, `failed`, `canceled`, `interrupted`, `superseded`) separate from per-probe observations and current applicability. Applicability additionally requires current active revision, enabled installation/profile, matching test-protocol version and matching destination-policy fingerprint. No guessed expiry TTL makes old observations a guarantee: display the timestamp, and require a deliberate retest after a known proxy/runtime/model change. A policy change invalidates applicability even if the model stayed the same.

Completion never mutates profile configuration, model, credential or approval. A compatible JSON-only result is evidence for later analysis mode selection, not authorization for automatic fallback or private-photo disclosure. Full server validation remains mandatory in CH09 even after a strict sample passes.

### 5. Persist one latest report per profile revision

Use the next additive Goose migration after the existing migration 018; verify the next free number at implementation. Add `ai_provider_capability_checks`, keyed by `(userID, profileID, revision)` with a composite cascading foreign key to `ai_provider_versions`. Store attempt ID, protocol version, policy fingerprint, start/deadline/completion timestamps, lifecycle state and bounded sanitized observations. Keep one latest attempt per revision; an explicit new attempt replaces the previous report for that same revision, not the immutable profile configuration. This is a compatibility snapshot, not a permanent call audit or analysis-job table.

The admission transaction persists `running` only for the owned enabled active revision and only when no unexpired attempt holds it. No transaction spans a network call. Completion compares the attempt ID before writing, so an older callback cannot replace a newer report. Before each dispatch and publication, revalidate session/profile authority. Completed facts from a superseded revision can remain attached to that historical revision but are not returned as current proof. New revision listings begin untested; late UI responses are rejected by profile/revision/request-generation checks.

On startup, mark residual running attempts interrupted before accepting new tests. On read/admission, expire a running row whose overall completion deadline has passed; never resume it or retransmit probes. If completion storage fails, return a storage error; a leftover running row eventually becomes interrupted. Use a short independent cleanup context after request cancellation solely to persist terminal state. Account deletion cascades all rows, and completion cannot recreate a deleted owner/version.

No images, data URLs, prompts, plaintext/ciphertext keys or raw model responses are stored in this table or logs. Store only allowlisted reason codes, bounded timestamps/identifiers and validated outcome fields. Safe historical provider observations do not grant future authority.

### 6. Accessible Settings with explicit use and cancellation

Add small feature-owned test/result components alongside the existing profile list/form, keeping all handwritten source/tests below 500 physical lines. A saved profile card identifies the endpoint/model and offers `Test provider` with adjacent synthetic-image and usage disclosure. Unsaved edits are not testable; save them explicitly first. Read stored summaries on normal profile loading and render textual image/JSON/strict labels and the tested revision/time.

Keep an operation-specific abort controller and generation identity across fetch and response parsing. Cancel/close/unmount, account change or a new profile revision aborts the pending operation and rejects stale results. Do not rely only on the shared helper's timeout, which currently ends at response headers. The feature owns its complete-operation deadline and result guard; preserve unrelated client defaults. Cancellation can release only local waiting/I/O and cannot reverse upstream usage.

Prevent duplicate test starts locally and handle server busy conflicts from another tab. Reload reads the last persisted report; it never resubmits the POST. If a response is lost, tell the user to reload the report before deliberately starting another test. Existing form secret clearing, masked input and conflict handling remain intact. Use live regions for pending/result status and alerts for errors, and preserve keyboard focus and the dialog's close behavior.

### 7. Verification, rollout and evidence

Every scenario under each requirement maps to the layers below; implementation records concrete test names when they exist.

| Delta requirement | Required evidence | Task groups |
|---|---|---|
| Protect configuration and keep saves offline | Existing handler/SQLite and browser configuration checks, plus zero probe calls on create/edit/list/disable/reload | 1, 4 |
| Test only an explicitly selected owned profile with synthetic input | Real authenticated handler/SQLite and receiving HTTP fixture; exact model/image/header assertions, injected-field rejection, tenant/origin/revision/policy denials | 1 |
| Record observed image and output-mode compatibility separately | Pure outcome tests plus real serialization/parsing fixtures for strict, JSON-only, strict-only, malformed/duplicate-key/refused/truncated/tool responses and missing metadata | 2 |
| Bound capability checks without hidden fallback or retries | Controlled clocks/barriers, counted fixture calls, global/user contention, cancellation, session revocation, revision changes and systemic errors | 2, 3 |
| Keep capability evidence private durable and revision-bound | Real SQLite fresh/upgrade/reopen, foreign-key cascade, atomic attempt replacement, late completion, restart/expiry and storage-failure tests | 1, 3 |
| Expose an accessible explicit test and honest results | Vitest/RTL pending/cancel/stale/error/keyboard coverage and real Next.js proxy → Go → SQLite → synthetic-provider Playwright journeys | 4 |

Expand the browser fixture with a deterministic local provider endpoint and explicit loopback CH02 policy. Record every request, validate synthetic payloads and reject unexpected routes. CH02 intentionally ignores environment proxies, so the existing smoke deny proxy alone cannot enforce this boundary; the exact local destination policy and receiving fixture assertions must cover provider traffic. Existing manual/GPX journeys stay AI-off and unchanged. No default test talks to NAS, OpenAI or the real Immich library.

Use Plus TDD for new behavior and direct characterization for preserved behavior. Installed gates cover size/imports, formatting, lint/type checks, Go vet/race tests with backend AI statement coverage, frontend AI line/branch coverage, builds and smoke journeys. Include migration/upgrade/container checks for the additive schema and package changes. GitNexus impact is required before actual symbol edits and change analysis before a commit; query the updated CH02 checkout instead of treating this planning index as future implementation evidence.

Roll out only after CH02, take a consistent database backup, apply the additive migration and leave AI default-off until configured. Normal rollback keeps the upgraded database, disables AI and retains safe observations; do not run a destructive down migration. The existing NAS proxy/network is reused through Dockhand-managed deployment when separately requested.

The operator guide's opt-in live procedure uses this same authorized test action with a saved Sol profile, records proxy/runtime image, endpoint, requested/reported model, protocol/revision, timestamp and three observed outcomes, and labels usage unknown when absent. Fixture success never closes GATE-03 for the NAS. If Sol is unavailable, report it and let the owner choose another model in a new profile revision; do not modify the proxy's global default, upgrade its container or switch models silently. No live invocation is required to complete ordinary CI; live evidence remains an explicit release gate.

## Risks / Trade-offs

- A synthetic visual answer may be guessed and a provider may ignore strict enforcement → report a validated observation, not proof of vision quality or permanent schema guarantees.
- Three calls can consume quota and exceed the time budget on a slow model → disclose the maximum, share one deadline, stop on failures and require an explicit retry; do not increase global HTTP timeouts.
- Account/runtime model access differs from public model documentation → verify through the existing proxy and retain manual model entry; its stale `/models` listing is not a capability gate.
- A browser disconnect may happen after provider work → persist safe completion when possible, offer read-only reload and never auto-resubmit.
- Latest-only storage loses earlier attempts for the same revision → keep bounded compatibility evidence appropriate to this feature; full analysis history belongs elsewhere.
- Planned CH02 interfaces can evolve during its implementation → reconcile this design against the implemented guarded dispatcher before applying CH03 without weakening its egress rules.
