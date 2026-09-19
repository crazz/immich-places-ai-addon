## Context

See [proposal](proposal.md) for intent and the [delta spec](specs/ai-provider-configuration/spec.md) for behavior. This design extends CH01 at checkout `dbc2a52dfae82b111b107dff702eb94c6d67cb31`, not the original package's `5e70c61` baseline.

CH01 has `providers.Validate`, owner-scoped profile/version storage, authenticated ciphertext, an origin-protected settings handler and migration 018. It has no dispatch reader or provider transport. `newAIProviderHandler` owns the `/ai/` mux; `main.go` mounts it separately from legacy protected routes. The existing Immich transport has its own retries and is not the provider transport.

The existing NAS `codex_proxy` container is the intended first destination. Read-only inspection on 19 September 2026 found image `homelab/codex-proxy:0.4.8-codex0.154.0-vision1`, network `npm_proxy`, alias `codex-proxy`, port 3466 and default model `gpt-5.6-luna`. The running Places backend already joins that network. Its profile base URL is `http://codex-proxy:3466/v1`; NAS loopback `127.0.0.1:3466` is for host-network clients and is not the address of a separate backend container. The repository's generic Compose file does not currently join `npm_proxy`; deployment-specific network attachment remains an operator concern.

GitNexus repository/context and provider queries used the matching current index. `save → updateAIProvider → resolveAIProviderSecret → decryptValue/encryptValue` was confirmed in source. A guessed route-registration symbol was absent; source established `newAIProviderHandler` as the actual seam. Graph results are discovery evidence, not a proof of cross-language route coverage.

## Goals / Non-Goals

**Goals:** A small backend dispatch boundary that combines pure destination rules, fresh owner/revision authority and a controlled HTTP adapter. Reuse Go, SQLite, the installed encryption mechanism and the existing NAS proxy within [architecture](../../../../docs/engineering/architecture.md), [testing](../../../../docs/engineering/testing.md), [coding standards](../../../../docs/engineering/coding-standards.md) and [ADR-07](../../../../docs/engineering/decisions/ADR-07-ai-internal-packages.md).

**Non-Goals:** No public generic forwarding route, capability-test UI, provider-specific model selection, new container, remote reconfiguration, private-photo processing, queue, writer, model discovery, custom headers or Responses adapter. No change to the generic application's network topology is required for development tests.

## Decisions

### 1. Exact installation policy, separate from saved profiles

Add a backend-only `AI_PROVIDER_EGRESS_POLICY` JSON setting through the existing configuration loader and Compose environment pass-through. Its value is an array of rules containing `baseURL`, `addressClass` (`public` or `local`), and `allowedCIDRs` for local rules. Empty/unset means an empty allowlist. Unknown fields, duplicates with conflicting policy and malformed values fail configuration validation when AI is enabled. With AI disabled, dispatch is unavailable and an unused policy cannot break legacy startup.

Canonical identity is lower-case ASCII hostname, scheme, effective numeric port and an unambiguous absolute base path without trailing slash. Reuse CH01's basic URL validation without tightening offline saves: dispatch policy additionally rejects percent-encoded paths, dot segments, repeated separators, backslashes, zone IDs and noncanonical numeric IP forms. Explicit default ports and absent default ports compare equally. International hostnames must be entered in ASCII form. The adapter appends only the fixed `chat/completions` operation to the approved base; callers cannot supply another URL, host, port or route.

Public rules require HTTPS and globally routable unicast addresses, excluding special-use/private ranges. Local rules permit HTTP or HTTPS and require nonempty CIDRs wholly contained in RFC1918, IPv6 ULA or loopback ranges; reject catch-all/mixed-space ranges. An exact hostname alone cannot authorize an entire private network. Canonicalize IPv4-mapped IPv6 before classification. A deny table excludes unspecified, multicast, link-local and metadata destinations, including `169.254.169.254` and `100.100.100.200`, regardless of a rule. Special-purpose address classification needs explicit positive/negative fixtures; `IsGlobalUnicast` alone is insufficient.

