# OpenSpec handoff

**Project:** `crazz/immich-places-ai-addon`  
**Planning baseline:** PRD 1.0 and Technical Design 1.0 with checkout reconciliation and adopted engineering standards, 17 September 2026  
**Checkout:** `5e70c6165777949c9d8b50ede3b2768bcaa5df87`

## Purpose and precedence

Use `PRD.md` for product outcomes, required scope, stable requirement IDs, and acceptance scenarios. Use `TECHNICAL_DESIGN.md` for proposed architecture and integration constraints. The JSON Schema is the proposed provider-output contract; server-owned identity, authorization, source records, and approval data are outside it.

Read and follow the adopted [architecture](../engineering/architecture.md), [testing](../engineering/testing.md) and [coding standards](../engineering/coding-standards.md) before artifact generation. They own engineering policy. [ADR-07](../engineering/decisions/ADR-07-ai-internal-packages.md) amends the original flat-file placement; use the updated Markdown design, not the historical Word export. The [roadmap](OPENSPEC_ROADMAP.md) owns candidate decomposition and prerequisites.

This package is reference planning material and does not claim any AI code has been implemented. The local checkout already has an `openspec/` workspace and repo-local OpenSpec/OpenSpec Plus skills from prior setup; preserve that setup. OpenSpec uses requirements/scenarios, design, and implementation tasks as separate planning artifacts. Generate the detailed change structure with the installed version when planning is requested, rather than copying potentially outdated CLI commands. [S01]

## Suggested location

The package now lives under `docs/ai-locate/`, with `contracts/` and `examples/` beside the Markdown documents and unchanged v1.0 Word files under `exports/`. Treat Markdown plus [RECONCILIATION.md](RECONCILIATION.md) as the current baseline. Do not overwrite the existing `openspec/` configuration or treat these reference documents as implemented specifications.

## Capability boundaries and change size

Use the [roadmap's eight capability areas and 26 candidate changes](OPENSPEC_ROADMAP.md#capability-areas), which supersede the package's initial capability grouping. Generate only the next bounded ready batch, with its prerequisites; do not turn an entire capability area into one large change or scaffold all V1 artifacts at once. Cross-cutting privacy and authorization requirements must be included in every applicable capability, not postponed to a final security task.

The [engineering tooling prerequisite](OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite) owns initial harnesses and enforcement before feature implementation. It does not add another AI product capability or mark an uninstalled check as passing.

## Non-negotiable planning constraints

Keep the existing Go/Next.js/SQLite architecture. Reuse the catalog, map, previews, filtering, and auth rather than create replacements. Preserve the entire V1 workflow, including nullable camera direction and multilingual descriptions. Research/search and multi-image sequences may remain later capabilities.

Accepting an AI result only saves a local draft. Every Immich mutation requires an explicit, revision-bound confirmation of exact assets and fields. The current writer expands to stack members; do not call it unchanged for AI-origin writes. Do not replace camera location with a landmark centroid, treat a model score as a calibrated probability, or let model output choose an authoritative asset ID.

No direct Immich database access, automatic writeback, public image URLs containing secrets, uncontrolled provider egress, or unbounded agent/retry loops. Keep before-values and handle ambiguous network outcomes through readback.

## Initial reconciliation

GATE-01 is complete for `5e70c6165777949c9d8b50ede3b2768bcaa5df87`; [RECONCILIATION.md](RECONCILIATION.md) records the evidence. The tracked source matched this commit. Local tooling setup was untracked and was not mistaken for committed application functionality. No application code was changed by reconciliation.

Carry these findings into later planning: separate exact-target writeback from both existing retry layers and stack expansion; reconcile timestamp filtering with day counts; preserve folder/tag/visibility selection scope; build consent-filtered context rather than forwarding existing suggestions; add durable drafts, fresh metadata reads, and explicit origin/CSRF protection. Migration `018` is available at the pinned commit. Existing Go tests were inventoried but not executed, and frontend test infrastructure is new work.

Record the target Immich version and provider/model capability results. Preserve unresolved integration gates as explicit validation work. A browser source review does not prove live server compatibility, authorization, image support, migration safety, or achievable geolocation quality.

## Ready-to-paste planning instruction

The following is a reference prompt for a future planning request. Importing or reconciling this package does not execute it.

> Read the adopted engineering standards, AI Locate PRD, Technical Design, RECONCILIATION.md, roadmap, source register and analysis-result contract. Confirm whether the implementation checkout still matches 5e70c6165777949c9d8b50ede3b2768bcaa5df87; if it differs, reconcile the changed integration paths and tests. Use OpenSpec and the required Plus skills to develop only the next requested bounded change or ready batch, honoring the tooling prerequisite and product dependencies. Maintain traceability to the applicable FR-01–14, NFR-01–08 and AC-01–12 requirements while retaining full V1 scope in the roadmap. Separate already-existing functionality from new work and unresolved compatibility gates. Preserve explicit approval, tenant isolation, camera-versus-subject separation, durable jobs, multilingual descriptions and exact stack-target semantics. Do not implement code or silently expand scope yet; first produce the planning artifacts for review. Do not claim that the supplied schema has been tested against a live model.

## Completion expectations for planning

Every product requirement should map to a specification and at least one verification scenario. Each new API and persistence transition needs authorization and failure-path coverage. Any departure from an ADR or required V1 behavior must be called out as a decision, not hidden in implementation tasks. No engineering effort percentage or delivery date is implied by these documents.

## Reference

[S01] OpenSpec official repository, inspected 17 September 2026. See `SOURCES.md`.
