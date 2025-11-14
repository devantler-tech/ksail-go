# KSail-Go

KSail is a Go-based CLI tool for managing Kubernetes clusters and workloads. It provides declarative cluster provisioning, workload management, and lifecycle operations for Kind, K3d, and EKS distributions.

**ALWAYS reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Core Software Engineering Principles

When working on KSail-Go, always adhere to these fundamental software engineering principles:

### KISS (Keep It Simple, Stupid)

**Simplicity over complexity**. Prefer straightforward solutions that are easy to understand and maintain. Avoid over-engineering or adding unnecessary abstraction layers. If a simple approach solves the problem effectively, use it.

### DRY (Don't Repeat Yourself)

**Eliminate duplication**. Extract common logic into reusable functions, packages, or interfaces. Every piece of knowledge should have a single, unambiguous representation in the codebase. Use Go's interface-based design to share behavior without duplicating code.

### YAGNI (You Aren't Gonna Need It)

**Implement only what's needed now**. Don't add functionality based on speculative future requirements. Focus on current, well-defined requirements. Additional features can be added later when they're actually needed.

### TDA (Tell, Don't Ask)

**Objects should do, not expose**. Instead of querying an object for its state and making decisions externally, tell the object what to do and let it manage its own state. This encapsulates behavior and maintains proper abstraction boundaries.

### SOLID Principles

**S - Single Responsibility Principle**: Each package, interface, and function should have one reason to change. Separate concerns clearly.

**O - Open/Closed Principle**: Code should be open for extension but closed for modification. Use interfaces to allow behavior extension without changing existing code.

**L - Liskov Substitution Principle**: Implementations of an interface should be interchangeable without breaking functionality. Ensure all implementations honor their interface contracts.

**I - Interface Segregation Principle**: Prefer small, focused interfaces over large, general-purpose ones. Clients shouldn't depend on methods they don't use.

**D - Dependency Inversion Principle**: Depend on abstractions (interfaces), not concrete implementations. High-level modules shouldn't depend on low-level modules.

These principles complement the constitutional requirements in `.specify/memory/constitution.md` and guide day-to-day coding decisions.

## Code Smells to Avoid

