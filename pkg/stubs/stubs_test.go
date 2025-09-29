package stubs_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/stubs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidatorStub demonstrates the validator stub can be used as an adapter.
func TestValidatorStub(t *testing.T) {
	t.Parallel()

	t.Run("default_success", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewValidatorStub[v1alpha1.Cluster]()
		config := &v1alpha1.Cluster{}
		
		result := stub.Validate(*config)
		
		require.NotNil(t, result)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("configured_error", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewValidatorStub[v1alpha1.Cluster]().
			WithValidationError("test.field", "test error message")
		config := &v1alpha1.Cluster{}
		
		result := stub.Validate(*config)
		
		require.NotNil(t, result)
		assert.False(t, result.Valid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, "test.field", result.Errors[0].Field)
		assert.Equal(t, "test error message", result.Errors[0].Message)
	})
}

// TestConfigManagerStub demonstrates the config manager stub can be used as an adapter.
func TestConfigManagerStub(t *testing.T) {
	t.Parallel()

	t.Run("default_nil_result", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewConfigManagerStub[v1alpha1.Cluster]()
		
		result, err := stub.LoadConfig()
		
		assert.NoError(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 1, stub.CallCount())
	})

	t.Run("configured_with_config", func(t *testing.T) {
		t.Parallel()
		
		expectedConfig := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Distribution: v1alpha1.DistributionKind,
			},
		}
		stub := stubs.NewConfigManagerStub[v1alpha1.Cluster]().
			WithConfig(expectedConfig)
		
		result, err := stub.LoadConfig()
		
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig, result)
		assert.Equal(t, v1alpha1.DistributionKind, result.Spec.Distribution)
	})

	t.Run("configured_with_error", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewConfigManagerStub[v1alpha1.Cluster]().
			WithLoadError("config load failed")
		
		result, err := stub.LoadConfig()
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config load failed")
		assert.Nil(t, result)
	})
}

// TestClusterProvisionerStub demonstrates the cluster provisioner stub can be used as an adapter.
func TestClusterProvisionerStub(t *testing.T) {
	t.Parallel()

	t.Run("default_success", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewClusterProvisionerStub()
		
		err := stub.Create(nil, "test-cluster")
		assert.NoError(t, err)
		assert.Equal(t, 1, stub.GetCreateCallsCount())
		assert.Equal(t, "test-cluster", stub.GetLastCreateCall())
		
		clusters, err := stub.List(nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{"test-cluster"}, clusters)
		
		exists, err := stub.Exists(nil, "test-cluster")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("configured_with_errors", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewClusterProvisionerStub().
			WithCreateError("create failed").
			WithListError("list failed")
		
		err := stub.Create(nil, "test-cluster")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "create failed")
		
		clusters, err := stub.List(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list failed")
		assert.Nil(t, clusters)
	})
}

// TestGeneratorStub demonstrates the generator stub can be used as an adapter.
func TestGeneratorStub(t *testing.T) {
	t.Parallel()

	t.Run("default_success", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewGeneratorStub[v1alpha1.Cluster, string]()
		config := v1alpha1.Cluster{}
		
		result, err := stub.Generate(config, "test-options")
		
		assert.NoError(t, err)
		assert.Contains(t, result, "Generated content")
		assert.Equal(t, 1, stub.CallCount())
	})

	t.Run("configured_with_custom_result", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewGeneratorStub[v1alpha1.Cluster, string]().
			WithResult("custom generated content")
		config := v1alpha1.Cluster{}
		
		result, err := stub.Generate(config, "test-options")
		
		assert.NoError(t, err)
		assert.Equal(t, "custom generated content", result)
	})

	t.Run("configured_with_error", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewGeneratorStub[v1alpha1.Cluster, string]().
			WithGenerationError("generation failed")
		config := v1alpha1.Cluster{}
		
		result, err := stub.Generate(config, "test-options")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generation failed")
		assert.Empty(t, result)
	})
}

// TestInstallerStub demonstrates the installer stub can be used as an adapter.
func TestInstallerStub(t *testing.T) {
	t.Parallel()

	t.Run("default_success", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewInstallerStub()
		
		err := stub.Install(nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, stub.InstallCalls)
		
		err = stub.Uninstall(nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, stub.UninstallCalls)
	})

	t.Run("configured_with_errors", func(t *testing.T) {
		t.Parallel()
		
		stub := stubs.NewInstallerStub().
			WithInstallError("install failed").
			WithUninstallError("uninstall failed")
		
		err := stub.Install(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "install failed")
		
		err = stub.Uninstall(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uninstall failed")
	})
}