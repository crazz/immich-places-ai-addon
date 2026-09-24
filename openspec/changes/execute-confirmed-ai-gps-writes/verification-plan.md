# CH19 verification plan

Status: planned, not executed. CH16 and CH18 must be applied, verified and synchronized first; CH11's implemented analysis runtime is supporting evidence, not proof of mutation recovery. All CH19 slices ship together before dispatch is enabled.

| Scenarios | Tasks | Required automated evidence |
|---|---|---|
| W01–W04 | 1.1–1.2 | Protected handler + file-backed SQLite approval/plan consumption, same-key/different-key races, storage failure around commit, restart and lost-ack lookup; zero sends before durable approval. |
| W05–W07 | 2.1–2.2 | Real transport to a recording fake Immich: exact path/method/one asset/latitude+longitude only, stack siblings unchanged, valid zero, no redirects/hidden retries/fallback on dropped sockets, timeout, 429 or 5xx. |
| W08, W26 | 1.3, 5.3 | Default/global/write disablement and unsupported profile tests prove zero mutation; automated checks validate the live-runner opt-in/fixture guard. Actual GATE-02 remains separate live evidence. |
| W09–W11 | 2.1, 5.1 | Fresh GPS/source/access/credential/installation/revision/expiry changes before reservation and just before send; known unchanged plan produces readback and zero mutations. |
| W12–W13 | 2.3, 3.1, 5.1 | Barrier-controlled concurrent SQLite connections/accounts/workers; runtime singleton lock; expired-generation late completion; reopen of every reserved/sent/verifying state; unresolved target guard survives expiry. |
| W14–W18 | 3.1–3.3 | Lost-response applied GPS, baseline unchanged with sender still running, definitive completed rejection, denied/failed readback, conflicting values, bounded read cycles and explicit second-attempt reservation races. |
| W19–W20, W22 | 4.1 | Real adapter readback with finite values inside/outside numerical tolerance; injected persistence failure after remote success; missing catalog row and read-only recovery without a second mutation. |
| W21 | 4.2 | Controlled stale sync response racing verified publication through the real sync pause/drain boundary; timeout/release paths and unchanged existing API-key settings behavior. |
| W23–W25 | 4.3, 5.1–5.2 | Disable/shutdown barriers, account deletion and installation rotation during work, private audit reopening/cleanup protection, other-owner survival and sanitized errors/logs. |
| R01–R02 | 2.4 | Draft edit/reject versus dispatch reservation in real SQLite, retryable cancellation, in-progress rejection with preserved client edits, and no stale approval dispatch. |
| R03–R06 | 1.2, 3.3, 4.2–4.3 | Built-browser staged draft→preview→confirm→readback→history journey; HTTP-origin UUID fallback, keyboard/narrow-screen/tile failure, missing-GPS refresh, manual pending conflict and account/result late-response fencing. |

Every scenario receives exact automated test references and actual execution results in `implementation-verification.md` during apply. Required additional branches are corrupted stored plans/digests, unknown request fields, oversized bodies, revision precondition failures, no-op versus causal-write labeling, account-scoped errors and installation-scoped target exclusion. Use deterministic clocks, channels and barriers; avoid timing sleeps as concurrency proof.

Apply uses OpenSpec Plus TDD for changed behavior. Existing manual/GPX and sync/settings behavior may be characterized by immediately passing tests. Use real temporary SQLite and the real Goose migration chain for fresh/025/CH16/CH18 upgrades, repeated migration, connection-pool foreign keys, crash boundaries and persistence claims. Existing Vitest/RTL and real built-application Playwright harnesses are installed; add exact mutation/readback and restart fixtures in their owning slices. No new test framework is required.

Required closure: strict OpenSpec validation, fresh GitNexus symbol impact before edits and complete pre-commit change analysis; `bun run check --base <recorded-pre-change-commit>` covering file size/import boundaries, format/lint/type checks, Go vet/race/AI coverage, frontend AI coverage, both builds and acceptance journeys. Container/deployment configuration changes additionally require image build and writer-disabled startup verification. Record actual results and limitations; a fixture pass is not proof of live compatibility or complete race freedom.

## Live release evidence

GATE-02 requires explicit authorization for a disposable Immich photo and the actual configured endpoint/version. Record read-only before values, exact approved GPS pair, one observed mutation, API readback, preservation of unrelated fields and private outcome audit. Do not send the three supplied analysis photographs through write tests. Do not assume GPS clearing/rollback support or use `0,0` to mean missing. Keep private IDs and coordinates out of committed public examples.

Live mutation is not authorized by this preparation request. If a fixture is not yet authorized during implementation, complete all deterministic checks and leave `AI_WRITE_ENABLED=false`; report GATE-02 as not run rather than passed. The complete implementation can be reviewed with dispatch disabled, but real-write rollout remains pending that evidence. codex-proxy is outside this change.
