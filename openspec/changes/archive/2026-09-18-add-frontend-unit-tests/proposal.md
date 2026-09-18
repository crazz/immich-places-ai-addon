## Why

The repository now checks source rules, types and builds, but it cannot automatically verify frontend logic or component interactions. Contributors need the adopted frontend test harness before adding the AI settings and review interfaces.

## What Changes

- Establish reproducible, offline frontend unit and component testing using the harness selected in the adopted testing standard.
- Demonstrate the harness with a small set of meaningful existing pure-logic and accessible component interactions, preserving application behavior.
- Provide coverage reporting and the adopted coverage policy for new AI TypeScript, including untested files; report absent AI code as unmeasured.
- Integrate the frontend suite into the shared local/CI verification command and document focused execution and remaining verification limits.
- Keep scope to this frontend test capability. Browser/application journeys, AI product implementation, backend coverage tooling, application upgrades and broad legacy refactoring or coverage targets are non-goals.

## Capabilities

### New Capabilities

None in the application's behavior specifications. This tooling-only change declares `skip_specs: true`; the harness verifies existing behavior without adding a product capability.

### Modified Capabilities

None. Existing browsing, manual location placement and GPX behavior remain unchanged.

## Impact

Frontend development dependencies and their frozen lockfile, test configuration and focused tests, the shared verification runner, CI execution and engineering documentation. F01 `enforce-engineering-checks` is the completed prerequisite. The Vitest/React Testing Library choice follows the adopted testing standard; application framework versions and deployment topology remain unchanged. No application API, schema or stored user data changes.
