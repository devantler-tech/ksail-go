# Contract: CNI Type Enum Extension

**Package**: `pkg/apis/cluster/v1alpha1`
**Type**: `CNI` (string enum)
**Date**: 2025-11-15

## Type Definition

```go
// CNI defines the CNI options for a KSail cluster.
type CNI string

const (
    // CNIDefault is the default CNI.
    CNIDefault CNI = "Default"
    // CNICilium is the Cilium CNI.
    CNICilium CNI = "Cilium"
    // CNICalico is the Calico CNI.
    CNICalico CNI = "Calico"
    // CNIFlannel is the Flannel CNI. (NEW)
    CNIFlannel CNI = "Flannel"
)
```

## Contract Requirements

### Value Constraints

- **Type**: String-based enum
- **Case-Sensitive**: "Flannel" must match exactly (capital F)
- **Immutable**: Constant values must not change once defined
- **Unique**: No duplicate constant values allowed

### Valid Values

The following values are valid for the CNI type:

| Value     | Description                 | Support Status | Distribution Behavior                                                                        |
| --------- | --------------------------- | -------------- | -------------------------------------------------------------------------------------------- |
| `Default` | Distribution's default CNI  | Existing       | Kind: kindnet, K3d: Flannel                                                                  |
| `Cilium`  | Cilium eBPF-based CNI       | Existing       | Disables default, installs Cilium                                                            |
| `Calico`  | Calico policy-based CNI     | Existing       | Disables default, installs Calico                                                            |
| `Flannel` | Flannel overlay network CNI | **NEW**        | **Kind**: Disables default, installs Flannel; **K3d**: Uses native Flannel (no installation) |

### String Representation

```go
func (c *CNI) String() string {
    return string(*c)
}
```

**Contract**:

- Returns the string value of the CNI enum
- Used for display and serialization
- Must preserve exact case and spelling

### Type Information

```go
func (c *CNI) Type() string {
    return "string"
}
```

**Contract**:

- Returns "string" (for flag/config parsing)
- Indicates this is a string-based type
- Required by flag parsing libraries

## Validation Contract

### Set Method

```go
func (c *CNI) Set(value string) error
```

**Purpose**: Validates and sets the CNI value from a string

**Preconditions**:

- `value` must be a non-empty string
- Pointer receiver `c` must not be nil

**Postconditions (Success)**:

- `*c` is set to the validated CNI value
- Returns `nil`

**Postconditions (Failure)**:

- `*c` is unchanged
- Returns `ErrInvalidCNI` wrapped with context

**Validation Rules**:

1. Value must exactly match one of the defined constants (case-sensitive)
2. No whitespace trimming or normalization
3. Empty string is invalid
4. Unknown values are rejected

**Implementation**:

```go
func (c *CNI) Set(value string) error {
    switch CNI(value) {
    case CNIDefault, CNICilium, CNICalico, CNIFlannel:  // Add CNIFlannel
        *c = CNI(value)
        return nil
    default:
        return fmt.Errorf("%w: %s (valid: %s, %s, %s, %s)",
            ErrInvalidCNI, value,
            CNIDefault, CNICilium, CNICalico, CNIFlannel)
    }
}
```

**Error Format**:

```text
invalid CNI: <value> (valid: Default, Cilium, Calico, Flannel)
```

**Examples**:

```go
var cni v1alpha1.CNI

// Valid
err := cni.Set("Flannel")  // nil, cni == CNIFlannel

// Invalid
err := cni.Set("flannel")   // error: case mismatch
err := cni.Set("FLANNEL")   // error: case mismatch
err := cni.Set("Weave")     // error: unknown CNI
err := cni.Set("")          // error: invalid CNI
```

### validCNIs Function

```go
func validCNIs() []CNI {
    return []CNI{CNIDefault, CNICilium, CNICalico, CNIFlannel}
}
```

**Purpose**: Returns slice of all valid CNI values

**Contract**:

- Returns exhaustive list of all supported CNI options
- Order is not significant
- Used for validation and documentation
- Must include **CNIFlannel** in returned slice

**Usage**:

```go
// Validation
validValues := validCNIs()
isValid := slices.Contains(validValues, userInput)

// Documentation generation
for _, cni := range validCNIs() {
    fmt.Printf("- %s\n", cni)
}
```

## YAML Serialization

