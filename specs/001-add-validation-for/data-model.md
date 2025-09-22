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

### ConfigurationSchema

Defines validation rules and constraints for configuration fields.

**Fields**:

- `RequiredFields` ([]string): List of fields that must be present
- `FieldTypes` (map[string]reflect.Type): Expected data types for each field
- `FieldConstraints` (map[string]FieldConstraint): Validation constraints per field
- `CrossFieldRules` ([]CrossFieldRule): Dependencies between fields

**Validation Rules**:

- RequiredFields must contain valid field paths
- FieldTypes must specify valid Go types
- FieldConstraints must be well-formed and testable
- CrossFieldRules must not create circular dependencies

### FieldConstraint

Represents validation constraints for individual configuration fields.

**Fields**:

- `Type` (ConstraintType): Type of constraint (enum, range, pattern, etc.)
- `AllowedValues` ([]interface{}): Valid values for enum constraints
- `MinValue` (interface{}): Minimum value for numeric constraints
- `MaxValue` (interface{}): Maximum value for numeric constraints
- `Pattern` (string): Regular expression pattern for string validation
- `CustomValidator` (func(interface{}) error): Custom validation function

**Validation Rules**:

- Type must be valid ConstraintType
- AllowedValues required for enum constraints
- Min/MaxValue required for range constraints
- Pattern must be valid regex for pattern constraints
- CustomValidator must be non-nil for custom constraints

### CrossFieldRule

Represents validation rules that depend on multiple configuration fields.

**Fields**:

- `Name` (string): Descriptive name for the rule
- `SourceField` (string): Primary field that triggers the rule
- `TargetFields` ([]string): Fields that must be validated based on source
- `Condition` (func(map[string]interface{}) bool): Condition function
- `ValidationFunc` (func(map[string]interface{}) []ValidationError): Validation logic

**Validation Rules**:

- Name must be descriptive and unique
- SourceField must exist in configuration schema
- TargetFields must be valid field paths
- Condition function must be deterministic
- ValidationFunc must return structured errors

## Entity Relationships

```
ValidationResult
├── contains []ValidationError
├── references ConfigurationSchema
└── validates against FileLocation

ValidationError
├── references FileLocation
└── describes FieldConstraint violation

ConfigurationSchema
├── contains []FieldConstraint
├── contains []CrossFieldRule
└── validates ConfigurationFile

FieldConstraint
└── validates individual field values

CrossFieldRule
├── references multiple fields
└── generates ValidationError instances
```

## Data Flow

1. **Configuration Loading**: File content parsed into structured data
2. **Schema Application**: ConfigurationSchema applied to parsed data
3. **Field Validation**: Individual FieldConstraints checked against values
4. **Cross-Field Validation**: CrossFieldRules evaluated for dependencies
5. **Error Collection**: ValidationErrors accumulated in ValidationResult
6. **Result Generation**: Final ValidationResult with status and errors returned

## Validation Contexts

### KSail Configuration Context

- Validates ksail.yaml structure and content
- Checks distribution compatibility
- Validates cross-references to other config files
- Ensures cluster naming consistency

### Kind Configuration Context

- Validates kind.yaml against Kind API schema
- Checks node configuration consistency
- Validates networking and port mappings
- Ensures image and version compatibility

### K3d Configuration Context

- Validates k3d.yaml against K3d API schema
- Checks server and agent configurations
- Validates registry and volume mappings
- Ensures network and security settings
