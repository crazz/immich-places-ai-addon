## Why

The application can prepare safe image copies and validate structured proposals, but cannot yet connect those boundaries through one controlled Visual analysis attempt. CH09 provides that internal analysis contract so the later durable worker can request honest, validated proposals without bypassing authorization, transport policy or call budgets.

## What Changes

- Analyze one explicitly bound prepared image using a private provider revision and its recorded capabilities, with no neighboring context, private source URL or Immich credential in the request.
- Keep each provider dispatch explicit and bounded; report unsupported structured-output mode separately so the durable caller can authorize a subsequent attempt rather than hide retries.
- Validate complete responses through the canonical result contract, preserving unknown/ambiguous outcomes, separate viewpoint/direction and exact requested language statuses.
- Return immutable validated output and bounded execution metadata for later persistence, with no writer authority and no public queue-bypassing endpoint.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-location-proposals`: add the internal single-image Visual invocation and validated response handoff.

## Impact

Adds a focused analysis workflow and integrates the existing private provider transport, capability observations, image preparation and canonical validator. CH03, CH07 and CH08 are prerequisites. CH11 owns durable execution and immutable storage; CH12 connects real submissions. CH10's context-assisted inputs, public launch/progress UI, geocoding/search, schema-repair loops and Immich writes are outside this change. No new library, service or database migration is expected.

Covers FR-04–08 and AC-02/03/10 in the [PRD](../../../../docs/ai-locate/PRD.md), following the [analysis design](../../../../docs/ai-locate/TECHNICAL_DESIGN.md#5-analysis-pipeline-and-evidence-model). Synthetic fixtures establish behavior; model quality and full-schema live compatibility remain separately measured release gates.
