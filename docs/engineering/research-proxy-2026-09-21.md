# Research proxy compatibility — 21 September 2026

The existing NAS `codex_proxy` container was inspected read-only and exercised with one authorized synthetic one-pixel image. No private photos, user hints, saved provider secrets or source pages were read. No proxy or container configuration was changed.

- Image: `homelab/codex-proxy:0.4.8-codex0.154.0-vision1`.
- Runtime: `pool`; sandbox: `read-only`; approval policy: `never`.
- Existing operation: `POST /v1/chat/completions`, `stream: false`, model `gpt-5.6-sol`, one image, Research instruction and strict v2 schema. The request used `max_completion_tokens: 4000`.
- Response: HTTP 200 in 8,085 ms; ordinary assistant message; `finish_reason: stop`; result schema `2.0`; outcome `unknown`; empty source list. A clue-free synthetic image should not establish a geographic location.
- The actual answer passed the addon's compiled Research JSON Schema and semantic validator with requested language `en`.
- No search-event metadata, proxy extension or new service was required for this request/answer path. This test does not prove provider-side web search availability or real-photo accuracy.

## Effective duration limit

The inspected environment has `CODEX_PROXY_TIMEOUT_MS=120000`. The installed `dist/server/config.js` reads that value into `defaultTimeoutMs`, and `dist/server/routes.js` passes it to inference. Therefore this deployment still has a two-minute upstream limit, even though the addon now supports up to ten minutes for Research. The measured successful request finished well within that limit; deliberate timeout exhaustion was not run.

Longer investigations require an authorized update of this existing runtime setting through the proxy's normal versioned configuration process. This change does not alter the proxy source or silently change its environment. Until that deployment setting is updated, a proxy timeout remains a technical failure, without automatic fallback or retransmission under the default one-call budget.

## Interpretation

This is protocol compatibility evidence for the inspected version/configuration, not a geolocation benchmark. The [comparison report](../ai-research-comparison.md) can retain real results for matched approved photos and distinguish model-estimated error from independently measured camera error. No approved matched private-photo dataset with independent camera references was executed for this change.
