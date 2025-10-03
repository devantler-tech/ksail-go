
# Implementation Plan: CLI Command Timing

**Branch**: `005-implement-timing-in` | **Date**: 2025-10-01 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/005-implement-timing-in/spec.md`

## Execution Flow (/plan command scope)

```text
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
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

**Primary Requirement**: Add timing visibility to all KSail CLI commands, displaying total elapsed time and per-stage duration to help users monitor performance of cluster operations.

**Technical Approach**: Create a new `pkg/ui/timer` package that provides timing tracking functionality. The timer will integrate with the existing `cmd/ui/notify` system by providing structured timing data (total duration, stage duration) that notify functions format and display. Timing will be shown progressively after each stage completes in the format `[stage: X|total: Y]` for multi-stage commands and `[stage: X]` for single-stage commands, using Go's Duration.String() method for consistent formatting.

## Technical Context

**Language/Version**: Go 1.24.0+
**Primary Dependencies**: Go standard library (`time` package), existing `cmd/ui/notify` package
**Storage**: N/A (in-memory timing state only)
**Testing**: Go testing framework (`testing` package), mockery for interface mocks
**Target Platform**: Linux (amd64/arm64), macOS (amd64/arm64)
**Project Type**: Single project (CLI tool with package-first design)
**Performance Goals**: <1ms overhead for timing mechanism itself
**Constraints**: Must use Go's Duration.String() for formatting; timing display only on successful completion; progressive updates after each stage
**Scale/Scope**: Single command execution scope (no persistence); timing tracked per-command invocation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Package-First Design

- ✅ **PASS**: Feature starts as `pkg/ui/timer` package before CLI integration
- ✅ **PASS**: Package will be self-contained, independently testable
- ✅ **PASS**: Designed as public API (timer can be used by external applications)
- ✅ **PASS**: Will include README.md and GoDoc comments

### II. CLI Interface

- ✅ **PASS**: Timer integrates with existing CLI commands via notify system
- ✅ **PASS**: Output to stdout (timing in success messages)
- ✅ **PASS**: Human-readable format using Duration.String()
- ⚠️ **DEFERRED**: Not a standalone CLI command; integrated into existing commands

### III. Test-First Development (NON-NEGOTIABLE)

- ✅ **PASS**: TDD workflow will be followed (tests before implementation)
- ✅ **PASS**: Tests will be generated via mockery
- ✅ **PASS**: Red-Green-Refactor cycle mandatory

### IV. Interface-Based Design

- ✅ **PASS**: Timer will be defined as interface before implementation
- ✅ **PASS**: Will use context.Context for lifecycle management
- ✅ **PASS**: Mockable via mockery tool

### V. Clean Architecture Principles

- ✅ **PASS**: Timer in `pkg/ui/` domain (UI utilities)
- ✅ **PASS**: No circular dependencies (timer → notify, not reverse)
- ✅ **PASS**: Context propagation for command lifecycle

### VI. Quality Gates

- ✅ **PASS**: All quality gates will be satisfied (lint, test, build, mocks)
- ✅ **PASS**: Pre-commit hooks will run automatically

### VII. Semantic Versioning & Conventional Commits

- ✅ **PASS**: Feature will be committed with conventional commit message
- ✅ **PASS**: Will trigger MINOR version bump (feat: new feature)

**Constitution Check Result**: ✅ **PASS** - No violations detected. All principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/005-implement-timing-in/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)

```text
pkg/ui/timer/            # New timer package
├── timer.go             # Timer interface and implementation
├── timer_test.go        # Unit tests
├── doc.go               # Package documentation
└── README.md            # Package README

cmd/ui/notify/           # Existing notify package (to be updated)
├── notify.go            # Updated to format timing display
└── notify_test.go       # Updated tests

cmd/                     # All CLI commands (to be updated)
├── init.go              # Updated to use timer
├── cluster/
│   ├── up.go            # Updated to use timer with stages
│   ├── down.go          # Updated to use timer
│   ├── status.go        # Updated to use timer
│   └── ...
└── workload/
    ├── reconcile.go     # Updated to use timer
    └── ...
```

