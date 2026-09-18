# Coding standards

**Status:** Adopted 17 September 2026. Binding for new and changed code. F01 repository checks and F02 frontend unit/component tests are installed; the [roadmap](../ai-locate/OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite) retains pending application smoke tests and backend AI coverage enforcement.

This document owns code construction and size rules. See [architecture](architecture.md) for boundaries and [testing](testing.md) for verification. Existing violations do not authorize new violations or an unrelated legacy rewrite.

## File size

- Handwritten source and test files contain at most **500 physical lines**, including blank lines and comments. Aim for 200–300 where responsibilities naturally fit.
- Count newline-delimited lines, including a final nonempty line without a newline. An empty file has zero lines; CRLF counts as one line ending. Apply the check to tracked files and new files intended for inclusion.
- Include Go, TypeScript/JavaScript, CSS, handwritten declarations, SQL migrations, shell/other executable scripts, Dockerfiles, Makefiles and executable CI configuration. Handwritten tests and test helpers follow the same limit.
- Exclude documentation, lockfiles, dependency manifests, binary/media assets and non-executable fixture data. An executable fixture or a test disguised as data is still code.
- Dependency/build/tool output under `node_modules/`, `.next/`, `out/`, `coverage/`, `.git/` and `.gitnexus/`, plus the generated root `next-env.d.ts`, is outside handwritten-source scope. The inherited `backend/coverage.out` is also excluded: it is a Go coverage profile, the output format produced by `go test -coverprofile=coverage.out` from `backend/`. Other generated/vendor exclusions must name their exact path and producer; do not exclude all declarations, tests, SQL, or configuration.
- Do not delete useful comments, compress statements or create `part1`/`part2` files to pass. Split by coherent responsibility. Documentation is exempt from the code limit but should remain focused.

## Inherited size baseline

The following exceptions were verified against `5e70c6165777949c9d8b50ede3b2768bcaa5df87` on adoption. They allow staged extraction while preserving current behavior.

| File | Adoption ceiling (physical lines) |
|---|---:|
| `backend/syncService_test.go` | 1467 |
| `backend/handlers_test.go` | 1373 |
| `backend/database_test.go` | 1301 |
| `backend/database.go` | 996 |
| `backend/handlers.go` | 932 |
| `backend/syncService.go` | 700 |
| `src/shared/services/backendApi.ts` | 589 |
| `src/features/map/overview/overviewLayerClusterSync.ts` | 568 |
| `backend/databaseLibraries_test.go` | 503 |

While a listed file exceeds 500 lines, a change must not increase its count relative to the change's base revision or exceed its adoption ceiling. Lower the recorded ceiling after extraction; remove the exception once the file reaches 500 or fewer lines. The normal 500-line limit then applies. New files cannot inherit an exception. A rename must carry the same reviewed exception explicitly, without increasing its ceiling.

Future exceptions must identify the exact file/rule, rationale and removal or review condition in this document. A wildcard exclusion, lint suppression or change-local claim does not amend the standard. Do not change a ceiling merely to make a check pass. Routine work within these rules needs no additional permission.

## Construction and readability

- Give each file/module one coherent responsibility. Avoid catch-all `utils`, broad service objects and forwarding layers created only to satisfy a size check.
- Approximately 60 nonblank lines per function, nesting beyond three levels or cyclomatic complexity above ten trigger review. These are diagnostic thresholds, not mandatory extraction or approval gates for a clear linear function.
- Keep TypeScript strict mode. New AI code must not introduce unchecked `any`, blanket assertions or suppressed type errors. Treat external values as unknown and validate them at the boundary.
- Handle or deliberately propagate Go errors with safe context. Do not ignore failures or include secrets/private payloads in error messages.
- Avoid new mutable globals and service locators. Pass clocks, clients, persistence and cancellation explicitly where needed. Define interfaces at real I/O or policy seams, owned by consumers; do not add an interface for every helper.
- Use existing formatting and naming conventions. Keep broad reformatting out of feature changes. Comments explain intent, invariants and surprising constraints.
- Add abstractions, dependencies and switches only for accepted requirements or demonstrated current needs. Future Research/sequence/provider-comparison work does not justify unused V1 interfaces.
- Split large test files by behavior, such as authorization, date filtering or recovery. Share small fixture helpers when they improve clarity; preserve behavior during extraction.

## Enforcement

`bun run check:size --base <revision>` enforces the physical-line limit and this document's exact inherited baseline/ratchet, including historical policy provenance. It examines tracked and non-ignored new files and fails on invalid inputs. `bun run check` runs it with the other shared local/CI gates. See [testing: available commands](testing.md#available-commands-and-remaining-setup) for comparison bases, formatting, frontend tests and remaining verification work. Code review remains responsible for meaningful boundaries and the construction rules that structural checks cannot prove.
