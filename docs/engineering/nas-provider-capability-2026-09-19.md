# NAS provider capability evidence — 19 September 2026

GATE-03 is complete for the endpoint, model, runtime and protocol recorded below. One explicitly authorized capability test observed support for the bundled image, JSON and strict-schema samples. This is synthetic compatibility evidence, not geolocation quality, a guarantee of strict enforcement for arbitrary output, or authorization to send private photographs.

## Configuration and execution

| Item | Observed value |
|---|---|
| Backend source | Release `v0.0.1`, commit `a78ad562f5786aa5b738d67bb47b9312687fb798` |
| Backend image index | `ghcr.io/crazz/immich-places-backend@sha256:cfaee955d53cdd59e3d2ba295880d58294f2f867f2b9efabf97d4a83f2117c15` |
| Running backend image ID | `sha256:0549437d251e1ac7d0b6d1299513ce0bfe50d7513d4f6abbd184c14d3e138ef5` |
| Existing proxy image | `homelab/codex-proxy:0.4.8-codex0.154.0-vision1` |
| Running proxy image ID | `sha256:960a39b86500568efa924287b578f8e40ae600fc37a9fd01c8fddebb60e0ed3f` |
| Endpoint | `http://codex-proxy:3466/v1` |
| Network / inspected CIDR | `npm_proxy` / `192.168.144.0/20` |
| Requested / reported model | `gpt-5.6-sol` / `gpt-5.6-sol` |
| Protocol / profile revision | `capability-v2` / `1` |
| Policy fingerprint | `11924537e71da6c26333b82a54a0f9df76b86e09f0d4fb14f757497a2e522685` |
| Attempt ID | `9fe93501-d349-47ab-b4d7-6f8023d33c43` |
| Started / completed (UTC) | `2026-09-19T20:58:26.85138988Z` / `2026-09-19T20:58:36.687431669Z` |
| Elapsed provider operation | Approximately 9.84 seconds; shared deadline 120 seconds |

Dockhand API created an isolated temporary backend on the existing proxy network, using the published image by digest. It had a fresh SQLite database in its disposable container layer, a newly generated encryption key, no host mounts and no published ports. `IMMICH_URL=http://127.0.0.1:9` prevented connection to the real library. A synthetic account registered through the normal HTTP endpoint, saved an enabled private provider profile with the proxy's non-secret `local` placeholder, and invoked the same authenticated `POST /ai/providers/{id}/test` action used by Settings exactly once with `expectedRevision: 1` and the configured Origin.

The backend's normal destination policy, session checks, revision checks, bounded dispatcher and capability runner remained active. The existing proxy forwarded only the built-in synthetic test input. No proxy configuration, global model default, production Places database or Immich data was changed.

## Result and persistence

| Observation | Result |
|---|---|
| Image | `supported` |
| JSON | `supported` |
| Strict-schema sample | `supported` |
| Lifecycle | `completed` |
| Compatibility summary | `strict-schema sample compatible` |
| Applicable to saved revision | `true` |
| Input may have consumed upstream usage | `true` |
| Reported aggregate usage | 38,339 prompt tokens; 37 completion tokens; 38,376 total tokens |

The subsequent provider-list GET returned the identical persisted report, including usage and consumption disclosure, without resubmitting a test. Usage is the proxy's reported aggregate; it is not independently verified billing or marginal cost. The reported model likewise identifies the proxy response, rather than independently attesting the upstream model implementation.

Dockhand API then removed the temporary container successfully, and a final inventory confirmed its absence. Production `immich-places` and `immich-places-backend` remained on upstream `1.0.18`, both healthy, and `codex_proxy` remained healthy on its existing image. Thus this test does not claim that the addon is deployed in the production Places UI. The local secret-free execution log is retained at `out/checks/nas-capability-2026-09-19.log` (ignored output).

## Remaining gates

Revalidate if the endpoint, model, runtime, policy or protocol changes. Production deployment and its own user-owned profile remain separate from this disposable evidence instance. GATE-02 (Immich compatibility), GATE-04 (operational decisions), analysis integration and release-quality acceptance remain open. See the [operator procedure](../ai-provider-settings.md) and [roadmap](../ai-locate/OPENSPEC_ROADMAP.md).
