
# Implementation Plan: KSail Cluster Provisioning Command

**Branch**: `005-implement-the-description` | **Date**: 2025-09-28 | **Spec**: `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/spec.md`
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

Deliver a production-ready `ksail cluster up` command that provisions Kind, K3d, or EKS clusters using the existing KSail configuration pipeline. The command will reuse current config managers and provisioners, execute explicit dependency and readiness checks, and emit inline telemetry (slowest stage vs total runtime) through the notify/provisioner helpers—all implemented inside `cmd/cluster/up` without introducing a new orchestration package.

## Technical Context

**Language/Version**: Go 1.25.1 (constitution requires ≥1.24)
**Primary Dependencies**: Cobra CLI, Viper, `pkg/config-manager/{kind,k3d,eks}`, `pkg/provisioner/cluster/{kind,k3d,eks}`, `pkg/provisioner/containerengine`, `k8s.io/client-go`, AWS SDK v2 via eksctl wrappers, `internal/utils/path`
**Storage**: N/A (reads project YAML configs only)
**Testing**: `go test` (with testify mocks/helpers), go tool cover, golangci-lint; benchmarks replaced by inline telemetry validations
**Target Platform**: macOS/Linux developer workstations provisioning local (Kind/K3d) or AWS EKS clusters
**Project Type**: Single CLI project driven from `cmd/` and `pkg/`
**Performance Goals**: CLI response <200 ms for dependency/validation stages, readiness capped at configurable timeout (default 5 min), telemetry must capture per-stage and total durations
**Constraints**: Reuse existing provisioners, honour FR-007 force semantics, keep orchestration confined to `cmd/cluster/up`, fail fast on missing prerequisites, merge kubeconfig and switch context automatically, instrumentation must not spam or require extra output modes
**Scale/Scope**: One cluster per invocation across Kind/K3d/EKS; minimal footprint for telemetry to keep UX responsive

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **I. Code Quality Excellence** → Plan keeps orchestration in `cmd/cluster/up` using well-factored helper functions for configuration loading, dependency checks, telemetry, and readiness, enabling focused unit coverage while avoiding new packages.
- **II. Testing Standards (TDD-First)** → All new behaviour (dependency checks, readiness waits, telemetry summaries, CLI wiring) will start with failing tests before implementation and maintain >90 % coverage.
- **III. User Experience Consistency** → Command continues to use existing notify/provisioner helpers, supports the clarified `--force` semantics, and surfaces actionable remediation messages.
ios/ or android/
- **IV. Performance Requirements** → Inline timers capture local-stage and total durations, enforcing thresholds without external benchmarks and giving operators immediate visibility into slow steps.

## Project Structure

### Documentation (this feature)

```text
specs/005-implement-the-description/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)

```text
cmd/
├── cluster/
│   ├── cluster.go
│   └── up.go              # houses all orchestration helpers and telemetry logic
pkg/
├── config-manager/
├── provisioner/
│   └── cluster/{kind,k3d,eks}
├── provisioner/containerengine
├── io/
└── validator/

internal/
└── testutils/

