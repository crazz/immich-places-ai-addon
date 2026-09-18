## Context

See proposal.md for intent. The adopted engineering documents govern size, dependencies and testing; this change implements the first part of their tooling prerequisite. The checkpoint baseline is recorded in `docs/engineering/verification-baseline.md`. Go race tests, vet, both application builds and TypeScript checking pass; frontend lint has five errors and five Go files need formatting. There are no AI core packages or frontend test harnesses yet.

This change is tooling-only and declares an empty product-spec delta. It introduces no runtime API, persistence, migration or deployment change. The existing Bun lockfile and Go module dependencies remain the application dependency authorities.

The first cumulative verification run also exposed a temporary-directory cleanup failure in `TestDoIncrementalSyncFallsBackToFull`. Source tracing identified five existing full/incremental sync tests that can close their SQLite/server fixtures before asynchronous frequent-location enrichment finishes. The bounded fixture repair below belongs to making this verification gate reliable; it does not change production sync behavior.

## Goals / Non-Goals

**Goals:** Make the adopted policies executable through one reproducible local/CI check surface, preserve precise inherited exceptions, and establish meaningful regression tests for the checker behavior itself.

**Non-Goals:** Implementing the later unit/browser harnesses, changing application behavior, repairing unrelated legacy architecture, adding empty AI packages, or claiming AI coverage before AI code exists. Remote branch-protection settings and deployment are outside this repository change.

## Decisions

### Tool ownership and dependencies

Keep verification tools under `scripts/checks/`, outside application modules. A small Node.js command coordinates focused checker modules and existing build/test commands. Use Node's built-in test runner for the tooling's temporary-repository fixtures; Vitest belongs to the following frontend change. Reuse the installed TypeScript compiler API for TypeScript/JavaScript dependency extraction and module resolution, and Go's standard parser for Go imports. Use a small standard-library Go helper invoked by the Node check rather than duplicating Go lexical parsing in JavaScript. No new production dependency is needed.

Separate inventory/base resolution, file-size policy, dependency analysis and command coordination. Share only actual common inputs. Keep each handwritten source and test file within the adopted limit. Node tooling receives an ESLint scope suited to JavaScript rather than inheriting TypeScript-only parser/type rules; application lint rules remain enabled.

### Source inventory, exclusions and revision comparison

Build the inventory from tracked and non-ignored new Git files, with NUL-safe path handling and deduplication. Exclude deleted paths from current-file reads while retaining base revisions where comparison requires them. Execute Git/tool subprocesses through argument arrays; do not evaluate filenames or refs as shell code.

Apply the handwritten-source classification and named exclusions in the coding standard, including SQL, scripts, Dockerfiles, Makefiles and executable CI configuration. The inherited `backend/coverage.out` contains a Go-generated coverage profile and is explicitly excluded in the standard by exact path and producer; do not infer a general `.out` exclusion from it. Ignore generated GitNexus output and the local Plus update timestamp through repository ignore rules so fresh clones behave consistently. Do not hide test helpers, source declarations or newly introduced languages behind broad exclusions.

The coding standard's inherited-size table remains the authoritative baseline. Parse its exact path/ceiling records and reject malformed, duplicate or invalid entries rather than silently treating them as absent. Do not maintain a competing exception list. Count physical lines including an unterminated last line; CRLF has one newline per line. An inherited oversized file cannot exceed either its documented ceiling or its count at the supplied change base. New paths do not inherit exceptions by size or name similarity. When a file drops below the normal limit, the exemption must be retired through the documented process.

Support the initial adoption comparison against `5e70c61`, which predates the engineering document. When the policy file has never existed in the selected base's available history, validate each current exception against the actual source at that base: the source must already exceed 500 lines, and neither the declared ceiling nor current size may exceed its base count. An explicitly carried Git rename uses the original source path for that evidence. Once a historical policy exists, its records and ceilings remain mandatory; a deleted, unreadable or malformed historical policy cannot trigger the initial-adoption fallback. Fail if required history or source provenance cannot be established. This handles first-PR review without granting an exception to new code or increasing an inherited ceiling.

Accept an explicit base revision for reproducible review and CI. Resolve it to a commit before comparing paths. Share option parsing across the focused size/dependency commands and the shared verifier; reject repeated options instead of silently replacing an earlier value, so every entry point has an unambiguous comparison base. Each command retains its own supported option set. For local convenience, default to the merge base against the available main branch reference; if that cannot be resolved, require the caller to supply a base instead of silently weakening the check. CI supplies the pull-request base or the previous push revision, handling an absent/zero push base through a documented valid ancestor. An invalid or unavailable base is a check failure.

### Dependency analysis

Resolve repository-local static frontend imports/re-exports using the existing TypeScript configuration, including aliases. Exclude type-only edges from runtime cycle detection; retain runtime imports and re-exports. External packages do not form repository-file edges. An unresolved repository-local import produces an actionable error. Compare detected runtime cycles with the same analysis at the selected base so existing cycles are reported and new cycles fail; any cycle involving new AI code fails. The current GitNexus check reports no runtime cycles, but the source-based check is authoritative for CI and tests, independent of graph-index availability.

