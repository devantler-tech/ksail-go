# Tasks: Configuration File Validation

**Input**: Design documents from `/specs/001-add-validation-for/`
**Prerequisites**: plan.md (✓), research.md (✓), data-model.md (✓), contracts/ (✓)

## Execution Flow (main)

```txt
1. Load plan.md from feature directory
   → Tech stack: Go 1.24.0+, upstream validators (kind, k3d, eksctl)
   → Libraries: sigs.k8s.io/kind, github.com/k3d-io/k3d/v5, eksctl
   → Structure: Single project CLI tool with pkg/validator/ structure
2. Load design documents:
   → data-model.md: ValidationError, ValidationResult, FileLocation entities
   → contracts/: 5 contract files (validator-interface + 4 validators)
   → research.md: API simplification to single Validate() method
3. Generate tasks by category:
   → Setup: Remove spec violations (K8sVersion), API simplification, dependencies
   → Tests: Contract tests for each validator (TDD)
   → Core: Simplified validator implementations
   → Integration: End-to-end validation workflows
   → Polish: Performance benchmarks, documentation
4. Apply task rules:
   → Different validator files = mark [P] for parallel
   → Same interface files = sequential (no [P])
   → Tests before implementation (TDD)
5. Spec Compliance: Remove K8sVersion field that violates "DO NOT ALTER" requirement
6. API Simplification Focus: Remove Validate([]byte), rename ValidateStruct→Validate
```

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Phase 3.1: Setup & API Simplification

- [x] T001 Remove K8sVersion field from KSail Spec struct in pkg/apis/cluster/v1alpha1/types.go (violates spec requirement)
- [x] T002 Update validator interface to remove Validate([]byte) method in pkg/validator/interfaces.go
- [x] T003 Rename ValidateStruct→Validate in pkg/validator/interfaces.go
- [x] T004 [P] Update go.mod dependencies for upstream validators (kind, k3d, eksctl)

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

> [!CAUTION]
> CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation

- [x] T005 [P] Contract test for simplified Validator interface in pkg/validator/interfaces_test.go
- [x] T006 [P] Contract test for KSail validator Validate() method in pkg/validator/ksail/config-validator_test.go
- [x] T007 [P] Contract test for Kind validator Validate() method in pkg/validator/kind/config-validator_test.go
- [x] T008 [P] Contract test for K3d validator Validate() method in pkg/validator/k3d/config-validator_test.go
- [x] T009 [P] Contract test for EKS validator Validate() method in pkg/validator/eks/config-validator_test.go
- [x] T010 [P] Integration test complete validation workflow in pkg/validator/manager_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

- [x] T011 [P] Update ValidationError struct in pkg/validator/types.go per data-model.md
- [x] T012 [P] Update ValidationResult struct in pkg/validator/types.go per data-model.md
- [x] T013 [P] Add FileLocation type in pkg/validator/types.go per data-model.md
- [x] T014 [P] Implement simplified KSail validator Validate() method in pkg/validator/ksail/config-validator.go
- [x] T015 [P] Implement simplified Kind validator Validate() method in pkg/validator/kind/config-validator.go
- [x] T016 [P] Implement simplified K3d validator Validate() method in pkg/validator/k3d/config-validator.go
- [x] T017 [P] Implement simplified EKS validator Validate() method in pkg/validator/eks/config-validator.go
- [x] T018 Update validator manager to use simplified interface in pkg/validator/manager.go
- [x] T019 Remove deprecated Validate([]byte) method implementations across all validators

## Phase 3.4: Integration & Error Handling

- [x] T020 [P] Implement detailed error messages with FixSuggestion in pkg/validator/ksail/config-validator.go
- [x] T021 [P] Implement detailed error messages with FixSuggestion in pkg/validator/kind/config-validator.go
- [x] T022 [P] Implement detailed error messages with FixSuggestion in pkg/validator/k3d/config-validator.go
- [x] T023 [P] Implement detailed error messages with FixSuggestion in pkg/validator/eks/config-validator.go
- [x] T024 Add file location tracking for validation errors in pkg/validator/manager.go
- [x] T025 Implement validation error aggregation in pkg/validator/manager.go

## Phase 3.5: Polish & Performance

- [ ] T026 [P] Performance benchmarks for <100ms validation time in pkg/validator/benchmarks_test.go
- [ ] T027 [P] Memory usage validation <10MB in pkg/validator/benchmarks_test.go
- [ ] T028 [P] Update validator package godoc comments in pkg/validator/interfaces.go
- [ ] T029 [P] Update types package godoc comments in pkg/validator/types.go
- [ ] T030 [P] Update README.md with simplified validation API examples
- [ ] T031 Run quickstart validation scenarios from quickstart.md
- [x] T032 [REMOVED] ~~Implement EKS GetSupportedTypes() in pkg/validator/eks/config-validator.go returning ["eks"]~~ - Method removed from interface

