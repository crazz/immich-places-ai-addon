# ADR-09: Useful approximate AI location proposals

Status: Accepted product direction on 20 September 2026; implementation is planned in [research-photo-locations-with-web-evidence](../../../openspec/changes/research-photo-locations-with-web-evidence/proposal.md).

## Context

The user wants camera coordinates and an estimated error from the photograph, their hints and available AI research. An estimate of ±500 meters, several kilometers or more can be more useful than no coordinates. The user decides whether to use it. Search logs, tool metadata and independently certified sources are not part of the requested outcome.

The historical v1 result contract accepts only backend-supplied source identities and disallows numeric city/region radii. Those restrictions are not an appropriate product gate for the new Research workflow.

## Decision

Return the best available meaningful coordinate estimate with nullable estimated error, explanation and optional answer-provided links. Do not impose an upper error threshold or minimum confidence. Permit an explicitly approximate representative point for a site, city or region. Keep true unknown outcomes and alternatives available without manufacturing a camera point.

Treat source URLs as untrusted answer content that can be rendered safely, not as independently verified evidence. No search-event, page-access or source-certification subsystem is required. Reuse the current configured provider and codex-proxy path; a proxy extension or separate search service is not a prerequisite.

Keep the existing Go/Next.js/SQLite ownership, authentication, egress and confirmed-writer boundaries. Add a versioned Research contract while preserving historical v1 validation and reads. Execution duration and payload bounds remain necessary operational limits; they are unrelated to geographic error.

## Alternatives and consequences

A strict evidence/accuracy gate would discard useful estimates and contradict the user's decision. Collecting tool-origin provenance would add a proxy and storage dependency without serving the requested output. Both are excluded from this change.

The application can retain inaccurate estimates or stale links. It presents them as AI proposals with estimated uncertainty and leaves acceptance to the user. Optional comparisons use independent reference coordinates when available and do not become a product cutoff. No automated write or draft approval is added.

## Affected documents

The AI planning index, PRD, technical design, roadmap, reconciliation annotation and semantic-validation notes link this decision. Existing maintained specs describe implemented behavior until the change is applied and synced. Archived v1 verification remains historical evidence, and this decision does not claim an implementation or deployment pass.
