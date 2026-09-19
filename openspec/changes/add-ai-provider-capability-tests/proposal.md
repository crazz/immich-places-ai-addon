## Why

A saved and approved provider destination does not prove that the selected model receives images or returns usable structured output. Users need an explicit, bounded compatibility check before relying on the existing NAS codex-proxy for AI Locate, with results that remain tied to the settings actually tested.

## What Changes

- Add an owner-requested provider test using only a bundled synthetic image, with visible usage disclosure and separate image, JSON-mode and strict-schema observations.
- Report actionable authentication, model, policy, parameter, refusal, truncation and transient failures without silently changing models, destinations or expanding shared data.
- Keep private revision-bound observations across reload/restart, invalidate their applicability when configuration changes, and distinguish interrupted tests from completed evidence.
- Add an accessible test action and result display to provider Settings while keeping profile saves offline and manual model entry available.
- Reuse the existing NAS proxy and select the intended stronger model in the addon profile; treat its availability and compatibility as observations to verify, not advertised guarantees.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-provider-configuration`: Add explicit capability tests and durable private observations; clarify that zero-outbound configuration operations remain separate from the deliberately invoked test.

## Impact

CH03 depends on implemented CH02 `enforce-ai-provider-egress-policy` and CH01 private profiles. It affects the backend provider use case and protected API, additive SQLite observation storage, feature-owned Settings UI, deterministic provider/browser fixtures and an opt-in live compatibility procedure. It covers FR-03, synthetic-only disclosure under FR-04, NFR-01–03, AC-09, the provider-output portion of AC-10 and GATE-03.

No new proxy container, proxy upgrade, default-model reconfiguration or remote deployment is included. No private Immich images, analysis-result schema implementation, geolocation benchmark, jobs, writer, model discovery, Responses adapter, automatic retry/fallback or new provider SDK is included. A successful synthetic check establishes observed compatibility only. No departure from adopted engineering standards is proposed.
