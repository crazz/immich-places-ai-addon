# Reviewed description translations

CH17 adds explicit text-only translation of a saved AI draft. Open **Translate reviewed text**, review or enter the exact basis, choose scene-only or candidate-based wording, languages and a provider, then acknowledge sending that text. Opening history, accepting a draft or editing text never starts translation. No image, neighboring-photo context or Immich mutation is included.

Each language has its own retained outcome. **Refresh translations** reads progress; **Cancel translation** cancels unfinished work. Successful suggestions remain available when another language fails. To retry, select unsuccessful languages and review a new request with fresh consent. Lost submission acknowledgements can be recovered using the same frozen request identity. There are no automatic provider retries.

Select successful suggestions explicitly before adopting them. Adoption creates a local draft revision, preserves geometry and unselected descriptions, and unstages the draft. It requires the exact current draft and factual revision. Older runs remain readable but cannot be adopted into changed facts. Pending write approvals are invalidated; a possibly acting writer blocks adoption. Unsaved translation text activates the existing navigation guard and requires explicit discard when closing the panel.

## Limits and authority

- One active run per draft; 1–8 distinct normalized language tags and at most 16 KiB of reviewed UTF-8 text per run.
- The request freezes owner, installation, draft/factual revisions, provider/model revision, execution policy, reviewed text, language set and idempotency key. Altered replay is rejected.
- Per-language calls have a 120-second deadline, 64 KiB request/response caps, 16 KiB output cap, token reservations and optional estimated-cost limits. Analysis and translation share global and per-owner provider concurrency. Pending work is capped at 100 items per owner and 500 per installation.
- Profile/policy authority is checked before dispatch, after destination resolution and before publication. Malformed, refused, wrong-language or oversized replies cannot become adoptable text. Failures contain safe categories rather than private provider output.
- Cancellation retains the in-flight reservation until the sender settles and discards late text. Startup marks interrupted reservations without resending; disabling AI retains queued work and readable history.

`/ai/translations` provides protected admission and bounded history; `/ai/translations/{id}` provides private readback, with explicit `/adopt` and `/cancel` actions. All responses are `no-store`. The frontend displays retained history in pages of twenty runs and fences replies after owner, draft or editing-state changes.

## Persistence and rollout

Migration 029 adds immutable run consent, per-language outcomes and usage reservations linked to draft revisions. Account deletion cascades private history; installation rotation interrupts old work and prevents late publication. Catalog reset and ordinary job retention preserve draft-linked runs. Take a consistent SQLite backup before deployment and retain additive tables when rolling back binaries. A destructive down migration is not a recovery procedure.

Verification uses synthetic local providers, real SQLite and built-browser review. It does not establish live provider compatibility or translation quality. This change does not enable Immich writing, close CH19 live GATE-02 or deploy the application. See the [verification plan](../openspec/changes/retranslate-reviewed-ai-descriptions/verification-plan.md) for scenario coverage and release boundaries.
