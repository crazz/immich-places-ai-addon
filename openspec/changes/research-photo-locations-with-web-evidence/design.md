## Context

See [proposal.md](proposal.md) for the outcome and scope. The user's clarification is decisive: useful approximate coordinates and estimated error take priority over search metadata, source certification and arbitrary precision thresholds. Existing codex-proxy integration is the starting point; no proxy extension is selected or required.

At base `58350f7`, the application sends one prepared image through a non-streaming Chat Completions request. `analysis/prompt.go` forbids external evidence; the v1 result validator accepts only pre-supplied sources and rejects a numeric city/region radius. Those application restrictions, rather than a demonstrated missing proxy feature, explain the concrete work here.

The Go analysis, jobs, provider transport and review packages already have authenticated composition adapters and SQLite persistence. AI result detail uses the typed validated document, so adding fields only to the prompt would lose or reject the new information at validation, persistence or frontend decoding. Existing worker, analysis and provider layers each impose a 120-second deadline; the HTTP client also has a 60-second response-header timeout. A longer Research deadline must be propagated through all applicable layers instead of changing one timer.

GitNexus was bound to the current repository and its matching indexed base. Concept queries exposed inconsistent symbol identities in some results; exact `RunOne` context and targeted source checks established the worker and transport seams. The graph findings are discovery evidence, not a complete blast-radius review. Apply must run fresh upstream impacts before symbol edits and the normal change analysis before a commit.

## Goals / Non-Goals

**Goals:** Extend the existing single-photo request/result path with a versioned Research answer, explicit approximate camera geometry and answer-provided references. Reuse the existing context preparation, durable job and read-only review boundaries. Keep execution bounds separate from geographic-error magnitude.

**Non-Goals:** A search orchestrator, source-fetching service, new proxy endpoint, source-event database, independent fact checker or quality-based acceptance policy. No new library or deployment topology is needed. Existing manual/GPX and confirmed-writer ownership remain unchanged.

## Decisions

### 1. Use the existing provider request and answer

Research is an additive analysis mode with a separate prompt version. It sends the same prepared photograph and approved model/profile revision through the current Chat Completions adapter. Extend the encoder to select the versioned schema rather than hard-coding the Visual schema. Retain the complete-response parser, supported JSON/strict modes, credential handling and egress policy.

The Research instruction asks the AI to identify likely camera locations, use its available research tools to compare public references, consider contradictory hints, and return its best available estimate even when coarse. It asks for concise reasoning summaries, estimated error and useful reference URLs, not private chain-of-thought or tool logs. A provider unable to research can still return an honest best-effort estimate; neither absent links nor absent telemetry proves a failed analysis. An explicit protocol, authentication or timeout failure remains a technical failure.

The earlier alternatives were a proxy extension exposing structured search events and an application-owned search integration. Both added machinery to satisfy a metadata requirement the user rejected. The selected design uses answer-provided links and existing inference; it does not infer live search support merely from a valid answer. A later infrastructure change needs a demonstrated limitation and its own concrete justification.

### 2. Add a v2 answer without changing historical v1 semantics

Keep the embedded v1.0 schema, fixtures and validation path for Visual/Context-assisted history. Add a v2.0 Research schema and typed document variant. Reuse the current candidate, camera/subject, language and outcome structures; add a bounded answer `sources` list containing local ID, URL, nullable title and concise relevance text. Candidate source references resolve within this answer or its explicitly supplied context. Do not add source-access, tool-event or search-log fields.

Research camera locations retain latitude, longitude, granularity, nullable `estimated_radius_m` and radius basis. Add `model_estimate` as a Research basis supported by the answer's concise explanation. Accept finite nonnegative radii without an upper quality limit, including city and region estimates. Null radius means unknown estimated error and does not erase a coordinate. A coarse representative point is explicitly approximate; the backend never copies subject coordinates into camera coordinates automatically.