.github/
└── workflows/ci.yaml      # enforces system-level coverage
```

**Structure Decision**: Option 1 (single CLI). All orchestration stays inside `cmd/cluster/up` behind file-scoped helpers (config loading, dependency checks, telemetry recording, provisioning, readiness, kubeconfig management) so no new packages are introduced.

## Phase 0: Outline & Research

- Reused provisioners/config managers confirmed to satisfy current interfaces (see `research.md`).
- Established dependency checks via `containerengine.AutoDetect` and AWS SDK credential resolution to honour FR-009.
- Selected client-go readiness polling (nodes Ready + namespace reachability) to meet FR-008/FR-011.
- Validated kubeconfig merging via `clientcmd` helpers and post-provision verification.
- Determined inline telemetry as primary performance signal: capture timestamps at dependency validation, provisioning start/stop, readiness wait, kubeconfig merge.

**Output**: `/specs/005-implement-the-description/research.md` (complete).

## Phase 1: Design & Contracts

Prerequisite: research.md complete.

### Orchestration Design

- Keep all orchestration logic inside `cmd/cluster/up`, introducing file-scoped helper functions (e.g., `loadConfig`, `checkDependencies`, `recordTelemetry`, `provisionCluster`, `waitForReadiness`, `finalizeKubeconfig`).
- Define small structs within the command file (such as `dependencyCheckResult`, `telemetrySummary`) to group related data without exporting new packages.
- CLI handler (`NewUpCommand`/`runClusterUp`) remains the entry point, wiring helpers together and delegating output to `notify` helpers.

### Control Flow

1. Load KSail project context (`ksail.yaml` + distribution-specific overlays) via config manager interfaces.
2. Resolve and validate dependencies: container engine availability for Kind/K3d, AWS credentials/profile for EKS.
3. Start telemetry scope (record `commandStart`).
4. Inspect provider state; if cluster exists and no `--force`, reuse; otherwise delete/create as needed.
5. Capture per-stage timestamps (`dependencyDuration`, `provisionDuration`, `readinessDuration`, `kubeconfigDuration`).
6. Merge kubeconfig and set active context using `clientcmd.ModifyConfig` or existing provisioner helpers.
7. Wait for readiness (poll nodes and namespace) respecting timeout.
8. On success, emit notify/provisioner output augmented with concise telemetry summary (`slowestStage`, `totalDuration`). On failure, emit actionable guidance and attempt cleanup when safe.

### Dependency Validation

- Kind/K3d: use `containerengine.AutoDetect` and check for running daemon socket. Provide remediation (`Start Docker/Podman`).
- EKS: load AWS config (`config.LoadDefaultConfig`), ensure credentials retrieved, and highlight missing profile/permission issues.

### Runtime Instrumentation

- Implement a lightweight telemetry helper struct inside `cmd/cluster/up` to capture labelled durations and compute slowest stage + total runtime.
- Ensure telemetry respects quiet/JSON modes by embedding summary within existing notify output (e.g., `success: ... • slowest stage: readiness=3m12s • total: 3m25s`).
- Provide hooks so tests can inject fake clocks (via function parameters or interfaces) for deterministic assertions.

### Readiness Waiter

- Build rest config using resolved kubeconfig context.
- Poll every 5 s: ensure all schedulable nodes Ready and default namespace accessible; fail with timeout after configured limit.
- Differentiate between transient API errors and fatal conditions for actionable messaging.

### Kubeconfig Management

- Confirm expected context exists after provisioning; create/update entries using `clientcmd.ModifyConfig` when necessary.
- Guard writes with file locking or safe write helper from `pkg/io`.

- `cmd/cluster/up_test.go`: covers flag wiring, dependency failure messaging, readiness timeout exit codes, telemetry summary content, force semantics, and kubeconfig switching (using helper functions where necessary).
- `cmd/cluster/up_internal_test.go`: exercises file-scoped helpers for dependency checks, readiness waits (with fake clientsets), telemetry calculations, and kubeconfig persistence using temp directories.

### Agent Context Update

- `.specify/scripts/bash/update-agent-context.sh copilot` executed on 2025-09-28 to sync guidance with the updated plan; rerun if significant dependencies change later.

**Output**: `/specs/005-implement-the-description/data-model.md`, `/specs/005-implement-the-description/contracts/`, `/specs/005-implement-the-description/quickstart.md` (all complete), plus agent context update (pending run).

## Phase 2: Task Planning Approach

This section describes what the /tasks command will do - DO NOT execute during /plan

- Load `.specify/templates/tasks-template.md` as base.
- Seed setup + shared fixtures tasks (command helper scaffolding, testdata for kubeconfig/AWS stubs).
- Create failing tests for dependency checks, readiness waiter, kubeconfig manager, telemetry helper, and CLI wiring—all within the `cmd/cluster` package.
- Add implementation tasks for helper functions, dependency logic, readiness waiter, kubeconfig updates, telemetry helper, and CLI integration inside `cmd/cluster/up.go`.
- Include integration tasks for wiring helper seams (dependency injection via function parameters) and ensuring existing provisioners are invoked correctly.
- Generate polish tasks for docs, gofmt/go test/go tool cover, golangci-lint, telemetry review, and manual quickstart validation capturing emitted telemetry summary.

**Ordering Strategy**:

- Strict TDD order: author failing tests before implementation for each component.
- Implement helpers in dependency order (config loader → dependency checker → readiness → kubeconfig → telemetry → runner → CLI).
- Mark independent test cases ([P]) when they target separate files.
- Reserve polish tasks until after integration, ensuring telemetry review occurs alongside manual quickstart run.

**Estimated Output**: ~21 tasks (already captured in `tasks.md`) covering setup, tests, implementation, integration, telemetry, docs, and validation.

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan.

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
*Based on Constitution v1.1.0 - See `/memory/constitution.md`*
