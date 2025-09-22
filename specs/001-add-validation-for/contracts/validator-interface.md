# Validator Interface Contract

```go
// Validator defines the interface for configuration file validators
type Validator interface {
    // Validate performs validation on the provided configuration data
    // Returns ValidationResult containing status and any errors found
    Validate(data []byte) *ValidationResult

    // ValidateStruct performs validation on an already parsed configuration struct
    // Returns ValidationResult containing status and any errors found
    ValidateStruct(config interface{}) *ValidationResult

    // GetSupportedTypes returns the configuration types this validator supports
    GetSupportedTypes() []string
}
```

## Contract Requirements

### Validate Method
- **Input**: Raw configuration file content as byte array
- **Output**: ValidationResult with status and errors
- **Behavior**:
  - MUST handle marshalling errors gracefully
  - MUST prioritize parsing errors over semantic validation
  - MUST return structured ValidationError instances
  - MUST complete within 100ms for files <10KB
  - MUST NOT perform file I/O operations
  - MUST be thread-safe for concurrent validation

### ValidateStruct Method
- **Input**: Already parsed configuration struct
- **Output**: ValidationResult with status and errors
- **Behavior**:
  - MUST validate semantic correctness of configuration
  - MUST check field constraints and dependencies
  - MUST return actionable error messages
  - MUST be idempotent (same input = same output)
  - MUST handle nil or malformed structs gracefully

### GetSupportedTypes Method
- **Input**: None
- **Output**: String slice of supported configuration types
- **Behavior**:
  - MUST return consistent list of supported types
  - MUST include all configuration formats the validator handles
  - MUST be deterministic across calls

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
