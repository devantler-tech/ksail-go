
# Implementation Plan: Configuration File Validation

**Branch**: `001-add-validation-for` | **Date**: 2025-09-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-add-validation-for/spec.md`

## Execution Flow (/plan command scope)

```txt
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

Configuration file validation system that validates all ksail configuration files (ksail.yaml, kind.yaml, k3d.yaml, eks.yaml) whenever they are loaded. The system prioritizes marshalling errors for efficient in-memory validation, provides actionable error messages with specific field information and fix examples, and fails fast to prevent destructive operations with invalid configurations. Uses separate validators for each configuration type with independent validation logic.

## Technical Context

**Language/Version**: Go 1.24.0+ (as per go.mod and constitution requirements)
**Primary Dependencies**: sigs.k8s.io/yaml, github.com/k3d-io/k3d/v5/pkg/config/v1alpha5, sigs.k8s.io/kind/pkg/apis/config/v1alpha4, github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1
**Storage**: File-based configuration files (ksail.yaml, kind.yaml, k3d.yaml, eks.yaml)
**Testing**: Go test with testify, go-snaps for snapshot testing, mockery for mocks
**Target Platform**: Linux (amd64/arm64), macOS (amd64/arm64) - cross-platform CLI
**Project Type**: Single project with pkg/ structure for validators
**Performance Goals**: Configuration validation <100ms for files <10KB, memory usage <10MB
**Constraints**: In-memory validation only, no file I/O during validation, fail-fast on errors
**Scale/Scope**: Individual configuration files up to 10KB, 3 separate validator packages with independent logic

**User-provided Implementation Details**: Add pkg/validator/ksail/config-validator.go, pkg/validator/kind/config-validator.go, and pkg/validator/k3d/config-validator.go to validate configurations. Each validator should be independent - ksail validator handles ksail.yaml and coordinates loading of other configs, while kind/k3d validators only validate their specific configuration formats.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**I. Code Quality First**: ✅ PASS

- New validator packages will follow Go best practices in pkg/validator/ structure
- Comprehensive godoc comments required for all public functions
- golangci-lint compliance mandatory
- Dependencies controlled via existing depguard rules

**II. Test-Driven Development (NON-NEGOTIABLE)**: ✅ PASS

- TDD approach: Write failing tests first for each validator
- *_test.go files required for each validator package
- Unit tests for in-memory validation logic (no file I/O)
- Integration tests for complete validation workflows
- Snapshot testing for error message consistency

**III. User Experience Consistency**: ✅ PASS

- Standardized error messages using existing notify package
- Consistent error format across all validators
- Actionable error messages with specific field paths and fix examples
- Human-readable output for basic Kubernetes users

**IV. Performance Excellence**: ✅ PASS

- Validation completes <100ms (within constitution's <10s status check requirement)
- Memory usage <10MB (within reasonable limits)
- In-memory validation eliminates inefficient file I/O
- No impact on build times (<90s) or test execution (<60s)

**Quality Standards Compliance**: ✅ PASS

- Go 1.24.0+ version requirement met
- Unit testable without file system operations
- Documentation for all public interfaces
- Security: No external network dependencies for validation

## Project Structure

### Documentation (this feature)

```txt
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)

```txt
# Go project structure (existing ksail-go structure)
pkg/
├── validator/                      # NEW: Validation packages
│   ├── ksail/
│   │   ├── config-validator.go    # KSail config validation
│   │   └── config-validator_test.go
│   ├── kind/
│   │   ├── config-validator.go    # Kind config validation
│   │   └── config-validator_test.go
│   └── k3d/
│       ├── config-validator.go    # K3d config validation
│       └── config-validator_test.go
├── apis/                          # Existing: API definitions
├── config-manager/                # Existing: Config management
├── provisioner/                   # Existing: Cluster provisioning
└── ...                           # Other existing packages

cmd/                               # Existing: CLI commands (integration points)
internal/                          # Existing: Internal utilities
```

**Structure Decision**: Go project with new pkg/validator/ packages following existing conventions

## Phase 0: Outline & Research

1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:

   ```txt
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts

> [!IMPORTANT]
> Prerequisites: research.md complete

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh copilot`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach

> [!NOTE]
> This section describes what the /tasks command will do - DO NOT execute during /plan

**Task Generation Strategy**:

- Load `.specify/templates/tasks-template.md` as base framework
- Generate tasks from Phase 1 design documents (contracts/, data-model.md, quickstart.md)
- Create TDD workflow: failing tests first, then implementation to pass tests
- Structure tasks by validator package (ksail, kind, k3d) for parallel development

**Task Categories and Ordering**:

1. **Setup Tasks**: Create package structure and basic types from data-model.md
2. **Contract Test Tasks**: Generate failing tests from contracts/ specifications
3. **Unit Test Tasks**: Create comprehensive unit tests for each validator [P]
4. **Implementation Tasks**: Implement validators to pass tests
5. **Integration Test Tasks**: Test complete validation workflows
6. **Documentation Tasks**: Update godoc and integration guides

**Parallelization Strategy**:

- Mark [P] for tasks within different validator packages (ksail, kind, k3d)
- Sequential dependencies: Setup → Tests → Implementation → Integration
- Each validator package can be developed independently after contracts are defined
- Integration tests require all validators to be complete

**Estimated Output**:

- 8-10 setup and structure tasks
- 15-20 test creation tasks (5-7 per validator)
- 15-20 implementation tasks to make tests pass
- 5-8 integration and documentation tasks
- **Total**: 45-60 numbered, ordered tasks with clear dependencies

**Integration Points**:

- Tasks will integrate with existing pkg/config-manager for validation triggers
- CLI command integration points defined for each validator
- Error message formatting aligned with existing cmd/ui/notify package

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation

> [!NOTE]
> These phases are beyond the scope of the /plan command

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking

> [!WARNING]
> Fill ONLY if Constitution Check has violations that must be justified

| Violation                  | Why Needed         | Simpler Alternative Rejected Because |
|----------------------------|--------------------|--------------------------------------|
| [e.g., 4th project]        | [current need]     | [why 3 projects insufficient]        |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient]  |

## Progress Tracking

> [!NOTE]
> This checklist is updated during execution flow

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
- [x] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