Code smells are patterns in code that indicate potential problems and suggest a need for refactoring. While not bugs themselves, they make code harder to understand, maintain, and extend. This section is inspired by [refactoring.guru](https://refactoring.guru/refactoring/smells) and adapted for Go and KSail-Go.

When reviewing or writing code, actively watch for these smells and refactor them when discovered:

### Bloaters

Code that has grown too large and unwieldy:

**Long Method/Function**: Functions that do too much, making them hard to understand and test.

- ❌ Avoid: Functions exceeding 50-100 lines or doing multiple unrelated tasks
- ✅ Prefer: Extract smaller, focused functions with clear single purposes
- Example: Split a 200-line cluster creation function into smaller steps like `validateConfig()`, `provisionNodes()`, `installComponents()`

**Large Struct**: Structs with too many fields or methods, indicating they have too many responsibilities.

- ❌ Avoid: Structs managing unrelated concerns (e.g., a Config struct handling both parsing and validation and persistence)
- ✅ Prefer: Split into focused types (e.g., `ConfigParser`, `ConfigValidator`, `ConfigStorage`)
- Example: Break apart a monolithic `ClusterManager` into `ClusterProvisioner`, `ClusterLifecycle`, `ClusterStatus`

**Primitive Obsession**: Overusing basic types instead of creating meaningful domain types.

- ❌ Avoid: Using `string` for everything (cluster names, distributions, statuses)
- ✅ Prefer: Create type aliases or structs: `type ClusterName string`, `type Distribution string`
- Example: Instead of `func Create(name string, dist string)`, use `func Create(name ClusterName, dist Distribution)`

**Long Parameter List**: Functions taking too many parameters, making them hard to call and understand.

- ❌ Avoid: Functions with 4+ parameters
- ✅ Prefer: Introduce parameter objects or options pattern
- Example: Instead of `Create(name, dist, version, nodes, memory, cpu, registry)`, use `Create(config ClusterConfig)`

**Data Clumps**: Groups of variables that always appear together but aren't encapsulated.

- ❌ Avoid: Repeatedly passing `kubeconfig, context, namespace` to every function
- ✅ Prefer: Create a `ClusterConnection` struct encapsulating these related fields
- Example: Group `registry, username, password, insecure` into a `RegistryConfig` struct

### Object-Orientation Abusers

Incorrect application of object-oriented principles (relevant to Go's composition model):

**Switch/Type Assertions on Types**: Using type switches instead of interfaces.

- ❌ Avoid: `switch v := provisioner.(type) { case *KindProvisioner: ..., case *K3dProvisioner: ... }`
- ✅ Prefer: Define common interface methods that each type implements
- Example: All provisioners implement `Provisioner` interface with `Create()`, `Delete()`, etc.

**Temporary Field**: Struct fields that are only used in certain circumstances.

- ❌ Avoid: Fields like `tempResult` that are only set during specific operations
- ✅ Prefer: Use local variables or return values instead
- Example: Pass intermediate results as return values rather than storing in struct fields

**Refused Bequest**: Embedding types but not using their methods.

- ❌ Avoid: Embedding a type just to satisfy an interface without using the embedded behavior
- ✅ Prefer: Only embed types when you actually want their behavior; otherwise implement explicitly
- Example: Don't embed `BaseProvisioner` if you override all its methods anyway

**Alternative Structs with Different Interfaces**: Types doing similar things with inconsistent method names.

- ❌ Avoid: `KindProvisioner.CreateCluster()` vs `K3dProvisioner.Create()` vs `EKSProvisioner.Provision()`
- ✅ Prefer: Consistent interface across all implementations
- Example: All provisioners implement same `Provisioner` interface with uniform method names

### Change Preventers

Code that makes changes difficult:

**Divergent Change**: One type is frequently modified for different reasons.

- ❌ Avoid: A single `ConfigManager` handling parsing, validation, persistence, and migration
- ✅ Prefer: Split responsibilities so each type has one reason to change
- Example: Separate `ConfigReader`, `ConfigValidator`, `ConfigWriter`, `ConfigMigrator`

**Shotgun Surgery**: A single change requires modifications across many files/packages.

- ❌ Avoid: Changing how clusters are named requiring updates in 10 different places
- ✅ Prefer: Centralize related logic; use interfaces to isolate changes
- Example: Define cluster naming logic in one place (`pkg/naming`) and import it everywhere

**Parallel Inheritance Hierarchies**: When adding a new type requires adding corresponding types elsewhere.

- ❌ Avoid: Adding `NewDistribution` requires adding `NewDistributionProvisioner`, `NewDistributionConfig`, `NewDistributionValidator`
- ✅ Prefer: Use composition and interfaces to reduce the need for parallel hierarchies
- Example: Use common `DistributionConfig` interface that all distributions implement

### Dispensables

Code that adds no value:

**Comments Explaining What Code Does**: Comments that restate what the code already says clearly.

- ❌ Avoid: `// Create cluster` above `func CreateCluster()` or `// Loop through nodes` above `for _, node := range nodes`
- ✅ Prefer: Self-documenting code; comments explain *why*, not *what*
- Example: Comment intent or non-obvious decisions: `// Skip validation in dry-run mode to show what would happen`

**Duplicate Code**: Identical or very similar code in multiple places.

- ❌ Avoid: Copy-pasting the same validation logic into multiple command handlers
- ✅ Prefer: Extract shared logic into reusable functions or packages
- Example: Create `pkg/validator` package with common validation functions used across commands

**Lazy Package**: Packages that do too little and don't justify their existence.

- ❌ Avoid: A package with only one small function
- ✅ Prefer: Merge into a related package or ensure sufficient functionality to warrant separation
- Example: Don't create `pkg/stringutil` for one `TrimSpace` wrapper; put it in a more substantial utility package

**Data Struct**: Types with only exported fields and no behavior.

- ❌ Avoid: Structs that are just data containers with no validation or behavior
- ✅ Prefer: Add validation methods, constructors, or move behavior closer to data
- Example: Add `Validate()` method to config structs rather than external validation functions

**Dead Code**: Code that is never executed.

- ❌ Avoid: Unused functions, parameters, or packages
- ✅ Prefer: Remove it; version control preserves history if needed
- Example: Use `golangci-lint` with `unused` linter to catch dead code

**Speculative Generality**: Code designed for hypothetical future needs that never materialize.

- ❌ Avoid: Complex abstraction layers "in case we need to support X someday"
- ✅ Prefer: Follow YAGNI; add abstractions when actually needed
- Example: Don't create plugin system if there are no plugins; add it when the second provider arrives

### Couplers

Problems with how code depends on other code:

**Feature Envy**: A method accessing another type's data more than its own.

- ❌ Avoid: Methods that reach deep into other types: `cluster.provisioner.config.nodes[0].name`
- ✅ Prefer: Move method closer to the data it uses, or ask the other type to do the work
- Example: Instead of `ValidateCluster(cluster)` accessing all cluster internals, add `cluster.Validate()`

**Inappropriate Intimacy**: Types knowing too much about each other's internal details.

- ❌ Avoid: One package directly accessing unexported fields of another package's types
- ✅ Prefer: Use interfaces and public methods; hide implementation details
- Example: Use getter methods rather than exposing internal fields directly

**Message Chains**: Long chains of calls across objects.

- ❌ Avoid: `app.config.cluster.provisioner.client.connection.endpoint`
- ✅ Prefer: Add delegating methods or pass needed data directly
- Example: Add `GetEndpoint()` method at appropriate level to avoid the chain

**Middle Man**: A type that just delegates all work to another type.

- ❌ Avoid: Wrapper types that add no value: `func (w *Wrapper) Create() { w.real.Create() }`
- ✅ Prefer: Remove the middle man and use the actual type directly
- Example: Don't wrap provisioners unless the wrapper adds validation, logging, or other value

**Incomplete Library**: External dependencies missing needed functionality, forcing workarounds.

- ❌ Avoid: Scattered workarounds for library limitations throughout codebase
- ✅ Prefer: Create a focused adapter/wrapper to centralize workarounds
- Example: If an external k8s client lacks retry logic, create `pkg/k8s/client` wrapper with retry

### When to Refactor vs. Accept the Smell

Not all code smells require immediate action:

- **Refactor immediately**: Code smells in critical paths, frequently changed code, or when adding new features nearby
- **Accept temporarily**: Code that rarely changes, is well-tested, and refactoring provides little benefit
- **Document if accepting**: Add a comment explaining why the smell is acceptable for now
- **Track for later**: File an issue to address technical debt when time permits

Remember: The goal is maintainable, understandable code. Use these smells as guidelines, not absolute rules. Context matters.

## Design Patterns

Design patterns are proven solutions to common software design problems. This section is inspired by [refactoring.guru](https://refactoring.guru/design-patterns/catalog) and adapted for Go and KSail-Go. Apply patterns judiciously—only when they solve a real problem, not for the sake of using patterns.

### Creational Patterns

Patterns for object creation mechanisms:

**Factory Method**: Define an interface for creating objects, but let implementations decide which type to instantiate.

- Use when: Different provisioner types (Kind, K3d, EKS) need to be created based on configuration
- Example: `NewProvisioner(distribution string) (Provisioner, error)` returns the appropriate provisioner implementation
- Go idiom: Use constructor functions returning interfaces

**Builder**: Construct complex objects step by step, allowing different representations.

- Use when: Creating complex configurations with many optional parameters
- Example: `ClusterConfig` with fluent builder methods: `NewClusterConfig().WithNodes(3).WithRegistry(reg).Build()`
- Go idiom: Use functional options pattern: `NewCluster(name string, opts ...Option)`

**Singleton**: Ensure a type has only one instance and provide global access to it.

- Use when: Managing shared resources like configuration managers or client connections
- Example: Kubernetes client instance shared across operations
- Go idiom: Use `sync.Once` for thread-safe initialization: `var (instance *Client; once sync.Once)`

**Prototype**: Clone existing objects without making code dependent on their concrete types.

- Use when: Creating variations of cluster configurations
- Example: Copying base cluster config and modifying specific fields
- Go idiom: Implement `Clone() *Config` methods or use struct copying with modifications

**Abstract Factory**: Produce families of related objects without specifying concrete types.

- Use when: Creating sets of related components (provisioner + validator + installer) per distribution
- Example: `DistributionFactory` that creates all components for a specific distribution
- Go idiom: Return multiple interfaces from factory methods

### Structural Patterns

Patterns for assembling objects and types into larger structures:

**Adapter**: Convert the interface of a type into another interface clients expect.

- Use when: Wrapping external libraries (Kind, K3d, eksctl) with unified interfaces
- Example: Adapting different cluster API clients to a common `ClusterClient` interface
- Go idiom: Define target interface and implement it with adapter structs

**Decorator**: Attach additional responsibilities to objects dynamically.

- Use when: Adding functionality like logging, metrics, or retries to provisioners
- Example: `LoggingProvisioner` wrapping actual provisioner to add operation logging
- Go idiom: Wrap interfaces with structs implementing the same interface

**Facade**: Provide simplified interface to complex subsystem.

- Use when: Simplifying complex workflows like cluster creation with multiple steps
- Example: `ClusterManager` facade hiding provisioner, validator, and installer complexity
- Go idiom: Create high-level packages that orchestrate lower-level ones

**Proxy**: Provide surrogate or placeholder to control access to an object.

- Use when: Adding access control, lazy initialization, or caching to clients
- Example: Caching proxy for expensive Kubernetes API calls
- Go idiom: Implement same interface as target, delegate with additional logic

**Composite**: Compose objects into tree structures to represent hierarchies.

- Use when: Managing hierarchical workload structures or nested configurations
- Example: Workload groups containing individual workloads, all implementing `Workload` interface
- Go idiom: Define common interface, implement for both leaf and composite types

**Bridge**: Separate abstraction from implementation so they can vary independently.

- Use when: Supporting multiple dimensions of variation (e.g., distributions × environments)
- Example: Abstract cluster operations from specific provisioner implementations
- Go idiom: Use interfaces for both abstraction and implementation sides

**Flyweight**: Share common state among many objects to reduce memory usage.

- Use when: Managing many similar objects with shared immutable data
- Example: Sharing read-only configuration templates across cluster instances
- Go idiom: Use shared structs with pointers for unique state

### Behavioral Patterns

Patterns for algorithms and responsibility assignment:

**Strategy**: Define family of algorithms, encapsulate each, and make them interchangeable.

- Use when: Different validation strategies for different distributions
- Example: `ValidationStrategy` interface with `KindValidator`, `K3dValidator` implementations
- Go idiom: Accept strategy interfaces as function/method parameters

**Command**: Encapsulate request as an object, enabling parameterization and queuing.

- Use when: Implementing undo/redo, queuing operations, or CLI command pattern
- Example: CLI commands as structs implementing `Execute()` method
- Go idiom: Use Cobra's Command pattern, or define `type Command interface { Execute() error }`

**Observer**: Define one-to-many dependency so when one object changes state, dependents are notified.

- Use when: Notifying UI or logging systems about provisioning progress
- Example: Progress observers receiving updates during cluster creation
- Go idiom: Use channels for event streams: `progressChan chan ProgressEvent`

**Template Method**: Define skeleton of algorithm, letting concrete types or embedded structs override specific steps.

- Use when: Common workflow with distribution-specific steps
- Example: Base cluster creation flow with override points for distribution differences
- Go idiom: Use embedding and method overriding, or function parameters for variations

**State**: Allow object to alter behavior when internal state changes.

- Use when: Managing cluster lifecycle states (Creating, Running, Stopped, Deleted)
- Example: `ClusterState` interface with different behavior per state
- Go idiom: Use state structs implementing common interface, switch on current state

**Chain of Responsibility**: Pass request along chain of handlers until one handles it.

- Use when: Processing validation rules where each validator handles specific checks
- Example: Chain of validators, each checking different aspects of configuration
- Go idiom: Use slice of validators or linked handler structs

**Iterator**: Access elements of collection sequentially without exposing representation.

- Use when: Iterating over workloads, nodes, or configuration items
- Example: `WorkloadIterator` for traversing workload collections
- Go idiom: Use Go's native range over slices/maps, or implement `Next()` pattern for complex cases

**Mediator**: Define object that encapsulates how set of objects interact.

- Use when: Coordinating complex interactions between provisioner, validator, and installer
- Example: `ClusterOrchestrator` mediating between components
- Go idiom: Create coordinator struct that owns and coordinates components

**Memento**: Capture and externalize object's internal state for later restoration.

- Use when: Implementing backup/restore or rollback functionality for configurations
- Example: Saving cluster state before modifications for potential rollback
- Go idiom: Export state to structs, use JSON serialization for persistence

**Visitor**: Represent operation to be performed on elements of an object structure.

- Use when: Performing operations on heterogeneous collections (e.g., different workload types)
- Example: Visiting each workload in a collection to apply transformations
- Go idiom: Define `Accept(Visitor)` method on elements, visitor implements operation for each type

### Applying Patterns in Go

**Go-Specific Considerations**:

- **Favor composition over inheritance**: Go doesn't have inheritance; use embedding and interfaces
- **Interfaces are implicit**: Types satisfy interfaces automatically without declaration
- **Keep interfaces small**: Go prefers many small interfaces over large ones
- **Use functional options**: For flexible constructors instead of Builder pattern overuse
- **Leverage goroutines and channels**: For Observer, Pipeline, and concurrent patterns
- **Avoid over-abstraction**: Apply YAGNI—implement patterns only when complexity justifies them

**When to Use Patterns**:

- **Use**: When pattern solves a current, concrete problem you face
- **Use**: When pattern improves code clarity and maintainability
- **Avoid**: Using patterns "just because" or for speculative future needs
- **Avoid**: Forcing patterns where simple solutions work better

**KSail-Go Pattern Usage**:

- **Command**: CLI command structure (Cobra framework)
- **Strategy**: Multiple provisioner implementations for different distributions
- **Factory Method**: Creating provisioners based on distribution type
- **Adapter**: Wrapping external tools (Kind, K3d, eksctl) with unified interfaces
- **Decorator**: Adding logging, metrics, and retry logic to core operations
- **Facade**: Simplified high-level operations hiding complex multi-step workflows

Remember: Patterns are tools, not goals. Focus on solving problems clearly and maintainably first, then recognize where patterns emerge naturally.

## Task Suitability for GitHub Copilot

### ✅ Tasks Well-Suited for Copilot

Copilot excels at focused, well-defined tasks such as:

- **Bug fixes**: Addressing specific, reproducible bugs with clear acceptance criteria
- **Test improvements**: Adding unit tests, improving test coverage, fixing flaky tests
- **Documentation updates**: Updating README, API docs, code comments, or contribution guidelines
- **Code refactoring**: Improving code structure, removing duplication, optimizing performance
- **Dependency updates**: Updating Go modules, addressing security vulnerabilities
- **CLI enhancements**: Adding new commands, flags, or improving command output
- **Technical debt**: Addressing linting issues, improving error handling, cleaning up deprecated code

### ❌ Tasks Better Handled by Humans

Reserve these tasks for human developers:

- **Architecture decisions**: Major design changes, new subsystem designs, API redesigns
- **Complex integrations**: Deep cross-system changes requiring domain expertise
- **Security-critical changes**: Authentication, authorization, encryption implementations
- **Production incidents**: Critical bug fixes in production requiring deep understanding
- **Business logic**: Changes requiring business domain knowledge or stakeholder input

### 📝 Writing Issues for Copilot

When creating issues to assign to Copilot:

- **Be specific**: Clearly describe the problem and expected outcome
- **Include context**: Reference related files, functions, or documentation
- **Define acceptance criteria**: Specify tests that should pass, expected behavior
- **Provide examples**: Include code snippets, error messages, or expected output
- **Limit scope**: Keep issues focused on a single, well-defined change

### 💬 Providing Feedback to Copilot

When reviewing Copilot's pull requests:

- **Use PR comments**: Tag @copilot in comments on specific lines or files
- **Be specific**: Clearly describe what needs to change and why
- **Iterate**: Copilot will update the PR based on your feedback
- **Approve when ready**: Merge the PR once all feedback is addressed

## **CRITICAL: Always Use Serena First (#serena MCP server)**

**For ALL analysis, investigation, and code understanding tasks, use Serena semantic tools:**

### **Standard Serena Workflow**

1. **Start with Serena memories**: Use Serena to list memories and read relevant ones for context #serena
2. **Use semantic analysis**: Use Serena to find [symbols/functions/patterns] related to [issue] #serena
3. **Get symbol-level insights**: Use Serena to analyze [specific function] and show all referencing symbols #serena
4. **Create new memories**: Use Serena to write a memory about [findings] for future reference #serena

### **Serena-First Examples**

1. Instead of: "Search the codebase for database queries"
   Use: "Use Serena to find all database query functions and analyze their performance patterns #serena"

2. Instead of: "Find all admin functions"
   Use: "Use Serena to get symbols overview of admin files and find capability-checking functions #serena"

3. Instead of: "How do the three systems integrate?"
   Use: "Use Serena to read the system-integration-map memory and show cross-system dependencies #serena"

## Working Effectively

### Prerequisites

Install these exact tools before starting development:

- Go (v1.23.9+) - `go version` should show go1.23.9 or higher
- mockery (v3.x) - for generating test mocks
- golangci-lint - for code linting
- mega-linter-runner - for comprehensive validation
- Docker - required for cluster provisioning and system tests

### Bootstrap and Build Commands

Run these commands in sequence to set up the development environment:

```bash
# Download dependencies (very fast if cached)
go mod download

# Build all packages - takes ~2 seconds
go build ./...

# Build the main CLI binary - takes ~1.4 seconds
go build -o ksail .

# Generate mocks - takes ~1.2 seconds
mockery

# Run unit tests - takes ~37 seconds. NEVER CANCEL - set timeout to 60+ seconds
go test ./...

# Run linter - takes ~1m16s. NEVER CANCEL - set timeout to 90+ seconds
golangci-lint run
```

### Mega-Linter (Comprehensive Validation)

```bash
# Run comprehensive linting with go flavor - takes 5+ minutes
# NEVER CANCEL: This is thorough validation. Set timeout to 10+ minutes
mega-linter-runner -f go
```

## Validation

### Always Run Before Committing

Execute these commands before any commit to ensure CI will pass:

```bash
# Essential pre-commit validation (run all of these):
mockery                    # Generate fresh mocks
go test ./...             # Run all tests (~37s)
golangci-lint run         # Lint code (~1m16s)
go build -o ksail .       # Ensure clean build
```

### Manual Testing Scenarios

**ALWAYS test actual CLI functionality after making changes by running these scenarios:**

#### Basic CLI Validation

```bash
# Test CLI help and version
./ksail --help
./ksail --version

# Test all command help outputs
./ksail cluster init --help
./ksail up --help
./ksail down --help
./ksail status --help
```

#### Complete Cluster Lifecycle Test

Run this complete scenario in a temporary directory to validate changes:

```bash
# Create test directory and navigate to it
mkdir -p /tmp/ksail-test && cd /tmp/ksail-test

# Test Kind distribution (most common)
./ksail cluster init --distribution Kind
./ksail up
./ksail status
./ksail list
./ksail start
./ksail reconcile
./ksail down

# Clean up test files
rm -rf k8s kind.yaml ksail.yaml
```

#### Alternative Distribution Testing

```bash
# Test K3d distribution
./ksail cluster init --distribution K3d

# Test EKS distribution
./ksail cluster init --distribution EKS
```

### System Tests

The CI runs comprehensive system tests that validate:

- `init --distribution Kind`
- `init --distribution K3d`
- `init --distribution EKS`

Each runs the complete lifecycle: init → create → info → list → start → stop → delete

## Project Structure and Navigation

### Repository Layout

```txt
/home/runner/work/ksail-go/ksail-go/
├── cmd/                    # CLI commands using Cobra framework
│   ├── *.go               # Command implementations (init.go, root.go, etc.)
│   ├── cipher/            # Cipher command implementations
│   ├── cluster/           # Cluster command implementations (create.go, delete.go, etc.)
│   ├── workload/          # Workload command implementations
│   └── internal/          # Command helper utilities
├── pkg/                   # Core business logic packages
│   ├── apis/              # Kubernetes API definitions
│   ├── config-manager/    # Configuration management
│   ├── installer/         # Component installation utilities
│   ├── io/                # Safe file I/O operations
│   ├── provisioner/       # Cluster provisioning and lifecycle
│   ├── scaffolder/        # Project scaffolding
│   └── validator/         # Validation utilities
├── internal/              # Internal utility packages
│   └── utils/             # Common utilities (k8s, path)
├── main.go               # Application entry point
├── go.mod                # Go module definition
├── .github/workflows/    # CI/CD pipeline definitions
└── scripts/              # Build and development scripts
```

### Key Files to Review When Making Changes

- **Command changes**: Always check corresponding test files (`*_test.go`)
- **API changes**: Review `pkg/apis/cluster/v1alpha1/` definitions
- **Configuration changes**: Check `pkg/io/config-manager/` implementations
- **Provisioner changes**: Validate against `pkg/provisioner/cluster/` providers
- **Build/CI changes**: Review `.github/workflows/ci.yaml`

### Common Development Locations

- **Adding new CLI commands**: `cmd/*.go` + corresponding tests
- **Cluster provider logic**: `pkg/provisioner/cluster/{kind,k3d,eks}/`
- **Configuration handling**: `pkg/io/config-manager/`
- **File generation**: `pkg/io/generator/`
- **Test utilities**: `internal/testutils/` and package-specific `testutils/`

## Timing Expectations and Timeouts

### Command Timing Reference (based on validation)

- `go mod download`: ~0.045s (when cached)
- `go build ./...`: ~2.1s
- `go build -o ksail .`: ~1.4s
- `go test ./...`: ~37s (full test suite)
- `golangci-lint run`: ~1m16s
- `mockery`: ~1.2s
- `mega-linter-runner -f go`: 5+ minutes

### Recommended Timeout Settings

> [!CAUTION]
> CRITICAL: NEVER CANCEL these operations prematurely

- Build commands: 60+ seconds timeout
- Test commands: 90+ seconds timeout
- Linter commands: 120+ seconds timeout
- Mega-linter: 600+ seconds (10+ minutes) timeout

## CI Workflow Information

### GitHub Actions Pipeline

The CI pipeline (`.github/workflows/ci.yaml`) runs:

1. **Standard Go CI**: Build, test, lint using reusable workflows
2. **System Tests**: Matrix testing across Kind, K3d, and EKS distributions
3. **Full lifecycle validation**: Each distribution tested through complete workflow

### Pre-commit Hooks

Pre-commit hooks automatically run:

- `golangci-lint-fix`: Runs `golangci-lint run --fix` to automatically fix linting issues via `.github/scripts/run-golangci-lint.sh`
- `mockery`: Mock generation via `.github/scripts/run-mockery.sh`
- `generate-schema`: Generates or updates schema files via `.github/scripts/generate-schema.sh`

Install pre-commit hooks: `pre-commit install`

## Dependencies and Requirements

### Go Version

- **Required**: Go 1.24.0+ (specified in go.mod)
- **Validated**: Works with Go 1.25.1

### External Tools

- **Docker**: Required for cluster provisioning (Kind, K3d containers)
- **mockery v3.x**: Critical for test mock generation
- **golangci-lint**: Code quality enforcement
- **mega-linter**: Comprehensive project validation

### Key Go Dependencies

- `github.com/spf13/cobra`: CLI framework
- `sigs.k8s.io/kind`: Kind cluster management
- `github.com/k3d-io/k3d/v5`: K3d cluster management
- `github.com/weaveworks/eksctl`: EKS cluster management
- `k8s.io/client-go`: Kubernetes client libraries

## Common Tasks Reference

### Building the Application

```bash
# Development build
go build -o ksail .

# Cross-platform build (example)
GOOS=linux GOARCH=amd64 go build -o ksail-linux .
GOOS=darwin GOARCH=amd64 go build -o ksail-darwin .
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package tests
go test ./cmd
go test ./pkg/provisioner/cluster/kind

# Verbose test output
go test -v ./cmd

# Test with coverage
go test -cover ./...
```

### Mock Management

```bash
# Generate all mocks (uses .mockery.yml config)
mockery

# Check mockery configuration
mockery showconfig
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix

# Comprehensive validation
mega-linter-runner -f go
```

## CLI Command Timing Feature

### Overview

All KSail CLI commands display timing information on successful completion to help users monitor performance of cluster operations.

### Timer Package (`pkg/ui/timer`)

**Location**: `pkg/ui/timer/`
**Purpose**: Provides timing tracking functionality for CLI command execution.

#### Usage Pattern (Single-Stage Command)

```go
package cmd

import (
 "github.com/devantler-tech/ksail-go/pkg/ui/notify"
 "github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

func HandleCommandRunE(cmd *cobra.Command, ...) error {
 // Create and start timer
 tmr := timer.New()
 tmr.Start()

 // Execute command logic
 err := doSomething()
 if err != nil {
  // NO timing on error paths
  return fmt.Errorf("operation failed: %w", err)
 }

 // Get timing and format (false = single-stage)
 total, stage := tmr.GetTiming()
 timingStr := notify.FormatTiming(total, stage, false)

 // Display success with timing
 notify.Successf(cmd.OutOrStdout(), "operation complete %s", timingStr)
 return nil
}
```

#### Usage Pattern (Multi-Stage Command)

```go
func HandleMultiStageCommandRunE(cmd *cobra.Command, ...) error {
 // Create and start timer
 tmr := timer.New()
 tmr.Start()

 // Stage 1
 notify.Titleln(cmd.OutOrStdout(), "🚀", "Starting...")
 err := doStage1()
 if err != nil {
  return fmt.Errorf("stage 1 failed: %w", err)
 }

 // Transition to stage 2
 tmr.NewStage()
 notify.Titleln(cmd.OutOrStdout(), "📦", "Deploying...")
 err = doStage2()
 if err != nil {
  return fmt.Errorf("stage 2 failed: %w", err)
 }

 // Get timing and format (true = multi-stage)
 total, stage := tmr.GetTiming()
 timingStr := notify.FormatTiming(total, stage, true)

 notify.Successf(cmd.OutOrStdout(), "operation complete %s", timingStr)
 return nil
}
```

### Timing Display Formats

- **Single-stage**: `[1.2s]`
- **Multi-stage**: `[5m30s total|2m15s stage]`
- **Sub-second**: `[500ms]` or `[123µs]`
- **Long durations**: `[1h23m45s total|15m0s stage]`

### Constitutional Compliance

- ✅ **Package-First Design**: Timer is a standalone `pkg/ui/timer` package
- ✅ **Test-First Development**: All contract tests written before implementation
- ✅ **Interface-Based**: Timer interface with mockery support
- ✅ **<1ms Overhead**: Timer adds negligible performance impact
- ✅ **Clean Architecture**: Timer has no dependency on notify (one-way integration)

### Testing Timer Integration

```bash
# Run timer package tests
go test ./pkg/ui/timer/... -v

# Run notify format timing tests
go test ./pkg/ui/notify/... -run FormatTiming -v

# Test CLI command with timing
./ksail cluster init --distribution Kind
# Expected output: "✔ initialized project [1.2s]"
```
