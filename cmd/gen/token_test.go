package gen //nolint:testpackage // Tests need access to unexported helpers

// NOTE: Token command requires an actual cluster connection to work,
// unlike other gen commands that use --dry-run=client.
// Token generation is tested at the integration/e2e level.
