# Feature Specification: Configuration File Validation

**Feature Branch**: `001-add-validation-for`
**Created**: 2025-09-22
**Status**: Not Implemented
**Input**: User description: "Add validation for configuration files, to ensure issues are highlighted for users, with actionable steps to fix their configuration files. The validation must be run whenever a config file is loaded, and fail in case the contents of the file is invalid with ksail or invalid in general. Marshalling errors should take precedence, such that validation can happen in memory, and not via read/write operations that are inefficient and hard to test."

## ⚠️ IMPLEMENTATION STRATEGY

**CRITICAL**: This validation implementation MUST follow a specific validator architecture where each configuration file gets its own validator, and validation covers both internal misconfigurations and cross-configuration misconfigurations. The implementation MUST prioritize upstream Go package validators over custom validation logic. Use official APIs from:

- `sigs.k8s.io/kind` for Kind configuration validation
- `github.com/k3d-io/k3d/v5` for K3d configuration validation
- `github.com/weaveworks/eksctl` for EKS configuration validation
- Existing KSail `v1alpha1.Cluster` APIs (DO NOT ALTER)

**Validator Architecture Requirements**:

- Each configuration type has dedicated config managers that handle file loading and parsing into structs
- Validators operate on pre-loaded configuration structs, not raw files or bytes
- Validation focuses purely on semantic correctness and cross-configuration consistency
- Custom validation is only implemented when upstream validators don't provide the needed functionality

**DO NOT** implement custom validation logic that duplicates functionality already available in upstream packages. This ensures compatibility with official tools and reduces maintenance burden.

## User Scenarios & Testing

### Primary User Story

As a KSail user, when I create or modify configuration files (ksail.yaml, kind.yaml, k3d.yaml, eks.yaml), I want immediate feedback about any errors or invalid configurations so I can quickly fix them before attempting cluster operations. The system should use existing config managers to load configurations into memory, then validate both individual configuration correctness and cross-configuration consistency, providing specific error messages with actionable fix suggestions.

### Acceptance Scenarios

1. **Given** a user has an invalid ksail.yaml file with malformed YAML syntax, **When** the config manager attempts to load it, **Then** the system displays a clear error message indicating the YAML syntax error with actionable fix instructions

2. **Given** a user has a loaded ksail.yaml configuration with a context name that doesn't match the expected pattern for their distribution (e.g., "wrong-context" instead of "kind-cluster-name"), **When** ksail validates the loaded configuration struct, **Then** the system shows a specific error explaining the expected context pattern with examples

3. **Given** a user has loaded kind.yaml and ksail.yaml configurations where the cluster names don't match, **When** ksail validates the configurations, **Then** the system fails with a detailed explanation of the naming inconsistency and how to fix it

4. **Given** a user has valid loaded ksail.yaml and distribution configuration structs, **When** they run validation, **Then** validation passes silently and operations proceed normally

5. **Given** a user has a loaded k3d.yaml configuration with CNI settings that conflict with their loaded ksail.yaml CNI configuration, **When** validation runs, **Then** the system detects the cross-configuration conflict and suggests specific changes to align both configurations

### Edge Cases

- What happens when configuration structs are loaded but contain invalid field combinations?
- How does system handle when a config struct has valid syntax but semantically invalid field values?
- What happens when required distribution config managers fail to load their configurations?
- How does validation behave when cross-configuration dependencies exist but one config is missing?
- What happens when upstream validators are not available or fail during validation?
- How does the system handle when loaded configurations reference non-existent resources?
- What happens when context names in loaded configs are valid but point to non-existent clusters?

## 🎯 **Validation Architecture: Clear Responsibility Separation**

**CRITICAL DESIGN PRINCIPLE**: This validation system uses a clear separation of concerns between different validator types:

### **KSail Validator: Cross-Configuration Orchestrator**

- **Primary Role**: Validates consistency **across multiple loaded configurations**
- **Scope**: Coordinates validation between KSail, Kind, K3d, EKS configs
- **Responsibility**: Ensures naming consistency, compatibility checks, orchestrates validation
- **Cross-Config Authority**: **ONLY** the KSail validator handles cross-configuration validation

### **Distribution Validators: Single-Config Specialists**

- **Primary Role**: Validates **individual configuration semantics** using upstream APIs
- **Scope**: Each validator handles only its own distribution's configuration struct
- **Responsibility**: Leverage official Kind/K3d/EKS validation APIs, no cross-config logic
- **Limited Scope**: **DO NOT** implement cross-configuration consistency checks

This architecture ensures:

- Clear ownership of validation responsibilities
- Proper use of upstream APIs for distribution-specific validation
- Centralized cross-configuration logic in KSail validator
- Maintainable and testable validation components

