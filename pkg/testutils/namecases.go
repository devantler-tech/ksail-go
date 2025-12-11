package testutils

// NameCase represents a common test case shape for commands that accept an optional name
// parameter and default to a configured name when the input name is empty.
type NameCase struct {
	Name         string // Test case name for identification
	InputName    string // Name argument provided to the command
	ExpectedName string // Expected name after default resolution
}