### Marshaling

```yaml
apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  cni: Flannel  # String value serialized directly
```

**Contract**:

- Serializes as plain string (not quoted unless needed)
- Case is preserved exactly as defined
- Field name is lowercase `cni`

### Unmarshaling

```go
// Automatically handled by yaml.Unmarshal
var spec ClusterSpec
err := yaml.Unmarshal(data, &spec)
// spec.CNI is CNIFlannel if yaml has "cni: Flannel"
```

**Contract**:

- String value is directly assigned to CNI type
- No validation during unmarshal (happens during Set or explicit validation)
- Case-sensitive matching

## CLI Flag Integration

### Flag Definition

```go
cmd.Flags().Var(&cni, "cni", "CNI to use (Default, Cilium, Calico, Flannel)")
```

**Contract**:

- `--cni Flannel` sets CNI to CNIFlannel
- Tab completion should include "Flannel"
- Help text lists all valid options including Flannel
- Error on invalid value shows all valid options

### Flag Usage Examples

```bash
# Valid
ksail cluster init --cni Flannel
ksail cluster init --cni=Flannel

# Invalid
ksail cluster init --cni flannel   # Error: case mismatch
ksail cluster init --cni FLANNEL   # Error: case mismatch
```

## Testing Contract

### Unit Tests Required

```go
func TestCNI_Set_ValidValues(t *testing.T)
func TestCNI_Set_Flannel(t *testing.T)  // NEW
func TestCNI_Set_InvalidValues(t *testing.T)
func TestCNI_String(t *testing.T)
func TestValidCNIs_IncludesFlannel(t *testing.T)  // NEW
```

### Test Table Structure

```go
func TestCNI_Set(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    CNI
        wantErr bool
    }{
        // Existing tests...
        {
            name:    "valid Flannel",
            input:   "Flannel",
            want:    CNIFlannel,
            wantErr: false,
        },
        {
            name:    "invalid flannel lowercase",
            input:   "flannel",
            want:    "",
            wantErr: true,
        },
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var cni CNI
            err := cni.Set(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "invalid CNI")
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, cni)
            }
        })
    }
}
```

### Snapshot Tests

CLI output snapshots must be updated to include Flannel:

```bash
ksail cluster init --help
# Output should list: Default, Cilium, Calico, Flannel
```

## Backward Compatibility

### Existing Configurations

All existing ksail.yaml files with `cni: Default`, `cni: Cilium`, or `cni: Calico` continue to work unchanged.

### Existing Code

All existing code that switches on CNI values must add Flannel case:

```go
switch spec.CNI {
case v1alpha1.CNIDefault:
    // ...
case v1alpha1.CNICilium:
    // ...
case v1alpha1.CNICalico:
    // ...
case v1alpha1.CNIFlannel:  // NEW case required
    // ...
default:
    return fmt.Errorf("unsupported CNI: %s", spec.CNI)
}
```

### Schema Generation

JSON Schema must be regenerated to include Flannel:

```bash
go run .github/scripts/generate-schema.go
```

**Expected schema.json update**:

```json
{
  "cni": {
    "type": "string",
    "enum": ["Default", "Cilium", "Calico", "Flannel"],
    "description": "CNI to use for cluster networking"
  }
}
```

## Integration Points

### Files Requiring Updates

1. **pkg/apis/cluster/v1alpha1/types.go**
   - Add `CNIFlannel` constant
   - Update `validCNIs()` function
   - Update `Set()` method switch

2. **pkg/apis/cluster/v1alpha1/types_test.go**
   - Add Flannel validation tests

3. **cmd/cluster/create.go**
   - Add Flannel case to installer factory

4. **pkg/io/scaffolder/scaffolder.go**
   - Handle Flannel in CNI configuration logic

5. **schemas/ksail-config.schema.json**
   - Regenerate with Flannel in enum

6. **docs/cni.md** (if exists)
   - Document Flannel usage

## Dependencies

- No new dependencies required
- Uses existing error types (`ErrInvalidCNI`)
- Fully compatible with existing validation framework

## Migration Impact

**Version**: MINOR (v1.X.0 → v1.Y.0)

- **Breaking Changes**: None
- **Deprecations**: None
- **Migration Required**: No
- **Rollback**: Safe (removing Flannel constant doesn't break existing configs using other CNIs)