The NAS operator example approves only `http://codex-proxy:3466/v1` with `addressClass: local` and `allowedCIDRs: [192.168.144.0/20]`. A read-only inspection of the existing `npm_proxy` network on 19 September 2026 confirmed that subnet; recheck it at deployment rather than treating it as a portable application default. Do not commit a broad private-network allowance or auto-populate approval from a user profile. The operator guide explains how to supply the existing network's value at deployment. The same policy supports explicit loopback fixtures for offline tests. Administrator changes are loaded on backend restart, closing old connections; runtime policy hot reload is outside this change.

### 2. Keep workflow and concrete I/O ownership separate

| Owner | Responsibility |
|---|---|
| `backend/internal/ai/providers/` | Pure canonical destination/address decisions, dispatch inputs and owner/revision-aware coordination through small consumer-owned interfaces |
| New focused `backend/aiProviderDispatch*.go` adapters | Read exact immutable versions joined to current profile/user state, enforce encrypted-only secret access and translate legacy database types |
| `backend/internal/aiadapters/providerhttp/` | DNS lookup, approved-address dialing, TLS, fixed operation construction, safe header assembly, bounded request/response and error translation |
| `backend/config.go`, `backend/main.go` and focused AI composition | Validate installation configuration and construct dependencies; no inference logic in handlers |

Core policy/workflow code imports neither SQL nor HTTP clients nor Immich/writer packages. The concrete transport implements its consumer's narrow interface. CH02 exercises the internal dispatch entry through integration tests but adds no endpoint or startup probe. CH03 is its first public caller. This is a real transport boundary, not an empty hierarchy.

Use Go's existing standard `net/http`, `net/url`, `net/netip`, `net`, `context` and JSON packages, plus existing database/encryption dependencies. No new library is needed. The standard transport provides cancellation and TLS while leaving retries, proxy use and dialing explicit. An SDK or the installed retrying Immich client would add behavior that this boundary must suppress; an external egress service would change the accepted deployment. Pure application policy is project-specific glue around these established primitives.

### 3. Resolve, pin and authorize each request

The dispatch flow is: validate limits and server-established owner/profile/revision; load exact version and current enablement; match destination policy; resolve addresses with a deadline; reject the entire result if any address is disallowed; perform the final state/credential read; admit and execute one bounded request to a selected validated IP. Missing/foreign profiles fail before DNS. The selected IP is dialed directly without another hostname lookup; the HTTP Host and TLS ServerName remain the approved hostname.

Use a dedicated direct transport with environment proxies, cookie jars, redirects, connection reuse and transparent compression disabled. A fresh connection per dispatch prevents an earlier connection from bypassing later DNS/policy validation and removes retry-on-reused-connection behavior. Leave request replay support and idempotency headers unset. A dial failure is returned to the caller rather than triggering another application request. All 3xx responses, including 307/308 and same-origin responses, are terminal redirect failures.

The final owner-scoped read is the dispatch-admission point. It rechecks installation/profile enablement and loads the named revision's current encrypted secret state after DNS; no stale decrypted credential cache survives removal. It does not select the active revision. A consumer-owned authority check can impose stricter requirements at this same point: CH03 requires a live session and the still-current revision. If secret material was erased, the call can proceed credential-free only when the caller still authorizes that state. A state change committed after admission affects subsequent calls, while that admitted request is already in flight. No SQLite transaction or database lock is held over network I/O.

Only the provider Authorization header, JSON Content-Type/Accept and a fixed application User-Agent are assembled by the adapter. A missing secret produces no Authorization header; a deliberately configured placeholder such as `local` is treated as a provider value, not as application authentication. Reject secrets with forbidden header characters without echoing them. Browser headers, cookies, Immich keys, debug dumps and arbitrary caller headers never enter this boundary.

### 4. Fixed initial bounds and safe errors

