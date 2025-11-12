package installer_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/svc/installer"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/devantler-tech/ksail-go/pkg/testutils/cnihelpers"
)

func TestCNIInstallerBaseBuildRESTConfig(t *testing.T) {
	t.Parallel()

	t.Run("ErrorWhenKubeconfigMissing", testBuildRESTConfigErrorWhenKubeconfigMissing)
	t.Run("UsesCurrentContext", testBuildRESTConfigUsesCurrentContext)
	t.Run("OverridesContext", testBuildRESTConfigOverridesContext)
	t.Run("MissingContext", testBuildRESTConfigMissingContext)
}

func testBuildRESTConfigErrorWhenKubeconfigMissing(t *testing.T) {
	t.Helper()
	t.Parallel()

	base := installer.NewCNIInstallerBase(helm.NewMockInterface(t), "", "", time.Second, nil)
	_, err := base.BuildRESTConfig()

	testutils.ExpectErrorContains(t, err, "kubeconfig path is empty", "buildRESTConfig empty path")
}

func testBuildRESTConfigUsesCurrentContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := cnihelpers.WriteKubeconfig(t, t.TempDir())
	base := installer.NewCNIInstallerBase(helm.NewMockInterface(t), path, "", time.Second, nil)

	restConfig, err := base.BuildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig current context")
	cnihelpers.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-one.example.com",
		"rest config host",
	)
}

func testBuildRESTConfigOverridesContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := cnihelpers.WriteKubeconfig(t, t.TempDir())
	base := installer.NewCNIInstallerBase(helm.NewMockInterface(t), path, "alt", time.Second, nil)

	restConfig, err := base.BuildRESTConfig()

	testutils.ExpectNoError(t, err, "buildRESTConfig override context")
	cnihelpers.ExpectEqual(
		t,
		restConfig.Host,
		"https://cluster-two.example.com",
		"rest config host override",
	)
}

func testBuildRESTConfigMissingContext(t *testing.T) {
	t.Helper()
	t.Parallel()

	path := cnihelpers.WriteKubeconfig(t, t.TempDir())
	base := installer.NewCNIInstallerBase(
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
