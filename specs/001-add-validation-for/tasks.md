# Tasks: Configuration File Validation

**Input**: Design documents from `/specs/001-add-validation-for/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## ⚠️ CRITICAL IMPLEMENTATION REQUIREMENTS

**UPSTREAM VALIDATOR PRIORITY**: Use upstream Go package validators wherever available to avoid duplicating validation logic that already exists in well-tested packages:

- **Kind Validator**: MUST use `sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster` struct and leverage Kind's official validation methods
- **K3d Validator**: MUST use `github.com/k3d-io/k3d/v5/pkg/config/v1alpha5.SimpleConfig` struct and K3d's validation patterns
- **EKS Validator**: MUST use `github.com/weaveworks/eksctl` and AWS SDK Go v2 packages for EKS configuration validation
- **KSail Validator**: Use existing `github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1.Cluster` - **DO NOT ALTER CONFIG STRUCTURE**
- **Custom Logic**: Only implement custom validation for KSail-specific requirements NOT covered by upstream packages

This ensures validation behavior matches upstream tools exactly and reduces maintenance overhead.

## Execution Flow (main)

```txt
1. Load plan.md from feature directory
   → Extract: Go 1.24.0+, pkg/validator structure, sigs.k8s.io/yaml dependencies
2. Load optional design documents:
   → data-model.md: Extract ValidationError, ValidationResult, ConfigurationSchema entities
   → contracts/: validator-interface.md, ksail-validator.md, kind-validator.md, k3d-validator.md, eks-validator.md
   → research.md: Independent validator packages, in-memory validation strategy
3. Generate tasks by category:
   → Setup: pkg/validator structure, dependencies, interfaces
   → Tests: contract tests for each validator, integration tests
   → Core: ValidationError, ValidationResult models, validator implementations
   → Integration: config-manager integration, CLI command integration
   → Polish: unit tests, performance validation, documentation
4. Apply task rules:
   → Different validator packages = mark [P] for parallel
   → Same package files = sequential (no [P])
   → Tests before implementation (TDD)
5. Number tasks sequentially (T001, T002...)
6. Generate dependency graph
7. Create parallel execution examples
8. Validate task completeness:
   → All validator contracts have tests? ✓
   → All entities have models? ✓
   → All validators implemented? ✓
