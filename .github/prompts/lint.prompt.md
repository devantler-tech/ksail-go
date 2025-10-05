---
description: Execute comprehensive linting analysis and fix all golangci-lint, jscpd, and cspell issues following task-driven approach based on implementation tasks and linting best practices
---

The user input can be provided directly by the agent or as a command argument - you **MUST** consider it before proceeding with the prompt (if not empty).

User input:

$ARGUMENTS

1. Collect baseline linting context:
    - Identify the repository root (directory containing `go.mod`) and work with absolute paths.
    - **REQUIRED**: Read `CONTRIBUTING.md`, `.golangci.yml`, `.jscpd.json`, and `.cspell.json` to understand linting, duplication, and spelling expectations.
    - **IF AVAILABLE**: Review `README.md`, any files under `report/` (for example `report/jscpd-report.md`), and documentation under `docs/` or `notes/` that describe quality standards.

2. Load and analyze the linting context:
    - Summarize mandatory code quality, style, and tooling requirements gathered in step 1.
    - Identify helper scripts, make targets, or reusable workflows that support linting (record their absolute paths).
    - Capture active linter configurations, thresholds, and exclusions for golangci-lint, jscpd, and cspell.

3. Consolidate linting-focused tasks:
   - **Linting compliance tasks**: golangci-lint, code quality, style enforcement
   - **Code organization tasks**: Function length, complexity, duplication reduction
   - **Code duplication tasks**: jscpd violations, extract common functions, reduce copy-paste
   - **Spelling compliance tasks**: cspell violations, fix typos, maintain dictionary
   - **Test quality tasks**: Test structure, helper functions, best practices
   - **Documentation tasks**: Comments, godoc, formatting standards

4. Execute comprehensive linting analysis and fixes:
   - **Initial analysis**: Run `golangci-lint run --timeout=5m` to identify all issues
   - **Code duplication analysis**: Run `jscpd` to identify duplicated code blocks
   - **Spelling analysis**: Run `cspell` to identify spelling errors in code and comments
   - **Categorize issues**: Group by linter type (funlen, cyclop, wsl_v5, jscpd, cspell, etc.)
   - **Prioritize fixes**: Critical issues first, then duplication, spelling, then style and formatting
   - **Apply automatic fixes**: Use `golangci-lint run --fix` and `golangci-lint fmt`
   - **Manual fixes**: Address complex issues requiring code restructuring, extract duplicated code, fix spelling

5. Linting execution rules and priorities:
   - **Auto-fixes first**: Apply all automated formatting and style fixes
   - **Function restructuring**: Break down long functions (funlen), reduce complexity (cyclop, gocognit)
   - **Code organization**: Extract helper functions, eliminate duplication, improve readability
   - **Duplication elimination**: Extract common code patterns identified by jscpd into shared functions
   - **Spelling corrections**: Fix all cspell violations, add technical terms to project dictionary
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
    - Run `jscpd` after duplication fixes to verify reduction
    - Run `cspell` after spelling fixes to verify corrections
    - Ensure `go test ./...` passes after each batch
    - Track remaining issue count and types for all linters
    - **IMPORTANT**: Maintain a running checklist in `report/linting-progress.md` (create if missing) and mark completed linting tasks as `[X]` with links or descriptions.

8. Completion validation and quality gates:
   - Verify **zero golangci-lint issues** remain (`golangci-lint run` exits with code 0)
   - Verify **zero jscpd duplications** remain (`jscpd` reports no violations)
   - Verify **zero cspell errors** remain (`cspell` reports no spelling mistakes)
   - Confirm all tests pass (`go test ./...` successful)
   - Validate code coverage is maintained or improved
   - Check that functionality is preserved (no breaking changes)
   - Report final linting status with summary of fixes applied for all linters

## Linting Issue Priority Matrix

### **Critical (Fix First)**
- **funlen**: Function length violations (break into smaller functions)
- **cyclop/gocognit**: Cyclomatic/cognitive complexity (extract helper functions)
- **exhaustive**: Missing switch cases (add required cases)
- **jscpd**: Code duplication violations (extract common functions, reduce copy-paste)

### **High Priority**
- **cspell**: Spelling errors in code and comments (fix typos, update dictionary)
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

### **Code Duplication (jscpd)**
```go
// Before: Duplicated validation logic
func ValidateKindConfig(config KindConfig) error {
    if config.Name == "" {
        return errors.New("name is required")
    }
    if config.Image == "" {
        return errors.New("image is required")
    }
    return nil
}

func ValidateK3dConfig(config K3dConfig) error {
    if config.Name == "" {
        return errors.New("name is required")
    }
    if config.Image == "" {
        return errors.New("image is required")
    }
    return nil
}

// After: Extract common validation
func validateRequiredFields(name, image string) error {
    if name == "" {
        return errors.New("name is required")
    }
    if image == "" {
        return errors.New("image is required")
    }
    return nil
}

func ValidateKindConfig(config KindConfig) error {
    return validateRequiredFields(config.Name, config.Image)
}

func ValidateK3dConfig(config K3dConfig) error {
    return validateRequiredFields(config.Name, config.Image)
}
```

Note: This approach focuses exclusively on linting compliance while maintaining existing functionality. Code quality improvements are achieved through systematic issue resolution following Go best practices, duplication elimination, and spelling accuracy while maintaining zero-issue status.
