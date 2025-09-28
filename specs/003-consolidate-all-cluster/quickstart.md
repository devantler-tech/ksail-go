# Quickstart – Validating KSail Cluster Command Consolidation

1. **Build the CLI**

   ```sh
   go build ./...
   ```

2. **Run unit tests** (ensures new command wiring covered by updated tests)

   ```sh
   go test ./cmd -run TestCluster
   go test ./...
   ```

3. **Verify help output**

   ```sh
   ./ksail --help | grep -A2 "cluster"
   ./ksail cluster --help
   ```

4. **Smoke test lifecycle commands** (stub behavior but validates routing)

   ```sh
   ./ksail cluster up --help
   ./ksail cluster status
   ```

5. **Check legacy commands removed**

   ```sh
   ./ksail up
   ```

   Expected: Cobra "unknown command" error referencing the new `cluster` namespace.
