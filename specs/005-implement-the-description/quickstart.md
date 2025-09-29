# Quickstart – KSail Cluster Up Command

## Prerequisites

- KSail project initialised with `ksail init` and populated `ksail.yaml`.
- Distribution config present (e.g., `kind.yaml`, `k3d.yaml`, or `eks.yaml`).
- Docker or Podman running for Kind/K3d workflows.
- AWS CLI configured with working credentials/profile for EKS workflows.

## Steps (Kind example)

1. Ensure Docker is running: `docker ps`.
2. Confirm KSail config targets Kind:
   - `spec.distribution: Kind`
   - `spec.distributionConfig: kind.yaml`
3. Run the command:

   ```bash
   ksail cluster up --distribution Kind
   ```

4. Wait for the success message summarising distribution, context, kubeconfig, slowest stage, and total duration.
5. Verify the kube context switched:

   ```bash
   kubectl config current-context
   ```

6. Inspect cluster nodes to confirm readiness:

   ```bash
   kubectl get nodes
   ```

## Steps (K3d example)

1. Start Docker or Podman.
2. Populate `spec.distribution: K3d` and provide `k3d.yaml`.
3. Execute:

   ```bash
   ksail cluster up --distribution K3d
   ```

4. Confirm the CLI reports success with the timing summary and the context `k3d-<name>` is active.

## Steps (EKS example)

1. Export the AWS profile if required:

   ```bash
   export AWS_PROFILE=infra-admin
   ```

2. Set `spec.distribution: EKS` and supply `eks.yaml` with region + node groups.
3. Run:

   ```bash
   ksail cluster up --distribution EKS --timeout 10m
   ```

4. After success, verify the timing summary, confirm the context, and describe cluster health:

   ```bash
   kubectl cluster-info
   ```

## Force Recreation Flow

- Use `--force` when you need a clean cluster reset:

   ```bash
   ksail cluster up --force
   ```

- The command deletes the existing cluster, reprovisions it, then waits for readiness before exiting.
