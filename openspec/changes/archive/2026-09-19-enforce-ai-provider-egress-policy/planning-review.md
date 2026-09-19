# CH02 planning review

19 September 2026 · Planning only · Source baseline `dbc2a52dfae82b111b107dff702eb94c6d67cb31`

## Process and scope

OpenSpec Propose with Plus proposal, specification, design and task checklists. Reviews were performed inline in this conversation, as the user explicitly requested no subagents. Existing user decisions and adopted standards supplied the scope and architectural direction; routine choices within those standards did not require another permission round. The Plus update check was already current within its seven-day window.

The five proposal lenses resolve to: prevent unapproved provider disclosure; deliver one guarded dispatch boundary; add destination/authority/credential/limit rules to the existing provider capability; exclude testing UI, analysis and infrastructure deployment; affect only provider backend/configuration, documentation and verification. CH01 is complete; CH03 consumes this contract later.

## Artifact findings

| Artifact | Review result |
|---|---|
| Proposal | Intent, scope, non-goals, capability path and dependency align with the roadmap and existing-proxy instruction. |
| Specification | Approval, actual destination, credentials, authority and bounded transport each have success/failure/edge scenarios. Corrected credential-removal wording to match the explicit dispatch-admission boundary rather than promise recall of an in-flight request. |
| Design | Source seams confirmed after GitNexus discovery. Covers ownership, flow, failures, verification, configuration rollout and rollback; no schema or new library required. |
| Tasks | Three dependency-ordered outcomes and eight unchecked tasks cover every requirement; verification is attached to each delivered behavior rather than a separate test-only change. |

Design trade-offs considered: in-process standard-library transport versus an SDK/retrying client or an external egress service; exact URL plus local CIDR approval versus host-only approval; per-dispatch connections versus connection reuse. The selected design stays within the existing architecture, uses no new dependency, and makes authority and network checks directly testable. Host-only approval is insufficient for the agreed DNS protection.

## Verification boundary

Strict OpenSpec validation passed for CH02. Final workspace validation passed all four items (two active changes and two maintained specs). CH02 contains five requirements, eighteen Gherkin scenarios and eight unchecked tasks. Local link/placeholder/numbering checks passed across all twelve changed Markdown files, and the ordered CH02 → CH03 delta merge preserves fourteen total provider requirements without editing the maintained spec. Scenario-to-test-layer and task-group mapping is in the design; implementation must record concrete test names as they are written. No task is implemented by this planning work. No application tests, NAS inference, proxy reconfiguration, deployment or Immich operation were performed.

Read-only NAS network inspection confirmed `npm_proxy` uses `192.168.144.0/20`; the design records it as a dated operator example to recheck before deployment, not a permissive application default. CH03 owns actual endpoint/model capability evidence. GitNexus index identity matched the source checkout; source verified an absent guessed registration symbol rather than treating an empty graph answer as proof. Cross-change review made configuration-only zero-call wording explicit and retained a consumer authority check at final dispatch admission for CH03's session/current-revision requirements.

GitNexus change analysis reported low risk for the two tracked planning-index updates and no affected execution processes. The new untracked OpenSpec artifacts were not mapped by that graph result; their direct artifact/link/delta checks are the relevant evidence. The working-tree inventory contains documentation and OpenSpec planning metadata only.
