# Implementation Plan: KSail Cluster Provisioning Command

**Branch**: `005-implement-the-description` | **Date**: 2025-09-28 | **Spec**: `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/spec.md`
**Input**: Feature specification from `/specs/005-implement-the-description/spec.md`

## Execution Flow (/plan command scope)

```text
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

Deliver a production-ready `ksail cluster up` command that provisions Kind, K3d, or EKS clusters from an existing KSail project configuration. The command will reuse established config managers and cluster provisioner implementations, add explicit dependency and readiness verification, and guarantee kubeconfig activation before returning success. Research confirms we can rely on container engine detection utilities, eksctl-backed provisioners, and client-go readiness polling to satisfy the clarified requirements.

## Technical Context

**Language/Version**: Go 1.24+ (current toolchain go1.25.1)
**Primary Dependencies**: Cobra CLI, Viper, KSail config managers, `pkg/provisioner/cluster` (Kind/K3d/EKS), `pkg/provisioner/containerengine`, `k8s.io/client-go`, AWS SDK v2 config loader (transitively via eksctl)
**Storage**: N/A — reads YAML configuration only
**Testing**: `go test` with `testify` mocks/helpers; maintain TDD discipline
**Target Platform**: Developer workstations (macOS/Linux) provisioning local Kind/K3d or AWS EKS clusters
**Project Type**: Single CLI project (Option 1 structure)
**Performance Goals**: Command should complete within configured timeout (default 5 min) and keep dependency checks <200 ms per constitution
**Constraints**: Enforce readiness timeout, ensure dependency failures surface actionable errors, switch kubeconfig without manual edits, keep CLI output consistent with notify helpers
**Scale/Scope**: One cluster per invocation across Kind/K3d/EKS distributions with optional `--force` recreation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Code Quality Excellence** → Plan confines orchestration to an injectable runner, enabling high test coverage and lint compliance.
- **Testing Standards (TDD-First)** → All new functionality will start with failing tests (command wiring, runner logic, dependency failure, readiness timeout, kubeconfig switching) before implementation.
- **User Experience Consistency** → CLI output will continue using `notify` helpers; flag naming aligns with existing conventions; error messages include remediation guidance.
- **Performance Requirements** → Readiness waiter uses context deadlines with 5 s polling; dependency checks short-circuit quickly; no unnecessary blocking operations introduced.

## Project Structure

### Documentation (this feature)

```text
specs/005-implement-the-description/
├── plan.md              # This file (/plan output)
├── research.md          # Phase 0 output (/plan)
├── data-model.md        # Phase 1 output (/plan)
├── quickstart.md        # Phase 1 output (/plan)
├── contracts/           # Phase 1 output (/plan)
└── tasks.md             # Phase 2 output (/tasks)
```

### Source Code (repository root)

```text
# Option 1: Single project (DEFAULT)
cmd/
├── cluster/
│   └── up.go
pkg/
├── clusterup/           # new orchestration package
├── provisioner/
├── config-manager/
└── ...

internal/
└── testutils/

