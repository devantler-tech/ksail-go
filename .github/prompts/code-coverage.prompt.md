---
description: Execute comprehensive code coverage analysis and test improvements based on implementation tasks, focusing on test quality, organization, and coverage metrics without altering source code.
---

Always evaluate the user input before acting:

User input:

$ARGUMENTS

## Fast-track coverage workflow

1. **Prep & baseline**
   - Locate the repository root (contains `go.mod`) and work with absolute paths.
   - Read `CONTRIBUTING.md`, `README.md`, and `.golangci.yml` to internalize testing strategy, linting, and quality bars.
   - Scan `report/`, `docs/`, or `notes/` (when present) for prior testing goals or metrics, and collect existing artifacts such as `coverage.out` or other reports.

2. **Spot high-value coverage wins**
   - Summarize coverage targets, supported test types, and architecture constraints from the prep step.
   - Record helper scripts, make targets, or workflows that relate to testing, noting their absolute paths.
   - Convert findings into actionable task buckets: organization, coverage, quality, and validation.

3. **Implement well-structured tests**
   - Keep one `_test.go` per source file; merge or rename duplicates when needed.
   - Add a focused test per constructor/function/method using `t.Run()` for scenarios; cap each test function at ~60 lines and delegate shared logic to helpers marked with `t.Helper()`.
   - Improve coverage through meaningful scenarios only—never alter production code.
   - Skip all `testutils` packages during analysis or test creation; those utilities are intentionally uncovered.

4. **Validate outcomes**
   - Run the relevant test, lint, and formatting commands to ensure compliance and reliability. Regression is not acceptable.
   - Ensure added tests are are actually adding code coverage.
   - Confirm testutils packages remain excluded from reports and that all changes respect Go testing best practices.

## Non-negotiable conventions

- Use only Go's standard `testing` package; avoid third-party testing libraries.
- Use Go snapshots for expected outputs; avoid hardcoded values in tests.
- One `<src-file-name>_test.go` per source file, no exceptions.
- Table-driven tests with `t.Run()` for subtests; max ~60 lines per test function.
- Shared test logic in helpers marked with `t.Helper()`.
- Parallelize tests with `t.Parallel()` where applicable.

## Non-negotiable guardrails

- Modify only test code or test utilities; production source must stay untouched and all existing tests must continue to pass.
- Preserve coding standards, readability, and maintainability while aiming for meaningful coverage gains, not just higher percentages.

## Coverage analysis exclusions

- `**/testutils/**`
- `**/*mocks*/**`

Test utility packages supply shared helpers for other tests and are intentionally left out of coverage requirements.
