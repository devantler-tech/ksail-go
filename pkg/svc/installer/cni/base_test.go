package cni_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
)

func TestInstallerBaseBuildRESTConfig(t *testing.T) {
	t.Parallel()

	t.Run("ErrorWhenKubeconfigMissing", testBuildRESTConfigErrorWhenKubeconfigMissing)
	t.Run("UsesCurrentContext", testBuildRESTConfigUsesCurrentContext)
	t.Run("OverridesContext", testBuildRESTConfigOverridesContext)
	t.Run("MissingContext", testBuildRESTConfigMissingContext)
}

func testBuildRESTConfigErrorWhenKubeconfigMissing(t *testing.T) {
	t.Helper()
	t.Parallel()

	base := cni.NewInstallerBase(helm.NewMockInterface(t), "", "", time.Second, nil)
	_, err := base.BuildRESTConfig()

	testutils.ExpectErrorContains(t, err, "kubeconfig path is empty", "buildRESTConfig empty path")
}

func testBuildRESTConfigUsesCurrentContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := testutils.WriteKubeconfig(t, t.TempDir())
	base := cni.NewInstallerBase(helm.NewMockInterface(t), path, "", time.Second, nil)

	restConfig, err := base.BuildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig current context")
	testutils.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-one.example.com",
		"rest config host",
	)
}

func testBuildRESTConfigOverridesContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := testutils.WriteKubeconfig(t, t.TempDir())
	base := cni.NewInstallerBase(helm.NewMockInterface(t), path, "alt", time.Second, nil)

	restConfig, err := base.BuildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig override context")
	testutils.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-two.example.com",
		"rest config host override",
	)
}

func testBuildRESTConfigMissingContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := testutils.WriteKubeconfig(t, t.TempDir())
	base := cni.NewInstallerBase(
		helm.NewMockInterface(t),
		path,
		"missing",
		time.Second,
		nil,
	)
	_, err := base.BuildRESTConfig()

	testutils.ExpectErrorContains(
		t,
		err,
		"context \"missing\" does not exist",
		"buildRESTConfig missing context",
	)
}
