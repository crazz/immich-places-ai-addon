## Why

AI Locate needs private, durable provider settings before any authorized provider request can exist. Users need to configure a model and manage credentials without exposing secrets, affecting another account, or silently redirecting queued work when a profile changes.

## What Changes

- Add private provider creation, revision-checked editing and logical enable/disable, with redacted settings UI and manual model entry.
- Store credentials encrypted in immutable configuration revisions, support explicit replacement/removal, and remove profile credentials when the owning account is deleted.
- Add default-off installation enablement, authenticated owner-scoped operations and explicit origin protection for settings mutations.
- Preserve existing browsing/manual/GPX behavior when AI is disabled. Profile saves perform no provider or Immich request and do not claim destination approval or capability support.
- Include the first AI package's required backend coverage enforcement, container packaging, migration/reopen tests and frontend behavior coverage.

## Capabilities

### New Capabilities

- `ai-provider-configuration`: Private revisioned provider settings, credential handling, installation enablement and accessible settings management.

### Modified Capabilities

None.

## Impact

Backend configuration, authenticated API integration, additive SQLite provider storage, focused AI core validation and persistence adapters, and an AI settings entry in the existing frontend. Deployment documentation and verification expand for the first internal AI package.

This is CH01's bounded portion of FR-03/FR-04 and the security/lifecycle requirements. Destination approval/egress is CH02; explicit synthetic capability testing is CH03. No model discovery, provider transport, photo analysis, jobs, results, drafts or Immich writeback is introduced here.
