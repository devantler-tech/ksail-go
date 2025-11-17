# Feature Specification: Flux OCI Integration for Automated Reconciliation

**Feature Branch**: `001-flux-oci-integration`  
**Created**: 2025-11-17  
**Status**: Draft  
**Input**: User description: "Integrate Flux for automated OCI artifact reconciliation"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Bootstrap Cluster with Flux (Priority: P1)

A cluster administrator initializes a new cluster with Flux as the GitOps engine. During cluster creation, Flux components are automatically installed and configured, establishing the foundation for continuous deployment workflows.

**Why this priority**: This is the foundational capability that enables all subsequent GitOps workflows. Without Flux installed, no automated reconciliation can occur.

**Independent Test**: Can be fully tested by creating a cluster with Flux enabled and verifying Flux controllers are running and healthy in the cluster.

**Acceptance Scenarios**:

1. **Given** a fresh cluster configuration specifying Flux as the GitOps engine, **When** the administrator runs cluster creation, **Then** Flux controllers are installed and report ready status
2. **Given** a cluster configuration with Flux selected, **When** cluster creation completes, **Then** the Flux namespace exists with all required custom resource definitions
3. **Given** an existing cluster without Flux, **When** the administrator updates configuration to enable Flux and reconciles, **Then** Flux components are installed without disrupting existing workloads

---

### User Story 2 - Provision Local OCI Registry (Priority: P1)

A cluster administrator provisions a local OCI-compliant registry during cluster setup. The registry is accessible from both the cluster nodes and the administrator's workstation, enabling image storage and retrieval for local development workflows.

**Why this priority**: The local registry is essential infrastructure for storing OCI artifacts that Flux will reconcile. Without it, there's nowhere to push or pull artifacts from.

**Independent Test**: Can be fully tested by provisioning the registry and verifying image push/pull operations succeed from both host and cluster.

**Acceptance Scenarios**:

1. **Given** a cluster initialization with registry provisioning enabled, **When** cluster creation completes, **Then** a local OCI registry is running and accessible at a known endpoint
2. **Given** a running local registry, **When** the administrator pushes a test image from their workstation, **Then** the image is successfully stored and retrievable
3. **Given** a running cluster with local registry, **When** a pod in the cluster references an image in the local registry, **Then** the image pulls successfully
4. **Given** registry connection failures, **When** the administrator attempts to push an artifact, **Then** clear error messages indicate the connectivity issue

---

### User Story 3 - Build and Push OCI Artifacts (Priority: P2)

A cluster administrator packages workload manifests as OCI artifacts and pushes them to repositories in the local registry. The artifacts contain all necessary Kubernetes resources and are versioned appropriately for tracking changes.

**Why this priority**: This capability bridges workload definitions and the GitOps reconciliation loop. It's essential but depends on having both Flux and the registry operational.

**Independent Test**: Can be tested by packaging sample manifests into OCI artifacts and pushing them to the local registry, then verifying artifact metadata and contents.

**Acceptance Scenarios**:

1. **Given** a directory containing Kubernetes manifests, **When** the administrator builds an OCI artifact from those manifests, **Then** the artifact is created with correct metadata and versioning
2. **Given** a built OCI artifact, **When** the administrator pushes it to a repository in the local registry, **Then** the artifact is stored and can be listed in the repository
3. **Given** multiple versions of an artifact in the registry, **When** the administrator queries available versions, **Then** all versions are listed with their metadata
4. **Given** an artifact push operation, **When** network issues occur, **Then** the operation retries or fails with actionable error messages

---

### User Story 4 - Configure Flux Custom Resources (Priority: P2)

A cluster administrator generates and applies Flux custom resources that track OCI artifact repositories. These resources configure Flux to monitor for new artifact versions and automatically synchronize changes to the cluster.

**Why this priority**: This establishes the automated reconciliation loop. It's critical for the GitOps workflow but requires Flux, registry, and artifacts to be in place first.

