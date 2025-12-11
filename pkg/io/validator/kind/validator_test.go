package kind_test

import (
	"testing"

	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	"github.com/devantler-tech/ksail-go/pkg/io/validator"
	kindvalidator "github.com/devantler-tech/ksail-go/pkg/io/validator/kind"
	"github.com/devantler-tech/ksail-go/pkg/io/validator/testutils"
)

// TestNewValidator tests the NewValidator constructor.
func TestNewValidator(t *testing.T) {
	t.Parallel()

	testutils.RunNewValidatorConstructorTest(t, func() validator.Validator[*kindapi.Cluster] {
		return kindvalidator.NewValidator()
	})
}

// TestValidate tests the main Validate method with comprehensive scenarios.
func TestValidate(t *testing.T) {
	t.Parallel()

	testutils.RunValidateTest[*kindapi.Cluster](t, testKindValidatorContract)
}

// Helper function for contract testing.
func testKindValidatorContract(t *testing.T) {
	t.Helper()

	// This test MUST FAIL initially to follow TDD approach
	validatorInstance := kindvalidator.NewValidator()
	testCases := createKindTestCases()

	testutils.RunValidatorTests(
		t,
		validatorInstance,
		testCases,
		testutils.AssertValidationResult[*kindapi.Cluster],
	)
}

func createKindTestCases() []testutils.ValidatorTestCase[*kindapi.Cluster] {
	return []testutils.ValidatorTestCase[*kindapi.Cluster]{
		{
			Name: "valid_kind_config",
			Config: &kindapi.Cluster{
				TypeMeta: kindapi.TypeMeta{
					APIVersion: "kind.x-k8s.io/v1alpha4",
					Kind:       "Cluster",
				},
				Name: "test-cluster",
				Nodes: []kindapi.Node{
					{Role: kindapi.ControlPlaneRole},
					{Role: kindapi.WorkerRole},
				},
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "valid_kind_config_no_name",
			Config: &kindapi.Cluster{
				TypeMeta: kindapi.TypeMeta{
					APIVersion: "kind.x-k8s.io/v1alpha4",
					Kind:       "Cluster",
				},
				Nodes: []kindapi.Node{
					{Role: kindapi.ControlPlaneRole},
				},
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		testutils.CreateNilConfigTestCase[*kindapi.Cluster](),
	}
}
