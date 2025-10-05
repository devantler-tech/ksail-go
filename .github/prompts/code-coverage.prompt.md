---
description: Execute comprehensive code coverage analysis and test improvements based on implementation tasks, focusing on test quality, organization, and coverage metrics without altering source code.
---

The user input can be provided directly by the agent or as a command argument - you **MUST** consider it before proceeding with the prompt (if not empty).

User input:

$ARGUMENTS

1. Collect baseline testing context:
   - Locate the repository root (directory containing `go.mod`) and use absolute paths for subsequent steps.
   - **REQUIRED**: Read `CONTRIBUTING.md`, `README.md`, and `.golangci.yml` to understand test strategy, coding standards, and lint enforcement.
   - **IF AVAILABLE**: Review files under `report/`, `docs/`, or `notes/` (for example prior coverage summaries or architecture notes) that describe testing goals and constraints.

2. Load and analyze the test coverage context:
   - Summarize coverage targets, supported test types, and architectural constraints captured in step 1.
   - Identify helper scripts, make targets, or reusable workflows related to testing (record their absolute paths).
   - Capture existing coverage reports such as `coverage.out` or artifacts in `report/` to establish current baselines.

3. Consolidate test-focused tasks:
   - **Test organization tasks**: Test file consolidation, structure improvements
   - **Coverage improvement tasks**: Identify low-coverage areas, add comprehensive tests
   - **Test quality tasks**: Code style, helper functions, maintainability improvements
   - **Test validation tasks**: Linting, formatting, best practices compliance

4. Execute test coverage improvements following task-based approach:
   - **Test file organization**: One _test.go file per source file, merge duplicates
   - **Test function structure**: One test per constructor/function/method with sub-tests via t.Run()
   - **Test function size limits**: Maximum 60 lines per test function (funlen compliance)
   - **Helper function extraction**: Break down large tests into reusable helper functions
   - **Coverage analysis**: Measure and improve coverage without altering source code
   - **EXCLUDE testutils**: Skip coverage analysis for all testutils packages - these are test utilities, not production code

5. Test improvement execution rules:
   - **Organization first**: Consolidate test files, eliminate duplication
   - **Coverage analysis**: Identify gaps and create comprehensive test scenarios
   - **Quality improvements**: Apply linting fixes, add helper functions, improve readability
   - **Validation**: Ensure all tests pass, coverage targets met, linting compliance

6. Test-specific guidelines and constraints:
   - **No source code changes**: Only modify test files and test utilities
   - **Maintain functionality**: All existing tests must continue to pass
   - **Coverage targets**: Aim for high coverage without sacrificing test quality
   - **Code style compliance**: Follow Go testing best practices and linting rules
   - **Helper function patterns**: Use t.Helper(), avoid code duplication, maintain readability
   - **CRITICAL: Ignore testutils packages**: testutils packages MUST BE IGNORED for coverage analysis - do not create tests for testutils directories as they are test helper utilities, not production code requiring coverage

7. Progress tracking for test improvements:
   - Report coverage metrics before and after changes
   - Track test file consolidation progress
   - Monitor linting compliance improvements
   - Validate test execution time and reliability
   - **IMPORTANT**: Maintain a running checklist in `report/test-coverage-progress.md` (create if missing) and mark completed test tasks as `[X]` with links or descriptions.

8. Test coverage validation and completion:
   - Verify all test files follow organizational standards
   - Confirm coverage improvements meet or exceed targets
   - Validate all tests pass with proper parallel execution
   - Check linting compliance for test files
   - Report final coverage metrics and improvement summary
   - **CONFIRM testutils exclusion**: Ensure testutils packages are not included in coverage reports or analysis

## Coverage Analysis Exclusions

**IMPORTANT**: The following package patterns MUST BE EXCLUDED from coverage analysis:
- `**/testutils/` - Test utility packages containing shared test helpers
- `**/*testutils*/` - Any package with "testutils" in the name
- Test utility packages provide shared helpers for other tests but do not require their own test coverage

Note: This approach focuses exclusively on test improvements while maintaining existing functionality. Source code remains unchanged while test quality and coverage are enhanced through better organization and comprehensive test scenarios. Testutils packages are explicitly excluded from coverage requirements as they are test infrastructure, not production code.
