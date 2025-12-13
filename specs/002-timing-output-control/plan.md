# Implementation Plan: Timing Output Control

**Branch**: `002-timing-output-control` | **Date**: 2025-12-13 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-timing-output-control/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.github/prompts/speckit.plan.prompt.md` for the execution workflow.

## Summary

Add a global/root persistent `--timing` flag (default false) that makes the CLI emit per-activity timing output. When enabled, each existing `✔` success/completion message is followed immediately by a timing block using the `⏲` glyph with `current` and `total` durations formatted as Go `time.Duration` strings. No configuration file support is added.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**: Cobra CLI (`github.com/spf13/cobra`), DI runtime (`github.com/samber/do/v2`), output styling (`github.com/fatih/color`)
**Storage**: N/A
**Testing**: `go test ./...`; snapshot testing via `github.com/gkampitakis/go-snaps` where appropriate
**Target Platform**: Cross-platform CLI (macOS/Linux/Windows)
**Project Type**: Go CLI application
**Performance Goals**: No behavior/format changes when timing is off; minimal overhead when timing is on
**Constraints**: Keep `cmd/` thin; implement behavior in `pkg/`; tests must validate public APIs only
**Scale/Scope**: Small surface change (root flag + notify formatting), but impacts user-visible output across commands

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

- KISS: solution stays simple; complexity justified if needed
- DRY: no duplicate logic across packages
- YAGNI: no speculative features or abstractions
- Interface-first: define/confirm interfaces before implementations; mocks are possible
- Test-first: tests planned first and cover public APIs only (no white-box)
- Package-first: feature work lives in `pkg/`; `cmd/` stays thin
- Quality gates: plan includes `mockery`, `go test ./...`, `golangci-lint run`, `go build ./...`

**Post-design re-check**: PASS (no justified violations)

## Project Structure

### Documentation (this feature)

```text
specs/002-timing-output-control/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
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

**Structure Decision**: Root persistent flag + shared renderer in `pkg/ui/notify` with existing `pkg/ui/timer` as the timing source.

Implement the feature as an opt-in behavior driven by a root persistent CLI flag and applied at message rendering time:

- `cmd/`: add a persistent `--timing` flag on the root command and expose it to command handlers in the established DI/executor flow.
- `pkg/ui/notify`: change how timing is rendered so a timer produces a multi-line timing block (using `⏲ current:` / `total:`) that appears immediately after the `✔` success message.
- `pkg/ui/timer`: continue using the existing `Timer` interface and `GetTiming()` values (`stage` as current, `total` accumulated).

Testing remains black-box:

- Public output formatting: update/extend `pkg/ui/notify` tests to assert the rendered strings (including newlines/indentation) using only exported identifiers.
- CLI behavior: add/adjust Cobra/command-level tests or snapshots to assert that `--timing` turns the blocks on and default runs keep output unchanged.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| _None_    | _N/A_      | _N/A_                                |
