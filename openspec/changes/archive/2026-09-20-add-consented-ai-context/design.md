## Context

See [proposal.md](proposal.md). This plan is grounded at `23c6ca6` in the [architecture](../../../../docs/engineering/architecture.md), [testing](../../../../docs/engineering/testing.md), [coding standards](../../../../docs/engineering/coding-standards.md) and ADR-07. GitNexus identifies the legacy suggestion flow and `RunVisual`; source confirms that suggestions fall back to upload time and include frequent locations, while the Visual runner validates against an empty source bundle. CH07 already supports Context-assisted validation and server-owned source IDs. CH11 persists Visual-only jobs; this change does not enable that consumer.

## Goals / Non-Goals

**Goals:** A separately bound context input, reusable pure evidence rules, exact provider serialization and a transient provenance handoff that CH13 can persist.

**Non-Goals:** No shared mutable catalog context, generic evidence framework, new library, migration, public endpoint, scheduler or external search. Preserve the existing Visual entry point and manual suggestions behavior.

## Decisions

### Context ownership and integration

Use a focused `backend/internal/ai/contextual/` package for consent-class validation, time/lineage policy and an opaque bounded bundle. Root `aiContext*.go` adapters own owner-qualified SQLite discovery and narrow authenticated metadata reads. Extend the existing `internal/ai/analysis` workflow with a Context-assisted entry point; keep shared one-call authority, framing and canonical validation mechanics behind the real two-mode consumer. Concrete HTTP and SQL remain outside core packages. This keeps policy independently testable without modifying legacy suggestion semantics.

The bundle binds owner, installation, target asset/source, consent version/classes and selected album. It exposes independent copies, a deterministic content digest and a minimal provider projection using opaque attempt-local source IDs. Private source asset IDs and retrieval bindings stay in the server envelope. The read adapter must inspect only allowlisted metadata; the existing image metadata parser does not yet expose capture time/GPS/album facts and must not be mistaken for this projection.

### Bounded source selection and time semantics

Use separate consent classes for `capture_time`, `selected_album`, `user_hint` and `nearby_locations`; image consent remains distinct. Unselected classes are neither fetched solely for disclosure nor serialized. Hint text is limited to 2,000 UTF-8 bytes, the single album label to 256 bytes, and the complete serialized context projection to 16 KiB. Reject malformed or oversized caller text instead of silently truncating it. Bound candidate discovery to 64 records and authorized metadata-only neighbors to six, ordered deterministically by temporal distance and ID. A bounded search may return fewer than six sources; it does not promise exhaustive neighbor discovery.

Use the existing six-hour neighbor-window value as the initial internal default, with an explicit valid range of greater than zero through 24 hours for this consumer. Preserve CH04's recorded calendar date and raw source offset. Distinguish absent, invalid, offsetless/local-unknown and offset-bearing capture times. Never substitute `fileCreatedAt` or the server timezone. Compare elapsed neighbor distance only when both timestamps have explicit offsets; otherwise omit temporal neighbors and retain the appropriate capture-time status when consented.

Candidate discovery is a local, read-only hint. Recheck each retained source through current per-user upstream access plus local visibility/library/type policy, exact installation and valid coordinates. Do not request any neighbor image. A selected album must belong to the authorized current scope and still contain the target; do not enumerate or disclose other album labels. Exclude the target itself, duplicates, hidden/inaccessible sources, sources outside the time window and known AI-origin coordinates. Unknown lineage remains explicitly unknown and is not independent verification. Do not infer GPS lineage merely because an asset has an AI analysis record.

### Evidence authority and failure boundaries

Build the authorized source set from the final provider projection. User hints, capture time, album labels and neighbor coordinates retain separate kinds; text is untrusted data, not instructions. Ordinary context inputs do not grant `SourceReportedRadius` or `ViewpointAlignment` authority. Set CH07 evidence-support flags only for actually established support; this initial metadata contract grants neither alignment nor reported-radius evidence and does not invent context-extent precision.

Unavailable optional sources are omitted during preparation with safe reason categories. Missing target authority, invalid consent, malformed bounds or unresolved storage/transport failures fail the preparation. All eligible inputs being absent is a valid empty context bundle with explicit status; it must not silently turn the requested mode into Visual. After a bundle is frozen, observed source/consent/owner/provider changes at pre-encoding, after-resolution or publication checks invalidate the attempt rather than substituting new evidence mid-call. Reads cannot provide an atomic upstream snapshot; retain this limit explicitly.

Context-assisted requests reuse one prepared target image, applicable strict/explicit JSON format rules, 120-second overall attempt deadline, 15 MiB request and 1 MiB response ceilings and exactly one caller reservation. Context construction/read work shares the caller deadline. Serialize versioned controlled instructions plus the bounded context as data, never raw suggestion objects. Validate returned references against the exact Context-assisted bundle. No repair, downgrade or additional call is hidden inside the attempt.

### Result handoff, verification and rollout

Extend the transient result metadata with mode, context/consent-policy versions, bundle digest, source records, omission categories and copied authorized evidence support. Keep canonical model JSON unchanged. CH13 will extend CH11's completion envelope and revalidation context when durable Context-assisted jobs are introduced; CH10 cannot merely pass this result to the current Visual-only completion path.

Use pure policy/serialization tests, real SQLite source-isolation fixtures and synthetic authenticated HTTP services. Cover date offsets and missing values, consent minimization, six-source/resource limits, lineage, authority changes, canonical source validation and no neighbor-image/provider-follow-up/Immich-write calls. Preserve existing Visual/capability/manual/GPX suites. Follow Plus TDD, map every scenario to a test, enforce shared gates and Docker packaging for the new core package. A paired synthetic or owner-approved evaluation must report Visual versus Context-assisted coverage, camera error, direction and misleading-context cases separately; no live quality claim follows from fixtures.

No schema migration is required. Keep the new internal entry point unreachable from public/startup composition until CH13's durable consent path exists. Rollback removes the unused context consumer without changing stored Visual records. No production photo transmission or new provider compatibility claim is authorized by this plan.

## Risks / Trade-offs

- Context may strengthen a wrong guess → preserve lineage, uncertainty and source separation; never promote it to verified truth.
- Unknown upstream lineage is common → label it unknown, exclude known AI-origin sources and make no independent-confirmation claim.
- Access can change between reads → recheck at the established dispatch/publication boundaries and discard stale bundles.
- Bounded discovery can omit useful clues → expose omission status; avoid unbounded catalog or network searches.
- Compatible providers may reject larger prompts → retain strict limits and explicit failures; verify full-schema compatibility separately.
