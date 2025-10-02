
# Implementation Plan: KSail Init Command

**Branch**: `002-add-the-ksail` | **Date**: 2025-09-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-add-the-ksail/spec.md`

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

Primary requirement: Implement `ksail init` command that scaffolds new Kubernetes projects with intuitive CLI UX for easy onboarding. The command creates a ksail.yaml config file with Kind as default distribution, generates distribution-specific configuration files, creates basic Kustomize structure, and provides encouraging feedback during 5-second initialization. Must work completely offline, detect conflicts and require --force flag, and offer customization through CLI flags.

## Technical Context

**Language/Version**: Go 1.24+ (from go.mod and constitution requirements)

**Primary Dependencies**:

- `github.com/spf13/cobra` for CLI framework (existing pattern)
- `github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1` for cluster APIs
- `github.com/devantler-tech/ksail-go/pkg/scaffolder` for file generation
- `github.com/devantler-tech/ksail-go/pkg/config-manager/ksail` for CLI input handling

**Storage**: Runtime template generation via pkg/io/generator system (no embedded files needed), output to local filesystem

**Testing**: go test with testify framework, contract tests for CLI interface

**Target Platform**: Linux/macOS/Windows CLI tool (cross-platform Go binary)

**Project Type**: Single project (CLI tool extension to existing codebase)

**Performance Goals**: <5 second initialization, <50MB memory usage, <200ms CLI response time

**Constraints**: Completely offline operation, backward compatibility with existing KSail patterns, must follow constitutional TDD and code quality requirements

**Scale/Scope**: Single command implementation, ~5-10 template files, integration with existing scaffolder package

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Code Quality Excellence**: ✅ PASS

- CLI implementation follows existing cobra patterns in codebase
- Error handling provides actionable user guidance (FR-010)
- golangci-lint compliance mandatory per existing workflow
- Comprehensive godoc documentation required for new public APIs

**Testing Standards (TDD-First)**: ✅ PASS

- TDD approach: Contract tests for CLI interface written first
- Unit tests for scaffolder integration with >90% coverage requirement
- System tests for complete init workflow validation
- End-to-end CLI testing for user scenarios

**User Experience Consistency**: ✅ PASS

- Consistent cobra CLI patterns with existing commands
- Unified flag naming conventions (--force, --distribution)
- Clear error messages with actionable suggestions (FR-007, FR-010)
- Comprehensive help text explaining all options (FR-012)

**Performance Requirements**: ✅ PASS

- <5 second initialization meets <200ms CLI response requirement
- <50MB memory usage during operation (within <50MB normal limit)
- File I/O operations optimized for runtime template generation
- Performance benchmarking for critical path operations

**Post-Design Constitution Re-Check**: ✅ PASS

- Design artifacts maintain all constitutional compliance
- TDD contracts established for all components
- User experience patterns consistent with existing commands
- Performance requirements achievable with chosen architecture
- No new constitutional violations introduced

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

**Structure Decision**: [DEFAULT to Option 1 unless Technical Context indicates web/mobile app]

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

> [IMPORTANT]
> *Prerequisites: research.md complete*

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
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh copilot`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: ✅ data-model.md, ✅ contracts/cli-interface.md, ✅ contracts/scaffolder-service.md, ✅ quickstart.md, ✅ agent file updated

## Phase 2: Task Planning Approach

> [IMPORTANT]
> *This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:

- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- CLI contract → cobra command implementation and tests [P]
- Scaffolder contract → service implementation and tests [P]
- Data model entities → struct definitions and validation [P]
- Quickstart scenarios → test tasks
- Template creation → embed template files [P]

**Ordering Strategy**:

- TDD order: Contract tests → Unit tests → Implementation
- Dependency order: Data models → Services → CLI commands
- Mark [P] for parallel execution (independent files/modules)
- Sequential for dependent implementations

**Implementation Approach**: Enhancement of existing functionality (not new development)

- **Existing**: `cmd/init.go` already implements basic init command with scaffolder integration
- **Existing**: `pkg/scaffolder` provides complete Kind/K3d/EKS project generation
- **Existing**: Runtime template system via `pkg/io/generator` eliminates embedded files
- **Enhancement Focus**: Add missing UX features (progress spinner, enhanced error messages, additional CLI flags)

**Integration Points**:

- Enhance existing `cmd/init.go` with progress feedback and additional flags
- Extend existing `pkg/scaffolder` with conflict detection and validation
- Use existing `pkg/config-manager/ksail` for flag handling and configuration loading
- Leverage existing `pkg/apis/cluster/v1alpha1` APIs and validation

**Estimated Output**: 30 numbered, ordered tasks in tasks.md (enhancement-focused approach building on 60% existing functionality)

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation

> [IMPORTANT]
> *These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking

> [IMPORTANT]
> *Fill ONLY if Constitution Check has violations that must be justified*

| Violation                  | Why Needed         | Simpler Alternative Rejected Because |
|----------------------------|--------------------|--------------------------------------|
| [e.g., 4th project]        | [current need]     | [why 3 projects insufficient]        |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient]  |

## Progress Tracking

> [IMPORTANT]
> *This checklist is updated during execution flow*

**Phase Status**:

- [x] Phase 0: Research complete (/plan command) - ✅ research.md generated
- [x] Phase 1: Design complete (/plan command) - ✅ data-model.md, contracts/, quickstart.md generated
- [x] Phase 2: Task planning complete (/plan command - describe approach only) - ✅ Strategy documented
- [x] Phase 3: Tasks generated (/tasks command) - ✅ Tasks generated
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:

- [x] Initial Constitution Check: PASS - ✅ All constitutional requirements validated
- [x] Post-Design Constitution Check: PASS - ✅ Design maintains compliance
- [x] All NEEDS CLARIFICATION resolved - ✅ No unknowns in technical context
- [x] Complexity deviations documented - ✅ No violations, complexity justified

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
