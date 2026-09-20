## Context

See [proposal.md](proposal.md), [CH14 detail](../archive/2026-09-20-browse-persistent-ai-results/design.md), the [canonical schema](../../../backend/internal/ai/results/ai-analysis-result.v1.schema.json), and adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md) and [coding standards](../../../docs/engineering/coding-standards.md). At `23c6ca6`, GitNexus traces map interaction into pending-selection and group-move handlers. Source confirms that the ordinary `MapView` installs click/drop/reset actions and `LocationConfirm`; putting AI markers into that composition without isolating those actions would grant unintended manual-edit behavior. Canonical proposals already distinguish candidate camera, subject, direction, evidence references and per-language status.

## Goals / Non-Goals

**Goals:** Read-only, honest, accessible interpretation of a retained proposal with separate geometry, evidence and descriptions, including unknown/ambiguous outcomes and unavailable maps/images.

**Non-Goals:** No result mutation, candidate acceptance, draggable markers, draft creation, translation regeneration, provider/geocoder/search calls, schema migration or new library. CH16 owns actual editing and CH17 owns translation operations.

## Decisions

### Composition and trust boundaries

Keep the immutable detail DTO from CH14 as server state. Build pure typed view models under `src/features/ai/` for candidate selection, nullable geometry, heading, evidence and language status. Inspection state consists only of focused candidate ID, selected language and expanded sections; it is not a draft. Clear it on result/owner change and never infer an approved candidate from a locally focused row.

Provide a small read-only map surface in the existing map slot, reusing installed Leaflet, tile configuration and safe base map initialization through a narrow public map contract. AI owns its overlay data and presentation. The ordinary `MapView` cannot be reused wholesale: do not install its manual click/drop/group-move, draggable-marker, reset/remove-location or `LocationConfirm` consumers in AI review. Preserve existing manual state while switching surfaces and restore it on return. All AI layer/event resources are removed on unmount. Avoid a broad map rewrite or a generic overlay framework; extract only the base reuse actually needed by these two consumers.

Use non-draggable, explicitly labeled camera/subject/alternative markers and optional supported radius/heading layers. Never populate `TPendingLocation`, invoke `setLocationAction`, call manual save helpers or expose an acceptance/save button. Pan/zoom/focus are presentation-only. Model text is escaped plain text in both React and Leaflet popups; never insert model HTML or auto-open/fetch source URLs. CH14 remains responsible for current image access and private detail authorization.

### Camera, subject and candidate inspection

Initial focus uses the canonical selected candidate only when present. Ambiguous proposals start with no chosen candidate; show candidate rows and an overview of candidate markers labeled as alternatives. Focusing one is explicitly **Inspect candidate**, not approval or a persisted selection. Unknown or subject-only results show no invented camera point. Latitude/longitude zero remain valid values; null means absent.

Separate camera and subject by icon shape, text label and legend, not just color. A landmark/subject marker or approximate administrative label never becomes a camera marker. Display the candidate's granularity and warnings. Show supported non-null estimated camera radius in meters, labeled as an estimate with its basis and without a calibrated confidence claim. Null radius produces an explicit unknown-precision label and no fabricated circle; zero is displayed as supplied with the same caution, not silently treated as absent or proof of accuracy.

Keep original numeric coordinates available at all times. Account for poles and antimeridian when choosing map bounds; invalid/non-renderable projection regions fall back to numeric data without altering the stored coordinate. Auto-fit on initial result/candidate focus only, not on every render or tile retry; allow normal pan/zoom afterward.

### Heading, evidence and languages

Camera heading is the canonical horizontal optical-axis azimuth clockwise from true north in `[0,360)`. Render an arrow only for a camera point with a non-null supported direction, with degrees, method and nullable uncertainty shown as text. A camera-to-subject line, if displayed for orientation, is labeled a bearing and never supplies missing heading. Null heading remains **Unknown**; do not coerce it to north or derive pitch/roll/field of view. Large/unknown direction uncertainty must not look like a precise verified ray.

Resolve `evidence_refs` against observations inside the canonical result and `source_refs` against CH14's trusted stored source provenance; these are distinct namespaces. Distinguish model-only Visual claims, user hints and contextual sources with their available lineage. Unknown lineage is not independent verification, and omitted/private source metadata is not fabricated into a public citation. Show concise reasons, limitations and warnings; do not expose private chain-of-thought or invent research. Missing references in corrupt detail fail safely rather than producing clickable fabricated sources. No reverse-geocode or external map-search call enriches the proposal.

Render one tab per requested normalized language, with primary-language indication and the canonical `complete` or `unavailable` status plus `unavailable_reason`. Do not invent separate missing/failed translation records: a missing required language is an invalid detail, while later translation operations belong to CH17. Use the exact text and `scene_only` or candidate-specific basis; never silently copy another language. If the inspected candidate differs from a description's `candidate_id`, explicitly label the description's actual basis rather than reassigning its text. Switching tabs performs no provider call, translation retry or write. A locale change does not change requested result languages. Show stored uncertainty instead of upgrading claims through presentation.

### Accessibility and degraded presentation

Candidate controls, language tabs, evidence disclosure and map-focus actions need accessible names, keyboard operation and predictable focus. Provide a textual candidate table with camera/subject coordinate labels, granularity, radius/basis and heading/method/uncertainty; all meaningful map information is readable without pointing at a marker. Numeric values are read-only in CH15; numeric editing arrives with CH16. Use text labels for all statuses and a restrained live region for actual selection/load changes.

Tile load errors or absent current image access leave the full result panel and numeric presentation usable. Missing camera geometry produces an explanatory state rather than zooming to `0,0`. Keep panel content responsive and scrollable on narrow screens. Switching result or account aborts stale loads, removes stale layers and restores focus; source unavailability does not hide retained local evidence.

### Verification and rollout

Use pure view-model tests for null versus zero, camera/subject separation, ambiguous focus, radius/heading semantics, antimeridian/poles, evidence provenance and language status. RTL tests cover keyboard tabs/candidate controls, inert malicious text, unavailable-image/tile states and no implied approval. Browser journeys inspect known/unknown/ambiguous proposals, change language, lose tiles and return to manual mode; assert zero provider/Immich mutations and unchanged manual pending coordinates even after map clicks, drags, drops and context menus. Preserve manual/GPX map interactions after leaving AI review. Run shared frontend, backend regression, browser, boundary/size/coverage and build gates during apply.

Apply/sync CH14 first so the capability and detail contract exist; this delta only adds requirements to that capability. No new persisted format or migration is needed. Rollback removes the inspection surface while CH14 history remains readable. Geographic accuracy and live map availability are not established by synthetic presentation tests.

## Risks / Trade-offs

- Familiar map controls can imply editability; isolated read-only composition and explicit inspection labels avoid creating silent pending changes.
- Visual prominence can overstate certainty; marker/uncertainty labels and equal access to alternatives preserve the proposal's limits.
- Maps distort extreme latitudes and wrap longitude; keep canonical numeric values authoritative and degrade presentation safely.
- Retained descriptions may disagree or be incomplete; expose their exact status and warnings rather than inventing repaired text.
