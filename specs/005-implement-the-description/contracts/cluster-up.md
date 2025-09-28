# Contract: `ksail cluster up`

## Invocation

- **Command**: `ksail cluster up`
- **Required config**: Valid `ksail.yaml` with `spec.distribution`, `spec.distributionConfig`, and distribution-specific YAML present.
- **Flags**:
  - `--distribution` (string, optional) – overrides `spec.distribution`; allowed values: `Kind`, `K3d`, `EKS`.
  - `--distribution-config` (string, optional) – overrides `spec.distributionConfig` path.
  - `--context` (string, optional) – overrides target kube context name.
  - `--timeout` (duration, optional, default `5m`) – upper bound for provisioning + readiness.
  - `--force` (bool, optional, default `false`) – delete/recreate existing cluster when true; otherwise reuse.

## Inputs

| Source | Description |
|--------|-------------|
| KSail cluster spec | Provides distribution selection, kubeconfig path, context, timeout, and optional AWS profile. |
| Distribution config file | Supplies provider-specific parameters (Kind: `kind.yaml`, K3d: `k3d.yaml`, EKS: `eks.yaml`). |
| Environment | Determines Docker/Podman availability (Kind/K3d) and AWS credentials/profile (EKS). |

## Behavior

1. Load KSail config via `configmanager.NewConfigManager` and validate through `ksailvalidator`.
2. Resolve distribution-specific config using the matching config manager and validator.
3. Perform dependency checks:
   - Kind/K3d: verify Docker/Podman reachable via `containerengine.GetAutoDetectedClient()`.
   - EKS: verify AWS credentials using default credential chain or optional `awsProfile`.
4. Instantiate the matching `clusterprovisioner.ClusterProvisioner` implementation.
5. If `--force` is true and the cluster exists, delete and recreate; otherwise reuse existing cluster when found.
6. Call `Create` (or `Start` if reusing) and wait for kubeconfig/context to become available.
7. Wait for Kubernetes readiness (nodes Ready and API responsive) within the timeout window.
8. Ensure kubeconfig merged and set the command context as current.
9. Emit structured success output including cluster name, distribution, context, kubeconfig path, duration.

## Outputs

### Human-readable (stdout)

```text
✅ Cluster "kind-kind" is ready
   Distribution : Kind
   Context      : kind-kind
   Kubeconfig   : /Users/alex/.kube/config
   Duration     : 2m41s
```

### Machine-readable (stdout when `--output json` is supported in future)

```json
{
  "clusterName": "kind-kind",
  "distribution": "Kind",
  "context": "kind-kind",
  "kubeconfigPath": "/Users/alex/.kube/config",
  "ready": true,
  "durationSeconds": 161
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Provisioning succeeded and readiness verified. |
| `1` | Validation failure (ksail config or distribution config invalid). |
| `2` | Dependency failure (container engine or AWS credentials missing). |
| `3` | Provisioner returned error while creating cluster. |
| `4` | Readiness check timed out. |

## Error Messaging Conventions

- Include actionable remediation (e.g., "Start Docker and rerun `ksail cluster up`").
- Surface original provider errors as contextual detail while preserving KSail-specific prefix.
- On timeout, recommend rerunning with `--timeout` override and include partial status gathered during wait.
