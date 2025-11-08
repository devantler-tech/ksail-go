package testutils

// NameCase represents a common test case shape for commands that accept an optional name
// and default to a configured name when empty.
type NameCase struct {
	Name         string
	InputName    string
	ExpectedName string
}
