## 1. Only explicitly approved destinations receive an internal provider request

- [x] 1.1 Deliver exact administrator destination approval with default denial, fail-closed enabled startup, unchanged offline profile saves and operator configuration for the existing NAS proxy; cover all scenarios of “Require administrator approval for provider destinations”.
- [x] 1.2 Deliver the owner/revision-bound internal dispatch entry with no public forwarding endpoint, initially exercising credential-free approved requests through real local HTTP fixtures; retain missing/foreign/disabled profile rejection from “Recheck profile authority at provider dispatch”.
- [x] 1.3 Enforce the actual resolved destination, local CIDR restrictions, metadata denial, pinned connection, proxy isolation and TLS verification; cover every scenario of “Validate the actual provider network destination” with controlled resolution and receiving-server evidence.

## 2. Only current authority can release the exact revision's credential

- [x] 2.1 Deliver encrypted-only credential loading for the exact owned revision, safe credential-free dispatch and a closed set of outbound headers; cover credential isolation and corrupt/removed-secret scenarios with real SQLite versions and receiver assertions.
- [x] 2.2 Recheck owner, installation/profile enablement and credential state at dispatch admission, including disable/remove/delete races and historical revision selection; complete all scenarios of “Recheck profile authority at provider dispatch” using deterministic barriers and the race suite.
- [x] 2.3 Reject every redirect without replaying the body or credential and without selecting another endpoint/model; complete “Confine credentials to a single approved request”, proving that redirect targets receive zero follow-up calls.

## 3. Provider failures remain bounded and existing workflows stay usable

- [x] 3.1 Deliver request/response/header/deadline bounds, cancellation and safe typed failures with a single application request per dispatch; cover limit, upstream-error and lost-response scenarios of “Bound provider transport and preserve offline settings”.
- [x] 3.2 Preserve offline settings, AI-disabled browsing/manual/GPX behavior and the absence of a public dispatch route with installed handler/browser regression evidence; finish the operator guide's existing-container configuration and rollback guidance, with applicable size/dependency, formatting/lint/types, race/coverage, build/container and Compose checks from the design passing and live compatibility explicitly unverified.