9. Return: SUCCESS (tasks ready for execution)
```

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions

Go project structure with pkg/validator packages:

- `pkg/validator/ksail/` - KSail configuration validator
- `pkg/validator/kind/` - Kind configuration validator
- `pkg/validator/k3d/` - K3d configuration validator
- `pkg/validator/eks/` - EKS configuration validator

## Phase 3.1: Setup

- [x] T001 Create pkg/validator directory structure with ksail/, kind/, k3d/, eks/ subdirectories
- [x] T002 Create core validation types file pkg/validator/types.go with ValidationError, ValidationResult, FileLocation structs
- [x] T003 Create validator interface file pkg/validator/interface.go with Validator interface from contracts
- [x] T004 [P] Configure testing setup with testify and go-snaps dependencies in go.mod

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

> [!CAUTION]
> CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation

### Contract Tests (Parallel - Different Validators)

- [x] T005 [P] Contract test for KSail validator in pkg/validator/ksail/config-validator_test.go - test Validate() method with ksail.yaml content
- [x] T006 [P] Contract test for Kind validator in pkg/validator/kind/config-validator_test.go - test Validate() method with kind.yaml content
- [x] T007 [P] Contract test for K3d validator in pkg/validator/k3d/config-validator_test.go - test Validate() method with k3d.yaml content
- [ ] T008 [P] Contract test for EKS validator in pkg/validator/eks/config-validator_test.go - test Validate() method with eksctl configuration content

### Interface Contract Tests (Parallel - Different Validators)

- [x] T009 [P] Interface compliance test for KSail validator in pkg/validator/ksail/config-validator_test.go - test ValidateStruct() and GetSupportedTypes()
- [x] T010 [P] Interface compliance test for Kind validator in pkg/validator/kind/config-validator_test.go - test ValidateStruct() and GetSupportedTypes()
- [x] T011 [P] Interface compliance test for K3d validator in pkg/validator/k3d/config-validator_test.go - test ValidateStruct() and GetSupportedTypes()
- [ ] T012 [P] Interface compliance test for EKS validator in pkg/validator/eks/config-validator_test.go - test ValidateStruct() and GetSupportedTypes()

### Validation Scenario Tests (Based on quickstart.md)

- [x] T011 [P] YAML syntax error test in pkg/validator/ksail/config-validator_test.go - test malformed YAML detection
- [x] T012 [P] Invalid field values test in pkg/validator/ksail/config-validator_test.go - test invalid distribution enum
- [x] T013 [P] Missing required fields test in pkg/validator/ksail/config-validator_test.go - test missing spec section
- [x] T014 [P] Cross-configuration validation test in pkg/validator/ksail/config-validator_test.go - test name mismatch between configs
- [x] T015 [P] Performance validation test in pkg/validator/ksail/config-validator_test.go - test <100ms validation time

### Integration Tests

- [ ] T016 [P] Integration test for complete validation workflow in pkg/validator/integration_test.go - test ksail → kind coordination
- [ ] T017 [P] Integration test for error message format in pkg/validator/integration_test.go - test structured ValidationError output

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Core Types Implementation

- [x] T018 Implement ValidationError struct in pkg/validator/types.go with Field, Message, CurrentValue, ExpectedValue, FixSuggestion, Location fields
- [x] T019 Implement ValidationResult struct in pkg/validator/types.go with Valid, Errors, Warnings, ConfigFile fields
- [x] T020 Implement FileLocation struct in pkg/validator/types.go with FilePath, Line, Column fields

### Validator Implementations (Parallel - Independent Validators)

- [x] T021 [P] Implement KSail validator in pkg/validator/ksail/config-validator.go - Validate() method for ksail.yaml using existing v1alpha1.Cluster (DO NOT ALTER)
- [x] T022 [P] Implement Kind validator in pkg/validator/kind/config-validator.go - Validate() method for kind.yaml **USING sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster**
- [x] T023 [P] Implement K3d validator in pkg/validator/k3d/config-validator.go - Validate() method for k3d.yaml **USING github.com/k3d-io/k3d/v5/pkg/config/v1alpha5.SimpleConfig**
- [ ] T024 [P] Implement EKS validator in pkg/validator/eks/config-validator.go - Validate() method for eksctl configuration **USING github.com/weaveworks/eksctl APIs**

### Validator Struct Methods (Parallel - Independent Validators)

- [x] T025 [P] Implement KSail ValidateStruct() in pkg/validator/ksail/config-validator.go for v1alpha1.Cluster validation using existing struct
- [x] T026 [P] Implement Kind ValidateStruct() in pkg/validator/kind/config-validator.go **LEVERAGING official v1alpha4.Cluster validation methods**
- [x] T027 [P] Implement K3d ValidateStruct() in pkg/validator/k3d/config-validator.go **LEVERAGING official v1alpha5.SimpleConfig validation patterns**
- [ ] T028 [P] Implement EKS ValidateStruct() in pkg/validator/eks/config-validator.go **LEVERAGING official eksctl ClusterConfig validation**

### Validator Support Methods (Parallel - Independent Validators)

- [x] T029 [P] Implement KSail GetSupportedTypes() in pkg/validator/ksail/config-validator.go returning ["ksail"]
- [x] T030 [P] Implement Kind GetSupportedTypes() in pkg/validator/kind/config-validator.go returning ["kind"]
- [x] T031 [P] Implement K3d GetSupportedTypes() in pkg/validator/k3d/config-validator.go returning ["k3d"]
- [ ] T032 [P] Implement EKS GetSupportedTypes() in pkg/validator/eks/config-validator.go returning ["eks"]

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

### Unit Tests (Parallel - Comprehensive Coverage)

- [ ] T041 [P] Unit tests for ValidationError methods in pkg/validator/types_test.go
- [ ] T042 [P] Unit tests for ValidationResult methods in pkg/validator/types_test.go
- [ ] T043 [P] Unit tests for FileLocation methods in pkg/validator/types_test.go
- [ ] T044 [P] Unit tests for KSail validator edge cases in pkg/validator/ksail/config-validator_test.go
- [ ] T045 [P] Unit tests for Kind validator edge cases in pkg/validator/kind/config-validator_test.go
- [ ] T046 [P] Unit tests for K3d validator edge cases in pkg/validator/k3d/config-validator_test.go
- [ ] T047 [P] Unit tests for EKS validator edge cases in pkg/validator/eks/config-validator_test.go

### Performance and Documentation

- [ ] T044 Performance benchmarks in pkg/validator/benchmark_test.go - validate <100ms and <10MB constraints
- [ ] T045 [P] Add godoc comments to all public interfaces and types in pkg/validator/
- [ ] T046 [P] Update README.md with configuration validation feature documentation
- [ ] T047 Validate complete quickstart.md scenarios manually - run all test cases from quickstart guide

## Dependencies

```txt
Setup Phase:
T001 → T002 → T003 → T004

Test Phase (TDD):
T001-T004 → T005-T017 (all tests must be written and failing)

Core Implementation:
T005-T017 → T018-T020 (core types)
T018-T020 → T021-T023 (validator implementations)
T021-T023 → T024-T026 (struct validation)
T024-T026 → T027-T029 (support methods)
T027-T029 → T030-T033 (validation logic)

Integration:
T018-T033 → T034-T037 (config-manager and CLI integration)

Polish:
T034-T037 → T038-T047 (unit tests, performance, docs)
```

## Parallel Execution Examples

### Phase 3.2: Contract Tests (Run Together)

```bash
# All validator contract tests can run in parallel
go test -v ./pkg/validator/ksail/ -run TestConfigValidator_Validate &
go test -v ./pkg/validator/kind/ -run TestConfigValidator_Validate &
go test -v ./pkg/validator/k3d/ -run TestConfigValidator_Validate &
go test -v ./pkg/validator/eks/ -run TestConfigValidator_Validate &
wait
```

### Phase 3.3: Validator Implementations (Run Together)

```bash
# Implement all validators in parallel (different packages)
# T021-T023: Core Validate() methods
# T024-T026: ValidateStruct() methods
# T027-T029: GetSupportedTypes() methods
```

### Phase 3.5: Unit Testing (Run Together)

```bash
# All unit tests can run in parallel
go test -v ./pkg/validator/ksail/ &
go test -v ./pkg/validator/kind/ &
go test -v ./pkg/validator/k3d/ &
go test -v ./pkg/validator/eks/ &
go test -v ./pkg/validator/ &
wait
```

## Completion Criteria

✅ **All Validators Implemented**: KSail, Kind, K3d, EKS validators with full interface compliance
✅ **TDD Compliance**: All tests written first and failing before implementation
✅ **Performance Validated**: <100ms validation time, <10MB memory usage verified
✅ **Integration Complete**: Validation integrated into config-manager and CLI commands
✅ **Documentation Updated**: Godoc comments, README, and quickstart scenarios validated
✅ **Constitutional Compliance**: Code quality, testing standards, and user experience requirements met
