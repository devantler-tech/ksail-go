package kind_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	kindvalidator "github.com/devantler-tech/ksail-go/pkg/validator/kind"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// TestNewValidator tests the NewValidator constructor.
func TestNewValidator(t *testing.T) {
	t.Parallel()

	t.Run("constructor", func(t *testing.T) {
		t.Parallel()

		validator := kindvalidator.NewValidator()
		if validator == nil {
			t.Fatal("NewValidator should return non-nil validator")
		}
	})
}

// TestValidate tests the main Validate method with comprehensive scenarios.
func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("contract_scenarios", func(t *testing.T) {
		t.Parallel()
		testKindValidatorContract(t)
	})
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
