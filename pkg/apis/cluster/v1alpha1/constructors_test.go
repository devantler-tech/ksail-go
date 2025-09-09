// Package v1alpha1_test provides test definitions for the KSail cluster v1alpha1 model.
package v1alpha1_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewCluster(t *testing.T) {
	t.Parallel()

	// Test with no options - should create empty cluster structure
	cluster := v1alpha1.NewCluster()
	assert.Equal(t, v1alpha1.Kind, cluster.Kind)
	assert.Equal(t, v1alpha1.APIVersion, cluster.APIVersion)
	assert.Equal(t, "", cluster.Metadata.Name) // No defaults in NewCluster anymore

	// Test with options - covers all WithSpec* functions
	testTimeout := metav1.Duration{Duration: time.Duration(10) * time.Minute}
	cluster = v1alpha1.NewCluster(
		v1alpha1.WithMetadataName("test"),
		v1alpha1.WithSpecDistribution(v1alpha1.DistributionK3d),
		v1alpha1.WithSpecConnectionKubeconfig("/test"),
		v1alpha1.WithSpecConnectionContext("test-ctx"),
		v1alpha1.WithSpecConnectionTimeout(testTimeout),
		v1alpha1.WithSpecCNI(v1alpha1.CNICilium),
		v1alpha1.WithSpecCSI(v1alpha1.CSILocalPathStorage),
		v1alpha1.WithSpecIngressController(v1alpha1.IngressControllerTraefik),
		v1alpha1.WithSpecGatewayController(v1alpha1.GatewayControllerCilium),
		v1alpha1.WithSpecReconciliationTool(v1alpha1.ReconciliationToolFlux),
	)
	assert.Equal(t, "test", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)
}

func TestDistribution_Set(t *testing.T) {
	t.Parallel()

	validCases := []struct{ input, expected string }{
		{"Kind", "Kind"},
		{"k3d", "K3d"},
		{"TIND", "Tind"},
	}
	for _, validCase := range validCases {
		var dist v1alpha1.Distribution

		require.NoError(t, dist.Set(validCase.input))
	}

	err := func() error {
		var dist v1alpha1.Distribution

		return dist.Set("invalid")
	}()
	testutils.AssertErrWrappedContains(t, err, v1alpha1.ErrInvalidDistribution, "invalid", "Set(invalid)")
}

func TestReconciliationTool_Set(t *testing.T) {
	t.Parallel()

	validCases := []struct{ input, expected string }{
		{"kubectl", "Kubectl"},
		{"FLUX", "Flux"},
		{"ArgoCD", "ArgoCD"},
	}
	for _, validCase := range validCases {
		var tool v1alpha1.ReconciliationTool

		require.NoError(t, tool.Set(validCase.input))
	}

	err := func() error {
		var tool v1alpha1.ReconciliationTool

		return tool.Set("invalid")
	}()
	testutils.AssertErrWrappedContains(t, err, v1alpha1.ErrInvalidReconciliationTool, "invalid", "Set(invalid)")
}

func TestStringAndTypeMethods(t *testing.T) {
	t.Parallel()

	// Test String() and Type() methods for pflags interface
	dist := v1alpha1.DistributionKind
	assert.Equal(t, "Kind", dist.String())
	assert.Equal(t, "Distribution", dist.Type())

	tool := v1alpha1.ReconciliationToolKubectl
	assert.Equal(t, "Kubectl", tool.String())
	assert.Equal(t, "ReconciliationTool", tool.Type())
}