```

**Structure Decision**: Option 1 (single CLI project). New orchestration helpers live in `pkg/clusterup` with minimal updates to existing command files.

## Phase 0: Outline & Research

- Reused provisioner/config manager approach validated; no new provider abstractions required.
- Resolved dependency handling by leveraging existing container engine detection and AWS SDK profile loading.
- Selected client-go readiness polling strategy to meet FR-008/FR-011.
- Outcomes and alternatives captured in `research.md` (complete).

## Phase 1: Design & Contracts

Prerequisite: `research.md` complete.

### Orchestration Design

- Introduce `pkg/clusterup.Runner` composed of configuration loader, distribution config factory, dependency validator, provisioner resolver, readiness waiter, and kubeconfig manager.
- Runner invoked by `cmd/cluster/up` with CLI options (timeout override, force recreation).

### Control Flow

1. Load KSail project context (`ksail.yaml` + distribution config).
2. Run dependency checks (container engine for Kind/K3d, AWS credentials for EKS).
3. Acquire provisioner; inspect existing cluster state via `Exists`/`Status`.
4. Branch on `--force` + existence: reuse (`Start`) or delete/create accordingly.
5. Merge kubeconfig and activate context using `clientcmd.ModifyConfig`.
6. Track timing checkpoints for each major stage (dependency checks, provisioning, readiness wait, kubeconfig merge) while executing them.
7. Wait for readiness: poll nodes Ready + API healthy within timeout.
8. Emit success through existing provisioner logging/notify helpers, appending the recorded total duration and slowest stage, and surface remediation-focused errors when something fails.

### Dependency Validation

- Kind/K3d: `containerengine.AutoDetect` ensures Docker/Podman reachable.
- EKS: AWS SDK `config.LoadDefaultConfig` (respecting profile) and `cfg.Credentials.Retrieve` to guarantee access before provisioning.

### Readiness Waiter

- Build rest config from kubeconfig path/context.
- Poll every 5 seconds using client-go until nodes Ready and namespace accessible.
- Respect context deadline; timeout yields exit code 4 per contract.

### Kubeconfig Management

- Use `clientcmd.Load` to inspect contexts, add/update context entries, and persist via `clientcmd.ModifyConfig`.
- Maintain safe write semantics via existing IO helpers.

### Testing Strategy (TDD)

- `cmd/cluster/up_test.go`: flag wiring, runner invocation, force path, timeout parsing, error propagation, confirmation that success output leverages existing notify/provisioner helpers without echoing flag values, and assertions around the timing summary content (local vs total durations).
- `pkg/clusterup/runner_test.go`: mocks for config loaders, provisioners, dependency checker, readiness waiter, kubeconfig updater, and verification that runner delegates success messaging to the notify/provisioner helpers while emitting the timing summary.
- `pkg/clusterup/dependencies_test.go`: container engine/AWS credential scenarios produce actionable errors.
- `pkg/clusterup/readiness_test.go`: fake clientset covers ready/timeout cases.
- `pkg/clusterup/kubeconfig_test.go`: ensure context switching persists correctly (temp filesystem).

System-level scenarios are validated by CI workflows in `.github/workflows/ci.yaml`, so this feature focuses on unit and contract coverage in-repo, with local execution enforcing >90% coverage via `go test` against a generated profile.

### Performance Instrumentation

- Introduce lightweight timers inside the runner to capture per-stage durations and total command runtime.
- Surface a concise summary (e.g., `slowest stage: readiness=3m12s • total=3m25s`) via existing notify helpers on success to help developers spot bottlenecks without stand-alone benchmarks.
- Ensure instrumentation paths remain opt-out capable for scripting (respect quiet/JSON modes) while still recording metrics internally for potential future aggregation.

### Artifacts Generated

- `data-model.md`, `contracts/cluster-up.md`, and `quickstart.md` complete and lint clean.
- Agent context update pending after finalizing plan.

### Agent Context Update

- Run `.specify/scripts/bash/update-agent-context.sh copilot` post-plan to keep assistant guidance current.

## Phase 2: Task Planning Approach

This section describes what the /tasks command will do. Do not execute it during `/plan`.

### Task Generation Strategy

- Seed tasks from contracts, data model, and design notes (tests first, implementation second).
- Create dedicated tasks for new packages (`pkg/clusterup`) covering dependency checker, runner, readiness waiter, and kubeconfig manager.
- Include CLI wiring, documentation updates, and validation commands (build, test, lint).

### Ordering Strategy

- Begin with unit tests for CLI and runner, then dependency/readiness helper tests.
- Implement helper packages in dependency order to satisfy the failing tests.
- Mark independent suites (readiness vs dependency checker) with `[P]` for potential parallel execution.

**Estimated Output**: Approximately 20–25 ordered tasks captured in `tasks.md`.

## Phase 3+: Future Implementation

Beyond the scope of the `/plan` command:

- **Phase 3**: Generate `tasks.md` via /tasks command.
- **Phase 4**: Execute implementation tasks adhering to constitution (TDD, lint, docs).
- **Phase 5**: Run full validation (tests, quickstart walkthrough, readiness smoke).

## Complexity Tracking

Fill this section only if the constitution check has violations that must be justified.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None      | —          | —                                   |

## Progress Tracking

This checklist is updated during execution flow.

### Phase Status

- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [ ] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

### Gate Status

- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented (none required)

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
