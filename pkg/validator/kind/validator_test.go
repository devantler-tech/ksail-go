package kind_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/validator"
	kindvalidator "github.com/devantler-tech/ksail-go/pkg/validator/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// TestKindValidatorContract tests the contract for Kind configuration validator.
func TestKindValidatorContract(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach
	t.Parallel()

	validator := kindvalidator.NewValidator()
	require.NotNil(t, validator, "Kind validator constructor must return non-nil validator")

	testCases := createKindTestCases()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := validator.Validate(testCase.config)
			require.NotNil(t, result, "Validation result cannot be nil")

			assertKindValidationResult(t, testCase, result)
		})
	}
}

func createKindTestCases() []kindTestCase {
	return []kindTestCase{
		createValidKindConfigCase(),
		createValidKindConfigNoNameCase(),
		createValidKindConfigNoNodesCase(),
		createValidKindConfigMinimalCase(),
		createNilKindConfigCase(),
	}
}

type kindTestCase struct {
	name         string
	config       *kindapi.Cluster
	expectValid  bool
	expectErrors []string
}

func createValidKindConfigCase() kindTestCase {
	return kindTestCase{
		name: "valid_kind_config",
		config: &kindapi.Cluster{
			TypeMeta: kindapi.TypeMeta{
				APIVersion: "kind.x-k8s.io/v1alpha4",
				Kind:       "Cluster",
			},
			Name: "test-cluster",
			Nodes: []kindapi.Node{
				{
					Role: kindapi.ControlPlaneRole,
				},
				{
					Role: kindapi.WorkerRole,
				},
			},
		},
		expectValid:  true,
		expectErrors: []string{},
	}
}

func createValidKindConfigNoNameCase() kindTestCase {
	return kindTestCase{
		name: "valid_kind_config_no_name",
		config: &kindapi.Cluster{
			TypeMeta: kindapi.TypeMeta{
				APIVersion: "kind.x-k8s.io/v1alpha4",
				Kind:       "Cluster",
			},
			Nodes: []kindapi.Node{
				{
					Role: kindapi.ControlPlaneRole,
				},
			},
		},
		expectValid:  true,
		expectErrors: []string{},
	}
}

func createValidKindConfigNoNodesCase() kindTestCase {
	return kindTestCase{
		name: "valid_kind_config_no_nodes",
		config: &kindapi.Cluster{
			TypeMeta: kindapi.TypeMeta{
				APIVersion: "kind.x-k8s.io/v1alpha4",
				Kind:       "Cluster",
			},
			Name: "test-cluster",
		},
		expectValid:  true,
		expectErrors: []string{},
	}
}

func createValidKindConfigMinimalCase() kindTestCase {
	return kindTestCase{
		name: "valid_kind_config_minimal",
		config: &kindapi.Cluster{
			TypeMeta: kindapi.TypeMeta{
				APIVersion: "kind.x-k8s.io/v1alpha4",
				Kind:       "Cluster",
			},
		},
		expectValid:  true,
		expectErrors: []string{},
	}
}

func createNilKindConfigCase() kindTestCase {
	return kindTestCase{
		name:         "nil_config",
		config:       nil,
		expectValid:  false,
		expectErrors: []string{"configuration cannot be nil"},
	}
}

func assertKindValidationResult(t *testing.T, testCase struct {
	name         string
	config       *kindapi.Cluster
	expectValid  bool
	expectErrors []string
}, result *validator.ValidationResult,
) {
	t.Helper()

	assert.Equal(t, testCase.expectValid, result.Valid, "Expected validation to pass")

	if testCase.expectValid {
		assert.Empty(t, result.Errors, "Expected no validation errors")

		return
	}

	// Check that expected error messages are found
	for _, expectedError := range testCase.expectErrors {
		found := false

		for _, resultErr := range result.Errors {
			if resultErr.Message == expectedError {
				found = true

				break
			}
		}

		assert.True(
			t,
			found,
			"Expected error message '%s' not found in validation errors",
			expectedError,
		)
	}
}
