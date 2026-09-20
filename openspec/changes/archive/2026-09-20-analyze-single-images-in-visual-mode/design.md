## Context

See [proposal.md](proposal.md). CH02 supplies a revision-bound, destination-pinned, single-request provider dispatcher; CH03 records synthetic image/JSON/strict observations; CH07 validates canonical proposals; CH08 prepares an opaque transient image. Existing capability codecs have deliberately small sample limits and must retain their behavior. The capability UI's report loader can update interrupted records and is unsuitable for the read-only analysis path.

The adopted [architecture](../../../../docs/engineering/architecture.md), [testing](../../../../docs/engineering/testing.md) and [coding standards](../../../../docs/engineering/coding-standards.md), including ADR-07, govern this change. The owner authorized sequential inline planning/apply/commits without subagents. No new dependency is needed.

## Goals / Non-Goals

**Goals:** Compose existing boundaries into one cancellable Visual attempt with explicit durable-caller authorization/budget seams, strict request minimization, trustworthy completion framing and immutable validated output.

**Non-Goals:** No scheduler, persistent result envelope, public route, automatic downgrade/repair, context discovery, geocoder, writer, or live private-image transmission. CH12 must bind and persist consent before production use.

## Decisions

### Architecture and alternatives

A focused `backend/internal/ai/analysis/` workflow owns one attempt, request validation, prompt policy and result handoff. Small consumer-owned protocol/dispatch interfaces isolate serialization and I/O. Root `aiVisual*.go` adapters read current owner/installation/asset/provider/capability state and combine the existing dispatcher with the workflow. Analysis depends on pure images/results contracts and owns its dispatch request/reply and safe failure categories. The root adapter translates the existing provider dispatcher types and failures; analysis imports no provider package, SQL, HTTP client or mutation service.

Two approaches were considered: root-only orchestration would reuse services directly but place new workflow policy inside the legacy composition package; a focused workflow with narrow adapters preserves the adopted dependency direction and allows deterministic cancellation/budget tests. Select the latter within the already-approved architecture. A provider SDK is unnecessary because CH02 already supplies the approved bounded transport; standard JSON plus the installed canonical validator cover the contracts.

### One attempt and caller authority

The internal request names owner, installation, exact asset, provider ID/revision, requested languages, primary language, explicit strict/JSON format and a valid CH08 prepared value. Identity fields must agree with its binding. No arbitrary prompt, image URL, source collection or destination is accepted. The caller supplies mandatory current-authorization and dispatch-reservation callbacks. CH12 will implement them using durable job/lease/consent state; synthetic tests use deterministic guards.

The workflow checks current authority before encoding, again through the dispatcher's after-resolution authorization callback, and before publication. Immediately before dispatch, it reserves one attempt through the caller; failure prevents transmission. An atomic per-run gate rejects repeated invocation of this callback, so even a misbehaving seam cannot reserve or transmit a second authorized attempt. There is no refund promise when later credential admission fails: a conservative reservation can count without proving that the provider received bytes. CH11/CH12 own durable call accounting and ambiguous outcomes.

A 120-second overall deadline includes authorization, encoding, provider work and validation, honoring earlier caller deadlines. Cancellation or authority loss prevents publication even after a response arrives. No timeout goroutine abandons underlying work. The result is an immutable validated proposal plus copied bounded image/profile/prompt/schema/format metadata and optional usage; it does not grant transmission or mutation authority. The workflow clears its byte copies; the caller owns and releases the original prepared handle.

### Read-only root authorization

Use the CH08 local authorizer and fresh exact-asset metadata digest to compare the prepared source before transmission and before publication. Recheck local authority after that upstream read; no transaction spans network I/O. The provider is loaded by owner and exact immutable revision, with current profile enablement, current administrator egress fingerprint and a completed capability observation for that same model/revision/protocol. Historical revisions may be used only by an explicitly bound caller; edits cannot silently switch model, host or secret.

