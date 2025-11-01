package gen //nolint:testpackage // Tests need access to unexported helpers

// NOTE: Secret command returns a parent command with subcommands (generic, tls, docker-registry).
// Users should use the subcommands directly, e.g., `ksail gen secret generic ...`
// Testing is done at the integration level rather than unit level.
