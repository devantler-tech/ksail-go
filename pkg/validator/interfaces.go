// Package validator provides interfaces for configuration file validation.
package validator

// Validator defines the interface for configuration file validators.
// All implementations must be thread-safe and suitable for concurrent use.
type Validator interface {
	// Validate performs validation on an already parsed configuration struct.
	// This method focuses on semantic validation of the configuration content.
	//
	// Parameters:
	//   - config: Already parsed configuration struct
	//
	// Returns:
	//   - ValidationResult containing status and any errors found
	//
	// Contract Requirements:
	//   - MUST validate semantic correctness of configuration
	//   - MUST check field constraints and dependencies
	//   - MUST return actionable error messages
	//   - MUST be idempotent (same input = same output)
	//   - MUST handle nil or malformed structs gracefully
	Validate(config interface{}) *ValidationResult

	// GetSupportedTypes returns the configuration types this validator supports.
	// This helps callers determine which validator to use for a given file.
	//
	// Returns:
	//   - String slice of supported configuration types (e.g., ["ksail", "kind", "k3d"])
	//
	// Contract Requirements:
	//   - MUST return consistent list of supported types
	//   - MUST include all configuration formats the validator handles
	//   - MUST be deterministic across calls
	GetSupportedTypes() []string
}

// ValidatorFactory creates validator instances for specific configuration types.
// This interface supports the factory pattern for validator creation.
type ValidatorFactory interface {
	// CreateValidator creates a new validator instance for the specified type.
	//
	// Parameters:
	//   - configType: The type of configuration to validate (e.g., "ksail", "kind", "k3d")
	//
	// Returns:
	//   - Validator instance for the specified type
	//   - Error if the configuration type is not supported
	CreateValidator(configType string) (Validator, error)

	// GetSupportedTypes returns all configuration types supported by this factory.
	//
	// Returns:
	//   - String slice of all supported configuration types
	GetSupportedTypes() []string
}

// SchemaProvider defines the interface for providing validation schemas.
// This allows for flexible schema definition and loading strategies.
type SchemaProvider interface {
	// GetSchema returns the validation schema for the specified configuration type.
	//
	// Parameters:
	//   - configType: The type of configuration (e.g., "ksail", "kind", "k3d")
	//
	// Returns:
	//   - ConfigurationSchema for the specified type
	//   - Error if the schema cannot be loaded or type is unsupported
	GetSchema(configType string) (*ConfigurationSchema, error)

	// ValidateSchema validates that a schema is well-formed and usable.
	//
	// Parameters:
	//   - schema: The ConfigurationSchema to validate
	//
	// Returns:
	//   - Error if the schema is malformed or invalid
	ValidateSchema(schema *ConfigurationSchema) error
}

// FileLocationProvider defines the interface for providing file location information.
// This is used by validators to include precise error location data.
type FileLocationProvider interface {
	// GetLocation returns the file location for a specific field path.
	//
	// Parameters:
	//   - filePath: Absolute path to the configuration file
	//   - fieldPath: Dot-notation path to the field (e.g., "spec.distribution")
	//
	// Returns:
	//   - FileLocation with line/column information
	//   - Error if location cannot be determined
	GetLocation(filePath, fieldPath string) (FileLocation, error)

	// GetLineContent returns the content of a specific line in the file.
	//
	// Parameters:
	//   - filePath: Absolute path to the configuration file
	//   - lineNumber: Line number (1-based)
	//
	// Returns:
	//   - String content of the specified line
	//   - Error if line cannot be read
	GetLineContent(filePath string, lineNumber int) (string, error)
}

// ValidationContextProvider defines the interface for providing validation context.
// This allows validators to access additional information needed for validation.
type ValidationContextProvider interface {
	// GetFilePath returns the file path being validated.
	GetFilePath() string

	// GetFileContent returns the raw file content being validated.
	GetFileContent() []byte

	// GetRelatedFiles returns paths to related configuration files.
	// For example, a ksail.yaml might reference kind.yaml or k3d.yaml files.
	GetRelatedFiles() []string

	// GetValidationOptions returns validation-specific options.
	// This might include flags like strict mode, warning levels, etc.
	GetValidationOptions() map[string]interface{}
}
