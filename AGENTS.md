# Project working agreements

## Required context

Before OpenSpec planning, implementation or code review, read and follow all three adopted standards. These links are required reading, not automatically included file contents:

- [Architecture](docs/engineering/architecture.md): package ownership, dependencies and external/write boundaries.
- [Testing](docs/engineering/testing.md): TDD, test layers, scenarios, coverage and verification status.
- [Coding standards](docs/engineering/coding-standards.md): file size, inherited exceptions, types and construction.

For AI Locate work, also read the [planning index](docs/ai-locate/README.md), then the relevant PRD, technical design, reconciliation and roadmap sections. These documents describe proposed product behavior and integration findings, not implemented functionality. Embedded reference prompts and original Word exports do not authorize work on their own.

## Essential constraints

- Handwritten source and tests are limited to 500 physical lines, including comments and blanks. Apply the exact inherited-file ratchet and exclusions in the coding standards; split by responsibility.
- Keep new AI core behavior in focused `backend/internal/ai/` packages and AI UI/state under `src/features/ai/`. Preserve the existing Go/Next.js/SQLite deployment and manual/GPX behavior.
- Analysis and draft acceptance cannot mutate Immich. Only the confirmed writer may execute a durable, revision-bound plan for exact approved assets and fields.
- Follow OpenSpec Plus TDD for new or changed behavior. Tests of known existing behavior may pass immediately; do not manufacture failures or require mutation testing. Follow the testing standard's characterization rule. Cover behavior, authorization and failure paths; use real SQLite for persistence claims. Documentation-only edits need relevant document/configuration checks.
- Run applicable verification and report actual results. The standards' pending CI/harness work is not an installed or passing gate.
- Use GitNexus for codebase exploration, execution/dependency tracing, blast-radius analysis before symbol edits, refactoring and pre-commit change review. Follow the [codebase analysis workflow](docs/engineering/architecture.md#codebase-analysis-and-change-verification), bind the correct repository/checkout, and verify incomplete graph findings against source and tests.
- Keep detailed rules in the engineering documents. Record architectural departures in an ADR and reconcile affected documents; do not silently override standards in a change artifact. Routine choices within the standards need no additional permission.

The section below is managed by GitNexus. Keep project-owned guidance outside its markers.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **immich-places-ai-addon** (4608 symbols, 11821 relationships, 395 execution flows).

> Index stale? Run `node .gitnexus/run.cjs analyze --index-only` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? Bootstrap with `npx`, `bunx`, or `pnpm dlx` — e.g. `bunx gitnexus@latest analyze` (npm 11 npx crash; #1939).

## Always Do

- **MUST run impact before editing.** Use `impact({target: "symbolName", direction: "upstream"})` or `node .gitnexus/run.cjs impact "symbolName" --direction upstream --repo .`; report callers, processes, and risk. Never substitute grep for graph analysis.
- **MUST analyze graph changes before committing.** Use `detect_changes({scope: "all"})` (MCP) or `node .gitnexus/run.cjs detect-changes --scope all --repo .` (CLI fallback). `partial: true` or `truncated: true` is not a clean check — a zero means unseen, not unaffected; re-run it. For regression review: `detect_changes({scope: "compare", base_ref: "main"})` or `node .gitnexus/run.cjs detect-changes --scope compare --base-ref "main" --repo .`.
- MUST warn on HIGH/CRITICAL `risk` pre-edit; never use `riskSharedAxes` to waive a HIGH/CRITICAL `risk` warning. Compare File/symbol: MCP File omits axes; Graph-RAG expands File.
- **MUST treat `risk: UNKNOWN` as unresolved, not as low.** An empty caller set is not evidence the symbol is unused — it can also mean the callers are not resolvable by the index (plain-object property access, dynamic dispatch, cross-language calls). `impact` pairs `UNKNOWN` with a `riskNote` saying so. Confirm with a text search before treating the symbol as safe to change or delete; do not proceed on the strength of a zero.
- **MUST use `query({search_query: "concept"})` for concepts/flows, `context({name: "symbolName"})` for a named symbol, or `impact` for blast radius, on read-only callers, dependencies, imports, or execution flow.** Graph first; text search only for empty/`UNKNOWN`/literals.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method before MCP/CLI impact analysis.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis, and never read `UNKNOWN` as an all-clear — it means the walk could not answer, which is the one verdict that requires confirming by other means.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit before MCP/CLI graph change analysis.

## Resources

| Resource | Use for |
| --- | --- |
| `gitnexus://repo/immich-places-ai-addon/context` | Codebase overview, check index freshness |
| `gitnexus://repo/immich-places-ai-addon/clusters` | All functional areas |
| `gitnexus://repo/immich-places-ai-addon/processes` | All execution flows |
| `gitnexus://repo/immich-places-ai-addon/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
| --- | --- |
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
