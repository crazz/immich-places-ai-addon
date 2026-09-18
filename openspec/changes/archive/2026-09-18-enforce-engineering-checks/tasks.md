## 1. Contributors receive reliable file-size policy decisions

- [x] 1.1 Identify tracked and non-ignored new handwritten code with the adopted exclusions, count physical lines correctly, and report unreadable inputs or invalid change bases as failures.
- [x] 1.2 Reject new files above 500 lines and growth in inherited oversized files relative to both the documented ceiling and change base; reject malformed exceptions and require retirement of resolved exceptions.
- [x] 1.3 Expose a focused size check with actionable per-file diagnostics and regression coverage of accepted, rejected, renamed, removed and malformed repository inputs.

## 2. Contributors can detect prohibited dependencies before integration

- [x] 2.1 Resolve repository frontend runtime dependencies, including aliases and re-exports, distinguish inherited cycles from new cycles, and reject new cycles, all AI cycles and unresolved local imports.
- [x] 2.2 Reject prohibited Go AI core dependencies and transitive analysis-to-writer access while allowing consumer-owned interfaces and pure types; reject frontend imports into private AI implementation outside its public integration surface.
- [x] 2.3 Expose a focused dependency check with actionable paths and regression coverage of accepted and prohibited imports, type-only edges, parser failures and revision comparisons.

## 3. Contributors and CI obtain the same reproducible verification result

- [x] 3.1 Provide one local verification entry point and focused commands that expose every required gate's status and fail when a command fails, is missing or terminates; include regression coverage for execution failure handling.
- [x] 3.2 Make the recorded lint errors, Go formatting differences and observed sync-test cleanup race pass their existing checks without changing application behavior, relaxing rules or concealing inherited warnings.
- [x] 3.3 Run the shared checks in read-only pull-request and push CI with pinned toolchains, frozen dependencies and an explicit valid comparison base; keep artifacts and local generated state out of source control.
- [x] 3.4 Document usable commands, comparison-base selection, formatting checks and the installed enforcement boundaries while retaining historical baseline evidence and identifying the remaining frontend harness and AI coverage work.
