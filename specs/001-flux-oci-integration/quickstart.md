# Quickstart: Flux OCI Integration with KSail-Go

1. **Initialize a KSail-Go project (if not already)**
   - Run `ksail init` in a new directory to create `ksail.yaml` and base config.

2. **Enable Flux and local registry in cluster config**
    - Edit `ksail.yaml` (or equivalent) to:
       - Set `gitOpsEngine: Flux`.
       - Enable the local registry and optionally tune the host port, for example:

          ```yaml
          localRegistry: Enabled
          options:
             localRegistry:
                hostPort: 5000
          ```

3. **Create or start the local cluster**
   - Run `ksail up` (or the existing cluster create command) to:
     - Provision a local Kind/K3d cluster.
     - Install Flux controllers.
     - Start a localhost-only `registry:3` instance with persistent storage.
   - Use the pinned `rancher/k3s:v1.29.4-k3s1` image for K3d clusters so Flux always sees a supported Kubernetes version (>=1.22). Update existing `k3d.yaml` files if they still reference older images.

4. **Package workloads as OCI artifacts**
   - Place Kubernetes manifests (or Kustomize bases) in a directory, for example `k8s/workloads/app`.
   - Use the planned KSail-Go workload command (e.g., `ksail workload build`) to:
     - Build an OCI artifact from that directory.
     - Tag it with a semantic version (e.g., `1.0.0`).
     - Push it to the local registry at `localhost:<hostPort>`.

5. **Configure Flux to track OCI artifacts**
   - Generate Flux `OCIRepository` and `Kustomization` resources that:
     - Point to `oci://localhost:<hostPort>/<project-name>`.
     - Use a 1-minute reconciliation interval by default.
   - Apply these resources to the cluster (e.g., via KSail-Go or `kubectl`).

6. **Trigger reconciliation and verify changes**
   - Wait for Flux’s automatic reconciliation (default ~1 minute), or run `ksail workload reconcile` to trigger it immediately.
   - Use `kubectl get kustomizations -n flux-system` and `kubectl get ocirepositories -n flux-system` to inspect status and errors.

7. **Iterate on workloads**
   - Update manifests, rebuild and push new artifact versions, and allow Flux to reconcile.
   - Use the same `ksail` commands and Flux CR inspection to validate each iteration.
