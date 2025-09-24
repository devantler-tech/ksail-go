package kind_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	kindvalidator "github.com/devantler-tech/ksail-go/pkg/validator/kind"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// TestKindValidatorContract tests the contract for Kind configuration validator.
func TestKindValidatorContract(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach
	t.Parallel()

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
