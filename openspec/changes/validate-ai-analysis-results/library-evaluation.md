# CH07 schema-validator evaluation

19 September 2026 · Owner approved `santhosh-tekuri/jsonschema/v6` after the comparison below. The selected implementation version is v6.0.3; this planning change does not install dependencies.

The existing Go 1.25 module has no Draft 2020-12 JSON Schema validator. `encoding/json` remains the parser and the already-installed `golang.org/x/text` can support language normalization; neither replaces canonical schema validation. Handwritten duplication of every schema constraint would create a second contract to maintain.

| Candidate | Fit and trade-offs | Assessment |
|---|---|---|
| `github.com/santhosh-tekuri/jsonschema/v6` v6.0.3 | Draft 2020-12, Apache-2.0, versioned v6 API; the pinned module declares Go 1.21 and is compatible with the repository's Go 1.25 baseline. Local resource registration and an explicit rejecting loader support offline compilation. Existing x/text is compatible with its lower minimum. | Selected and owner-approved. Adds one direct validator dependency; keep its raw errors and loaders behind the local validation boundary. |
| `github.com/kaptinlin/jsonschema` | Draft 2020-12, MIT, active repository; its current main module requires Go 1.27 and adds a broader dependency set. | Does not fit the current toolchain without evaluating/pinning an older version or expanding this change into a runtime upgrade. |
| `github.com/xeipuuv/gojsonschema` | Established Go validator with Apache license text, but documents Draft 4/6/7 support rather than 2020-12. | Not compatible with the canonical dialect; do not weaken or translate the canonical schema to accommodate it. |

Repository metadata observed during planning: santhosh-tekuri had about 1,274 stars and activity on 6 August 2026; kaptinlin about 227 and activity on 14 September 2026; xeipuuv about 2,736 and activity on 28 June 2024. These are maintenance/popularity signals, not security or correctness proofs.

No dependency vulnerability audit or external security audit was established in this planning turn. Implementation must verify the pinned module/checksums and run a dependency vulnerability check, recording tool availability and actual findings. Schema compilation must use only the embedded trusted canonical resource and reject all unregistered lookups; provider output cannot choose the schema, dialect, loader or reference destination.

Sources: [santhosh-tekuri documentation](https://github.com/santhosh-tekuri/jsonschema), [v6.0.3 release](https://github.com/santhosh-tekuri/jsonschema/releases/tag/v6.0.3), [pinned module](https://github.com/santhosh-tekuri/jsonschema/blob/v6.0.3/go.mod), [compiler API](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6#Compiler.UseLoader), [kaptinlin documentation](https://github.com/kaptinlin/jsonschema), [kaptinlin module](https://github.com/kaptinlin/jsonschema/blob/main/go.mod), [xeipuuv documentation](https://github.com/xeipuuv/gojsonschema). Context7 library resolution and documentation retrieval corroborated the selected candidate's draft support and local resource registration.

The owner answered “`santhosh-tekuri/jsonschema/v6` - ok” to the question recommending v6.0.3 with the above trade-off. This resolves the explicit dependency approval required by OpenSpec Plus Design. The selected design uses this library with the repository's existing parser, language library and toolchain.
