# Reviewed description writes

CH20 extends the confirmed single-photo writer with one primary-language description. Select GPS and description independently. Description-only staging does not require camera coordinates and does not consume manual pending GPS or change Missing GPS membership.

## Review and approval

Description policy defaults to **preserve**. Deliberately choose **replace** or **managed append**, select one complete current language, save the draft, review its current source baseline, and stage the selected fields. Policy, language, text, field selection and baseline changes create a new revision. A new comparison and confirmation are required; an old operation never marks a newer draft saved.

The comparison displays the complete exact before/after description, language, policy, photo, selected fields, revision, digest and expiry. No trimming, Unicode normalization, newline rewriting or truncation occurs. A selected draft text is limited to 16 KiB UTF-8; observed and final descriptions are limited to 64 KiB. Valid absent/null descriptions compare as empty text while retaining their presence in the audit. Malformed or unavailable text is not treated as empty.

Only selected fields participate in baseline conflict checks and mutation. An independent GPS change does not invalidate a description-only operation. A fresh source identity or selected-field change requires renewed review. Analysis, translation, draft acceptance and preview creation cannot mutate Immich.

## Managed append ownership

Append preserves surrounding text and adds one visible block with a stable UUID and language. Subsequent confirmed appends replace only the exact previously verified owned block, including when changing primary language. Ownership is private to the application account, installation and asset and is established by durable verified write evidence.

Missing, edited, duplicated, malformed or foreign marker blocks cause a conflict. The writer does not repair, take over or duplicate them. Replacing the entire description deliberately retires active append ownership while retaining historical approvals and outcomes. The preview exposes every marker and separator before confirmation.

## Execution and outcomes

Description plans use canonical `standard-preview-v2`, bounded to 1 MiB. Existing `gps-preview-v1` bytes, digests, approval history and recovery state remain unchanged. Additive migrations 030–035 retain exact description baselines, versioned previews/operations, per-field evidence and append lineage. Both versions share preview capacity/expiry, target exclusion, the existing writer runtime and OS lock.

Each durable reserved attempt sends one PATCH to the exact approved asset with only selected `latitude`/`longitude` and/or `description`. The transport has a 20-second request bound and no hidden retry, redirect or fallback. Readback compares GPS within `1e-7` degrees and description exactly.

Per-field outcomes distinguish pending, verified, baseline, conflict and unavailable. A mixed outcome is partial only when sender completion is known. Unavailable observations and possibly acting senders remain unresolved. Verified selected GPS can refresh the local catalog independently; description-only success needs no GPS refresh. Durable verification evidence commits before local catalog publication so a catalog failure followed by an external revert cannot restore replay authority.

Explicit retry is limited to two total reserved attempts, current authority and a known completed request whose selected fields are all still at their original baseline and have never been verified successful. An accepted retry identity is recovered locally before checking newly expired or disabled eligibility. Successful fields in a partial result cannot be replayed by retrying the combined operation. A remaining-field operation requires a new review/approval and waits until the previous sender is known unable to act.

## Capability admission

Ordinary writing remains disabled by default. `AI_WRITE_ENABLED=true` and `AI_WRITE_PROFILE=immich-v3.2.2` retain the existing GPS gate; they alone do not enable description writes.

Description dispatch additionally requires an operator-provided `AI_WRITE_CAPABILITIES` JSON object with exactly these fields:

```json
{
  "version": 1,
  "installation": "<actual installation UUID>",
  "profile": "immich-v3.2.2",
  "evidence": "<reference to authorized compatibility evidence>",
  "capabilities": ["description"]
}
```

The policy must bind the actual installation and supported profile; its evidence reference must be nonempty. Unknown/duplicate keys and unsupported capabilities are rejected. A preview binds the effective capability-policy identity, so changing evidence or configuration requires a new preview. Confirmation, reservation and the final transport boundary recheck capability authority. Retained operation history and reconciliation reads remain available after disablement.

This configuration is an attestation of separately obtained endpoint/version/rights/exact-readback evidence. It is not automatic compatibility detection. Synthetic tests do not authorize enabling it against a private library. CH19 live GATE-02 and separate description compatibility remain pending; no live fixture was changed for CH20.

See the [implementation verification](../openspec/changes/write-approved-ai-descriptions/implementation-verification.md) for actual test evidence and the [GPS writer guide](ai-gps-writes.md) for inherited lifecycle controls.
