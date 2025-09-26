# Tasks: KSail Init Command Enhancement

**Input**: Design documents from `/specs/002-add-the-ksail/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Implementation Reality Check ✅

**EXISTING FUNCTIONALITY** (already implemented):

- ✅ `cmd/init.go` - Full init command with Cobra setup
- ✅ `pkg/scaffolder/` - Complete scaffolding system with Kind/K3d/EKS support
- ✅ `pkg/io/generator/` - Runtime template generators for all distributions
- ✅ Distribution-specific config generation (kind.yaml, k3d.yaml, eks.yaml)
- ✅ KSail config generation (ksail.yaml) with existing cluster APIs
- ✅ Kustomization structure generation (k8s/kustomization.yaml)
- ✅ Comprehensive test coverage (cmd/init_test.go, pkg/scaffolder/scaffolder_test.go)
- ✅ Constitutional compliance (code quality, testing, UX patterns)

## Enhancement Focus (actual gaps to address)

```text
1. Audit existing implementation capabilities
   → Identify spec compliance gaps vs working functionality
   → Focus on missing UX features (spinner, --force, conflict detection)
2. Enhance rather than rebuild:
   → Add progress feedback to existing cmd/init.go
   → Add CLI flags to existing ConfigManager integration
   → Add edge case handling to existing scaffolder
3. Generate enhancement tasks:
   → Tests for new functionality only
   → Implementation for actual gaps only
   → Integration of new features with existing architecture
4. Validation approach:
   → Test enhancements don't break existing functionality
   → Verify spec compliance with enhanced implementation
```

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Paths relative to repository root: `<REPO_ROOT>`

## Phase 1: Analysis & Validation

- [x] T001 Audit existing `ksail init` functionality and compare with spec requirements
- [x] T002 [P] Test current CLI behavior and document gaps vs FR-001 to FR-016
- [x] T003 [P] Run existing test suite and validate coverage for current functionality

## Phase 2: Enhancement Tests (TDD for new features only)

### CRITICAL: Tests for NEW functionality only - existing tests already pass

- [x] T004 [P] Add test for progress spinner feedback in `cmd/init_test.go`
- [x] T005 [P] Add test for --force flag and conflict detection in `cmd/init_test.go`
- [x] T006 [P] Add test for direct CLI flags (--name, --distribution) in `cmd/init_test.go`
- [x] T007 [P] Add test for disk space validation in `pkg/scaffolder/scaffolder_test.go`
- [x] T008 [P] Add test for template integrity validation in `pkg/scaffolder/scaffolder_test.go`
- [x] T009 [P] Add test for interruption handling in `cmd/init_test.go`
- [x] T010 [P] Add test for directory name validation in `pkg/scaffolder/scaffolder_test.go`

## Phase 3: Enhancement Implementation (build on existing code)

- [x] T011: Add progress spinner/status indicator during init process
- [x] T012 Add --force flag handling and file conflict detection to existing scaffolder (check for existing files before generation, respect force parameter from Scaffold method)
- [x] T013 Enhance existing ConfigManager integration to support direct CLI flags (--name, --distribution flags with proper Viper binding and validation)
- [x] T014 Add disk space validation to existing `pkg/scaffolder/scaffolder.go` Scaffold method (check available space >10MB before file operations, provide specific error with breakdown)
- [ ] T015 Add template integrity validation to existing generator system (validate each generator can produce valid output before file operations)
- [ ] T016 Add signal handling (SIGINT/SIGTERM) for graceful interruption cleanup (implement cleanup of partial files, restore original directory state)
- [ ] T017 Add directory name validation to existing scaffolder input validation (validate against filesystem constraints, provide specific error messages for violations)

## Phase 4: Integration & Validation

- [ ] T018 Integrate progress feedback with existing file generation events
- [ ] T019 Enhance existing error messages to meet NFR-005 (actionable remediation)
- [ ] T020 Add performance benchmarking to validate <200ms CLI + <5s initialization
- [ ] T021 Update existing CLI help text to document new flags and options
- [ ] T022 Validate enhanced implementation against all spec requirements (FR-001 to FR-016)

## Phase 5: Quality Assurance

- [ ] T023 [P] Run enhanced test suite and ensure >90% coverage maintained
- [ ] T024 [P] Run golangci-lint and ensure zero issues (constitutional requirement)
- [ ] T025 [P] Performance validation: ensure <200ms CLI response time
- [ ] T026 [P] Memory usage validation: ensure <50MB during initialization (NFR-002)
- [ ] T027 [P] Atomic operations verification: test file creation rollback on interruption (NFR-003)
- [ ] T028 [P] Compatibility testing: validate generated files work with existing KSail commands (NFR-005)
- [ ] T029 Manual validation using quickstart.md scenarios on enhanced implementation
- [ ] T030 Regression testing: ensure existing functionality still works correctly

## Dependencies

- Phase 1 (T001-T003): Analysis must complete before any implementation
- Phase 2 (T004-T010): Enhancement tests before implementation (TDD)
- Phase 3 (T011-T017): Core enhancements can run in parallel except:
  - T012 (conflict detection) before T011 (spinner integration)
  - T013 (CLI flags) before T011 (spinner needs flag handling)
- Phase 4 (T018-T022): Integration depends on completed Phase 3
- Phase 5 (T023-T030): Quality assurance depends on completed implementation

## Parallel Execution Examples

### Phase 2: Enhancement Tests (Run Together)

```bash
# New CLI feature tests
Task: "Add test for progress spinner feedback in cmd/init_test.go"
Task: "Add test for --force flag and conflict detection in cmd/init_test.go"
Task: "Add test for direct CLI flags (--name, --distribution) in cmd/init_test.go"
Task: "Add test for interruption handling in cmd/init_test.go"