Enforce the adopted AI dependency direction by inspecting actual imports. Core workflow/domain packages under `backend/internal/ai/` cannot import the executable, concrete adapter packages, HTTP/SQL clients or database drivers. Analysis cannot reach writeback/mutation packages, directly or transitively through another core package. Consumer interfaces and pure shared types remain allowed. Go's build checks remain the authority for Go import cycles and language validity. The parser helper reports parse/read errors rather than emitting an empty-success result.

Keep frontend AI implementation under its feature boundary and use a declared public integration surface for inbound cross-feature imports. Internal AI files can import one another. Existing shared integration contracts remain usable; do not reinterpret all existing cross-feature imports as new violations. Add concrete rules alongside real AI packages in later changes and document the limitations of structural checking: an import check does not prove runtime authorization or absence of mutations.

### Check execution, formatting and initial corrections

Provide one documented package command that runs checker tests, size/dependency checks, formatting/lint, TypeScript, Go vet/race tests and both application builds. Offer focused subcommands for development. Report the name, status and error of each invoked gate; a failed, missing or terminated command makes the overall result nonzero. Do not convert unavailable tools, malformed configuration or failed tests into a pass. Build artifacts belong in ignored/temporary locations.

Pin Node.js 22.23.2, Go 1.25.14 and Bun 1.4.2 in CI and document those verified versions. CI installs from the frozen Bun lockfile and the Go module files, and invokes the same repository command with an explicit base. Run repository checks on pull requests and relevant pushes with read-only permissions; do not publish images or call live services from the check workflow. Existing release publishing remains unchanged; branch protection is configured separately by the owner if desired.

Use gofmt for Go source and preserve the existing frontend formatting conventions. Correct only the recorded baseline lint errors: remove the unused import, retain boolean-return semantics while removing the redundant comparison, and rename the local boolean consistently. Apply the recorded five gofmt changes, which do not increase an inherited size ceiling. Existing warning findings remain visible; no lint rule or failing test is disabled to achieve green checks.

Keep the five affected sync fixtures alive until their service's existing wait group has drained. Register that wait through `t.Cleanup` immediately after constructing the service, so cleanup's LIFO order joins enrichment before closing the local HTTP servers and database and removing the temporary directory, including on fatal test exits. Limit this repair to `TestDoFullSync`, `TestDoFullSyncDoesNotMarkBackfillDoneWhenLibrarySyncFails`, `TestDoIncrementalSyncFallsBackToFull`, `TestDoIncrementalSync`, and `TestDoIncrementalSyncForcesFullWhenBackfillNeeded`. Preserve their assertions and production lifecycle behavior. Extract these cohesive lifecycle tests into a focused test file to keep the handwritten-file policy, and lower the inherited source file's documented ceiling to its measured size after extraction. No runtime changes, new helper abstraction, sleeps, disabled enrichment or retry-until-green behavior are needed.

### Testing and verification

Use isolated temporary Git repositories and files to test the actual CLI boundaries, including filenames with spaces, an unterminated final line, changed base revisions, inherited/new/removed files, malformed policy input and deterministic failure exit statuses. Test import analysis with valid and forbidden Go/TypeScript fixtures, aliases, type-only imports, transitive writer access, runtime cycles and unreadable/unresolved inputs. Fake process execution only when verifying command orchestration; separately run the real repository gates before completion.

Each tool behavior follows per-test RED/GREEN/REFACTOR. Configuration and mechanical formatting corrections use relevant configuration, lint, type, build and existing regression checks; do not invent product scenarios for an empty spec delta. The later frontend harness change supplies component behavior coverage. No image upload or live Immich/provider access is involved in these gates.

Retain the observed cleanup failure as RED evidence for the fixture correction. Validate the affected lifecycle cases in one fixed, repeated race-enabled run, then run the complete shared gate. Repetition supplements the explicit cleanup-order proof; an unchanged rerun that happens to pass would not resolve the failure.

### Adoption and maintenance

Land the corrected baseline and executable checks together. Update the standards' tooling-status sections and available commands to distinguish completed enforcement from the still-pending frontend harnesses and AI coverage. Preserve the recorded baseline as historical evidence. Reverting this change removes the tooling and local mechanical corrections without any data migration; the adopted engineering policies remain valid.

## Risks / Trade-offs

- Source inventories and base resolution can silently miss violations → use Git's NUL-safe inventory, resolve a real base commit, test new/deleted/ignored paths, and fail on missing required inputs.
- Import structure is only a partial architecture guarantee → use real parsers, test transitive prohibited paths, and retain behavioral authorization/no-write tests in the owning feature changes.
- Inherited findings could become a blanket waiver → retain exact baselines, report inherited cycles, forbid new AI cycles and keep existing warning output visible.
- CI can pass while an owner bypasses it → provide a repository check workflow without claiming remote branch protection has been enabled.
- New harnesses or AI code do not exist yet → report those stages as pending work in documentation, never as successful tests or coverage in this change.