**Independent Test**: Can be tested by creating Flux OCIRepository and Kustomization resources, applying them to the cluster, and verifying Flux recognizes and processes them.

**Acceptance Scenarios**:

1. **Given** an OCI artifact repository in the local registry, **When** the administrator generates Flux OCIRepository resources, **Then** the resources correctly reference the repository with appropriate authentication settings
2. **Given** Flux OCIRepository resources applied to the cluster, **When** Flux reconciles, **Then** Flux successfully connects to the registry and fetches artifact metadata
3. **Given** a Flux Kustomization resource referencing an OCIRepository, **When** applied to the cluster, **Then** Flux extracts manifests from the artifact and applies them to the cluster
4. **Given** misconfigured Flux resources, **When** Flux attempts reconciliation, **Then** status conditions clearly indicate configuration errors

---

### User Story 5 - Trigger Reconciliation on Artifact Updates (Priority: P1)

A cluster administrator updates workload definitions, builds a new OCI artifact version, and pushes it to the registry. Flux automatically detects the new version and reconciles the cluster state to match the updated manifests.

**Why this priority**: This is the core value proposition - automated continuous deployment. It's what makes the entire system worthwhile and demonstrates end-to-end functionality.

**Independent Test**: Can be fully tested by pushing a new artifact version to the registry and observing Flux automatically applying changes to the cluster within the configured reconciliation interval.

**Acceptance Scenarios**:

1. **Given** a Flux OCIRepository monitoring an artifact repository, **When** a new artifact version is pushed, **Then** Flux detects the update within its configured polling interval
2. **Given** Flux detecting a new artifact version, **When** the artifact contains updated manifests, **Then** Flux applies the changes to the cluster
3. **Given** updated manifests causing resource changes, **When** Flux reconciles, **Then** existing resources are updated or replaced according to manifest differences
4. **Given** a reconciliation interval configured in Flux resources, **When** administrators check reconciliation frequency, **Then** Flux adheres to the configured schedule

---

### User Story 6 - Manual Reconciliation via CLI (Priority: P3)

A cluster administrator uses the `ksail workload reconcile` command to manually trigger Flux reconciliation without waiting for the automatic polling interval. This enables immediate testing of changes and troubleshooting of deployment issues.

**Why this priority**: Manual triggering is convenient but not essential - automated reconciliation is the primary workflow. This is useful for debugging and rapid iteration but not core functionality.

**Independent Test**: Can be tested by running the reconcile command and verifying it triggers immediate Flux reconciliation of configured sources.

**Acceptance Scenarios**:

1. **Given** a cluster with Flux configured, **When** the administrator runs `ksail workload reconcile`, **Then** Flux immediately reconciles all configured sources
2. **Given** pending artifact updates, **When** manual reconciliation is triggered, **Then** changes are applied without waiting for the polling interval
3. **Given** reconciliation errors in Flux, **When** manual reconciliation is triggered, **Then** error details are displayed to the administrator

---

### Edge Cases

- What happens when the local registry becomes unavailable during Flux reconciliation?
- How does the system handle OCI artifacts with invalid or corrupted manifests?
- What happens when multiple artifact versions are pushed rapidly in succession?
- How does Flux handle authentication failures when accessing the OCI registry?
- What happens when an artifact update causes resource conflicts or namespace collisions?
- How does the system handle registry disk space exhaustion?
- What happens when Flux is already installed when the administrator tries to provision it?
- How are existing Flux configurations preserved when updating cluster settings?
- What happens when an OCI artifact contains resources that require elevated permissions?
- How does the system handle reconciliation when cluster resources are manually modified outside of Flux?

## Requirements *(mandatory)*

### Assumptions

