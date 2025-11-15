# Feature Specification: Flannel CNI Implementation

**Feature Branch**: `003-flannel-cni`
**Created**: 2025-11-15
**Status**: Draft
**Input**: User description: "Implement Flannel as a Container Network Interface (CNI) option in ksail-go to provide reliable cluster networking compatible with standard Kubernetes setups"

## Clarifications

### Session 2025-11-15

- Q: When Flannel installation fails (e.g., network unavailable, insufficient permissions, incompatible Kubernetes version), what should the system do? → A: Fail gracefully, rollback cluster to pre-installation state, display diagnostic error message
- Q: Why was VXLAN chosen as the default Flannel backend, and should other backends (host-gw, WireGuard) be supported? → A: VXLAN only - most compatible, works across all network types, sufficient for initial release
- Q: What is the expected maximum cluster size (number of nodes) that should be tested with Flannel, and should multi-node communication testing be required? → A: Use default distribution settings

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configure Flannel During Cluster Initialization (Priority: P1)

A Kubernetes administrator wants to specify Flannel as the CNI during initial cluster setup using the `ksail cluster init` command, so they can quickly bootstrap a cluster with Flannel networking from the start.

**Why this priority**: This is the foundational capability - users must be able to select Flannel as a CNI option. Without this, no other Flannel functionality matters. This delivers immediate value as an independently usable feature.

**Independent Test**: Can be fully tested by running `ksail cluster init --cni Flannel` and verifying the generated ksail.yaml contains `cni: Flannel`. This delivers value even without a running cluster.

**Acceptance Scenarios**:

1. **Given** a user wants to create a new KSail project, **When** they run `ksail cluster init --cni Flannel`, **Then** the generated ksail.yaml file contains `spec.cni: Flannel`
2. **Given** a user wants to create a new KSail project with K3d distribution, **When** they run `ksail cluster init --distribution K3d --cni Flannel`, **Then** the generated configuration supports Flannel with K3d
3. **Given** a user wants to create a new KSail project with Kind distribution, **When** they run `ksail cluster init --distribution Kind --cni Flannel`, **Then** the generated Kind configuration disables the default CNI

---

### User Story 2 - Deploy Cluster with Flannel CNI (Priority: P2)

A Kubernetes administrator wants to create and start a cluster with Flannel CNI using the `ksail up` command, so they can have a fully functional cluster with Flannel networking handling pod-to-pod communication.

**Why this priority**: This builds on P1 by actually provisioning and configuring Flannel. It's the next logical step after configuration and represents the core operational capability.

**Independent Test**: Can be fully tested by running `ksail up` on a project configured with Flannel CNI, then verifying nodes reach Ready state and pods can communicate across nodes. This delivers a complete working cluster.

**Acceptance Scenarios**:

1. **Given** a ksail.yaml configured with `cni: Flannel`, **When** user runs `ksail up`, **Then** Flannel is installed and configured in the cluster
2. **Given** Flannel is installed in the cluster, **When** user checks node status, **Then** all nodes show Ready state with Flannel network plugin
3. **Given** a cluster with Flannel running, **When** user deploys pods across different nodes, **Then** pods can communicate across the cluster network
4. **Given** a cluster with Flannel running, **When** user checks DNS resolution, **Then** pods can resolve service names and external DNS successfully

---

### User Story 3 - Validate Flannel Configuration (Priority: P3)

A Kubernetes administrator wants to validate their ksail.yaml configuration using the `ksail validate` command before cluster creation, so they can catch configuration errors early and understand what will be deployed.

**Why this priority**: This is quality-of-life functionality that improves user experience but isn't required for basic Flannel usage. Users can still successfully deploy Flannel without explicit validation.

**Independent Test**: Can be fully tested by running `ksail validate` on projects with various Flannel configurations (valid and invalid) and verifying appropriate messages are displayed. This works without creating any clusters.

**Acceptance Scenarios**:

1. **Given** a ksail.yaml with `cni: Flannel`, **When** user runs `ksail validate`, **Then** validation confirms Flannel is a supported CNI option
2. **Given** a ksail.yaml with invalid CNI value, **When** user runs `ksail validate`, **Then** validation fails with clear error message listing supported CNI options including Flannel
3. **Given** a Kind cluster configuration with Default CNI and Flannel specified in ksail.yaml, **When** user runs `ksail validate`, **Then** validation warns about CNI configuration mismatch

