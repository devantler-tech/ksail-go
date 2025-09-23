package integration

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/validator/eks"
	"github.com/devantler-tech/ksail-go/pkg/validator/k3d"
	"github.com/devantler-tech/ksail-go/pkg/validator/kind"
	"github.com/devantler-tech/ksail-go/pkg/validator/ksail"
	k3dapi "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kindapi "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// TestCompleteValidationWorkflow tests the integration of all validators
// in a complete validation workflow scenario
func TestCompleteValidationWorkflow(t *testing.T) {
	// This test MUST FAIL initially to follow TDD approach

	t.Run("ksail_validator_integration", func(t *testing.T) {
		validator := ksail.NewValidator()

		validConfig := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			Metadata: metav1.ObjectMeta{
				Name: "test-cluster",
			},
			Spec: v1alpha1.Spec{
				Distribution: v1alpha1.DistributionKind,
				Connection: v1alpha1.Connection{
					Context: "kind-test-cluster",
				},
			},
		}

		result := validator.Validate(validConfig)
		require.NotNil(t, result, "Validation result cannot be nil")

		// This should pass once validators are implemented
		assert.True(t, result.Valid, "Valid KSail configuration should pass validation")
		assert.Empty(t, result.Errors, "Valid configuration should have no errors")
	})

	t.Run("kind_validator_integration", func(t *testing.T) {
		validator := kind.NewValidator()

		validConfig := &kindapi.Cluster{
			TypeMeta: kindapi.TypeMeta{
				APIVersion: "kind.x-k8s.io/v1alpha4",
				Kind:       "Cluster",
			},
			Name: "test-cluster",
			Nodes: []kindapi.Node{
				{Role: kindapi.ControlPlaneRole},
				{Role: kindapi.WorkerRole},
			},
		}

		result := validator.Validate(validConfig)
		require.NotNil(t, result, "Validation result cannot be nil")

		// This should pass once validators are implemented
		assert.True(t, result.Valid, "Valid Kind configuration should pass validation")
		assert.Empty(t, result.Errors, "Valid configuration should have no errors")
	})

	t.Run("k3d_validator_integration", func(t *testing.T) {
		validator := k3d.NewValidator()

		validConfig := &k3dapi.SimpleConfig{
			Servers: 1,
			Agents:  2,
		}

		result := validator.Validate(validConfig)
		require.NotNil(t, result, "Validation result cannot be nil")

		// This should pass once validators are implemented
		assert.True(t, result.Valid, "Valid K3d configuration should pass validation")
		assert.Empty(t, result.Errors, "Valid configuration should have no errors")
	})

	t.Run("eks_validator_integration", func(t *testing.T) {
		validator := eks.NewValidator()

		validConfig := &eks.EKSClusterConfig{
			Name:   "test-cluster",
			Region: "us-west-2",
		}

		result := validator.Validate(validConfig)
		require.NotNil(t, result, "Validation result cannot be nil")

		// This should pass once validators are implemented
		assert.True(t, result.Valid, "Valid EKS configuration should pass validation")
		assert.Empty(t, result.Errors, "Valid configuration should have no errors")
	})

	t.Run("validation_error_aggregation", func(t *testing.T) {
		// Test multiple validators with error aggregation
		ksailValidator := ksail.NewValidator()

		invalidConfig := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
			// Missing name and distribution - should cause errors
		}

		result := ksailValidator.Validate(invalidConfig)
		require.NotNil(t, result, "Validation result cannot be nil")

		// Should have validation errors
		assert.False(t, result.Valid, "Invalid configuration should fail validation")
		assert.NotEmpty(t, result.Errors, "Invalid configuration should have errors")
	})

	t.Run("validation_result_structure", func(t *testing.T) {
		// Test that validation results have proper structure
		validator := ksail.NewValidator()

		config := &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "ksail.dev/v1alpha1",
				Kind:       "Cluster",
			},
		}

		result := validator.Validate(config)
		require.NotNil(t, result, "Validation result cannot be nil")

		// Check result structure
		assert.NotEmpty(t, result.ConfigFile, "Result should have config file set")

		// Check error structure if errors exist
		if len(result.Errors) > 0 {
			// Verify that errors are present but don't check specific type
			assert.NotEmpty(t, result.Errors[0].Message, "Error should have a message")
		}
	})
}
