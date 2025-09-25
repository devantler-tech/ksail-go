---
description: Execute comprehensive linting analysis and fix all golangci-lint issues following task-driven approach based on implementation tasks and linting best practices
---

The user input can be provided directly by the agent or as a command argument - you **MUST** consider it before proceeding with the prompt (if not empty).

User input:

$ARGUMENTS

1. Run `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` from repo root and parse FEATURE_DIR and AVAILABLE_DOCS list. All paths must be absolute.

2. Load and analyze the linting context:
   - **REQUIRED**: Read tasks.md for linting-related tasks and code quality requirements
   - **REQUIRED**: Read plan.md for tech stack, architecture, and coding standards
   - **IF EXISTS**: Read contracts/ for API specifications and test requirements
   - **IF EXISTS**: Read research.md for technical decisions and constraints
   - **IF EXISTS**: Read quickstart.md for integration scenarios and quality gates

3. Parse tasks.md structure and extract linting-focused tasks:
   - **Linting compliance tasks**: golangci-lint, code quality, style enforcement
   - **Code organization tasks**: Function length, complexity, duplication reduction
   - **Test quality tasks**: Test structure, helper functions, best practices
   - **Documentation tasks**: Comments, godoc, formatting standards

4. Execute comprehensive linting analysis and fixes:
   - **Initial analysis**: Run `golangci-lint run --timeout=5m` to identify all issues
   - **Categorize issues**: Group by linter type (funlen, cyclop, wsl_v5, etc.)
   - **Prioritize fixes**: Critical issues first, then style and formatting
   - **Apply automatic fixes**: Use `golangci-lint run --fix` and `golangci-lint fmt`
   - **Manual fixes**: Address complex issues requiring code restructuring

5. Linting execution rules and priorities:
   - **Auto-fixes first**: Apply all automated formatting and style fixes
   - **Function restructuring**: Break down long functions (funlen), reduce complexity (cyclop, gocognit)
   - **Code organization**: Extract helper functions, eliminate duplication, improve readability
   - **Documentation quality**: Add missing periods (godot), improve comments, maintain consistency
   - **Test improvements**: Add t.Helper(), fix test package naming, improve test structure
   - **Error handling**: Address exhaustive switch cases, unused variables, variable naming

6. Linting-specific execution patterns:
   - **Batch similar fixes**: Group fixes by linter type for efficiency
   - **Preserve functionality**: Ensure all tests pass after each batch of fixes
   - **Follow Go idioms**: Maintain Go best practices and conventions
   - **Test helper compliance**: Add t.Helper() to test utility functions
   - **Package naming**: Use proper test package naming (_test suffix)
   - **Variable naming**: Use descriptive names for longer-lived variables

7. Progress tracking and validation:
   - Report linting progress after each batch of fixes
   - Run `golangci-lint run --timeout=5m` after each major change
   - Ensure `go test ./...` passes after each batch
   - Track remaining issue count and types
   - **IMPORTANT**: Mark completed linting tasks as [X] in tasks.md

8. Completion validation and quality gates:
   - Verify **zero golangci-lint issues** remain (`golangci-lint run` exits with code 0)
   - Confirm all tests pass (`go test ./...` successful)
   - Validate code coverage is maintained or improved
   - Check that functionality is preserved (no breaking changes)
   - Report final linting status with summary of fixes applied

## Linting Issue Priority Matrix

### **Critical (Fix First)**
- **funlen**: Function length violations (break into smaller functions)
- **cyclop/gocognit**: Cyclomatic/cognitive complexity (extract helper functions)
- **exhaustive**: Missing switch cases (add required cases)

### **High Priority**
- **goconst**: Repeated strings (extract constants)
- **gocritic**: Code quality issues (improve patterns)
- **testpackage**: Test package naming (use _test suffix)
- **thelper**: Missing t.Helper() in test utilities

### **Medium Priority**
- **godot**: Missing periods in comments (formatting)
- **nlreturn**: Missing blank lines (code style)
- **wsl_v5**: Whitespace formatting (readability)

### **Low Priority**
- **varnamelen**: Variable name length (descriptive naming)

## Common Fix Patterns

### **Function Length (funlen)**
```go
// Before: Long function
func TestLongFunction(t *testing.T) {
    // 100+ lines of test code
}

// After: Extract helpers
func TestRefactoredFunction(t *testing.T) {
    t.Run("scenario1", testScenario1)
    t.Run("scenario2", testScenario2)
}

func testScenario1(t *testing.T) {
    t.Helper()
    // Focused test logic
}
```

### **Cyclomatic Complexity (cyclop)**
```go
// Before: Complex function
func complexValidation(config Config) error {
    if condition1 && condition2 && condition3 {
        // nested logic
    }
}

// After: Extract helpers
func complexValidation(config Config) error {
    if err := validateBasicFields(config); err != nil {
        return err
    }
    return validateAdvancedFields(config)
}
```

### **Test Helper (thelper)**
```go
// Before: Missing t.Helper()
func testUtilityFunction(t *testing.T) {
    // test logic
}

// After: Add t.Helper()
func testUtilityFunction(t *testing.T) {
    t.Helper()
    // test logic
}
```

Note: This approach focuses exclusively on linting compliance while maintaining existing functionality. Code quality improvements are achieved through systematic issue resolution following Go best practices and maintaining zero-issue status.