## Integration with Existing Config Management

The validation system integrates with existing config managers that handle file loading and parsing:

```go
// Existing config managers handle file loading into typed structs
ksailManager := configmanager.NewConfigManager(fieldSelectors...)
ksailConfig, err := ksailManager.LoadConfig() // Returns *v1alpha1.Cluster

kindManager := kind.NewConfigManager("kind.yaml")
kindConfig, err := kindManager.LoadConfig() // Returns *v1alpha4.Cluster

k3dManager := k3d.NewConfigManager("k3d.yaml")
k3dConfig, err := k3dManager.LoadConfig() // Returns *v1alpha5.SimpleConfig

// Validators operate on the loaded structs
ksailValidator := validator.NewKSailValidator()
kindValidator := validator.NewKindValidator()
k3dValidator := validator.NewK3dValidator()

// Validation happens on loaded structs, not files
result := ksailValidator.Validate(ksailConfig)
result = kindValidator.Validate(kindConfig)
result = k3dValidator.Validate(k3dConfig)
```

This architecture ensures clear separation of concerns:

- **Config Managers**: Handle file discovery, loading, parsing, and struct creation with proper defaults
- **Validators**: Handle semantic validation and cross-configuration consistency on loaded structs
- **Upstream APIs**: Provide validation for distribution-specific configuration correctness

## Requirements

### Functional Requirements

- **FR-001**: System MUST leverage existing config managers to load configurations into memory as typed structs before validation begins
- **FR-002**: System MUST validate loaded configuration structs whenever they are accessed during any command execution
- **FR-003**: System MUST prioritize config manager loading errors (marshalling/parsing) before semantic validation to ensure efficient error reporting
- **FR-004**: System MUST provide dedicated validators for each configuration type that operate on loaded structs:
  - KSail validator for v1alpha1.Cluster structs loaded by KSail config manager
  - Kind validator for v1alpha4.Cluster structs loaded by Kind config manager
  - K3d validator for v1alpha5.SimpleConfig structs loaded by K3d config manager
  - EKS validator for EKS configuration structs loaded by EKS config manager
- **FR-005**: System MUST validate cross-configuration consistency between loaded structs, specifically:
  - Context names match expected patterns (kind-{name}, k3d-{name})
  - Cluster names are consistent between ksail and distribution configs
  - CNI/CSI/Ingress settings align between ksail and distribution configurations
- **FR-006**: System MUST fail fast when configuration validation errors are detected on loaded structs, preventing execution of potentially destructive operations
- **FR-007**: System MUST provide actionable error messages that include specific field names, current values, expected values, and fix examples for struct fields
- **FR-008**: System MUST leverage upstream validation APIs on loaded structs where available to ensure compatibility with official tools
- **FR-009**: System MUST validate that loaded configuration values are compatible with the target Kubernetes distribution capabilities
- **FR-010**: System MUST perform validation purely in memory on loaded structs without requiring additional file I/O operations

### Non-Functional Requirements

- **NFR-001**: Configuration validation MUST complete within 100ms for typical configuration files (< 10KB)
- **NFR-002**: Validation error messages MUST be human-readable and actionable by users with basic Kubernetes knowledge
- **NFR-003**: Validation logic MUST be unit testable without requiring file system operations
- **NFR-004**: Memory usage during validation MUST not exceed 10MB regardless of configuration file size

### Key Entities

- **ConfigValidator**: Main validator that orchestrates configuration validation across loaded configuration structs and coordinates between different specialized validators
- **KSailValidator**: Dedicated validator for loaded v1alpha1.Cluster structs that handles cross-configuration coordination and context name validation
- **KindValidator**: Dedicated validator for loaded v1alpha4.Cluster structs that leverages upstream Kind APIs for schema and semantic validation
- **K3dValidator**: Dedicated validator for loaded v1alpha5.SimpleConfig structs that leverages upstream K3d APIs for configuration validation
- **EKSValidator**: Dedicated validator for loaded EKS configuration structs that leverages upstream eksctl APIs for EKS-specific validation
- **ConfigManagers**: Existing components responsible for loading configurations from files into typed structs (KSail, Kind, K3d config managers)
- **ValidationError**: Represents specific validation failures with field paths, current/expected values, and actionable fix suggestions for struct fields
- **ValidationResult**: Contains overall validation status and collection of errors with contextual information

## Review & Acceptance Checklist

### Content Quality

- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Execution Status

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed
- [ ] Implementation started
- [ ] Core validation types implemented
- [ ] Validator interfaces implemented
- [ ] Configuration validators implemented
- [ ] Integration with config managers completed
- [ ] Testing and validation completed
