## Context

See [proposal](proposal.md). Plan after CH20's versioned standard-field extension; apply and synchronize it first. Current CH19 stores one `assetID`, one `ai_write_targets` row and a UNIQUE guard token per operation, so it cannot safely accept a target array by changing the frontend alone. Source/authority checks currently bind the analyzed asset. The legacy location path expands stacks and retries implicitly and remains excluded.

GitNexus navigation and source checks at `0c5bc0c` covered `aiWriteStore.confirm`, `aiWriteAttempt.Reserve`, `dispatchAuthority`, guard schema, `writeback.Executor` and the operation UI. Source verification is required where receiver relationships are unresolved. Apply must repeat impact at its changed base. Adopted [architecture](../../../docs/engineering/architecture.md), [testing](../../../docs/engineering/testing.md), [coding standards](../../../docs/engineering/coding-standards.md), ADR-07 and REC-01 govern this extension.

## Goals / Non-Goals

**Goals:** represent independently reviewed targets in one bounded durable approval, preserve target exclusion across every plan version and expose honest partial outcomes.

**Non-Goals:** a bulk Immich request, remote batch atomicity, all-library operations, stack editing, automatic viewpoint inference or text/direction propagation to siblings.

## Decisions

### D1. Introduce an explicit target-review boundary

Keep pure target selection/validation in `internal/ai/writepreview` and draft scope rules in `internal/ai/drafts`; root adapters perform fresh stack and member metadata reads with the existing credential. Catalog stack-primary filtering is a browsing rule, not authority to omit or include a member. A bounded server-held observation binds current draft revision, analyzed source, stack identity, selected member identities and individual GPS baselines. Selection starts with the analyzed photo alone; no automatic Select All action is introduced.

The user can add individually accessible image members and acknowledge each GPS baseline when the analyzed photo has GPS selected with a valid camera pair. A description-only decision cannot expand to stack targets. Exclude hidden/trashed/inaccessible/non-image members without leaking private values. A selected member that cannot be fully read makes preview creation fail rather than silently reducing scope. Limit a plan to 50 distinct targets including the analyzed photo, preserve deterministic ID ordering and reject an oversized selection without truncation. Resolve the chosen IDs against the current stack through a bounded read; reject incomplete or oversized stack responses without implying that unseen members do not exist. Reads share a 30-second preview deadline, at most four concurrent metadata reads, a 1 MiB response cap each and the v2 1 MiB total plan cap.

### D2. Version the target manifest and field matrix

Use a v3 plan discriminator with a canonical ordered target list. Each target binds source identity, acknowledged before-values and explicit field set. GPS is the same approved camera pair for every selected target; the analyzed target may independently carry CH20's description. Siblings can carry only GPS. Freeze the selected member list, stack identity and selection evidence in the digest; later additions cannot expand it. A selected sibling leaving the stack before dispatch becomes a target conflict; changes to unselected membership alone do not invalidate the approved set. No fresh stack response may replace the saved list.

Source identity and authorization are rechecked independently for each target. The original draft source must remain current before any new dispatch, while one sibling's conflict does not rewrite another member's approved baseline. The plan expires after the existing five minutes; unstarted targets cannot send after expiry. A large/slow plan may therefore finish partially and require a newly reviewed plan for remaining targets.

### D3. Make target reservation atomic and execution independent

Confirm the complete immutable manifest in one SQLite transaction and acquire all target guards or none, using a deterministic order. The existing guard table retains installation/asset uniqueness across v1/v2/v3 and accounts. Assign a unique opaque token per target and record its association in the new target state; never reuse the parent's ID under the existing UNIQUE token constraint. Release a guard only when that target's sender is known unable to act; account deletion may retain only the opaque exclusion necessary for unresolved senders.

Each target owns its standard-step attempt budget (at most two), generation, retry-from identity, sender-completion evidence, readback and refresh state. Root scheduling remains in the existing bounded writer runtime. Target-scoped fenced SQL replaces v1 operation-wide updates for v3; no shared attempt count or completion flag may settle sibling work. Aggregate operation status is a projection of target states, including partial and unresolved; it grants no send authority.

Retry/reconcile addresses an exact target and observed generation. Repeating an accepted generation returns stored state even after expiry/disablement; a new retry requires all fresh checks. Completed targets are never re-sent by a Retry failed action. Sequential per-plan execution simplifies audit and limits load; unrelated operations can use existing bounded concurrency. No database transaction spans network work.

### D4. Preserve edit and lifecycle safety across all targets

An edit cancels only safely undispatched/retryable work and invalidates the old plan; while any target has a reserved/possibly acting sender, the draft revision remains fixed. Global disablement, shutdown, owner deletion and installation rotation prevent new target sends without claiming to undo earlier ones. Hold private approval/history and relevant draft/result records through ordinary cleanup. A failed or expired target does not remove completed siblings' audit.

Use an additive migration for v3 target state and guard associations while retaining v1/v2 readers, payloads, counters and unknown-sender guards. Allocate after CH20 at apply time. Upgrade tests include overlapping old-version operations and account-deletion guards, not only an empty database.

### D5. Present and verify the exact scope

AI controls show a keyboard-operable list with photo identity, inclusion choice, each GPS before/after pair, source availability and a warning that a stack may contain different viewpoints. The full selected list and per-target fields are visible before confirmation. Resolve manual pending GPS for every overlapping selected target without altering unrelated manual work. After execution, refresh only verified targets and keep failed/unknown rows inspectable after catalog changes or reload.

The [verification plan](verification-plan.md) maps frozen membership, cross-account guards, partial outcomes, replay and compatibility to pure, real SQLite, local HTTP, RTL and built-browser tests. Run Plus TDD and full installed gates with the implementation base. Extend synthetic fixtures with mixed GPS, hidden/inaccessible and membership-changing stacks. Live stack read/write compatibility remains GATE-02; no production stack is probed through an unapproved mutation.

Back up before the additive migration. CH20's backend capability allowlist gains an independently default-off stack-GPS capability, bound to the plan's policy identity; single-photo activation never enables it implicitly. Rollback preserves every target guard and audit and requires a binary that understands v3; older GPS-only binaries cannot reconcile v3 work. Do not remove tables or restore old backups over newer operations.

## Risks / Trade-offs

- A stack does not prove equal viewpoints → require explicit member selection and per-member review.
- Membership or access can change mid-operation → stop the affected member and show partial completion; never expand or silently substitute scope.
- Acquiring all guards may reject a plan because one photo is busy → return a private-safe conflict without creating a partial approval.
- More targets amplify expiry and runtime bounds → cap scope and concurrency, expire unstarted sends honestly and require fresh review for the remainder.
