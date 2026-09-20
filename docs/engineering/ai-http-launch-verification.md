# AI launch on HTTP origins

20 September 2026. Follow-up to the [launch correction](ai-launch-simplification-verification.md), reviewed against `87d3325`.

## Correction

The NAS preview runs on HTTP. `crypto.randomUUID()` is restricted to secure contexts, so pressing Start threw before `submitJob` and incorrectly displayed the ambiguous-submission warning. Localhost browser tests did not reproduce the restriction.

Start now generates a 128-bit random hexadecimal idempotency key using `crypto.getRandomValues()`, which is available on HTTP. The backend's existing admission contract accepts a bounded string; it does not require UUID syntax. A retained submission keeps its exact key during reconciliation, while a new run gets a new key. Errors before submission display their actual error without suggesting that a run might already exist. Transport failures still retain the reconciliation warning and request.

Current API behavior was verified through Context7 against MDN's [randomUUID documentation](https://developer.mozilla.org/en-US/docs/Web/API/Crypto/randomUUID) and [getRandomValues documentation](https://developer.mozilla.org/en-US/docs/Web/API/Crypto/getRandomValues). No dependency, backend behavior, API contract or data migration changed.

## Test cycles

`src/features/ai/LaunchForm.test.tsx` initially had three tests: one-click Visual launch, optional advanced controls/provider defaults, and Context reconciliation.

| Added scenario | Observed RED | GREEN and refactor assessment |
|---|---|---|
| Launch and reconciliation without `randomUUID` | No request reached `submitJob`; the screenshot's exact exception appeared. Three existing tests passed. | Four tests passed after replacing ID creation. No extraction needed: one local key-generation expression has one consumer. |
| Local preparation failure followed by retry | The local error incorrectly included “The run may already exist.” Four existing tests passed. | Five tests passed after conditioning the warning on a retained submission. No refactor needed: existing state already distinguishes pre-submit and uncertain-submit failures. |

The existing complete browser workflow now disables `randomUUID` before page scripts run, keeps the real browser `getRandomValues`, and checks Visual launch, lost-response reconciliation with the same key, reload, and Context reanalysis with a new key. This characterizes the corrected behavior against the built application and real backend/SQLite with synthetic upstreams. It simulates the HTTP API surface on loopback; it is not a separate authenticated live NAS analysis.

## Verification

- Size and dependency checks against `87d3325`: passed.
- ESLint, route type generation and TypeScript: passed; the three documented inherited ESLint warnings remain.
- Frontend unit/component suite: 90 tests in 44 files passed. AI line coverage 94.15%, branch coverage 85.39%.
- Production frontend build: passed; the inherited Next.js middleware convention warning remains.
- `bun run check --base 87d3325 --gate smoke`: all 3 legacy and 11 AI browser journeys passed, including launch/reconciliation without `randomUUID`. The shared command rebuilt both applications as prerequisites.
- GitNexus query/context and upstream impact located the launch path through `LaunchEditor`, `LaunchForm` and `BatchLaunchSession`; pre-edit risk was LOW. After indexing the changed source, change analysis identified the launch functions and the test random-source stub, with no unexpected production symbols. Process enumeration has the existing bounded/cross-language gaps; the empty affected-process list is not evidence that no launch flow is affected. Source inspection and the browser journey cover that integration.

Go source is unchanged; this correction does not claim a new Go race/coverage run. No private photo analysis or live Immich mutation is part of verification or rollout. No subagents were used.