A preferred but uncertain candidate can remain `located` with alternatives; the UI calls it a proposed location, not a verified fix. Equally plausible alternatives remain `ambiguous`. `unknown` is reserved for an answer with no meaningful estimate, not a radius/confidence policy decision.

Keep bounded JSON parsing, duplicate-key rejection, coordinate ranges and complete-response validation. Use the existing `santhosh-tekuri/jsonschema/v6`, Go JSON/URL packages and existing geometry helpers. No SDK, agent framework or custom schema engine is needed.

### 3. Treat references as answer content

Sources travel inside the immutable validated answer rather than in a second authority bundle or table. A source ID is local to that answer and grants no application-resource access. A bounded URL string is not independently verified. Missing or unusable optional links do not invalidate otherwise valid coordinates.

At the presentation boundary, derive safe link DTOs using the existing URL facilities: allow absolute HTTP(S), reject embedded credentials, unsafe schemes, controls and clearly local/private literals or hostnames. Do not resolve DNS or fetch a page while validating or rendering a link. This provides inert application rendering, not a guarantee about a site's later redirects or DNS. Omit unsafe destinations from active links and ordinary diagnostics; retain an optional-source warning without exposing credential-bearing text.

Render source text normally, with no HTML injection. Deliberate external links use `noopener` and `noreferrer`. Do not embed remote source images or fetch favicons/thumbnails. The link and its short relevance explanation are sufficient for this change. Existing input-context provenance remains separate from model-provided references.

### 4. Keep ownership in the current packages

| Owner | Change |
|---|---|
| `backend/internal/ai/analysis` | Research prompt and mode, existing single-attempt/image lifecycle, selected result schema |
| `backend/internal/ai/results` | Versioned document/schema/validation, model-estimated radius and answer sources |
| `backend/internal/ai/jobs` | Research admission/configuration and mode-specific finite deadline |
| `backend/internal/ai/providers` and `aiadapters/providerhttp` | Server-selected duration policy and schema-aware encoding using existing transport |
| Root `aiProduction*`, `aiContext*`, `aiJob*`, `aiResult*` adapters | Compose current authorization, frozen context, execution, atomic completion and private historical reads |
| `backend/internal/ai/review` and `src/features/ai` | Versioned safe DTOs, launch inputs, prominent coordinates/error and answer links |

No new core package is needed merely because the mode is named Research. Keep SQL and transport outside pure workflows and maintain source/test file limits. Analysis and review cannot reach a mutation client.

### 5. Preserve durable execution and simple launch

New launches initially select Research unless the user explicitly chooses Visual or Context-assisted. Research shows a free-text hint, optional displayed selected-album label and recorded capture time. Reuse the existing context size limits and freeze only these selected classes. Do not include neighboring images or metadata. The Start action binds the displayed inputs, profile and selection; no second consent action or search-capability test is added.

The existing idempotency digest includes mode and exact hint/context choices. Preparing an image, freezing inputs, reserving a call, rechecking authority and committing a result retain their current order. No database transaction spans inference. Results remain immutable and separate from future draft/approval state.

Use one provider dispatch per item by default. Preserve existing finite payload/image/answer limits and allow at most 16 answer sources, each with a bounded ID, a URL up to 2,048 UTF-8 bytes, title up to 512 bytes and relevance text up to 2,000 bytes. These are memory/input bounds, not evidence or precision requirements. Sources can be an empty list. An oversized structural answer fails normally; an individual in-bounds unusable URL only affects that link's presentation.

Set an application Research provider-work default of ten minutes, subject to stricter explicit installation restrictions. Resolve this server-side and carry the same absolute deadline through the worker, analysis, dispatcher and HTTP response wait; preserve the existing Visual and synthetic capability-test timeouts. Keep short connection/TLS limits, the existing renewable lease/heartbeat cadence and current concurrency caps. A heartbeat cannot extend the absolute provider-work deadline. With one dispatch reserved, restart cannot silently resend the photograph under default budgets.

