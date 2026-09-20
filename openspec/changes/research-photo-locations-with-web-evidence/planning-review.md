# Planning review

Date: 20 September 2026. Base: `58350f7cd8bc6f43d62b6279bd55d479fe3aca36`, branch `codex/plan-ai-selection-and-validation`.

## Status

The proposal, five capability deltas, design and implementation tasks are complete. They reflect the user's clarification: useful approximate coordinates and estimated error, with no maximum-error or minimum-confidence threshold, no search metadata requirement and no proxy-extension prerequisite. Implementation tasks remain unchecked. No application code or container configuration changed.

## Scope and review

OpenSpec Propose and the Plus artifact workflows were used. Reviews were conducted inline because the user explicitly prohibited subagents; this is not an independent review. The user's clarification supersedes the earlier proposed metadata architecture and resolves that design choice. Routine integration choices stay within the adopted package, dependency and deployment boundaries. [ADR-09](../../../docs/engineering/decisions/ADR-09-approximate-ai-location-proposals.md) records the accepted product direction.

| Concern | Resolution |
|---|---|
| Main output | Best meaningful camera estimate, estimated error, concise explanation and optional useful links. |
| Large error | ±500 meters and multi-kilometer/city/region estimates remain valid; no quality cutoff or confidence threshold. |
| Missing information | Null error, missing sources and absent telemetry do not discard otherwise valid coordinates. Genuine unknown and ambiguous outcomes remain possible. |
| Sources | Accept bounded references from the AI answer. No pre-supplied source list, search audit, page-access status or independent certification is required. |
| Existing provider | Reuse the configured request/answer path. Any actual search/runtime limitation is an integration finding, not an assumed requirement to extend the proxy. |
| Durable execution | Existing authority, leases, one-dispatch defaults and atomic result/history ownership remain. A longer finite Research timeout must cross all applicable timer layers. |
| Historical compatibility | Retain v1 schemas/fixtures and dual-version reads. City/region radius changes belong to the new Research contract. |
| User decision | Show coordinates and estimated uncertainty clearly; no proposal or review operation writes to Immich. |
| Verification | Every retained or new scenario is mapped in verification-plan.md; actual implementation checks and live quality measurements remain pending. |
| Baseline reconciliation | The planning index, PRD, technical design, roadmap and historical annotations link the accepted amendment without claiming implementation. |

Inline review corrected the following inconsistencies: the previous tool-source authority requirement; the Research-specific capability test; the metadata-based proxy selection; city/region radius rejection; loss of coordinates due to unusable links; and the old review requirement that recognized only backend-supplied provenance. Existing scenarios were retained with version-specific context where behavior intentionally differs.

## Source and verification limits

GitNexus was bound to the matching base checkout. Some concept-query symbol identities were inconsistent and a caller listing was truncated; exact symbol context and actual source verified the prompt, result geometry, worker, transport and result projection seams. This is planning discovery, not a complete pre-edit impact pass.

The earlier read of the user's [shared conversation](https://chatgpt.com/share/6ab0496c-77f4-83eb-88f0-392f757beba8) showed photograph/hint-driven coordinate answers with public reference links. Its claimed accuracy was not independently measured. During this update, the public web reader returned only the title. Automatic browser approval review rejected reopening it because browser origin permission could expose other signed-in ChatGPT content. No workaround was attempted. This update uses the previously read examples and the user's explicit current requirements; it does not claim a fresh browser read.

Strict OpenSpec and document checks validate planning structure, capability coverage, inherited scenarios, scenario/test mapping, English text, local links and whitespace. OpenSpec planning-complete status does not mean applied or tested application behavior. No live research request, private-photo transmission, Git push or deployment was performed for this planning update.
