# ADR-07: Put new AI core behavior in internal Go packages

**Status:** Accepted 17 September 2026.  
**Baseline:** `5e70c6165777949c9d8b50ede3b2768bcaa5df87`.  
**Amends:** The original flat `package main` placement in section 3 of the [AI technical design](../../ai-locate/TECHNICAL_DESIGN.md).

## Context

The existing backend is flat `package main`. The planning package initially extended that arrangement with AI handler, service, worker, repository and writer files. Durable analysis, review and confirmed writeback introduce distinct ownership and test boundaries. Existing large files also need a no-growth policy rather than more responsibilities.

## Decision

New AI core behavior belongs in small packages under `backend/internal/ai/`, following [architecture rules](../architecture.md). Introduce only the packages needed by each accepted change. Keep existing services in place and integrate them through narrow adapters and explicit construction in `package main`.

Workflows own their small consumer interfaces. Analysis has no mutation capability. Confirmed writeback owns the AI mutation path. No speculative search/candidate interfaces or broad generic store are required merely to prepare for future features.

## Alternatives

- Extend flat `package main`: fewer initial package edits, but it relies on convention alone to preserve the new core boundaries; superseded for new AI core behavior.
- Reorganize the whole backend: unnecessary regression surface and unrelated work; rejected for this adoption.
- Separate AI service or queue broker: adds deployment and persistence coordination without a current requirement; deferred.

## Consequences and verification

- Legacy unexported types/functions remain behind adapters; core packages never import the executable package.
- The first subpackage implementation updates the backend Docker build to include those sources and verifies the resulting image. Documentation adoption itself does not change the image build.
- Import checks enforce dependency direction and analysis/writer separation; behavior tests also prove accepting a proposal performs zero mutations.
- Go/Next.js/SQLite deployment, product scope, provider contract, existing manual behavior and migration location remain unchanged.
- Automated import checks and test harnesses are tracked by the [tooling prerequisite](../../ai-locate/OPENSPEC_ROADMAP.md#engineering-tooling-prerequisite); accepting this ADR does not claim they exist.
