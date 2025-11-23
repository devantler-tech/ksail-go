# Data Model: Flux OCI Integration

## Entities

### ClusterConfig (existing, extended)

- **Fields (relevant additions/usage)**:
  - `GitOpsEngine` (enum): `Kubectl` | `Flux`
  - `RegistryEnabled` (bool): whether to provision a local OCI registry
  - `RegistryPort` (int, optional): host port for localhost registry binding
  - `FluxInterval` (duration, optional): reconciliation interval for OCI sources (default: 1m)

### OCIRegistry

- **Description**: Local `registry:3` instance bound to localhost, storing OCI artifacts.
- **Fields**:
  - `Name`: logical name (e.g., `local-registry`)
  - `Endpoint`: `localhost:<port>` endpoint used by workstation and Flux
  - `Port`: integer port exposed on host
  - `DataPath` or `VolumeName`: location/volume for persistent storage
  - `Status`: enum (`NotProvisioned`, `Provisioning`, `Running`, `Error`)
  - `LastError`: optional string describing last failure

### OCIArtifact

- **Description**: Versioned bundle of Kubernetes manifests packaged as an OCI artifact.
- **Fields**:
  - `Name`: logical artifact name (e.g., `workloads/app`)
  - `Version`: semantic version string (e.g., `1.0.3`)
  - `RegistryEndpoint`: registry base (e.g., `localhost:5000`)
  - `Repository`: repository path in registry (e.g., `ksail/workloads/app`)
  - `Tag`: computed or explicit tag (usually matches `Version`)
  - `SourcePath`: local filesystem directory with manifests
  - `CreatedAt`: timestamp (optional, used in metadata)

### FluxOCIRepository (CRD model)

- **Description**: Flux `OCIRepository` custom resource tracking an OCI artifact repository.
- **Key Fields (modeled for generation)**:
  - `metadata.name`: name of the source
  - `metadata.namespace`: Flux namespace (typically `flux-system`)
  - `spec.url`: `oci://<endpoint>/<repository>`
  - `spec.interval`: reconciliation interval (from `ClusterConfig.FluxInterval`)
  - `spec.ref.tag`: desired version tag
  - `status.conditions`: used by operators to see fetch/reconciliation status

### FluxKustomization (CRD model)

- **Description**: Flux `Kustomization` resource applying manifests from an `OCIRepository`.
- **Key Fields (modeled for generation)**:
  - `metadata.name`
  - `metadata.namespace`
  - `spec.sourceRef`: reference to `OCIRepository` (name, kind, namespace)
  - `spec.path`: path within artifact (usually `/`)
  - `spec.prune`: bool, whether to prune resources
  - `spec.targetNamespace`: namespace where workloads are applied
  - `spec.interval`: reconciliation interval (can mirror source)
  - `status.conditions`: used for apply/health status

### ReconciliationRequest

- **Description**: Logical request to reconcile one or more Flux sources.
- **Fields**:
  - `Targets`: list of `OCIRepository`/`Kustomization` identifiers
  - `Mode`: `Automatic` | `Manual`
  - `RequestedAt`: timestamp

## Relationships

- `ClusterConfig` **owns** a single `OCIRegistry` configuration when `RegistryEnabled=true`.
- `OCIRegistry` **hosts** many `OCIArtifact` repositories.
- Each `FluxOCIRepository` **points to** a specific `OCIArtifact` repository in `OCIRegistry`.
- Each `FluxKustomization` **references** exactly one `FluxOCIRepository`.
- `ReconciliationRequest` **targets** one or more `FluxOCIRepository` and/or `FluxKustomization` resources.

## Validation Rules

- `ClusterConfig.GitOpsEngine == Flux` is required to install Flux components.
- `OCIRegistry.Endpoint` MUST be `localhost:<port>` (per clarification).
- `OCIArtifact.Version` MUST be a valid semver string for FR-007.
- `FluxOCIRepository.spec.url` MUST follow `oci://<endpoint>/<project-name>` pattern.
- `FluxKustomization.spec.sourceRef` MUST reference an existing `OCIRepository` in the same namespace.
- Reconciliation interval defaults to 1 minute if not explicitly set and MUST be >0.