# New scaffolder feature tests
Task: "Add test for disk space validation in pkg/scaffolder/scaffolder_test.go"
Task: "Add test for template integrity validation in pkg/scaffolder/scaffolder_test.go"
Task: "Add test for directory name validation in pkg/scaffolder/scaffolder_test.go"
```

### Phase 3: Core Enhancements (Partial Parallel)

```bash
# Independent enhancements (can run together)
Task: "Add disk space validation to existing pkg/scaffolder/scaffolder.go Scaffold method"
Task: "Add template integrity validation to existing generator system"
Task: "Add directory name validation to existing scaffolder input validation"
Task: "Add signal handling (SIGINT/SIGTERM) for graceful interruption cleanup"

# Sequential dependencies
# 1. First: "Add --force flag and file conflict detection to existing scaffolder"
# 2. Then: "Enhance existing ConfigManager or add direct CLI flags"
# 3. Finally: "Add progress spinner to existing cmd/init.go HandleInitRunE function"
```

### Phase 5: Quality Assurance (Run Together)

```bash
Task: "Run enhanced test suite and ensure >90% coverage maintained"
Task: "Run golangci-lint and ensure zero issues (constitutional requirement)"
Task: "Performance validation: ensure <200ms CLI response time"
Task: "Memory usage validation: ensure <50MB during initialization"
Task: "Atomic operations verification: test file creation rollback on interruption"
Task: "Compatibility testing: validate generated files work with existing KSail commands"
```

## Key Implementation Files

### Existing Files to Enhance

- `cmd/init.go` - **EXISTS**: Add spinner, signal handling, enhanced flags
- `cmd/init_test.go` - **EXISTS**: Add tests for new functionality only
- `pkg/scaffolder/scaffolder.go` - **EXISTS**: Add conflict detection, validation
- `pkg/scaffolder/scaffolder_test.go` - **EXISTS**: Add tests for enhancements
- `pkg/io/generator/` - **EXISTS**: Runtime template system (no embedding needed)
- `pkg/apis/cluster/v1alpha1/` - **EXISTS**: Complete API models
- `pkg/config-manager/` - **EXISTS**: May need minor enhancements for direct flags

### Files That DON'T Need Creation

- ❌ Models: Already exist in `pkg/apis/cluster/v1alpha1/`
- ❌ Templates: Runtime generation via `pkg/io/generator/` already implemented
- ❌ Core scaffolder: `pkg/scaffolder/` already has full Kind/K3d/EKS support
- ❌ File I/O: `pkg/io/` already handles safe operations
- ❌ Basic tests: Existing test suite already provides coverage

## Constitutional Compliance

- **TDD Required**: All tests (T004-T010) must be written and failing before implementation
- **Test Coverage**: >90% coverage required for all new code (T023)
- **Performance**: <200ms CLI response + <5s initialization (T025)
- **Memory Usage**: <50MB during operation (T026)
- **Error Handling**: Fail-fast with user-friendly messages (T019)
- **Code Quality**: golangci-lint must pass on all new code (T024)

## Notes

- [P] tasks target different files and can run in parallel
- Runtime template generation enables offline operation per FR-011
- Atomic file operations prevent partial state on errors (T027)
- All CLI patterns follow existing Cobra conventions
- Integration tests validate complete user scenarios from quickstart.md (T029)
- Enhancement approach builds on existing cmd/init.go and pkg/scaffolder/ implementation
