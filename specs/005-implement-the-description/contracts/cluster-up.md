# Contract: `ksail cluster up`

## Invocation

- **Command**: `ksail cluster up`
- **Required config**: Valid `ksail.yaml` with `spec.distribution`, `spec.distributionConfig`, and distribution-specific YAML present.
- **Configuration precedence**: Effective settings resolve in this order – CLI flag overrides → environment variables (`KSAIL_`) → configuration files (`ksail.yaml`, `<distribution>.yaml`) → CLI defaults.
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
| CLI Flags | Immediate overrides such as `--context`, `--timeout`, `--distribution`. |
| Defaults | Applied only when not provided by higher-precedence sources (e.g., timeout = 5m). |

## Behavior

1. Initialise Viper via KSail config manager, ensuring the precedence chain (flags → env → files → defaults) is enforced.
2. Load KSail config through `configmanager.ConfigManager` and validate distribution choice.
3. Resolve distribution-specific config using the matching config manager and validator.
4. Perform dependency checks:
   - Kind/K3d: verify Docker/Podman reachable via `containerengine` helper.
   - EKS: verify AWS credentials using default credential chain or optional `awsProfile`.
5. Instantiate the matching `clusterprovisioner.ClusterProvisioner` implementation.
6. If `--force` is true and the cluster exists, delete and recreate; otherwise reuse existing cluster when found.
7. Call `Create` (or `Start` when reusing) and wait for kubeconfig/context to become available.
8. Wait for Kubernetes readiness (nodes Ready and API responsive) within the timeout window.
9. Merge kubeconfig data, set the desired context as current, and persist to disk.
10. Capture telemetry for each stage (dependencies, provisioning, readiness, kubeconfig) plus total duration, then emit success/failure output including remediation hints where applicable.

## Outputs

### Human-readable (stdout)

```text
✅ Cluster "kind-kind" is ready
   Distribution : Kind
   Context      : kind-kind
   Kubeconfig   : /Users/alex/.kube/config
   SlowestStage : readiness (92s)
   TotalTime    : 2m41s
```

### Machine-readable (stdout when `--output json` is supported in future)

```json
{
  "clusterName": "kind-kind",
  "distribution": "Kind",
  "context": "kind-kind",
  "kubeconfigPath": "/Users/alex/.kube/config",
  "ready": true,
  "totalDurationSeconds": 161,
  "stageDurationsSeconds": {
    "dependencies": 4,
    "provision": 55,
    "readiness": 92,
    "kubeconfig": 10
  }
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
- Always report the stage that failed and the elapsed time before failure for telemetry parity.
