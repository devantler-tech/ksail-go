package k3d_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	k3dvalidator "github.com/devantler-tech/ksail-go/pkg/validator/k3d"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	configtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// TestK3dValidatorContract tests the contract for K3d configuration validator.
func TestK3dValidatorContract(t *testing.T) {
	t.Parallel()

	// This test MUST FAIL initially to follow TDD approach
	validatorInstance := k3dvalidator.NewValidator()
	testCases := createK3dTestCases()

	testutils.RunValidatorTests(
		t,
		validatorInstance,
		testCases,
		testutils.AssertValidationResult[*k3dapi.SimpleConfig],
	)
}

func createK3dTestCases() []testutils.ValidatorTestCase[*k3dapi.SimpleConfig] {
	return []testutils.ValidatorTestCase[*k3dapi.SimpleConfig]{
		{
			Name: "valid_k3d_config",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					APIVersion: "k3d.io/v1alpha5",
					Kind:       "Simple",
				},
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 1,
				Agents:  2,
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "valid_k3d_config_zero_servers",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					APIVersion: "k3d.io/v1alpha5",
					Kind:       "Simple",
				},
				ObjectMeta: configtypes.ObjectMeta{
					Name: "test-cluster",
				},
				Servers: 0,
				Agents:  1,
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "valid_k3d_config_no_name",
			Config: &k3dapi.SimpleConfig{
				TypeMeta: configtypes.TypeMeta{
					APIVersion: "k3d.io/v1alpha5",
					Kind:       "Simple",
				},
				Servers: 1,
				Agents:  2,
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		testutils.CreateNilConfigTestCase[*k3dapi.SimpleConfig](),
	}
}