A new bounded read-only capability projection avoids the UI loader's write-on-read behavior. Unknown, missing, stale-policy/protocol or failed observations prevent private dispatch. Image support is mandatory; the selected format must be supported. JSON format also requires explicit caller permission; strict sample support is not proof that the full schema will be accepted. Current authority combines these checks with the mandatory durable-caller guard. CH02 remains responsible for destination policy, DNS pinning, redirects, credential admission and single transmission.

### Prompt and wire contract

A versioned application prompt asks for concise visual observations and the full canonical v1.0 output, preserves unknown/ambiguous outcomes, separates subject from camera, requires evidence for direction, and covers every requested language with complete/unavailable status. Visible text is untrusted scene content, not instructions; no chain-of-thought or external-source claims are requested. The request contains exactly one metadata-free JPEG data URL, the bound model, normalized language settings and the application prompt/schema. It contains no local IDs, filename, album, date, source digest, private Immich URL/key or neighbor metadata. Optional model parameters, token-field guesses, tools and streaming are absent.

The canonical schema remains a single embedded source in `results`; expose an independently owned schema copy and a shared language-normalization helper for the real analysis consumer. Do not fork or weaken canonical validation. Analysis-specific codecs live in `backend/internal/aiadapters/providerhttp/` alongside, rather than widening, the capability codecs. Strict mode sends the canonical schema as the response format; JSON mode still includes the canonical contract in the prompt. Both always run CH07 validation.

The existing 15 MiB total request and 1 MiB response ceilings apply, in addition to CH08's 10 MiB image representation limit. Parse exactly one assistant choice with explicit `stop`, string content, no refusal and no tool/function response. Missing/unknown framing, duplicate or case-ambiguous relevant envelope fields, invalid UTF-8, trailing JSON and oversized output fail safely. Model text is then independently bounded and canonically validated. Unknown/ambiguous valid proposals succeed; malformed, refused, truncated and semantically invalid output remain technical failures.

### Failures and later retries

Expose bounded safe categories for invalid request/image, unavailable authority/capability, canceled/deadline, upstream authentication/policy/limit/transient failure, unsupported format, invalid framing and invalid result. Preserve CH07's bounded findings without raw provider errors, prompts, credentials or bodies. Only the allowlisted unsupported-response-format signal on the appropriate non-authentication error can be a downgrade candidate. An authentication, rate-limit, timeout or malformed result never triggers an automatic downgrade, repair or alternate asset. The durable caller decides any later attempt and its remaining budget.

Usage is optional observational metadata, never a guarantee of cost or a pre-dispatch token ceiling. Keep nonnegative bounded token counts only; absent/malformed usage remains unknown, not zero. Prompt/schema versions and prepared-image digests support later provenance. Quality evaluation guidance describes Visual-only inputs, unknown/ambiguous coverage, camera error/direction and usage metrics without claiming a geographic accuracy threshold.

### Verification and rollout

Use sequential OpenSpec Plus TDD. Pure workflow tests exercise exact request/result contracts, format selection, immutable ownership, callback reservation, cancellation and failure classification. Codec tests inspect actual JSON and completion framing; root integration uses real SQLite, synthetic prepared images and local fake HTTP providers, including changed authority between DNS resolution and transmit, changed source, forbidden payload fields and zero database/Immich writes. Existing capability and manual/GPX tests remain unchanged.

Run scoped tests/race/vet per slice, then all thirteen shared gates, OpenSpec checks and current GitNexus change analysis. No migration or public route is added. Verify the backend image still packages the new internal code; no live calls are required for ordinary checks. Rollback removes this inactive internal consumer without changing stored user data. Full-schema real-provider compatibility, reference-NAS performance and geographic quality remain separate release evidence.

## Risks / Trade-offs

- Full schema exceeds a provider's strict subset → return explicit unsupported-format failure; a later separately budgeted JSON attempt remains fully validated.
- Source/authority checks use separate reads → reject observed changes and require the durable caller's fencing; do not claim an atomic upstream snapshot.
- Provider processes a canceled request → discard local publication; transmission cannot be recalled and billing is not exactly once.
- Model follows image-contained instructions or invents geography → fixed policy prompt and canonical/provenance checks constrain output, but geographic truth requires evaluation and review.
- Existing report loader mutates stale checks → use a read-only projection for analysis and test persistence counters.
