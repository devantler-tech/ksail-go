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

func TestClusterDirectCreation(t *testing.T) {
	t.Parallel()

	// Test direct cluster creation without constructors
	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.Kind,
			APIVersion: v1alpha1.APIVersion,
		},
		Metadata: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.Spec{
			Distribution: v1alpha1.DistributionK3d,
			Connection: v1alpha1.Connection{
				Kubeconfig: "/test",
				Context:    "test-ctx",
				Timeout:    metav1.Duration{Duration: time.Duration(10) * time.Minute},
			},
			CNI:                v1alpha1.CNICilium,
			CSI:                v1alpha1.CSILocalPathStorage,
			IngressController:  v1alpha1.IngressControllerTraefik,
			GatewayController:  v1alpha1.GatewayControllerCilium,
			ReconciliationTool: v1alpha1.ReconciliationToolFlux,
		},
	}

	assert.Equal(t, v1alpha1.Kind, cluster.Kind)
	assert.Equal(t, v1alpha1.APIVersion, cluster.APIVersion)
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
	testutils.AssertErrWrappedContains(
		t,
		err,
		v1alpha1.ErrInvalidDistribution,
		"invalid",
		"Set(invalid)",
	)
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
	testutils.AssertErrWrappedContains(
		t,
		err,
		v1alpha1.ErrInvalidReconciliationTool,
		"invalid",
		"Set(invalid)",
	)
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