---

### Edge Cases

- What happens when Flannel is specified but the distribution's native CNI is not disabled (e.g., K3d with default CNI enabled)?
- When Flannel installation fails (network unavailable, insufficient permissions, incompatible Kubernetes version), system performs graceful rollback to pre-installation state and displays diagnostic error message with specific failure reason
- What happens when users try to change from another CNI to Flannel on an existing cluster?
- How does the system behave if Flannel manifests are unavailable or corrupted during installation?
- What happens when users specify Flannel with distributions that have specific CNI requirements or defaults (K3d includes Flannel by default)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST recognize "Flannel" as a valid CNI option in cluster configurations
- **FR-002**: System MUST support `--cni Flannel` flag in the `cluster init` command to generate configurations with Flannel CNI
- **FR-003**: System MUST validate "Flannel" as an accepted CNI value in ksail.yaml configuration files
- **FR-004**: System MUST install Flannel CNI components during cluster creation when `cni: Flannel` is specified in ksail.yaml
- **FR-005**: System MUST ensure distribution configurations disable default CNI when Flannel is selected
- **FR-006**: System MUST verify Flannel installation succeeds by checking that Flannel pods are running in kube-system namespace
- **FR-007**: System MUST verify nodes reach Ready state after Flannel installation
- **FR-008**: System MUST obtain and apply Flannel installation resources from official Flannel sources
- **FR-009**: System MUST support Flannel with Kind distribution clusters
- **FR-010**: System MUST support Flannel with K3d distribution clusters
- **FR-011**: System MUST provide clear error messages when Flannel installation fails, including diagnostic information
- **FR-011a**: System MUST rollback cluster to pre-installation state when Flannel installation fails, ensuring no partial or broken configuration remains
- **FR-012**: Documentation MUST describe how to configure and use Flannel CNI, including limitations and compatibility notes
- **FR-013**: System MUST include automated tests validating Flannel installation and basic networking functionality (pod-to-pod communication, DNS resolution)

### Dependencies and Assumptions

- **Dependency**: Flannel installation requires internet connectivity to download Flannel manifests from official sources
- **Dependency**: Kubernetes cluster must be at a version compatible with Flannel (assumes Kubernetes 1.20+)
- **Assumption**: Distribution configurations (Kind, K3d) support disabling default CNI to allow custom CNI installation
- **Assumption**: Flannel will use VXLAN backend exclusively for overlay networking (rationale: most compatible across the network topologies, works universally without special network configuration, sufficient for initial release)
- **Assumption**: Users have appropriate cluster permissions to install DaemonSets and modify cluster networking
- **Assumption**: Cluster sizing follows distribution defaults (Kind/K3d default configurations); no specific node count requirements beyond what distributions provide
- **Assumption**: Standard Flannel configuration is sufficient; custom network configurations are out of scope
- **Out of Scope**: Alternative Flannel backends (host-gw, WireGuard, UDP) are not supported in this initial implementation

### Key Entities

- **CNI Type Enum**: Defines supported CNI options (Default, Cilium, Calico, Flannel) in the cluster specification
- **Cluster Specification**: Contains the CNI field indicating which CNI to use for the cluster
- **Distribution Configuration**: Contains distribution-specific settings (e.g., Kind's disableDefaultCNI) that must be adjusted when using Flannel
- **Flannel Installation Manifest**: The Kubernetes resources required to deploy Flannel (DaemonSet, ConfigMap, RBAC) obtained from official Flannel sources

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can initialize a new KSail project with Flannel CNI in under 30 seconds using the init command
- **SC-002**: Cluster creation with Flannel CNI completes successfully within 3 minutes on standard hardware (4 CPU, 8GB RAM) using distribution default node configurations
- **SC-003**: All cluster nodes reach Ready state within 60 seconds after Flannel pods are running
- **SC-004**: Pod-to-pod communication across nodes succeeds with <10ms latency for same data center nodes
- **SC-005**: DNS resolution for services and external names succeeds within 100ms
- **SC-006**: E2E tests for Flannel CNI pass with 100% success rate in CI/CD pipeline
- **SC-007**: Flannel installation succeeds on both Kind and K3d distributions with 100% reliability in automated tests
- **SC-008**: Users can complete basic networking validation (pod deployment, connectivity test, DNS check) in under 5 minutes
