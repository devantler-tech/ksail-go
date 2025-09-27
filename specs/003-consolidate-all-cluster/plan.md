
# Implementation Plan: Consolidate Cluster Commands Under `ksail cluster`

**Branch**: `003-consolidate-all-cluster` | **Date**: 2025-09-27 | **Spec**: [`spec.md`](./spec.md)
**Input**: Feature specification from `/specs/003-consolidate-all-cluster/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary

Group all cluster lifecycle commands (`up`, `down`, `start`, `stop`, `status`, `list`)—excluding `reconcile`—beneath a new `ksail cluster` parent command while removing their root-level registrations. The `reconcile` command will remain at the top level until it is migrated to `ksail workloads reconcile` in a future change. The implementation will refactor existing command constructors in `cmd/` into a dedicated cluster command module, update Cobra wiring in `cmd/root.go`, refresh help text, and preserve current handler logic so behavior and messaging remain stable.

## Technical Context

**Language/Version**: Go 1.24.0 (per `go.mod`)
**Primary Dependencies**: Cobra CLI framework, Viper configuration manager, KSail internal `cmdhelpers` utilities
**Storage**: N/A (command dispatch only)
**Testing**: `go test` (unit + snapshot tests via `go-snaps` in `cmd/__snapshots__`)
**Target Platform**: Cross-platform CLI (macOS/Linux/Windows)
**Project Type**: Single Go CLI project
**Performance Goals**: CLI response <200 ms for help and command wiring (Constitution IV)
**Constraints**: Must maintain >90 % coverage, golangci-lint clean, and consistent Cobra UX (Constitution I–III)
**Scale/Scope**: Eight existing cluster lifecycle commands plus root help updates

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Code Quality Excellence**: Plan keeps existing lint and gofmt standards; refactors will reuse helpers and keep files formatted via gofmt and golangci-lint. No violations expected.
- **Testing Standards (TDD-First)**: Every command move will be accompanied by updated unit tests (e.g., `cmd/cluster/up_test.go`) before modifying implementation. Snapshot expectations updated last to preserve red→green.
- **User Experience Consistency**: Help text remains descriptive and consistent; root command only lists sanctioned commands, aligning with Cobra patterns.
- **Performance Requirements**: Refactor affects only command registration; no runtime-intensive logic introduced, so performance targets remain satisfied.

✅ Initial Constitution Check: PASS (ready for Phase 0)

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)

```text
# Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure]
```

**Structure Decision**: Option 1 (single project). Existing Go CLI layout in `cmd/`, `pkg/`, and `internal/` already conforms.

## Phase 1: Design & Contracts

Prerequisite: research.md complete

Design artifacts produced:

1. [`data-model.md`](./data-model.md) captures the new CLI command hierarchy, flag inheritance expectations, and state transitions so implementation can confirm each subcommand remains wired with existing selectors.
2. [`contracts/cluster-cli.md`](./contracts/cluster-cli.md) defines the command contract—help output requirements, flag expectations, and error handling (including guaranteed absence of legacy commands).
3. [`quickstart.md`](./quickstart.md) outlines build, test, help verification, and regression steps to validate the consolidation quickly.
4. `.specify/scripts/bash/update-agent-context.sh copilot` executed to sync agent guidance (will rerun after final edits to reflect populated plan values if needed).

Next steps during implementation will update unit and snapshot tests to encode these contracts before modifying command wiring, honoring TDD.

## Phase 2: Task Planning Approach

This section describes what the /tasks command will do - DO NOT execute during /plan

**Task Generation Strategy**:

- Load `.specify/templates/tasks-template.md` as base.
- Derive contract tasks from `contracts/cluster-cli.md` (e.g., write/update unit tests for `NewClusterCmd`, legacy command removal checks, help snapshot updates).
- Derive data-model tasks for reorganizing command constructors into `cmd/cluster/` and ensuring flag wiring remains intact.
- Quickstart steps translate into validation tasks (build, `go test`, CLI smoke tests).
- Ensure all tests are listed before implementation tasks to enforce TDD.

**Ordering Strategy**:

- Write failing unit and snapshot tests first, then move constructors, then adjust root wiring, followed by documentation/help updates, finishing with cleanup and validation.
- Mark independent subcommand test updates as `[P]` for parallel execution where files do not overlap.

**Estimated Output**: 18–22 ordered tasks in `tasks.md` (fewer than template default because only cluster commands are affected).

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan.

## Phase 3+: Future Implementation

These phases are beyond the scope of the /plan command

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking

No constitutional deviations identified; table intentionally left empty.


## Progress Tracking

This checklist is updated during execution flow

**Phase Status**:

- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:

- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
