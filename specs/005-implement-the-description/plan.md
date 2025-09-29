
# Implementation Plan: KSail Cluster Provisioning Command

**Branch**: `005-implement-the-description` | **Date**: 2025-09-29 | **Spec**: [/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/spec.md](spec.md)
**Input**: Feature specification from `/specs/005-implement-the-description/spec.md`

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
Implement a production-ready `ksail cluster up` command that provisions Kind, K3d, or EKS clusters using existing provisioner packages, waits for readiness, merges kubeconfig, and emits actionable telemetry-informed feedback. The solution will reuse current lifecycle components, enforce dependency checks, and honor the configuration priority (CLI flags → environment → config files → defaults).

## Technical Context
**Language/Version**: Go 1.24+
**Primary Dependencies**: Cobra, Viper, sigs.k8s.io/kind, github.com/k3d-io/k3d/v5, github.com/weaveworks/eksctl, Kubernetes client-go, internal provisioner packages
**Storage**: Local filesystem for kubeconfig and project manifests
**Testing**: go test with testify, go-snaps for snapshots, mockery-generated mocks
**Target Platform**: Developer workstations (macOS/Linux) managing local Docker/Podman engines and AWS accounts
**Project Type**: single
**Performance Goals**: CLI response under 200ms for non-provisioning steps; readiness polling completes within default 5-minute timeout; inline telemetry for stage timings
**Constraints**: Must honor constitutional telemetry requirement, keep memory footprint <50 MB, operate within dependency prerequisites (Docker/Podman/AWS)
**Scale/Scope**: Single cluster per invocation; supports multi-node Kind/K3d setups and existing AWS EKS clusters

## Constitution Check
> GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.

- **Code Quality Excellence**: Plan must enforce gofmt, golangci-lint, reuse existing packages, and avoid introducing debt. ✔️ Strategy: leverage internal provisioners, require lint/test gates per plan before merge.
- **Testing Standards (TDD-First)**: All new behavior covered by pre-written unit/contract tests before implementation. ✔️ Plan includes dependency checks, readiness waiters, kubeconfig merge tests prior to code.
- **User Experience Consistency**: Maintain Cobra UX, consistent messaging, actionable errors. ✔️ Ensure outputs go through existing notify/provisioner channels and include remediation hints.
- **Performance Requirements**: Provisioning stages instrumented with lightweight telemetry (stage + total durations); readiness timeout respected. ✔️ Include telemetry plan and budget.

Initial Constitution Check: PASS

Post-Design Constitution Check: PASS (design maintains telemetry, TDD, and UX mandates)

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

**Structure Decision**: Option 1 (single project structure)

## Phase 0: Outline & Research

1. Extract unknowns and validation topics from the Technical Context:
   - Configuration precedence (flags → env → config files → defaults)
   - Provisioner reuse and dependency checks per distribution
   - Readiness verification and kubeconfig management
   - Telemetry obligations under Constitution IV

2. Capture findings in `research.md` following the decision/rationale/alternatives format.
   - Confirm Viper precedence satisfies the required configuration priority.
   - Document reuse of existing provisioners (kind, k3d, eks) and dependency helpers.
   - Validate readiness and kubeconfig merge strategy via client-go helpers.
   - Outline telemetry recorder usage for per-stage and total timings.

3. Ensure all NEEDS CLARIFICATION markers are resolved before moving forward.

Output: `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/research.md`

## Phase 1: Design & Contracts

Prerequisite: `research.md` complete

1. Map domain entities and configuration inputs to `data-model.md`:
   - Cluster specification (distribution, name, configuration priority sources)
   - Dependency result/state objects (engine readiness, AWS credentials)
   - Telemetry summary structures (stage durations)

2. Produce CLI contract definitions for success/error outputs in `/contracts/`:
   - Describe expected notify/provisioner messages, including timing summaries and remediation hints.
   - Capture error contracts for missing dependencies, timeout failures, and provisioning errors.

3. Create failing contract tests aligned with the above contracts.
   - Use Go test scaffolding in `cmd/cluster` to express expected behaviours before implementation.

4. Outline integration scenarios in `quickstart.md` mapping to acceptance criteria:
   - Local (Kind/K3d) provisioning happy path
   - EKS provisioning with credentials
   - Force recreation flow and dependency failure guidance

5. Detail telemetry instrumentation strategy:
   - Stage boundaries (dependency check, provisioning, readiness wait, kubeconfig merge)
   - Ensure outputs surface total vs per-stage durations consistent with Constitution IV.

6. Update agent guidance by running `.specify/scripts/bash/update-agent-context.sh copilot` and recording new technologies or patterns.

Outputs: `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/data-model.md`, `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/contracts/`, `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/quickstart.md`, updated agent context file.

## Phase 2: Task Planning Approach

This section describes what the /tasks command will do - DO NOT execute during /plan

**Task Generation Strategy**:

- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- Each contract → contract test task [P]
- Each entity → model creation task [P]
- Each user story → integration test task
- Add instrumentation tasks to capture runtime telemetry when constitution requires it
- Implementation tasks to make tests pass

**Ordering Strategy**:

- TDD order: Tests before implementation
- Dependency order: Models before services before UI
- Mark [P] for parallel execution (independent files)

**Estimated Output**: 25-30 numbered, ordered tasks in tasks.md

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation

These phases are beyond the scope of the /plan command

**Phase 3**: Task execution (/tasks command creates tasks.md)
**Phase 4**: Implementation (execute tasks.md following constitutional principles)
**Phase 5**: Validation (run tests, execute quickstart.md, review telemetry outputs against performance thresholds)

## Complexity Tracking

Fill ONLY if Constitution Check has violations that must be justified

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking

This checklist is updated during execution flow

**Phase Status**:

- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [ ] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:

- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [ ] Complexity deviations documented

---
*Based on Constitution v1.1.0 - See `/memory/constitution.md`*
