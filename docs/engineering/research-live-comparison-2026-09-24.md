# Live Research comparison — 24 September 2026

## Scope

The owner supplied three photographs and authorized Visual/Research comparison.
Each photograph received one request in each mode, sequentially, using
`gpt-5.6-sol` through the existing NAS codex-proxy. No geographic hint, album label
or capture date was supplied. Requests used the committed addon's image
preparation, prompts, strict-schema codecs and complete-result validator at
`46e8fa8`, with English output and a requested 4,000-token output budget.

Clipboard PNGs contained a metadata-only EXIF chunk without an orientation tag,
which the current image preparation contract rejects. The local comparison
helper decoded/re-encoded the displayed pixels without that metadata before
normal application preparation. Both modes received identical prepared pixels.
Original file digests bind each comparison. This is an explicit fixture
normalization, not evidence that the production importer accepts these PNGs.

The temporary helper executed on the NAS because SSH forwarding was unavailable.
It used private temporary inputs and the loopback proxy, with no Immich upload,
application job insertion or mutation. This checks the production analysis
protocol and validator, not the complete authenticated job/UI path. The preview
[rollout record](nas-research-preview-2026-09-24.md) covers deployment separately.

## Actual results

| Photograph | Visual | Research | Research HTTP duration |
|---|---|---|---:|
| 1 | Valid `unknown`, 19.178 s | Rejected `invalid_json` | 212.739 s |
| 2 | Rejected `semantic_violation`, 21.928 s | Rejected `invalid_json` | 96.844 s |
| 3 | Rejected `semantic_violation`, 18.443 s | Rejected `invalid_json` | 83.056 s |

All six HTTP requests returned 200. Visual answers 2 and 3 declared `located`
while their selected candidates had no camera point; rejection preserves the
existing v1 outcome contract. Each Research response contained commentary followed
by two identical complete v2 JSON documents. The addon's complete-object parser
correctly rejected those wire responses. There was no automatic retry.

For diagnosis only, each standalone Research document was extracted and passed
through the unchanged compiled v2 schema and semantic validator with empty input
context. All three passed. Each proposed camera coordinates, a model-estimated
radius, alternatives and five answer-provided references. Those extracted
proposals are not successful application results and were not inserted into
history. Private reports preserve both the failures and separately labeled
extracted values rather than converting failures into passes.

No independent camera reference coordinates were provided. Actual geographic
error is unmeasured; model-estimated radii are not calibrated accuracy evidence.
References have not been independently certified. Photographs, exact locations,
raw responses and comparison reports are excluded from Git.

## Confirmed proxy defect

The deployed image remains `homelab/codex-proxy:0.4.8-codex0.154.0-vision1`.
Read-only inspection found the same response aggregation in the deployed adapter
and its pinned source. `submitTurnOnThread` concatenates all agent-message deltas
into one string and feeds completed messages into `appendAssistantText`.
That helper deduplicates a completion only when it equals or starts with the
whole accumulated string. A preceding commentary message defeats that check:

```text
accumulated = commentary + final JSON deltas
completed item = final JSON
returned content = commentary + final JSON + final JSON
```

A synthetic reproduction against the deployed helper produced exactly that
invalid JSON shape. Correct repair belongs at the proxy's message/event boundary:
track messages by identity, separate commentary from final answer content and
avoid appending a completed message a second time. Searching for braces or
silently accepting the last object in the addon would conceal an invalid wire
response and would not repair other proxy clients.

The existing numeric timeout setting is 120,000 ms; a complete HTTP response was
nevertheless observed after 212.739 seconds. That wall-clock measurement does
not identify queue time, per-attempt timing or internal runtime fallback. The
previous synthetic test's inferred two-minute overall ceiling is therefore not
established. No proxy configuration or source was changed during this comparison.

The addon change remains active while the separate shared-proxy repair is scoped.
The maintained Research specs have been synchronized and pass strict validation.
