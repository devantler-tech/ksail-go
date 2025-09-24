package eks_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	eksvalidator "github.com/devantler-tech/ksail-go/pkg/validator/eks"
	"github.com/devantler-tech/ksail-go/pkg/validator/testutils"
	eksctlapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// TestEKSValidatorContract tests the contract for EKS configuration validator.
func TestEKSValidatorContract(t *testing.T) {
	t.Parallel()

	// This test MUST FAIL initially to follow TDD approach
	validatorInstance := eksvalidator.NewValidator()
	testCases := createEKSTestCases()

	testutils.RunValidatorTests(
		t,
		validatorInstance,
		testCases,
		testutils.AssertValidationResult[*eksctlapi.ClusterConfig],
	)
}

func createEKSTestCases() []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig] {
	return []testutils.ValidatorTestCase[*eksctlapi.ClusterConfig]{
		{
			Name: "valid_eks_config",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				Metadata: &eksctlapi.ClusterMeta{
					Name:   "test-cluster",
					Region: "us-west-2",
				},
			},
			ExpectedValid:  true,
			ExpectedErrors: []validator.ValidationError{},
		},
		{
			Name: "invalid_eks_config_missing_name",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				Metadata: &eksctlapi.ClusterMeta{
					Region: "us-west-2",
				},
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{Field: "metadata.name", Message: "cluster name is required"},
			},
		},
		{
			Name: "invalid_eks_config_missing_region",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				Metadata: &eksctlapi.ClusterMeta{
					Name: "test-cluster",
				},
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{Field: "metadata.region", Message: "region is required"},
			},
		},
		{
			Name: "invalid_eks_config_missing_metadata",
			Config: &eksctlapi.ClusterConfig{
				TypeMeta: eksctlapi.ClusterConfigTypeMeta(),
				// No metadata
			},
			ExpectedValid: false,
			ExpectedErrors: []validator.ValidationError{
				{Field: "metadata", Message: "metadata is required"},
			},
		},
		testutils.CreateNilConfigTestCase[*eksctlapi.ClusterConfig](),
	}
}
