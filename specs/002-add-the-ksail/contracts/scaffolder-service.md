# Scaffolder Service Contract

## Interface Definition

### Primary Method


```go
func (s *Scaffolder) Scaffold(output string, force bool) error
```

## Input Contract


### Parameters

- `output` (string): Target directory path for generated files
- `force` (bool): Whether to overwrite existing files without conflict detection



### Scaffolder Configuration

The scaffolder is initialized with `v1alpha1.Cluster` configuration loaded by ConfigManager in the init command:

```go
// ConfigManager loads cluster configuration with CLI flag integration
cluster, err := configManager.LoadConfig()
if err != nil {
    return fmt.Errorf("failed to load cluster config: %w", err)
}

// Scaffolder is created with the loaded cluster configuration

scaffolderInstance := scaffolder.NewScaffolder(*cluster)
```

**ConfigManager Integration**:

- Uses existing `cmd/config-manager` for loading cluster configuration
- Supports CLI flag binding via Viper: `configManager.Viper.BindPFlag()`
- Provides standard field selectors for distribution, source directory, etc.
- Handles configuration validation before scaffolding begins

### Validation Requirements

- Output path must be accessible directory
- v1alpha1.Cluster configuration must be valid
- Distribution must be supported (Kind, K3d, EKS)
- Invalid input returns validation error before any file operations


## Output Contract

### Success Response

Returns `error = nil` on successful scaffolding. Generated files include:

- ksail.yaml (KSail cluster configuration)

- Distribution-specific config (kind.yaml, k3d.yaml, or eks.yaml)
- k8s/kustomization.yaml (base Kustomize structure)

### Error Response

Returns wrapped errors with:

- Error type classification (validation, filesystem, generation)

- Human-readable message with context
- Actionable remediation suggestions
- No partial state left behind (existing force behavior)

## File Generation Contract

### Template Processing

**Note**: Templates are generated at runtime by pkg/io/generator system, not loaded from embedded files.

1. **Runtime Generation**: Templates created dynamically by distribution-specific generators
2. **Variable Substitution**: Uses v1alpha1.Cluster configuration values:
   - Cluster metadata (name, labels)
   - Distribution-specific configuration
   - Source directory paths

   - Generated timestamps
3. **Validation**: Each generator validates output before writing:
   - YAML syntax validation
   - Distribution-specific config validation
   - File path safety checks
4. **Atomic Writing**: Files written atomically to prevent partial states



### File System Operations

- Create directories as needed with appropriate permissions
- Write files atomically (temp file + rename)
- Verify written content matches generated content
- Set appropriate file permissions (644 for configs, 755 for directories)

### Rollback Behavior


On any error during generation:

- Remove any partially created files
- Remove any created directories if empty
- Restore filesystem to pre-operation state
- Return clear error describing failure point

## Template Contract

### Required Generators


**Note**: Uses runtime generators from pkg/io/generator, not template files.

The scaffolder provides generators for all distributions:

1. **KSailYAMLGenerator** - Generates ksail.yaml from v1alpha1.Cluster
2. **KindGenerator** - Generates kind.yaml from Kind configuration
3. **K3dGenerator** - Generates k3d.yaml from K3d configuration

4. **EKSGenerator** - Generates EKS configs from EKS configuration
5. **KustomizationGenerator** - Generates k8s/kustomization.yaml structure

### Generator Implementation

Runtime generation characteristics:

- **Configuration-Driven**: Uses v1alpha1.Cluster fields for all content generation

- **Type-Safe**: Strongly-typed generator interfaces prevent runtime errors
- **Conditional**: Only generates files for selected distribution
- **Validated**: Each generator validates output syntax before returning

### Output Format


- **Valid YAML**: All generated content must parse correctly
- **Tool Compatibility**: Generated configs work with target distributions
- **Consistent Style**: Standard formatting across all generators
- **Documented**: Generated files include explanatory comments where helpful


## Performance Contract

### Time Limits

- Template loading: <100ms
- Template processing: <500ms per template
- File writing: <1 second total

- Total operation: <5 seconds

### Memory Usage

- Maximum 50MB during operation (constitutional requirement)
- Release resources after completion

- No memory leaks in template processing

### Concurrency

- Thread-safe template processing
- Safe for concurrent use by multiple commands
- No shared mutable state between operations


## Integration Contract

### Dependencies

- Must use existing `v1alpha1.Cluster` APIs
- Must integrate with existing input validation
- Must respect existing file system abstractions
- Must support existing error handling patterns


### Observability

- Log all file operations at appropriate levels
- Emit metrics for operation timing
- Provide progress callbacks for UI updates
- Include operation context in all log messages

## Error Scenarios


### Template Errors

```go
type TemplateError struct {
    Template string
    Line     int
    Message  string
}
```

### Filesystem Errors


```go
type FilesystemError struct {
    Operation string
    Path      string
    Cause     error
}

```

### Validation Errors

```go
type ValidationError struct {
    Field   string
    Value   string
    Rule    string
    Message string
}
```

## Testing Contract

### Unit Test Requirements

- Test each template individually
- Mock filesystem operations for error scenarios
- Validate generated content against schemas
- Test rollback behavior on failures

### Integration Test Requirements

- End-to-end scaffolding in temporary directories
- Real filesystem operations with cleanup
- Cross-platform compatibility testing
- Performance benchmarking for all scenarios
