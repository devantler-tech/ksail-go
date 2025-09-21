# Scaffolder Interface Contract

## Scaffolder.Scaffold Method

### Method Signature

```go
func (s *Scaffolder) Scaffold(output string, force bool) error
```

### Contract

**Preconditions:**

- output path must be a valid directory path
- scaffolder must be properly initialized with valid configuration

**Postconditions:**

- All required project files are generated if no error
- Existing files are overwritten only if force=true
- Directory structure is created as needed
- Returns specific error types for different failure modes

**Error Handling:**

- Returns ErrKSailConfigGeneration if ksail.yaml generation fails
- Returns distribution-specific errors for config generation failures
- Returns ErrTindNotImplemented for Tind distribution
- Returns ErrUnknownDistribution for unsupported distributions

### Test Requirements

**Success Cases:**

- Generate complete project structure for each supported distribution
- Handle force overwrite correctly
- Create directory structure as needed

**Error Cases:**

- Invalid distribution types
- File system permission errors
- Generator failures

## Generator Interface Contract

### Interface Definition

```go
type Generator[T any, O any] interface {
    Generate(input T, options O) (string, error)
}
```

### Interface Contract

**Preconditions:**

- Input must be valid configuration object
- Options must specify valid output parameters

**Postconditions:**

- Returns generated YAML content as string
- Writes to file if output path specified in options
- Generated content is syntactically valid YAML

**Error Handling:**

- Returns wrapped errors with context
- Preserves original error information
- Provides clear failure reasons
