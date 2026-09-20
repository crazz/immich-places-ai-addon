# ADR-08: Start AI analysis with ordinary defaults

Status: Accepted 20 September 2026, following the user's correction of the preview workflow.

## Problem

CH12/CH13 made a manually authored, revision-bound execution attestation and a separate image-disclosure checkbox prerequisites for ordinary analysis. The preview consequently displayed `policy_required` and unusable token values despite a configured provider. This was more administrative work than the requested product workflow needed.

## Decision

**Start analysis** is the explicit action authorizing the displayed selection, provider and settings. Show the destination beside the action and retain the existing exact configuration/selection record in the admission API. Do not require a second checkbox for sending the images the action explicitly analyzes. Context classes remain separately selected because they add data beyond those images. Merely opening the dialog, previewing or changing fields never dispatches inference.

When no operator restriction exists for the owner/profile, resolve deterministic application defaults bound to owner, installation, provider revision/model and egress fingerprint. Readiness reports `application-defaults`; it does not invent an operator attestation. An explicitly configured restriction still takes precedence, including blocking stale bindings until the operator updates or removes that restriction. Capability observations, destination rules, current source access and exact admission checks remain required.

Default token reservations are scheduling estimates, and the output-token field is a request to the provider, not a verified billing ceiling. The application enforces call counts, attempts, payload sizes, response sizes, concurrency and deadlines. Actual reported usage remains independent and can exceed default estimates without invalidating the provider. An explicitly attested allowance retains its existing violation behavior. Prices remain unknown without configured tariffs.

The normal form uses working defaults and keeps token/cost controls under Advanced settings. Changing the provider refreshes its default format and allowances.

## Alternatives and consequences

Keeping the mandatory JSON and checkbox preserves the previous implementation but leaves a basic feature unnecessarily difficult to launch. Fabricating an operator attestation would conceal uncertainty: the deployed codex-proxy adapter does not forward maximum-token fields. Removing all server authorization would lose exact scope and restart safety. The chosen approach simplifies the interaction while preserving those existing contracts.

No new service, dependency, schema migration or Immich mutation is introduced. Existing jobs and explicit policies keep their stored identity. Default-policy changes require a version change, so queued work cannot silently inherit different settings. This decision amends the maintained production/batch contracts, launch requirements, PRD, technical design and environment guidance; archived change records remain historical.
