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

func TestDistributionSet(t *testing.T) {
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

func TestReconciliationToolSet(t *testing.T) {
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

func TestCNISet(t *testing.T) {
	t.Parallel()

	validCases := []struct{ input, expected string }{
		{"Default", "Default"},
		{"cilium", "Cilium"},
		{"CILIUM", "Cilium"},
	}
	for _, validCase := range validCases {
		var cni v1alpha1.CNI

		require.NoError(t, cni.Set(validCase.input))
	}

	err := func() error {
		var cni v1alpha1.CNI

		return cni.Set("invalid")
	}()
	testutils.AssertErrWrappedContains(
		t,
		err,
		v1alpha1.ErrInvalidCNI,
		"invalid",
		"Set(invalid)",
	)
}

func TestCSISet(t *testing.T) {
	t.Parallel()

	validCases := []struct{ input, expected string }{
		{"Default", "Default"},
		{"localpathstorage", "LocalPathStorage"},
		{"LOCALPATHSTORAGE", "LocalPathStorage"},
	}
	for _, validCase := range validCases {
		var csi v1alpha1.CSI

		require.NoError(t, csi.Set(validCase.input))
	}

	err := func() error {
		var csi v1alpha1.CSI

		return csi.Set("invalid")
	}()
	testutils.AssertErrWrappedContains(
		t,
		err,
		v1alpha1.ErrInvalidCSI,
		"invalid",
		"Set(invalid)",
	)
}

func TestIngressControllerSet(t *testing.T) {
	t.Parallel()

	validCases := []struct{ input, expected string }{
		{"Default", "Default"},
		{"traefik", "Traefik"},
		{"NONE", "None"},
	}
	for _, validCase := range validCases {
		var ic v1alpha1.IngressController

		require.NoError(t, ic.Set(validCase.input))
	}

	err := func() error {
		var ic v1alpha1.IngressController

		return ic.Set("invalid")
	}()
	testutils.AssertErrWrappedContains(
		t,
		err,
		v1alpha1.ErrInvalidIngressController,
		"invalid",
		"Set(invalid)",
	)
}

func TestGatewayControllerSet(t *testing.T) {
	t.Parallel()

	validCases := []struct{ input, expected string }{
		{"Default", "Default"},
		{"traefik", "Traefik"},
		{"cilium", "Cilium"},
		{"NONE", "None"},
	}
	for _, validCase := range validCases {
		var gc v1alpha1.GatewayController

		require.NoError(t, gc.Set(validCase.input))
	}

	err := func() error {
		var gc v1alpha1.GatewayController

		return gc.Set("invalid")
	}()
	testutils.AssertErrWrappedContains(
		t,
		err,
		v1alpha1.ErrInvalidGatewayController,
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

	cni := v1alpha1.CNIDefault
	assert.Equal(t, "Default", cni.String())
	assert.Equal(t, "CNI", cni.Type())

	csi := v1alpha1.CSIDefault
	assert.Equal(t, "Default", csi.String())
	assert.Equal(t, "CSI", csi.Type())

	ic := v1alpha1.IngressControllerDefault
	assert.Equal(t, "Default", ic.String())
	assert.Equal(t, "IngressController", ic.Type())

	gc := v1alpha1.GatewayControllerDefault
	assert.Equal(t, "Default", gc.String())
	assert.Equal(t, "GatewayController", gc.Type())
}

// Tests for constructor functions

func TestNewCluster(t *testing.T) {
	t.Parallel()

	cluster := v1alpha1.NewCluster()

	require.NotNil(t, cluster)
	assert.Equal(t, v1alpha1.Kind, cluster.Kind)
	assert.Equal(t, v1alpha1.APIVersion, cluster.APIVersion)
	assert.NotNil(t, cluster.Metadata)
	assert.NotNil(t, cluster.Spec)
}

func TestNewClusterMetadata(t *testing.T) {
	t.Parallel()

	t.Run("with_name", func(t *testing.T) {
		t.Parallel()

		metadata := v1alpha1.NewClusterMetadata("test-cluster")

		assert.Equal(t, "test-cluster", metadata.Name)
		assert.NotNil(t, metadata.OwnerReferences)
		assert.NotNil(t, metadata.Finalizers)
		assert.NotNil(t, metadata.ManagedFields)
	})

	t.Run("with_empty_name", func(t *testing.T) {
		t.Parallel()

		metadata := v1alpha1.NewClusterMetadata("")

		assert.Empty(t, metadata.Name)
		assert.NotNil(t, metadata.OwnerReferences)
		assert.NotNil(t, metadata.Finalizers)
		assert.NotNil(t, metadata.ManagedFields)
	})
}

func TestNewClusterSpec(t *testing.T) {
	t.Parallel()

	spec := v1alpha1.NewClusterSpec()

	assert.Empty(t, spec.DistributionConfig)
	assert.Empty(t, spec.SourceDirectory)
	assert.NotNil(t, spec.Connection)
	assert.Equal(t, v1alpha1.Distribution(""), spec.Distribution)
	assert.Equal(t, v1alpha1.CNI(""), spec.CNI)
	assert.Equal(t, v1alpha1.CSI(""), spec.CSI)
	assert.Equal(t, v1alpha1.IngressController(""), spec.IngressController)
	assert.Equal(t, v1alpha1.GatewayController(""), spec.GatewayController)
	assert.Equal(t, v1alpha1.ReconciliationTool(""), spec.ReconciliationTool)
	assert.NotNil(t, spec.Options)
}

func TestNewClusterConnection(t *testing.T) {
	t.Parallel()

	connection := v1alpha1.NewClusterConnection()

	assert.Empty(t, connection.Kubeconfig)
	assert.Empty(t, connection.Context)
	assert.Equal(t, metav1.Duration{Duration: 0}, connection.Timeout)
}

func TestNewClusterOptions(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptions()

	assert.NotNil(t, options.Kind)
	assert.NotNil(t, options.K3d)
	assert.NotNil(t, options.Tind)
	assert.NotNil(t, options.EKS)
	assert.NotNil(t, options.Cilium)
	assert.NotNil(t, options.Kubectl)
	assert.NotNil(t, options.Flux)
	assert.NotNil(t, options.ArgoCD)
	assert.NotNil(t, options.Helm)
	assert.NotNil(t, options.Kustomize)
}

func TestNewClusterOptionsKind(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsKind()

	// OptionsKind is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsK3d(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsK3d()

	// OptionsK3d is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsTind(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsTind()

	// OptionsTind is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsEKS(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsEKS()

	assert.Empty(t, options.AWSProfile)
}

func TestNewClusterOptionsCilium(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsCilium()

	// OptionsCilium is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsKubectl(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsKubectl()

	// OptionsKubectl is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsFlux(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsFlux()

	// OptionsFlux is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsArgoCD(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsArgoCD()

	// OptionsArgoCD is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsHelm(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsHelm()

	// OptionsHelm is an empty struct, just verify it's created
	assert.NotNil(t, options)
}

func TestNewClusterOptionsKustomize(t *testing.T) {
	t.Parallel()

	options := v1alpha1.NewClusterOptionsKustomize()

	// OptionsKustomize is an empty struct, just verify it's created
	assert.NotNil(t, options)
}
