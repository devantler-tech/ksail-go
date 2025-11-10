// Package cnitesthelpers provides shared test utilities for CNI installer tests.
package cnitesthelpers

import (
	"context"
	"reflect"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	installertestutils "github.com/devantler-tech/ksail-go/pkg/svc/installer/testutils"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/stretchr/testify/mock"
)

// CNIInstaller defines the minimal interface needed for testing CNI installers.
type CNIInstaller interface {
	Install(context.Context) error
	Uninstall(context.Context) error
	WaitForReadiness(context.Context) error
	SetWaitForReadinessFunc(func(context.Context) error)
	GetWaitFn() func(context.Context) error
	SetWaitFn(func(context.Context) error)
}

// InstallerScenario defines a test scenario for installer operations.
type InstallerScenario[T CNIInstaller] struct {
	Name       string
	Setup      func(*testing.T, *helm.MockInterface)
	ActionName string
	Action     func(context.Context, T) error
	WantErr    string
}

// RunInstallerScenarios runs a set of installer test scenarios.
func RunInstallerScenarios[T CNIInstaller](
	t *testing.T,
	scenarios []InstallerScenario[T],
	newInstaller func(*testing.T) (T, *helm.MockInterface),
) {
	t.Helper()

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel()

			installer, client := newInstaller(t)
			scenario.Setup(t, client)

			err := scenario.Action(context.Background(), installer)

			installertestutils.ExpectInstallerResult(t, err, scenario.WantErr, scenario.ActionName)
		})
	}
}

// TestSetWaitForReadinessFunc tests the SetWaitForReadinessFunc method for any CNI installer.
func TestSetWaitForReadinessFunc[T CNIInstaller](
	t *testing.T,
	newInstaller func(*testing.T) T,
) {
	t.Helper()

	t.Run("InvokesCustomFunction", func(t *testing.T) {
		t.Parallel()

		installer := newInstaller(t)
		called := false

		installer.SetWaitForReadinessFunc(func(context.Context) error {
			called = true

			return nil
		})

		testutils.ExpectNoError(
			t,
			installer.WaitForReadiness(context.Background()),
			"WaitForReadiness with custom func",
		)
		testutils.ExpectTrue(t, called, "custom wait function invocation")
	})

	t.Run("RestoresDefaultWhenNil", func(t *testing.T) {
		t.Parallel()

		installer := newInstaller(t)
		defaultFn := installer.GetWaitFn()
		testutils.ExpectNotNil(t, defaultFn, "default wait function")
		defaultPtr := reflect.ValueOf(defaultFn).Pointer()

		installer.SetWaitForReadinessFunc(func(context.Context) error { return nil })

		replacedPtr := reflect.ValueOf(installer.GetWaitFn()).Pointer()
		if replacedPtr == defaultPtr {
			t.Fatal("expected custom wait function to replace default")
		}

		installer.SetWaitForReadinessFunc(nil)
		restoredPtr := reflect.ValueOf(installer.GetWaitFn()).Pointer()
		installertestutils.ExpectEqual(
			t,
			restoredPtr,
			defaultPtr,
			"wait function pointer after restore",
		)
	})
}

// TestWaitForReadinessNoOpWhenUnset tests behavior when wait function is unset.
func TestWaitForReadinessNoOpWhenUnset[T CNIInstaller](
	t *testing.T,
	newInstaller func(*testing.T) T,
) {
	t.Helper()

	installer := newInstaller(t)
	installer.SetWaitFn(nil)

	err := installer.WaitForReadiness(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when waitFn unset, got %v", err)
	}
}

// HelmRepoExpectation configures expectations for Helm repository operations.
type HelmRepoExpectation struct {
	RepoName string
	RepoURL  string
}

// ExpectAddRepository sets up mock expectations for adding a Helm repository.
func ExpectAddRepository(t *testing.T, client *helm.MockInterface, expect HelmRepoExpectation, err error) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				t.Helper()
				installertestutils.ExpectEqual(t, entry.Name, expect.RepoName, "repository name")
				installertestutils.ExpectEqual(t, entry.URL, expect.RepoURL, "repository URL")

				return true
			}),
		).
		Return(err)
}

// HelmChartExpectation configures expectations for Helm chart operations.
type HelmChartExpectation struct {
	ReleaseName     string
	ChartName       string
	Namespace       string
	RepoURL         string
	CreateNamespace bool
	SetJSONVals     map[string]string
}

// ExpectInstallChart sets up mock expectations for installing a Helm chart.
func ExpectInstallChart(t *testing.T, client *helm.MockInterface, expect HelmChartExpectation, err error) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				t.Helper()
				installertestutils.ExpectEqual(t, spec.ReleaseName, expect.ReleaseName, "release name")
				installertestutils.ExpectEqual(t, spec.ChartName, expect.ChartName, "chart name")
				installertestutils.ExpectEqual(t, spec.Namespace, expect.Namespace, "namespace")
				installertestutils.ExpectEqual(t, spec.RepoURL, expect.RepoURL, "repository URL")
				testutils.ExpectTrue(t, spec.Wait, "Wait flag")
				testutils.ExpectTrue(t, spec.WaitForJobs, "WaitForJobs flag")

				if expect.CreateNamespace {
					testutils.ExpectTrue(t, spec.CreateNamespace, "CreateNamespace flag")
				}

				for key, expectedVal := range expect.SetJSONVals {
					actualVal, ok := spec.SetJSONVals[key]
					if !ok {
						t.Fatalf("expected SetJSONVals[%s] to exist", key)
					}
					installertestutils.ExpectEqual(t, actualVal, expectedVal, key)
				}

				return true
			}),
		).
		Return(nil, err)
}

// ExpectUninstall sets up mock expectations for uninstalling a Helm release.
func ExpectUninstall(
	t *testing.T,
	client *helm.MockInterface,
	releaseName, namespace string,
	err error,
) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, releaseName, namespace).
		Return(err)
}
