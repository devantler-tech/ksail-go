package clustertestutils

import (
	"errors"

	"github.com/devantler-tech/ksail-go/internal/testutils"
)

// ErrBoom is a common test error used across cluster provisioner tests.
var ErrBoom = errors.New("boom")

// DefaultDeleteCases returns standard test cases for testing delete operations with name handling.
func DefaultDeleteCases() []testutils.NameCase {
	return []testutils.NameCase{
		{Name: "without name uses cfg", InputName: "", ExpectedName: "cfg-name"},
		{Name: "with name", InputName: "custom", ExpectedName: "custom"},
	}
}