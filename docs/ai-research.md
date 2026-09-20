# Research photo locations

AI Locate starts in Research mode. Select a provider and images, add an optional hint, and start analysis. Optional album labels and recorded capture times show their exact retained values; select only the context to include. The immutable selection preview and admission preserve these values even if the catalog changes later. Research does not send nearby-photo metadata or neighboring images. Visual and Context-assisted remain explicit alternatives.

AI Results shows proposed camera coordinates, model-estimated error, concise support, alternatives and available reference links. A 500-meter, multi-kilometer, city or region estimate remains usable. Unknown error does not hide coordinates. Large radii that cannot be drawn reliably remain numeric. No radius or confidence threshold approves or rejects a location. Analysis and inspection never write to Immich.

References come from the answer, without a backend source allowlist or search-log requirement. They are not certified evidence. Missing or unsafe destinations do not discard coordinates; credential-bearing or clearly local/private links are inactive. Opening a public HTTP(S) link is deliberate. Rendering does not fetch reference pages, thumbnails or favicons. Retained history works without the current photo and when execution is disabled.

## Execution and compatibility

Research uses `research-v1` and result schema `2.0`. Visual and Context-assisted keep their v1 contracts. The existing configured inference endpoint carries one prepared image, the prompt, selected languages, schema and selected context. No proxy protocol extension, search service or source table is required. Available provider-side research tools can help; a valid answer alone does not prove live web searching.

`AI_RESEARCH_TIMEOUT_SECONDS` defaults to `600` and accepts `1` through `600`. The server resolves this limit; parent cancellation or an earlier deadline always wins. Ordinary Visual/capability limits stay unchanged. Renewable leases cannot extend the absolute work deadline. Default call reservations prevent automatic retransmission after uncertain delivery/restart.

The current NAS proxy has its own shorter limit; see the [synthetic compatibility check](engineering/research-proxy-2026-09-21.md). Configure longer proxy execution only through its normal versioned configuration process. An addon timeout setting cannot override an upstream timeout.

## Rollout and rollback

Deploy frontend and backend together so v2 producers and readers agree. This change requires no SQL migration or result rewrite. Existing real SQLite migration/reopen checks remain applicable. Use the isolated preview data path and preserve production data separation.

Before rollback, disable execution and cancel or drain Research work. Prefer a rollback build retaining v2 readers; an old v1-only build cannot promise to render Research history. Preserve the database until compatible readers return. Do not run destructive down migrations.

Use the [comparison report](ai-research-comparison.md) for matched approved inputs. Estimated error and independently measured camera error remain separate. The committed fixtures are synthetic and do not establish live geolocation accuracy.
