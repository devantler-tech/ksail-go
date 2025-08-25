// Package testutils provides testing utilities to aid with name-based test cases.
package testutils

import "testing"

// NameCase represents a common test case shape for commands that accept an optional name
// and default to a configured name when empty.
type NameCase struct {
	Name         string
	InputName    string
	ExpectedName string
}

// DefaultNameCases returns the standard two cases used throughout the repo:
// 1) explicit name; 2) empty name falls back to default/cfg name.
func DefaultNameCases(defaultCfgName string) []NameCase {
	return []NameCase{
		{Name: "with name", InputName: "my-cluster", ExpectedName: "my-cluster"},
		{Name: "without name uses cfg", InputName: "", ExpectedName: defaultCfgName},
	}
}

// RunNameCases runs the provided function for each NameCase with parallel subtests.
func RunNameCases(t *testing.T, cases []NameCase, run func(t *testing.T, c NameCase)) {
	t.Helper()

	for _, c := range cases {
		// capture
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			run(t, c)
		})
	}
}
