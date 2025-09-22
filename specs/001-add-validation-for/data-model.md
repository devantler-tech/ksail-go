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

### Validator[T any]

Defines a type-safe interface for configuration file validators.

**Methods**:

- `Validate(config T) *ValidationResult`: Validates typed configuration data and returns validation result

**Type Parameters**:

- `T`: The specific configuration type this validator handles (e.g., `*v1alpha1.Cluster`, `*kindapi.Cluster`)

### ValidatorManager

Provides centralized configuration validation across multiple types with hybrid validator architecture.

**Methods**:

- `ValidateFile(filePath string, data []byte) *ValidationResult`: Validates a configuration file, automatically detecting the type
- `ValidateFiles(files map[string][]byte) map[string]*ValidationResult`: Validates multiple configuration files and aggregates results

**Implementation Details**:

- Uses embedded validator instances for internal file type detection and routing
- Standalone validator packages provide direct access for unit testing and specific use cases
- Eliminates constructor parameters through self-contained embedded design
- Provides automatic file type detection and routing
- Supports multiple validation attempts with best-result selection

### Hybrid Validator Architecture

The system provides both standalone validators and embedded manager implementations:

**Standalone Validators** (for direct use and testing):

- `pkg/validator/ksail.ConfigValidator`: Direct KSail configuration validation
- `pkg/validator/kind.ConfigValidator`: Direct Kind configuration validation  
- `pkg/validator/k3d.ConfigValidator`: Direct K3d configuration validation
- `pkg/validator/eks.ConfigValidator`: Direct EKS configuration validation

**Embedded Validators** (within ValidatorManager):

- `KSailValidator`: Embedded `*v1alpha1.Cluster` validator
- `KindValidator`: Embedded `*kindapi.Cluster` validator  
- `K3dValidator`: Embedded `map[string]any` validator
- `EKSValidator`: Embedded `*v1alpha5.ClusterConfig` validator

## Entity Relationships

```txt
ValidationResult
├── contains []ValidationError
└── validates against FileLocation

ValidationError
├── references FileLocation
└── describes validation failure

Validator[T]
├── validates typed configuration T
└── generates ValidationResult

ValidatorManager
├── embeds KSailValidator, KindValidator, K3dValidator, EKSValidator
├── auto-detects file types from raw bytes
├── routes validation to appropriate embedded Validator
└── provides centralized validation orchestration

Standalone Validators (ConfigValidator)
├── provide direct validation access for testing
├── implement same Validator[T] interface  
└── enable isolated unit testing and development
```

## Data Flow

1. **Configuration Loading**: File content provided as raw []byte data
2. **File Type Detection**: ValidatorManager analyzes content to detect configuration type
3. **Parsing and Validation**: Each embedded validator attempts to parse and validate the data
4. **Best Result Selection**: Manager selects the result with fewest errors (or first valid result)
5. **Error Collection**: ValidationErrors accumulated in ValidationResult with file location context
6. **Result Generation**: Final ValidationResult with status and errors returned

### Alternative Direct Validation Flow

For unit testing and direct validation:

1. **Direct Instantiation**: Create standalone ConfigValidator instance
2. **Type-Safe Validation**: Call Validate() with pre-parsed configuration struct
3. **Immediate Results**: Get ValidationResult without file detection overhead

## Validation Architecture

### Embedded Validator Pattern

The ValidatorManager uses embedded validator instances rather than a registry pattern:

```txt
DefaultValidatorManager {
    ksailValidator: &KSailValidator{}
    kindValidator:  &KindValidator{}
    k3dValidator:   &K3dValidator{}
    eksValidator:   &EKSValidator{}
}
```

This design:

- Eliminates constructor parameter redundancy
- Provides type-safe validation routing
- Simplifies dependency management
- Enables concurrent validation attempts

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
