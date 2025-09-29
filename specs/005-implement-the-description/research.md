# Phase 0 Research – KSail Cluster Up Command

## Decision: Honour configuration priority through Viper

- **Rationale**: `pkg/config-manager/ksail.InitializeViper` already establishes the precedence chain (defaults < config files < environment variables < flags) and exposes bindings for nested fields. Reusing this instance ensures the command applies overrides in the mandated order (CLI flags → env → configuration files → defaults) without duplicating merge logic.
- **Alternatives Considered**:
  - Implementing manual merge of config structs in the command. Rejected to avoid divergence from global behaviour and reduce maintenance burden.
  - Introducing a bespoke config overlay layer. Rejected because Viper already handles hierarchical sources and the constitution favours minimal dependencies.

## Decision: Reuse existing distribution provisioners

- **Rationale**: The repository already exposes dedicated implementations in `pkg/provisioner/cluster/{kind,k3d,eks}` that satisfy the shared `ClusterProvisioner` interface. Wiring these packages keeps logic consistent with other cluster lifecycle commands and avoids duplicating provider-specific code paths.
- **Alternatives Considered**:
  - Implementing distribution-specific logic directly inside the Cobra command. Rejected because it would splinter provisioning concerns and make dependency injection/testing harder.
  - Introducing a new abstraction layer above `ClusterProvisioner`. Rejected for now; existing interface is sufficient and already covered by tests.

## Decision: Load distribution configs via config managers

- **Rationale**: `pkg/config-manager/{kind,k3d,eks}` already encapsulate YAML loading, defaulting, and validation against upstream schemas. Using them aligns with constitution requirements on validation and prevents bypassing safety checks the CLI depends on.
- **Alternatives Considered**:
  - Manually reading YAML using Viper or `os.ReadFile`. Rejected; would skip validation helpers and duplicate loader logic.
  - Assuming defaults only for missing config files. Rejected because users expect explicit error messaging when config is malformed.

## Decision: Dependency checks before provisioning

- **Rationale**: `pkg/provisioner/containerengine` can detect Docker/Podman availability, and AWS profile/credential presence can be checked via environment using the standard AWS SDK credential chain. Failing fast honours the spec (FR-009) and the constitution’s reliability principle.
- **Alternatives Considered**:
  - Deferring failures to provider libraries (kind, k3d, eksctl). Rejected as it produces hard-to-action errors and violates the spec’s requirement for clear guidance.
  - Attempting automatic installation of dependencies. Rejected for now; outside scope and risky. Guidance will direct users to install/enable prerequisites.

## Decision: Cluster readiness verification strategy

- **Rationale**: We can use client-go’s kubeconfig loader plus a lightweight wait loop over core APIs (e.g., `Nodes` ready condition) with the configured kubeconfig/context. This gives us a distribution-agnostic success signal and ensures the kubeconfig has been written before exit, satisfying FR-008 and FR-011.
- **Alternatives Considered**:
  - Shelling out to `kubectl wait`. Rejected to avoid external command dependencies and keep output consistent.
  - Polling provider-specific health endpoints. Rejected because it fragments logic and fails to guarantee Kubernetes API readiness.

## Decision: Kubeconfig merge and context switching

- **Rationale**: Kind and k3d provisioners already update kubeconfig automatically when run with default options. To guarantee FR-010, we will verify after provisioning that the expected context exists and optionally call `clientcmd.ModifyConfig` to set it as current if needed. For EKS, we can rely on eksctl behaviour and then confirm context selection.
- **Alternatives Considered**:
  - Writing custom kubeconfig merge logic from scratch. Rejected in favour of using `clientcmd` helpers.
  - Leaving context unchanged and printing manual steps. Rejected; contradicts clarified requirement.

## Decision: Telemetry & performance instrumentation

- **Rationale**: The `telemetryRecorder` helper tracks per-stage and total durations with negligible overhead. Instrumenting stages like dependency checks, provisioning, readiness wait, and kubeconfig merge satisfies the constitution’s performance mandate while keeping output actionable.
- **Alternatives Considered**:
  - Integrating external tracing or metrics collectors. Rejected as unnecessary for CLI workflows and would bloat dependencies.
  - Omitting telemetry in favour of manual timing. Rejected; explicitly violates Constitution IV and weakens observability.
