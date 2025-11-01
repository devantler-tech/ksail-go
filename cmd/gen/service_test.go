package gen //nolint:testpackage // Tests need access to unexported helpers

// NOTE: Service command returns a parent command with subcommands (clusterip, nodeport, loadbalancer, externalname).
// Users should use the subcommands directly, e.g., `ksail gen service clusterip ...`
// Testing is done at the integration level rather than unit level.