### Validation Logic Implementation

- [x] T033 Schema validation for KSail config in pkg/validator/ksail/config-validator.go - required fields, enum constraints (use existing v1alpha1.Cluster, DO NOT ALTER)
- [x] T034 Cross-configuration coordination in pkg/validator/ksail/config-validator.go - load and validate distribution configs **USING UPSTREAM VALIDATORS**
- [x] T035 Context name validation in pkg/validator/ksail/config-validator.go - kind-{name}, k3d-{name}, EKS ARN/name patterns
- [x] T036 Error message formatting in pkg/validator/ksail/config-validator.go - actionable ValidationError creation

## Phase 3.4: Integration

- [x] T037 Integrate validation into existing config-manager in pkg/config-manager/manager.go - call validators during config loading
- [x] T038 Add validation hooks to CLI commands in cmd/ - ensure validation runs before command execution
- [x] T039 Update error handling in cmd/ui/notify package to display ValidationError messages consistently
- [x] T040 Add fail-fast behavior to config loading - prevent command execution on validation errors

## Phase 3.5: Polish

## Dependencies

- API Cleanup (T001) before API updates (T002-T003) before dependency updates (T004)
- API Updates (T001-T004) before tests (T005-T010)
- Tests (T005-T010) before implementation (T011-T019)
- Core types (T011-T013) before validator implementations (T014-T017)
- Validator implementations before manager updates (T018-T019)
- Core implementation before error handling (T020-T025)
- All implementation before performance testing (T026-T027)

## Parallel Execution Examples

```txt
# Phase 3.2: Launch contract tests together
Task: "Contract test for KSail validator Validate() method in pkg/validator/ksail/validator_test.go"
Task: "Contract test for Kind validator Validate() method in pkg/validator/kind/validator_test.go"
Task: "Contract test for K3d validator Validate() method in pkg/validator/k3d/validator_test.go"
Task: "Contract test for EKS validator Validate() method in pkg/validator/eks/validator_test.go"

# Phase 3.3: Launch validator implementations together
Task: "Implement simplified KSail validator Validate() method in pkg/validator/ksail/validator.go"
Task: "Implement simplified Kind validator Validate() method in pkg/validator/kind/validator.go"
Task: "Implement simplified K3d validator Validate() method in pkg/validator/k3d/validator.go"
Task: "Implement simplified EKS validator Validate() method in pkg/validator/eks/validator.go"
```

## API Simplification Focus Areas

> [!IMPORTANT]
> PRIMARY OBJECTIVE: Simplify validator interface from dual-method to single-method

1. **Interface Simplification**:
   - Remove: `Validate(data []byte) *ValidationResult`
   - Rename: `ValidateStruct(config interface{}) *ValidationResult` → `Validate(config interface{}) *ValidationResult`
   - Remove: `GetSupportedTypes() []string` - Simplified to single-method interface

2. **Performance Benefits**:
   - Eliminates unnecessary marshaling/unmarshaling cycles
   - Reduces memory allocations
   - Improves testability with struct inputs
   - Removes auto-discovery overhead in favor of explicit registration

3. **User Experience**:
   - Cleaner API for consumers who already have parsed configurations
   - Consistent with KSail's existing configuration loading patterns
   - Explicit validator registration provides better control
   - Actionable error messages with FixSuggestion field

## Validation Checklist

> [!IMPORTANT]
> GATE: Checked before task execution

- [x] All contracts have corresponding tests (T005-T009)
- [x] All entities have implementation tasks (T011-T013)
- [x] All tests come before implementation (T005-T010 → T011-T019)
- [x] Parallel tasks truly independent ([P] tasks use different files)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] API simplification emphasized throughout task descriptions
- [x] Spec compliance enforced (K8sVersion field removal)

## Performance Targets

- **Validation Time**: <100ms per configuration file
- **Memory Usage**: <10MB during validation operations
- **Concurrency**: Thread-safe validation for parallel operations
- **Error Quality**: Actionable messages with specific fix suggestions

## Notes

- Focus on API simplification: single `Validate(config interface{})` method
- Leverage upstream validators to avoid custom validation logic duplication
- Maintain backward compatibility during transition
- All tests must fail initially (TDD approach)
- Commit after each completed task for progress tracking
