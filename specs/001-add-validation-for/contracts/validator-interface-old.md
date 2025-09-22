# Validator Interface Contract

```go
// Validator defines a type-safe interface for configuration file validators
// T represents the specific configuration type that this validator handles
type Validator[T any] interface {
    // Validate performs validation on a typed configuration object
    // Returns ValidationResult containing status and any errors found
    Validate(config T) *ValidationResult
}

// GenericValidator provides type-erased validation for the ValidatorManager
// This allows different validator types to be used through a common interface
type GenericValidator interface {
    // ValidateAny performs validation on a configuration of unknown type
    // Handles type conversion and parsing internally before validation
    ValidateAny(config any) *ValidationResult
}
```

## Contract Requirements

### Type-Safe Validate Method

- **Input**: Specific configuration type (e.g., `*v1alpha1.Cluster`, `*v1alpha4.Cluster`)
- **Output**: ValidationResult with status and errors
- **Behavior**:
  - MUST validate semantic correctness of the typed configuration
  - MUST check field constraints and dependencies
  - MUST return structured ValidationError instances
  - MUST complete within 100ms for typical configurations
  - MUST NOT perform file I/O operations
  - MUST be thread-safe for concurrent validation
  - MUST return actionable error messages
  - MUST be idempotent (same input = same output)
  - MUST handle nil input gracefully

### Type-Erased ValidateAny Method

- **Input**: Configuration data (any type: []byte, map[string]any, or typed struct)
- **Output**: ValidationResult with status and errors
- **Behavior**:
  - MUST handle both raw []byte data and parsed structs
  - MUST handle marshalling errors gracefully for byte input
  - MUST prioritize parsing errors over semantic validation
  - MUST convert data to appropriate type before calling Validate()
  - MUST return structured ValidationError instances
  - MUST complete within 100ms for files <10KB
  - MUST NOT perform file I/O operations
  - MUST be thread-safe for concurrent validation
  - MUST return actionable error messages
  - MUST be idempotent (same input = same output)
  - MUST handle nil or malformed input gracefully

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
