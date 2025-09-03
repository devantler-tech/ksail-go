// Package v1alpha1_test provides test definitions for the KSail cluster v1alpha1 model.
package v1alpha1_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewCluster(t *testing.T) {
	t.Parallel()

	// Test with defaults
	cluster := v1alpha1.NewCluster()
	assert.Equal(t, v1alpha1.Kind, cluster.Kind)
	assert.Equal(t, v1alpha1.APIVersion, cluster.APIVersion)
	assert.Equal(t, "ksail-default", cluster.Metadata.Name)

	// Test with options - covers all WithSpec* functions
	testTimeout := metav1.Duration{Duration: time.Duration(10) * time.Minute}
	cluster = v1alpha1.NewCluster(
		v1alpha1.WithMetadataName("test"),
		v1alpha1.WithSpecDistribution(v1alpha1.DistributionK3d),
		v1alpha1.WithSpecContainerEngine(v1alpha1.ContainerEnginePodman),
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

func TestSetDefaults(t *testing.T) {
	t.Parallel()

	// Test all defaults applied
	cluster := createTestClusterWithDefaults()
	cluster.SetDefaults()
	assertDefaultValues(t, cluster)

	// Test custom values preserved
	cluster = createTestClusterWithCustomValues()
	cluster.SetDefaults()
	assertCustomValues(t, cluster)
}

func createTestClusterWithDefaults() *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Metadata: createDefaultObjectMeta(""),
		Spec:     createDefaultSpec(),
	}
}

func createTestClusterWithCustomValues() *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Metadata: createDefaultObjectMeta("custom"),
		Spec:     createCustomSpec(),
	}
}

func createDefaultObjectMeta(name string) metav1.ObjectMeta {
	return clustertestutils.CreateDefaultObjectMeta(name)
}

func createDefaultSpec() v1alpha1.Spec {
	return clustertestutils.CreateDefaultSpec()
}

func createCustomSpec() v1alpha1.Spec {
	return v1alpha1.Spec{
		Distribution:       v1alpha1.DistributionK3d,
		DistributionConfig: "",
		SourceDirectory:    "",
		Connection: v1alpha1.Connection{
			Kubeconfig: "/custom",
			Context:    "custom-ctx",
			Timeout:    metav1.Duration{Duration: time.Duration(15) * time.Minute},
		},
		ContainerEngine:    "",
		CNI:                "",
		CSI:                "",
		IngressController:  "",
		GatewayController:  "",
		ReconciliationTool: "",
		Options:            createDefaultOptions(),
	}
}

func createDefaultOptions() v1alpha1.Options {
	return clustertestutils.CreateDefaultSpecOptions()
}

func assertDefaultValues(t *testing.T, cluster *v1alpha1.Cluster) {
	t.Helper()
	assert.Equal(t, "ksail-default", cluster.Metadata.Name)
	assert.Equal(t, "kind.yaml", cluster.Spec.DistributionConfig)
	assert.Equal(t, "k8s", cluster.Spec.SourceDirectory)
	assert.Equal(t, v1alpha1.DistributionKind, cluster.Spec.Distribution)
	assert.Equal(t, "~/.kube/config", cluster.Spec.Connection.Kubeconfig)
	assert.Equal(t, "kind-ksail-default", cluster.Spec.Connection.Context)
	assert.Equal(t, time.Duration(5)*time.Minute, cluster.Spec.Connection.Timeout.Duration)
}

func assertCustomValues(t *testing.T, cluster *v1alpha1.Cluster) {
	t.Helper()
	assert.Equal(t, "custom", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)
	assert.Equal(t, "/custom", cluster.Spec.Connection.Kubeconfig)
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

func TestContainerEngine_Set(t *testing.T) {
	t.Parallel()

	validCases := []struct{ input, expected string }{
		{"docker", "Docker"},
		{"PODMAN", "Podman"},
	}
	for _, validCase := range validCases {
		var engine v1alpha1.ContainerEngine

		require.NoError(t, engine.Set(validCase.input))
	}

	err := func() error {
		var engine v1alpha1.ContainerEngine

		return engine.Set("invalid")
	}()
	testutils.AssertErrWrappedContains(t, err, v1alpha1.ErrInvalidContainerEngine, "invalid", "Set(invalid)")
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

	engine := v1alpha1.ContainerEngineDocker
	assert.Equal(t, "Docker", engine.String())
	assert.Equal(t, "ContainerEngine", engine.Type())
}
