# Phase 0: Research - Configuration File Validation

## Research Areas and Decisions

### 1. API Design Simplification (UPDATED)

**Decision**: Generic `Validate[T any](config T) *ValidationResult` method interface
**Rationale**:

- Configurations are already unmarshaled from files during normal KSail operations
- Generic interface provides compile-time type safety for all validation logic
- Eliminates dual validation paths and type erasure complications
- Struct validation is more testable and efficient than byte validation
- In-memory validation avoids unnecessary marshaling/unmarshaling cycles
- Cleaner API for consumers who already have parsed configuration structs
- Better IDE support with autocomplete and type checking

**Alternatives considered**:

- Dual Interface Approach: Keep both `Validate([]byte)` and `ValidateStruct(interface{})`
  - Rejected: Adds complexity without significant benefit, creates confusion
- Byte-only Validation: Only `Validate([]byte)` method
  - Rejected: Forces unnecessary marshaling, less efficient, harder to test

### 2. Validation Architecture Patterns

**Decision**: Independent validator packages with common interfaces
**Rationale**:

- Each configuration type (ksail, kind, k3d) has unique validation rules and schemas
- Independent packages allow for isolated testing and development
- Follows Go package organization best practices
- Enables future extension for additional distributions (EKS, etc.)

**Alternatives considered**:

- Single monolithic validator: Rejected due to complexity and testing difficulties
- Interface-based plugin system: Overkill for current scope, adds unnecessary complexity

### 3. Error Handling and Message Format

**Decision**: Structured ValidationError type with actionable messages
**Rationale**:

- Consistent error format across all validators improves user experience
- Structured errors enable better testing and debugging
- Actionable messages with fix examples reduce user frustration
- Follows constitution's User Experience Consistency principle

**Alternatives considered**:

- Simple string errors: Rejected due to lack of structure and actionability
- Complex error hierarchies: Rejected due to unnecessary complexity for current scope

### 4. In-Memory Validation Strategy

**Decision**: Parse configuration into structs, then validate in memory
**Rationale**:

- Faster than file-based validation operations
- Easier to unit test without file I/O dependencies
- Marshalling errors naturally take precedence
- Meets performance requirements (<100ms)
- Enables snapshot testing of error messages

**Alternatives considered**:

- File-based validation with re-reading: Rejected due to performance and testing concerns
- Streaming validation: Overkill for typical config file sizes

### 4. Validation Timing and Integration Points

**Decision**: Validate during config loading in existing config-manager
**Rationale**:

- Fail-fast approach prevents invalid configurations from causing issues
- Natural integration point in existing codebase
- Consistent validation across all ksail commands
- Minimal changes to existing command structure

**Alternatives considered**:

- Separate validation command: Rejected as it doesn't prevent runtime errors
- Validation on file write: Rejected as it doesn't catch manual file edits

### 5. Testing Strategy

**Decision**: TDD with unit tests per validator + integration tests
**Rationale**:

- Unit tests validate specific validation rules in isolation
- Integration tests validate complete validation workflows
- Snapshot testing ensures consistent error message format
- Mocking enables testing without file dependencies
- Follows constitution's TDD requirement

**Alternatives considered**:

- End-to-end testing only: Rejected due to slow feedback and complexity
- Property-based testing: Deferred as additional enhancement, not core requirement

### 6. Dependencies and External Libraries

**Decision**: Prioritize upstream Go package validators, avoid custom validation that duplicates upstream logic
**Rationale**:

- **Upstream authenticity**: Using official sigs.k8s.io/kind/pkg/apis/config/v1alpha4 ensures Kind validation matches exactly what Kind tool expects
- **Upstream consistency**: Leveraging github.com/k3d-io/k3d/v5/pkg/config ensures K3d validation is identical to K3d tool behavior
- **EKS integration**: Using github.com/weaveworks/eksctl APIs ensures EKS configuration validation matches eksctl behavior
- **AWS SDK integration**: Leveraging AWS SDK Go v2 packages for authentic AWS resource validation
- **Reduced maintenance**: Upstream packages handle edge cases, version compatibility, and schema updates automatically
- **No duplicate logic**: Avoids reinventing validation that already exists in well-tested upstream packages
- Leverages existing yaml parsing capabilities (sigs.k8s.io/yaml)
- Minimizes dependency bloat and security surface
- Follows constitution's dependency control requirements

**Alternatives considered**:

- Custom validation logic: Rejected due to duplication of upstream efforts and potential inconsistency
- JSON Schema validation libraries: Rejected due to additional dependency and YAML focus
- Custom YAML parsing: Rejected due to reinventing existing functionality

### 6.1. Upstream Validator Package Research

**Kind Validator Strategy**:

- Use `sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster` struct for parsing and validation
- Leverage Kind's built-in validation methods where available
- Only add custom validation for KSail-specific requirements not covered by upstream

**K3d Validator Strategy**:

- Use `github.com/k3d-io/k3d/v5/pkg/config/v1alpha5.SimpleConfig` struct for parsing
- Leverage K3d's configuration validation patterns where available
- Ensure validation behavior matches K3d tool expectations

**KSail Validator Strategy**:

- Use existing `github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1.Cluster` - **DO NOT ALTER**
- Focus on KSail-specific validation logic that coordinates with upstream validators

### 7. Validator Independence and Coordination

**Decision**: KSail validator orchestrates, others are independent
**Rationale**:

- KSail validator handles ksail.yaml and coordinates loading other configs when needed
- Kind/K3d validators focus solely on their specific configuration format validation
- Clear separation of concerns and responsibilities
- Enables independent development and testing of each validator

**Alternatives considered**:

- Circular dependency between validators: Rejected due to complexity
- Central orchestrator separate from all validators: Rejected as unnecessary abstraction

## Technical Decisions Summary

| Aspect              | Decision                               | Justification                       |
|---------------------|----------------------------------------|-------------------------------------|
| Architecture        | Independent validator packages         | Separation of concerns, testability |
| Error Handling      | Structured ValidationError type        | Consistency, actionability          |
| Validation Strategy | In-memory after parsing                | Performance, testability            |
| Integration Point   | During config loading                  | Fail-fast, consistency              |
| Testing Approach    | TDD with unit + integration tests      | Constitution compliance             |
| Dependencies        | Use existing project deps              | Minimize bloat, security            |
| Coordination        | KSail orchestrates, others independent | Clear responsibilities              |

All decisions align with constitution requirements for code quality, performance, testing, and user experience.

## Research Status: COMPLETE

All technical decisions have been made and documented. No NEEDS CLARIFICATION items remain. Ready for Phase 1 design.

Key API simplification decision: Single `Validate(config interface{})` method replaces dual-method approach for simpler, more efficient validation.
