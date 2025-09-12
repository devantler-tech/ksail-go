package clustertestutils_test

import (
	"testing"

	clustertestutils "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorVariables(t *testing.T) {
	t.Parallel()

	// Test all common error variables
	assert.Equal(t, "create cluster failed", clustertestutils.ErrCreateClusterFailed.Error())
	assert.Equal(t, "delete cluster failed", clustertestutils.ErrDeleteClusterFailed.Error())
	assert.Equal(t, "list clusters failed", clustertestutils.ErrListClustersFailed.Error())
	assert.Equal(t, "start cluster failed", clustertestutils.ErrStartClusterFailed.Error())
	assert.Equal(t, "stop cluster failed", clustertestutils.ErrStopClusterFailed.Error())
	assert.Equal(t, "scale node group failed", clustertestutils.ErrScaleNodeGroupFailed.Error())
}

func TestDefaultDeleteCases(t *testing.T) {
	t.Parallel()

	cases := clustertestutils.DefaultDeleteCases()

	// Should return exactly 2 cases
	assert.Len(t, cases, 2)

	// First case: without name uses cfg
	assert.Equal(t, "without name uses cfg", cases[0].Name)
	assert.Equal(t, "", cases[0].InputName)
	assert.Equal(t, "cfg-name", cases[0].ExpectedName)

	// Second case: with name
	assert.Equal(t, "with name", cases[1].Name)
	assert.Equal(t, "custom", cases[1].InputName)
	assert.Equal(t, "custom", cases[1].ExpectedName)
}

func TestDefaultNameCases(t *testing.T) {
	t.Parallel()

	cfgName := "test-config"
	cases := clustertestutils.DefaultNameCases(cfgName)

	// Should return exactly 2 cases
	assert.Len(t, cases, 2)

	// First case: without name uses cfg
	assert.Equal(t, "without name uses cfg", cases[0].Name)
	assert.Equal(t, "", cases[0].InputName)
	assert.Equal(t, cfgName, cases[0].ExpectedName)

	// Second case: with name
	assert.Equal(t, "with name", cases[1].Name)
	assert.Equal(t, "custom", cases[1].InputName)
	assert.Equal(t, "custom", cases[1].ExpectedName)
}

func TestRunStandardSuccessTest(t *testing.T) {
	t.Parallel()

	t.Run("executes_test_runner_for_all_cases", func(t *testing.T) {
		// Don't use t.Parallel() here because we need to wait for completion

		cases := clustertestutils.DefaultNameCases("test-default")

		// This will create and run subtests
		clustertestutils.RunStandardSuccessTest(
			t,
			cases,
			func(innerT *testing.T, inputName, expectedName string) {
				// Just verify the parameters are passed correctly
				assert.Contains(innerT, []string{"", "custom"}, inputName)
				assert.Contains(innerT, []string{"test-default", "custom"}, expectedName)
			},
		)

		// The fact that it completed without panicking means it worked
		assert.Len(t, cases, 2)
	})
}

func TestRunCreateTest(t *testing.T) {
	t.Parallel()

	t.Run("executes_create_test_pattern", func(t *testing.T) {
		// Don't use t.Parallel() here because we need to wait for completion

		// This will create and run subtests with cfg-name cases
		clustertestutils.RunCreateTest(t, func(innerT *testing.T, inputName, expectedName string) {
			// Just verify the parameters are correct for cfg-name cases
			assert.Contains(innerT, []string{"", "custom"}, inputName)
			assert.Contains(innerT, []string{"cfg-name", "custom"}, expectedName)
		})

		// The fact that it completed without panicking means it worked
	})
}

func TestCreateTestEKSNodeGroupBase(t *testing.T) {
	t.Parallel()

	t.Run("with_minimal_options", func(t *testing.T) {
		t.Parallel()

		opts := clustertestutils.EKSNodeGroupBaseOptions{
			Name:         "test-node-group",
			InstanceType: "t3.medium",
		}

		nodeGroup := clustertestutils.CreateTestEKSNodeGroupBase(opts)

		assert.NotNil(t, nodeGroup)
		assert.Equal(t, "test-node-group", nodeGroup.Name)
		assert.Equal(t, "t3.medium", nodeGroup.InstanceType)
		assert.Nil(t, nodeGroup.ScalingConfig)
	})

	t.Run("with_scaling_config", func(t *testing.T) {
		t.Parallel()

		minSize := 1
		maxSize := 5
		desiredCapacity := 3

		opts := clustertestutils.EKSNodeGroupBaseOptions{
			Name:            "test-node-group",
			InstanceType:    "t3.medium",
			MinSize:         &minSize,
			MaxSize:         &maxSize,
			DesiredCapacity: &desiredCapacity,
		}

		nodeGroup := clustertestutils.CreateTestEKSNodeGroupBase(opts)

		assert.NotNil(t, nodeGroup)
		assert.Equal(t, "test-node-group", nodeGroup.Name)
		assert.Equal(t, "t3.medium", nodeGroup.InstanceType)

		require.NotNil(t, nodeGroup.ScalingConfig)
		assert.Equal(t, &minSize, nodeGroup.ScalingConfig.MinSize)
		assert.Equal(t, &maxSize, nodeGroup.ScalingConfig.MaxSize)
		assert.Equal(t, &desiredCapacity, nodeGroup.ScalingConfig.DesiredCapacity)
	})
}

func TestRunActionSuccess(t *testing.T) {
	t.Parallel()

	t.Run("successful_action", func(t *testing.T) {
		t.Parallel()

		// Simple test types
		type mockType struct{}
		type provisionerType struct{}

		setupFn := func(t *testing.T) (provisionerType, mockType) {
			return provisionerType{}, mockType{}
		}

		expectFn := func(mock mockType, name string) {
			// Mock setup - just verify name is passed correctly
			assert.Equal(t, "expected-name", name)
		}

		actionFn := func(provisioner provisionerType, name string) error {
			// Action should receive the input name
			assert.Equal(t, "input-name", name)
			return nil
		}

		// Should not panic or error
		clustertestutils.RunActionSuccess(
			t,
			"TestAction",
			"input-name",
			"expected-name",
			setupFn,
			expectFn,
			actionFn,
		)
	})
}

func TestRunCreateSuccessTest(t *testing.T) {
	t.Parallel()

	t.Run("runs_create_success_pattern", func(t *testing.T) {
		// Don't use t.Parallel() here because we need to wait for subtests to complete

		// Simple test types
		type mockType struct{}
		type provisionerType struct{}

		setupFn := func(t *testing.T) (provisionerType, mockType) {
			return provisionerType{}, mockType{}
		}

		expectFn := func(mock mockType, name string) {
			// Verify expected names are cfg-name or custom
			assert.Contains(t, []string{"cfg-name", "custom"}, name)
		}

		actionFn := func(provisioner provisionerType, name string) error {
			// Verify input names are empty or custom
			assert.Contains(t, []string{"", "custom"}, name)
			return nil
		}

		// Should complete successfully
		clustertestutils.RunCreateSuccessTest(
			t,
			setupFn,
			expectFn,
			actionFn,
		)
	})
}
