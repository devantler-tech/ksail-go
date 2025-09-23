# Configuration Validation Architecture Contract

## Validator Interface for Loaded Structs

```go
// Validator defines an interface for validating loaded configuration structs
type Validator[T any] interface {
    // Validate performs validation on a loaded configuration struct
    // Returns ValidationResult containing status and any errors found
    Validate(config T) *ValidationResult
}
```

## Contract Requirements

### Type-Safe Validation Method

- **Input**: Pre-loaded configuration struct of specific type (e.g., `*v1alpha1.Cluster`, `*v1alpha4.Cluster`)
- **Output**: ValidationResult with status and errors
- **Behavior**:
  - MUST validate semantic correctness of the loaded configuration struct
  - MUST check field constraints and dependencies within the struct
  - MUST return structured ValidationError instances for any issues found
  - MUST complete within 100ms for typical configuration structs
  - MUST NOT perform file I/O operations (configs are already loaded)
  - MUST be thread-safe for concurrent validation
  - MUST return actionable error messages with field paths
  - MUST be idempotent (same struct input = same validation output)
  - MUST handle nil or malformed structs gracefully

## Integration with Config Managers

The validation system integrates seamlessly with existing config managers that handle the file loading and struct creation:

### Config Manager Responsibilities

```go
// Each config manager handles file loading and struct creation
type ConfigManager[T any] interface {
    LoadConfig() (*T, error)
}

// Example implementations:
ksailManager := configmanager.NewConfigManager(fieldSelectors...)
kindManager := kind.NewConfigManager("kind.yaml")
k3dManager := k3d.NewConfigManager("k3d.yaml")
```

### Validator Responsibilities

```go
// Validators operate on the loaded structs from config managers
type KSailValidator struct{}
func (v *KSailValidator) Validate(config *v1alpha1.Cluster) *ValidationResult

type KindValidator struct{}
func (v *KindValidator) Validate(config *v1alpha4.Cluster) *ValidationResult

type K3dValidator struct{}
func (v *K3dValidator) Validate(config *v1alpha5.SimpleConfig) *ValidationResult
```

## Architecture Design Benefits

### Clear Separation of Concerns

- **Config Managers**: Handle file discovery, loading, parsing, defaults, and struct creation
- **Validators**: Handle semantic validation and cross-configuration consistency on loaded structs
- **Upstream APIs**: Provide distribution-specific validation for loaded configuration correctness

### Upstream API Integration

Each validator leverages official upstream APIs on the loaded structs:

- **Kind validation**: Uses `sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster` validation methods
- **K3d validation**: Uses `github.com/k3d-io/k3d/v5/pkg/config/v1alpha5.SimpleConfig` validation methods
- **EKS validation**: Uses `github.com/weaveworks/eksctl` configuration validation methods
- **KSail validation**: Uses existing `v1alpha1.Cluster` API validation (DO NOT ALTER)

### Error Handling Strategy

- **Loading errors take precedence**: Config manager YAML/parsing errors are reported first
- **Actionable validation messages**: Each error includes specific struct field, current value, expected value, and fix suggestion
- **Struct field context**: All errors reference struct field paths rather than file locations
- **Cross-config coordination**: Validators can load and compare multiple configuration structs for consistency validation
- **Simplified Dependencies**: No external validator parameters needed
- **Self-Contained**: ValidatorManager contains all necessary validator implementations
- **Better Testing**: Direct testing of typed validators without interface complications

## Type Safety Benefits

- **Compile-time validation**: Validators can only accept their intended configuration types
- **Clear contracts**: Each validator's expected input type is explicit
- **Improved IDE support**: Better autocomplete and type checking
- **Reduced runtime errors**: Type mismatches caught at compile time

## Error Handling Contract

### ValidationError Requirements

- Field path MUST be specific and navigable
- Error message MUST be human-readable
- Fix suggestion MUST be actionable
- Location information MUST be accurate when available

### ValidationResult Requirements

- Valid field MUST accurately reflect error state
- Errors slice MUST contain all validation failures
- Warnings MUST NOT affect validation status
- ConfigFile path MUST be provided for context

## Performance Contract

- Validation MUST complete within 100ms for typical files (<10KB)
- Memory usage MUST NOT exceed 10MB during validation
- No file I/O operations permitted during validation
- Must be suitable for unit testing without filesystem dependencies

## Thread Safety Contract

- All validator methods MUST be thread-safe
- Concurrent validation calls MUST NOT interfere
- No shared mutable state between validation calls
- Validators MUST be safe for use in concurrent CLI commands

## Removed Interfaces

The following interfaces have been removed in favor of the simplified architecture:

- `GenericValidator` - Eliminated type erasure for better type safety
- `ValidateAny(any)` methods - Removed from all validators to enforce type safety
- Registration-based ValidatorManager - Replaced with dependency injection pattern

This simplification ensures better code quality, compile-time safety, and clearer architecture.
