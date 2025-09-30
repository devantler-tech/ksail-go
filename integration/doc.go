// Package integration provides comprehensive integration tests for KSail CLI commands.
// These tests validate end-to-end command execution flows using stub implementations
// instead of mocks, ensuring proper integration between command layer and business logic.
//
// The integration directory follows the same structure as the source code:
//   - cmd/init_test.go: Tests ksail init command flows
//   - cmd/cluster/*_test.go: Tests all ksail cluster subcommand flows
//   - cmd/workload/*_test.go: Tests all ksail workload subcommand flows
//
// The tests use stub implementations from integration/stubs to avoid external dependencies
// while maintaining realistic execution paths.
package integration
