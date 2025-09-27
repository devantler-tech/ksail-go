# Data Model: KSail Init Command

## ScaffoldTemplate

Represents a runtime-generated template for project file generation.

**Attributes**:

- `Name` (string): Template identifier (e.g., "ksail.yaml", "kind.yaml")
- `Generator` (func): Runtime generation function for template content
- `TargetPath` (string): Relative path where file should be created
- `Required` (boolean): Whether template is mandatory for all projects
- `DistributionSpecific` (boolean): Whether template applies only to specific distribution

**Validation Rules**:

- Generator must produce valid template content with proper placeholder substitution
- TargetPath must be relative and safe (no directory traversal)
- Required templates cannot be skipped
- Distribution-specific templates only generated when matching distribution selected

## v1alpha1.Cluster (Existing)

**Note**: KSail init uses the existing v1alpha1.Cluster configuration structure, not a separate InitConfig.

The scaffolder is initialized with a complete v1alpha1.Cluster configuration loaded via ConfigManager:

**Key Attributes Used**:

- `Metadata.Name` (string): Project/cluster name
- `Spec.Distribution` (v1alpha1.Distribution): Kubernetes distribution (Kind, K3d, EKS)
- `Spec.SourceDirectory` (string): Directory for generated files
- `Spec.DistributionConfig` (string): Generated config filename

**Validation**:

- Uses existing v1alpha1.Cluster validation rules
- Distribution must be supported by scaffolder generators
- All validation handled by ConfigManager before scaffolding

**State Management**:

1. ConfigManager loads cluster configuration with defaults
2. Scaffolder.Scaffold() generates files based on configuration
3. No separate init-specific configuration model needed

**Relationships**:

- Scaffolder uses one Generator per distribution type
- Generates distribution-specific config files

## ProjectFile

Represents a generated configuration file in the new project.

**Attributes**:

- `Path` (string): Absolute file path where content was written
- `Content` (string): Final generated content after template processing
- `Size` (int64): File size in bytes
- `CreatedAt` (time.Time): When file was created
- `Checksum` (string): SHA256 hash for integrity verification

**Validation Rules**:

- Path must be within target directory (security constraint)
- Content must be valid YAML for configuration files
- Size must be reasonable (<1MB per file, <10MB total project)
- Checksum must match generated content

**State Transitions**:

1. Planned during template processing
2. Written to filesystem
3. Verified for integrity
4. Reported to user as created

**Relationships**:

- Generated from one ScaffoldTemplate
- Belongs to one ProjectStructure

## ProjectStructure

Represents the complete directory and file structure of a generated project.

**Attributes**:

- `RootPath` (string): Absolute path to project root directory
- `Files` ([]ProjectFile): List of all generated files
- `Directories` ([]string): List of created directories
- `TotalSize` (int64): Total size of all generated files
- `CreatedAt` (time.Time): When project was initialized

**Validation Rules**:

- RootPath must exist and be writable
- Must contain at least required files (ksail.yaml, distribution config)
- TotalSize must be within reasonable limits (<10MB)
- All directories must be within root path

**Relationships**:

- Contains multiple ProjectFiles
- Created by one InitConfig
- Has one primary distribution configuration

## Data Flow

```txt
InitConfig (user input + defaults)
    ↓ (validation)
ScaffoldTemplate[] (filtered by distribution)
    ↓ (template processing)
ProjectFile[] (generated content)
    ↓ (filesystem operations)
ProjectStructure (complete project)
```

## Constraints and Assumptions

### Scale Assumptions

- Maximum 20 files per project
- Maximum 1MB per individual file
- Maximum 10MB total project size
- Maximum 100 projects per day per user (no hard limit enforced)

### Performance Constraints

- Template processing: <100ms for all templates
- File generation: <2 seconds for all files
- Total operation: <5 seconds end-to-end
- Memory usage: <10MB during operation

### Security Constraints

- All file paths validated against directory traversal attacks
- Template content sanitized to prevent code injection
- File permissions set to user-only for sensitive files
- No external network access during operation

### Compatibility Constraints

- Generated files compatible with existing KSail command suite
- YAML format compliant with Kubernetes standards
- Directory structure follows established conventions
- Configuration schema versioned for backward compatibility
