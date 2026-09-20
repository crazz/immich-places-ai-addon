## Why

The existing preview transport does not establish a safe image payload for AI analysis. CH08 creates a bounded, metadata-free analysis copy from an asset the requesting user may currently access, so later Visual analysis cannot forward credentials, private source URLs or uncontrolled image data.

## What Changes

- Fetch only the exact requested, currently authorized eligible image through the server's configured Immich connection, with cancellation, finite time and byte limits, and no credential forwarding on redirects.
- Normalize orientation and size, remove source metadata, and produce one explicitly typed bounded raster payload without modifying the original asset.
- Reject unsupported, malformed, oversized or interrupted inputs without a partial prepared result, fallback to a different asset, or disclosure of private input in errors.
- Keep prepared data transient and independently owned, with cleanup on success and failure and no provider, persistence or Immich mutation authority.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-location-proposals`: add the authorized, bounded image-preparation prerequisite consumed by later analysis; retain CH07's canonical result-validation contract.

## Impact

Adds a focused internal image-preparation boundary and a narrow read-only Immich adapter, with synthetic image/HTTP tests and packaging verification. Reuses existing installation, user, catalog and selection eligibility boundaries without widening the legacy manual preview or mutation interfaces. No database migration, public analysis endpoint, provider request, worker queue, original-file modification or launch UI is introduced.

Covers FR-01/FR-04, the image boundary of FR-05/FR-06, and NFR-01–03/NFR-07 in the [PRD](../../../../docs/ai-locate/PRD.md), using the [image transport design](../../../../docs/ai-locate/TECHNICAL_DESIGN.md#42-image-preparation) and [roadmap](../../../../docs/ai-locate/OPENSPEC_ROADMAP.md). CH09 consumes the prepared payload; CH11/CH12 own durable admission and dispatch. Local fixtures establish implementation behavior; deployed Immich image compatibility requires separately recorded evidence.
