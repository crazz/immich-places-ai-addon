## Why

AI Locate currently restricts analysis to a photograph and rejects useful evidence or uncertainty that does not fit its narrow contract. Users need the best available camera-location estimate, an estimated error radius and a short explanation with useful links, then decide for themselves whether to use the coordinates.

## What Changes

- Add a normal Research analysis choice that combines one selected photograph per item with editable hints and optional displayed album/date context. Ask the configured AI to use available web research to compare reference photographs, maps and descriptions and challenge misleading hints.
- Return the best available camera coordinates and estimated error radius. An estimate of ±500 meters, several kilometers or more remains useful and eligible for review; there is no maximum acceptable error or minimum confidence threshold.
- Permit approximate representative points for a site, city or region when labeled as such with appropriate uncertainty. Preserve alternatives and allow unknown only when no meaningful location estimate is available; do not manufacture certainty or a point merely to fill a field.
- Accept source URLs and concise explanations returned in the AI answer. Keep safe links clickable without requiring search logs, tool-event provenance, page-access metadata, independent source verification or a proxy extension.
- Show the proposed coordinates and estimated error prominently in AI Results, with supporting links and alternatives nearby. The user decides whether to use them; this change does not apply them to Immich.
- Preserve durable jobs, private history and existing Visual/Context-assisted results. Provide ordinary application defaults and the existing single Start action, without another consent checkbox or research-metadata readiness gate.
- Compare results on the same owner-approved photographs and hints, measuring camera-location error when reference coordinates exist. Model-estimated error remains distinct from measured error; the comparison introduces no product acceptance cutoff.

## Capabilities

### New Capabilities

- `ai-web-research`: Best-effort photo-location research yielding useful coordinate estimates, estimated error and answer-provided references without collecting search metadata.

### Modified Capabilities

- `ai-provider-configuration`: Reuse the configured provider and normal image/output compatibility with finite Research defaults, without a new search-provenance test or service requirement.
- `ai-analysis-jobs`: Freeze Research inputs and retain its completed estimates and answer references under existing private durable execution controls.
- `ai-location-proposals`: Accept versioned Research results with coarse coordinates, unrestricted estimated-error magnitude and answer-provided sources while preserving structural validity and camera/subject distinctions.
- `ai-results-and-review`: Present estimated coordinates, estimated error and safe clickable references so users can judge usefulness without an automatic accuracy gate.

## Impact

The change covers the existing Go analysis, validation, provider transport configuration, jobs and SQLite result handoff, plus the Next.js AI launch and result review feature. It preserves the Go/Next.js/SQLite deployment and existing dependencies, provider credentials and destination controls. It requires no new service, key, SDK or subscription.

Reuse the existing codex-proxy request/answer path. Whether the deployed AI can use web search is an integration observation to establish during apply; missing search telemetry is not evidence that search is unavailable and is not a reason to reject a coordinate proposal. Any demonstrated runtime limitation must be reported concretely before proposing infrastructure changes. Extending codex-proxy is not a prerequisite of this change.

This revises the earlier product baseline that deferred Research, restricted city/region radius estimates and required sources to be supplied before inference. Reconcile the planning baseline with this user decision while keeping historical v1 contracts and verification records accurate. The work depends on implemented CH01–CH15, not the pending draft/writeback changes.

Non-goals: search metadata or tool-event storage; independently certifying sources or accuracy; automatic GPS/Immich writes; draft acceptance; new search infrastructure; multi-photo sequence analysis; implicit neighboring-image disclosure; guaranteed precision; or a mandatory confidence/error threshold. Returned coordinates remain proposals for the user's decision.
