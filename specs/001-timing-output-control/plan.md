# Implementation Plan: Timing Output Control

**Branch**: `001-timing-output-control` | **Date**: 2025-12-12 | **Spec**: `/specs/001-timing-output-control/spec.md`
**Input**: Feature specification from `/specs/001-timing-output-control/spec.md`

## Summary

Add a root-level `--timing` flag (default off). When enabled, print a spec-defined timing block after each timed activity completion message (`✔ ...`). Reuse the existing `pkg/ui/timer.Timer` semantics (`stage` → `current`, `total` → `total`) and centralize formatting in `pkg/ui/notify` to avoid drift across commands.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**: Cobra, Viper, samber/do (DI), fatih/color, go-snaps, testify
**Storage**: N/A
**Testing**: `go test ./...` (table-driven + snapshot tests); public APIs only (no white-box)
**Target Platform**: macOS + Linux CLI
**Project Type**: Go CLI (Cobra)
**Performance Goals**: No meaningful overhead when `--timing` is off
**Constraints**:

- Timing output must be OFF by default
- Flag-only control (no `ksail.yaml`)
- Output must match the spec format consistently across commands
- Avoid global mutable state; prefer DI or explicit plumbing
  **Scale/Scope**: Impacts most commands that emit `✔` completion messages (especially multi-stage workflows)

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

**Status**: PASS (no justified violations)

- KISS: solution stays simple; complexity justified if needed
- DRY: no duplicate logic across packages
- YAGNI: no speculative features or abstractions
- Interface-first: define/confirm interfaces before implementations; mocks are possible
- Test-first: tests planned first and cover public APIs only (no white-box)
- Package-first: feature work lives in `pkg/`; `cmd/` stays thin
- Quality gates: plan includes `mockery`, `go test ./...`, `golangci-lint run`, `go build ./...`

## Project Structure

### Documentation (this feature)

```text
specs/001-timing-output-control/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

This feature directory now contains:

```text
specs/001-timing-output-control/
├── plan.md
├── spec.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── cli.md
└── checklists/
  └── requirements.md
```

### Source Code (repository root)

```text
cmd/                 # CLI commands (thin wrappers)
pkg/                 # Public packages (business logic)
docs/                # Documentation site content
schemas/             # JSON schemas
specs/               # Feature specs & plans
k8s/                 # Example manifests (if applicable)

# Testing
# Go tests live alongside code as *_test.go in the same packages.
# Snapshot tests and fixtures live under the relevant package directories.
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

**Structure Decision**: Implement behavior in `cmd/` (flag definition + wiring) and in `pkg/` (shared formatting / helpers). Keep CLI command files thin by introducing a small helper in `pkg/cmd` or `pkg/ui/notify` for “completion message + optional timing block”.

## Phase 0 — Research (completed)

See `/specs/001-timing-output-control/research.md`.

Key findings:

- Existing timing output is already implemented as bracket suffixes (e.g. `[stage: 1ms]`) via `pkg/ui/notify.FormatTiming` and ad-hoc string concatenation.
- `pkg/ui/timer.Timer` already provides the correct data shape for the spec (`stage` and `total`).

## Phase 1 — Design

### User-visible output

When `--timing` is enabled, after each `✔ completion message` print:

```text
✔ completion message
⏲ current: <duration>
  total:  <duration>
```

Notes:

- “Timed activity” means each completion message emitted for a progress/spinner step.
- Use Go duration formatting for `<duration>` unless the spec is later tightened.

### Plumbing / control flow

Proposed design (minimal churn):

1. Add a root-level persistent flag `--timing` in `cmd.NewRootCmd(...)`.
2. Add a small helper (exported, testable via public API) for reading the flag:

- `pkg/cmd.GetTimingEnabled(cmd *cobra.Command) (bool, error)`

3. Consolidate completion printing:

- Prefer a single helper for emitting the completion line plus optional timing block.
- Remove direct uses of `notify.FormatTiming(...)` concatenated into the success line.

4. Use `timer.Timer` consistently:

- Call `Timer.Start()` at invocation start (existing)
- After each activity completes: read `(total, stage)` and emit the block
- Call `Timer.NewStage()` immediately after emitting the completion+timing for the next activity

## Phase 2 — Implementation Plan (high level)

1. Add `--timing` as a persistent flag on the root command.
2. Update `pkg/ui/notify`:

- Keep existing `FormatTiming(...)` for internal use if needed, but introduce a new exported formatter for the spec block (e.g. `FormatTimingBlock(current, total time.Duration) string`).
- Ensure multi-line formatting does not break existing indentation behavior.

3. Refactor key call sites to use the centralized completion+timing behavior:

- `pkg/cmd/lifecycle_helpers.go` (currently concatenates `FormatTiming` into the success content)
- Cluster and workload commands that currently emit `stage`/`total` suffixes directly

4. Tests (public API only):

- Update existing snapshot tests to remove default timing output.
- Add/extend snapshot coverage for `--timing` runs to verify the exact 3-line timing block format.
- Add unit tests for exported notify formatting helpers if needed.

5. Documentation:

- Update CLI docs/README to mention `--timing` and show the output format.

6. Validation gates:

- `mockery`
- `go test ./...`
- `golangci-lint run --timeout 5m --fix`
- `go build ./...`

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation                  | Why Needed         | Simpler Alternative Rejected Because |
| -------------------------- | ------------------ | ------------------------------------ |
| [e.g., 4th project]        | [current need]     | [why 3 projects insufficient]        |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient]  |