- Docker is available on the administrator's workstation for building and pushing OCI artifacts
- Administrators have basic understanding of Kubernetes manifests and GitOps principles
- Network connectivity exists between the administrator's workstation, the local registry, and cluster nodes
- Sufficient disk space is available for registry storage volumes
- The cluster has internet connectivity during initial Flux installation to pull Flux controller images (can be mirrored for air-gapped scenarios)
- Administrators understand OCI artifact versioning and semantic versioning practices
- The default reconciliation interval of 1 minute is acceptable for most use cases (configurable if needed)

### Functional Requirements

- **FR-001**: System MUST support installing Flux controllers during cluster bootstrap when Flux is specified as the GitOps engine
- **FR-002**: System MUST support installing Flux controllers on-demand to existing clusters that don't have Flux installed
- **FR-003**: System MUST provision a local OCI-compliant registry container during cluster creation when registry provisioning is enabled
- **FR-004**: System MUST configure network connectivity between the local registry and cluster nodes to enable image pulls
- **FR-005**: System MUST configure network connectivity between the local registry and the administrator's workstation to enable artifact pushes
- **FR-006**: System MUST support building OCI artifacts from directories containing Kubernetes manifests
- **FR-007**: System MUST support pushing OCI artifacts to repositories in the local registry with semantic versioning
- **FR-008**: System MUST generate Flux OCIRepository custom resources that reference repositories in the local registry
- **FR-009**: System MUST generate Flux Kustomization custom resources that source manifests from OCIRepository resources
- **FR-010**: System MUST configure Flux resources without authentication requirements for the local registry to optimize local development simplicity
- **FR-011**: System MUST configure Flux to reconcile OCIRepository sources at a configurable interval with a default of 1 minute for rapid feedback during development
- **FR-012**: System MUST provide a command to trigger immediate Flux reconciliation without waiting for the polling interval
- **FR-013**: System MUST display Flux reconciliation status and any errors encountered during synchronization
- **FR-014**: System MUST persist registry data across cluster stop/start cycles to preserve pushed artifacts
- **FR-015**: System MUST provide cleanup operations to remove registry volumes when deleting clusters
- **FR-016**: System MUST detect and handle scenarios where Flux is already installed, avoiding duplicate installations
- **FR-017**: System MUST validate OCI artifact structure before pushing to prevent pushing invalid artifacts
- **FR-018**: System MUST support pruning of old artifact versions from the registry to manage storage space

### Key Entities

- **OCI Registry**: A local container registry instance storing OCI artifacts, accessible from both workstation and cluster, with persistent storage volumes and configurable port mappings
- **OCI Artifact**: A versioned package containing Kubernetes manifest files, stored in an OCI-compliant format with metadata including version tags and creation timestamps
- **Flux OCIRepository**: A custom resource defining the connection to an OCI artifact repository, including registry endpoint, repository path, authentication credentials, and polling interval
- **Flux Kustomization**: A custom resource specifying which OCIRepository to sync from, target namespace for resources, pruning behavior, and health check configurations
- **Workload Manifest**: Kubernetes YAML files defining application resources such as Deployments, Services, and ConfigMaps, packaged within OCI artifacts
- **GitOps Engine Configuration**: Settings in the cluster configuration specifying which deployment tool to use (Kubectl or Flux) and associated parameters

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can bootstrap a new cluster with Flux fully operational in under 5 minutes
- **SC-002**: The local OCI registry successfully handles image push and pull operations within 10 seconds per operation
- **SC-003**: Flux detects new artifact versions and completes reconciliation within 2 minutes of pushing changes
- **SC-004**: Manual reconciliation triggered via CLI command completes within 30 seconds
- **SC-005**: The system provides clear error messages within 5 seconds when registry connectivity fails or artifacts are invalid
- **SC-006**: Administrators can complete the full workflow (cluster creation to workload deployment) in under 10 minutes following documentation
- **SC-007**: Registry persists artifacts across cluster stop/start cycles with 100% data retention
- **SC-008**: The system successfully handles 50+ artifact versions in a single repository without performance degradation