Initial transport ceilings are 120 seconds end to end, 15 MiB serialized request, 1 MiB response body and 32 KiB response headers. A caller can choose smaller limits/deadlines, never larger. DNS, dialing, TLS and response reading share the caller's bounded context. Disable transparent decompression so compressed responses cannot evade the byte cap; unsupported encodings fail explicitly. The 15 MiB request cap accommodates a later bounded 10 MiB image plus Base64 overhead; CH08 still owns actual image preparation.

The adapter reports typed categories for policy denial, unavailable/disabled profile, credential failure, resolution/TLS/connection failure, redirect, timeout/cancellation, body/header limits and upstream HTTP failure. The HTTP owner maps these to the existing safe `code/message/retryable/requestID` shape when CH03 exists. Retryability is advisory; this transport never retries. Retain numeric status and explicitly allowlisted structured provider error codes for capability interpretation, never arbitrary error text or raw bodies in ordinary logs/errors. Success bodies remain bounded untrusted bytes for the caller to validate.

### 5. Verification and scenario ownership

Each scenario under a requirement is covered by its row below; tests assert actual received requests and forbidden-call counts, not only validation return values.

| Delta requirement | Evidence required during implementation | Task group |
|---|---|---|
| Require administrator approval for provider destinations | Table-driven canonical URL/rule tests; configuration startup tests; real local HTTP destination receives only an approved request | 1 |
| Validate the actual provider network destination | Controlled resolver/dialer plus local HTTP/TLS servers: rebinding, mixed/empty answers, IPv4/IPv6 classification, metadata, proxy environment and certificate failure | 1 |
| Confine credentials to a single approved request | Real encrypted SQLite versions and receiving servers; redirected destination sees zero requests; stale/removed/corrupt keys and forbidden headers | 2 |
| Recheck profile authority at provider dispatch | Real file-backed SQLite with two users, exact versions and barriers around DNS/admission; disable/remove/delete races without sleeps | 2 |
| Bound provider transport and preserve offline settings | Limit/cancellation/failure fixtures and request counts; existing settings handlers and AI-off browser journeys remain unchanged | 3 |

Follow Plus TDD for new behavior. Known existing settings/manual/GPX checks are characterization tests and may pass immediately. Use installed race tests, explicit backend AI coverage including the new adapter path, dependency/500-line checks, formatting, lint/types, builds and full browser regression gate. Verify Compose configuration and backend container packaging when adding configuration/package paths. Bind GitNexus to the implementation checkout and run impact before symbol edits, then change analysis before any commit. Keep all source/test files within adopted limits. No live NAS inference is part of ordinary verification.

### 6. Rollout and recovery

No schema change is required; use migration 018's existing version records. Document the new backend setting with a default-deny example and a separately labeled existing-NAS example. Local tests use isolated approved loopback endpoints; generic Compose does not gain a hard dependency on the operator's external Docker network.

Deployment is a later Dockhand action, reusing the existing proxy and backend attachment to `npm_proxy`. Confirm the current CIDR and endpoint reachability there; do not recreate the proxy, publish another port or alter its default model. An administrator can disable AI or clear the policy and restart the backend to stop subsequent dispatch. Binary rollback retains existing profile data and requires no down migration. This planning change neither deploys nor runs provider calls.

## Risks / Trade-offs

- A hostname allowlist alone cannot stop rebinding → validate every resolved address and pin the actual connection, with no proxy bypass or connection reuse.
- Local HTTP has no TLS protection inside the Docker network → require exact endpoint plus narrow explicit local ranges and document reliance on the operator's network isolation. The proxy is local; model inference is still performed through the configured Codex service.
- Docker addresses/ranges can change → re-resolve each dispatch and update the explicit operator policy when the network changes; fail closed until it matches.
- Disablement cannot retract bytes already admitted for transmission → define admission and test races around it; repeat checks for every later call.
- Proxy API shape and model access can differ from advertised listings → CH03 records observed compatibility; CH02 grants no capability or quality status.
- Fresh connections add modest overhead → accept the simpler authority/network boundary for the initial low-concurrency deployment and measure before considering pooling.