**Structure Decision**: Single project structure (Option 1). This feature adds a new package `pkg/ui/timer` following the package-first design principle. The timer package will be integrated into existing CLI commands in `cmd/` directory, and the existing `cmd/ui/notify` package will be updated to format and display timing information.

## Phase 0: Outline & Research

**Status**: ✅ COMPLETE

### Research Completed

All technical unknowns from the Technical Context have been resolved and documented in [research.md](./research.md):

1. ✅ **Timer Design Pattern**: Stateful struct with Start(), NewStage(), Stop() methods
2. ✅ **Notify Integration**: Timer provides GetTiming() method; notify formats display
3. ✅ **Stage Tracking**: Explicit NewStage(title) calls mark transitions
4. ✅ **Timing Format**: Use Go's Duration.String() method directly
5. ✅ **Error Handling**: Timing tracked but only displayed on success
6. ✅ **Go time Package Best Practices**: Use time.Now(), store time.Time values
7. ✅ **CLI Integration Patterns**: Timer lifecycle matches command lifecycle
8. ✅ **Package-First Design**: Timer in pkg/ui/timer, independently testable

**Output**: ✅ research.md created and complete with no remaining NEEDS CLARIFICATION items.

## Phase 1: Design & Contracts

**Status**: ✅ COMPLETE

*Prerequisites: research.md complete* ✅

### Design Artifacts Created

1. ✅ **Data Model** ([data-model.md](./data-model.md)):
   - Timer entity with attributes and methods
   - TimingData struct for snapshot data
   - Type definitions and relationships
   - Data flow and testing considerations

2. ✅ **API Contracts** ([contracts/](./contracts/)):
   - Timer Interface Contract ([timer-interface.md](./contracts/timer-interface.md))
     - 7 contract requirements (CR-001 through CR-007)
     - Test scenarios for each requirement
     - Success criteria defined
   - Notify-Timer Integration Contract ([notify-integration.md](./contracts/notify-integration.md))
     - 5 integration requirements (IR-001 through IR-005)
     - Integration patterns and test scenarios
     - Clean architecture preservation

3. ✅ **Quickstart Guide** ([quickstart.md](./quickstart.md)):
   - 5 test scenarios from feature specification
   - Performance validation steps
   - Integration validation across all commands
   - Troubleshooting guide and validation checklist

4. ✅ **Agent Context Updated** (.github/copilot-instructions.md):
   - Added timing feature technical context
   - Updated with Go 1.24.0+ and dependencies
   - Preserved existing manual additions

### Contract Tests (To Be Implemented)

The following test files will be created during implementation (Phase 4):

- `pkg/ui/timer/timer_test.go` - Contract tests for Timer interface (CR-001 through CR-007)
- `pkg/ui/timer/integration_test.go` - Integration tests with notify package (IR-001 through IR-005)
- `cmd/ui/notify/notify_test.go` - Tests for FormatTiming() function

**Note**: Contract tests MUST FAIL initially (no implementation yet). This follows TDD principle from Constitution.

**Output**: ✅ All Phase 1 artifacts complete. Ready for Phase 2 (Task Planning).

## Phase 2: Task Planning Approach

> **Note**: This section describes what the /tasks command will do - DO NOT execute during /plan

**Task Generation Strategy**:

- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each entity → model creation task [P]
- Each user story → integration test task
- Implementation tasks to make tests pass

**Ordering Strategy**:

- TDD order: Tests before implementation
- Dependency order: Models before services before UI
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 25-30 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation

> **Note**: These phases are beyond the scope of the /plan command

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking

> **Note**: Fill ONLY if Constitution Check has violations that must be justified

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking

> **Note**: This checklist is updated during execution flow

**Current Status**: Phase 1 Complete - Ready for /tasks command

- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design & contracts complete (/plan command)
- [ ] Phase 2: Task planning (requires /tasks command)
- [ ] Phase 3+: Implementation (requires separate execution)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