The deployed proxy's previously observed 120-second runtime limit must be checked during integration. A provider that times out sooner is reported as such; changing an addon timer does not claim to extend the proxy's runtime. Align an existing documented runtime timeout setting if required for the deployment, through its normal versioned configuration process. This is configuration compatibility work, not a new proxy protocol or metadata feature. If the runtime cannot support longer calls without source changes, record that concrete limit and seek a separately justified adjustment rather than quietly altering a server.

### 6. Store one answer and render useful uncertainty

Persist v2 answer sources and estimated error in the existing immutable analysis JSON and continue the existing atomic terminal-history projection. Frozen hint/album/date data reuse the current context record. No source table, search-event migration or new database is planned. Prefer additive JSON configuration fields with explicit legacy defaults; if implementation proves a SQL change necessary, use the next unused additive migration and verify fresh/upgraded databases rather than editing an applied migration.

The detail view presents camera coordinates and estimated ±error first, then a short explanation, alternatives and available links. Use meters or kilometers without implying measurement precision from the number of displayed digits. Do not filter, hide or disable results by radius or confidence. For enormous radii or projection limits, keep numeric coordinates/error and explanatory text usable instead of drawing an unreliable circle. Null error displays as unknown. Existing camera/subject/heading separation and manual-map isolation remain.

Private history remains readable with execution disabled, after source-photo loss or after a cited page changes. Current photo access is independently checked. Account/history cleanup removes its answer references with the existing analysis lifecycle. Reads never rerun research or contact source sites.

### 7. Verify the outcome at existing test boundaries

Follow OpenSpec Plus TDD for changed behavior and the project's characterization exception for preserved v1 behavior. Use the installed Go race/SQLite, Vitest/RTL and built Playwright harnesses with synthetic local providers. Cover v1/v2 validation, 500-meter and multi-kilometer/city/region estimates, null and invalid radii, missing/unsafe links without coordinate loss, exact hint disclosure, idempotency, long-call cancellation/renewal, persistence reopen and owner isolation. Keep no-provider/no-source-fetch/no-Immich-write assertions at the relevant boundaries.

Use independent reference coordinates only for optional live comparison measurements. Record the same photos/hints, provider configuration, prompt/schema versions, estimated error, actual error where measurable, latency and useful links. Keep coarse and failed outcomes in the report. No accuracy or confidence threshold decides whether the UI may show a proposal. A deterministic fixture pass is not evidence of live geolocation quality.

### 8. Roll out without rewriting prior history

Ship dual-version readers and producers together, with Research admission enabled only on the compatible build. Verify legacy fixtures and a real SQLite upgrade/reopen before rollout. Use the existing isolated preview data path for an authorized integration check, preserving the production container and data separation.

A rollback first disables new Research admission and cancels or drains outstanding Research work. Prefer a rollback build retaining v2 readers so existing Research history remains usable. An older v1-only binary must not claim to support v2 history; keep the database and records intact until compatible readers are restored. No destructive down-migration or result rewrite is planned.

Reconcile the PRD, technical design, roadmap, reconciliation annotations and semantic-validation notes with the user's accepted approximate-estimate policy. Archived v1 evidence stays historical. This planning package changes no application or container behavior.

## Risks / Trade-offs

- Incorrect or overconfident AI estimates → show the supplied estimated error and explanation as a proposal; leave the decision to the user instead of imposing a cutoff.
- Invented or stale reference URLs → request useful real links, render safely and avoid independent-verification claims; do not discard useful coordinates because links are absent or unconfirmed.
- Prompted research may not be available through every configured endpoint → verify the actual existing deployment during apply and report limitations without inventing telemetry or infrastructure requirements.
- Long investigations can exceed an upstream timeout or consume shared capacity → use one dispatch, finite deadlines, existing concurrency/heartbeat controls and an explicit integration timeout check.
- Radius values can exceed useful map projection limits → retain readable numeric values and degrade the circle display without rejecting the estimate.
- Versioned result changes span persistence and frontend decoders → retain v1 fixtures, use typed variants and test real reopen plus built-browser journeys before rollout.
