# Data Model - Configuration File Validation

## Core Entities

### ValidationError

Represents a specific validation failure with detailed context and actionable remediation.

**Fields**:

- `Field` (string): The specific field path that failed validation (e.g., "spec.distribution", "metadata.name")
- `Message` (string): Human-readable description of the validation error
- `CurrentValue` (interface{}): The actual value that was found in the configuration
- `ExpectedValue` (interface{}): The expected value or constraint that was violated
- `FixSuggestion` (string): Actionable guidance on how to fix the error
- `Location` (FileLocation): File and line information where the error occurred

**Validation Rules**:

- Field path must be non-empty for specific field errors
- Message must be human-readable and actionable
- FixSuggestion must provide concrete steps to resolve the issue
- Location information must be accurate when available

**State Transitions**:

- Created → when validation rule fails
- Collected → when added to ValidationResult
- Displayed → when presented to user

### ValidationResult

Contains the overall validation status and collection of validation errors.

**Fields**:

- `Valid` (bool): Overall validation status (true if no errors)
- `Errors` ([]ValidationError): Collection of all validation errors found
- `Warnings` ([]ValidationError): Collection of validation warnings (non-blocking)
- `ConfigFile` (string): Path to the configuration file that was validated

**Validation Rules**:

- Valid must be false if any errors exist
- Errors slice can be empty for successful validation
- ConfigFile path must be provided for context
- Warnings do not affect overall validation status

**State Transitions**:

- Initialized → empty result created
- Populated → errors/warnings added during validation
- Finalized → validation complete, result returned

### FileLocation

Provides precise location information for validation errors.

**Fields**:

- `FilePath` (string): Absolute path to the configuration file
- `Line` (int): Line number where the error occurred (1-based)
- `Column` (int): Column number where the error occurred (1-based, optional)

**Validation Rules**:

- FilePath must be absolute path
- Line must be positive integer when specified
- Column must be positive integer when specified

### Validator

Defines the interface for configuration file validators.

**Methods**:

- `Validate(config interface{}) *ValidationResult`: Validates configuration data and returns validation result

**Validation Rules**:

- Must handle both raw []byte data and parsed structs
- Must validate semantic correctness of configuration
- Must check field constraints and dependencies  
- Must return actionable error messages
- Must be thread-safe for concurrent validation

## Entity Relationships

```txt
ValidationResult
├── contains []ValidationError
└── validates against FileLocation

ValidationError
├── references FileLocation
└── describes validation failure

Validator
├── validates ConfigurationFile
└── generates ValidationResult

ValidatorManager
├── registers multiple Validators
└── routes validation to appropriate Validator
```

## Data Flow

1. **Configuration Loading**: File content parsed into structured data
2. **Validator Selection**: ValidatorManager selects appropriate validator for configuration type
3. **Validation Execution**: Validator.Validate() performs semantic validation using upstream APIs
4. **Error Collection**: ValidationErrors accumulated in ValidationResult
5. **Result Generation**: Final ValidationResult with status and errors returned

## Validation Contexts

### KSail Configuration Context

- Validates ksail.yaml structure and content using upstream v1alpha1.Cluster APIs
- Checks distribution compatibility
- Validates cross-references to other config files
- Ensures cluster naming consistency

### Kind Configuration Context

- Validates kind.yaml against Kind API schema using sigs.k8s.io/kind
- Checks node configuration consistency
- Validates networking and port mappings
- Ensures image and version compatibility

### K3d Configuration Context

- Validates k3d.yaml against K3d API schema
- Checks server and agent configurations
- Validates registry and volume mappings
- Ensures network and security settings
