# Data Model: KSail Project Scaffolder

## Core Entities

### Scaffolder

The main orchestration entity that coordinates file generation.

**Attributes**:

- KSailConfig: The cluster configuration to scaffold
- KSailYAMLGenerator: Generator for ksail.yaml files
- KindGenerator: Generator for Kind distribution configs
- K3dGenerator: Generator for K3d distribution configs
- EKSGenerator: Generator for EKS distribution configs
- KustomizationGenerator: Generator for kustomization.yaml files

**Responsibilities**:

- Coordinate generation of all project files
- Handle distribution-specific configuration creation
- Manage file output and force overwrite options

### Cluster Configuration (v1alpha1.Cluster)

The input configuration defining the KSail project structure.

**Attributes**:

- Metadata: Name and metadata for the cluster
- Spec: Specification including distribution type, source directory, distribution config filename

**Relationships**:

- Input to Scaffolder
- Determines which distribution generator to use

### Distribution Configurations

#### Kind Configuration (v1alpha4.Cluster)

Minimal Kind cluster configuration for local development.

**Attributes**:

- Name: Cluster name from KSail config
- Nodes: Empty array for minimal config
- Networking: Default networking configuration

#### K3d Configuration (k3dv1alpha5.SimpleConfig)

Minimal K3d cluster configuration.

**Attributes**:

- TypeMeta: K3d API version and kind
- Basic configuration for simple cluster setup

#### EKS Configuration (eksv1alpha5.ClusterConfig)

AWS EKS cluster configuration with sensible defaults.

**Attributes**:

- Metadata: Cluster name and eu-north-1 region
- NodeGroups: Single node group with m5.large instances
- Default EKS settings for development

### Kustomization (ktypes.Kustomization)

Kubernetes kustomization configuration.

**Attributes**:

- TypeMeta: Kustomization API version and kind
- Resources: Empty array for base structure

## Data Flow

```text
1. Input: v1alpha1.Cluster configuration
2. Scaffolder processes configuration
3. Generate ksail.yaml from cluster config
4. Determine distribution type
5. Generate distribution-specific config file
6. Generate kustomization.yaml file
7. Output: Complete project file structure
```

## File Generation Patterns

### Configuration Generation

All generators follow the same pattern:

1. Create minimal configuration object
2. Set required fields from input
3. Use generator with output options
4. Handle errors with proper wrapping

### Error Handling Strategy

Structured error handling with specific error types:

- Distribution-specific generation errors
- Unknown distribution handling
- Not implemented feature errors (Tind)

### Output Structure

Generated files maintain KSail project conventions:

- ksail.yaml: Root project configuration
- {distribution}.yaml: Distribution-specific cluster config
- k8s/kustomization.yaml: Kubernetes resource management

## Validation Rules

### Input Validation

- Cluster configuration must have valid metadata
- Distribution type must be supported
- Output path must be accessible

### Output Validation

- Generated YAML must be syntactically valid
- Files must follow distribution-specific schemas
- Kustomization must include proper resources array
