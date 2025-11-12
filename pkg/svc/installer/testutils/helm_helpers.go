package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/stretchr/testify/mock"
)

// HelmRepoExpectation configures expectations for Helm repository operations.
type HelmRepoExpectation struct {
	RepoName string
	RepoURL  string
}

// ExpectAddRepository sets up mock expectations for adding a Helm repository.
func ExpectAddRepository(
	t *testing.T,
	client *helm.MockInterface,
	expect HelmRepoExpectation,
	err error,
) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				t.Helper()
				ExpectEqual(t, entry.Name, expect.RepoName, "repository name")
				ExpectEqual(t, entry.URL, expect.RepoURL, "repository URL")

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
func ExpectInstallChart(
	t *testing.T,
	client *helm.MockInterface,
	expect HelmChartExpectation,
	err error,
) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				t.Helper()
				ExpectEqual(
					t,
					spec.ReleaseName,
					expect.ReleaseName,
					"release name",
				)
				ExpectEqual(t, spec.ChartName, expect.ChartName, "chart name")
				ExpectEqual(t, spec.Namespace, expect.Namespace, "namespace")
				ExpectEqual(t, spec.RepoURL, expect.RepoURL, "repository URL")
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

					ExpectEqual(t, actualVal, expectedVal, key)
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
