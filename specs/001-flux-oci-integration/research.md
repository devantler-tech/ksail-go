# Research: Flux OCI Integration for Automated Reconciliation

## Decisions & Rationale

### 1. Local Registry Exposure & Auth

- **Decision**: Run a single local `registry:3` instance bound to localhost only (workstation + cluster), with no authentication.
- **Rationale**: Keeps local development simple while avoiding accidental exposure on the LAN. Localhost-only reduces security risk for an unauthenticated registry.
- **Alternatives Considered**:
  - **LAN-exposed, no-auth registry**: Rejected due to higher risk of unintended access from other machines on the network.
  - **Basic auth for local dev**: Rejected as overkill for single-developer local workflows and adds credential management complexity.
  - **Token/OIDC-based auth**: Rejected as unnecessary complexity for the initial local-only feature.

### 2. Flux Status & Error Visibility

- **Decision**: Primary visibility for Flux reconciliation status and errors is via Flux custom resource status/conditions, inspectable using `kubectl` and Flux tooling. `ksail` only needs to return high-level success/failure codes.
- **Rationale**: Aligns with Flux’s native UX and avoids duplicating its status presentation. Keeps `ksail` CLI thin and focused on orchestration, not full observability.
- **Alternatives Considered**:
  - **Rich `ksail` status surface**: Rejected for now to avoid repeating Flux’s logic and to keep the CLI simple (KISS/YAGNI).
  - **Dedicated log files/metrics dashboards**: Rejected as premature for this local-dev-focused feature; can be added later if needed.

### 3. OCI Artifact Implementation Approach

- **Decision**: Use `google/go-containerregistry` to build and push OCI artifacts that contain Kubernetes manifest bundles.
- **Rationale**: Widely adopted Go library with good support for OCI layout and registries, including local registries. Avoids shelling out to external tools and keeps logic testable.
- **Alternatives Considered**:
  - **Shelling out to `oras` CLI**: Rejected because it introduces an extra binary dependency and makes testing harder.
  - **Manual HTTP interactions with registry API**: Rejected as unnecessary complexity; `go-containerregistry` already provides robust primitives.

### 4. Flux Integration Mechanics

- **Decision**: Use the Flux Installer Go SDK (where available) to install Flux controllers, and generate `OCIRepository` + `Kustomization` YAML manifests applied via Kubernetes client-go or existing `pkg/k8s` helpers.
- **Rationale**: The Installer SDK encodes best practices for Flux controller installation. Generating manifests keeps everything declarative and easy to inspect in Git.
- **Alternatives Considered**:
  - **Shelling out to `flux install` CLI**: Rejected to avoid external binary dependencies and fragile process management.
  - **Embedding static YAML templates for Flux controllers**: Rejected because the Installer SDK better tracks Flux releases and options.

### 5. Registry Persistence Strategy

- **Decision**: Use a Docker volume or bind-mounted directory on the host to persist registry data across cluster stop/start.
- **Rationale**: Satisfies FR-014 (persistence) with minimal complexity, leveraging container engine primitives already in use by KSail-Go.
- **Alternatives Considered**:
  - **In-memory/ephemeral registry**: Rejected because artifacts would be lost between sessions.
  - **External managed registry (e.g., GHCR)**: Rejected; contradicts goal of fully local, offline-capable development.

### 6. Reconciliation Triggering

- **Decision**: Rely on Flux’s configurable reconciliation interval (default 1 minute) plus a KSail command that requests an immediate reconciliation using Flux APIs or by annotating relevant CRs.
- **Rationale**: Matches FR-011 and FR-012, and uses established Flux patterns for manual triggers.
- **Alternatives Considered**:
  - **Custom controller/operator to drive reconciliation**: Rejected as over-engineered for this feature.

## Open Questions (Deferred)

- Exact CLI UX for building and pushing artifacts (e.g., new `ksail workload build` vs. extending existing commands) – to be finalized in Phase 1 design.
- How aggressively to implement registry pruning (FR-018) in the first iteration (manual vs. automatic policies) – may be phased in.
