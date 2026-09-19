## Why

CH01 saves private provider profiles without contacting their destinations. Before AI Locate can test or use a provider, it needs an administrator-controlled boundary that prevents requests or credentials from reaching an unapproved service, including through redirects or changing DNS answers.

## What Changes

- Add default-deny destination approval and bounded provider dispatch, independent of whether a user has saved a profile.
- Permit the existing codex-proxy on the uGreen NAS through explicit local-network approval while retaining protection against unintended internal services and cloud metadata destinations.
- Bind dispatch to the owning user, an exact provider revision and current installation/profile enablement; keep credentials confined to the approved destination.
- Preserve offline profile saves and existing browsing, manual placement and GPX behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ai-provider-configuration`: Add administrator destination policy and the guarded internal provider-dispatch contract without treating saved settings as permission to send data.

## Impact

CH02 depends on completed CH01. It affects backend provider integration, installation configuration, operator documentation and deterministic transport/security verification. It covers FR-03, the dispatch boundary of FR-04, NFR-01–03 and the provider portion of AC-09.

The existing NAS proxy is reused. No new proxy container, proxy upgrade, global model change or remote deployment is included. Capability-test API/UI belongs to CH03; model discovery, Responses support, private-photo processing, consent collection, analysis jobs and Immich writes remain outside this change. No departure from the adopted architecture, testing or coding standards is proposed.
