package helm

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/stretchr/testify/mock"
)

// RepoExpectation configures expectations for Helm repository operations.
type RepoExpectation struct {
	RepoName string
	RepoURL  string
}

// ExpectAddRepository sets up mock expectations for adding a Helm repository.
func ExpectAddRepository(
	t *testing.T,
	client *MockInterface,
	expect RepoExpectation,
	err error,
) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *RepositoryEntry) bool {
				t.Helper()
				testutils.ExpectEqual(t, entry.Name, expect.RepoName, "repository name")
				testutils.ExpectEqual(t, entry.URL, expect.RepoURL, "repository URL")

				return true
			}),
		).
		Return(err)
}

// ChartExpectation configures expectations for Helm chart operations.
type ChartExpectation struct {
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
	client *MockInterface,
	expect ChartExpectation,
	err error,
) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *ChartSpec) bool {
				t.Helper()
				testutils.ExpectEqual(
					t,
					spec.ReleaseName,
					expect.ReleaseName,
					"release name",
				)
				testutils.ExpectEqual(t, spec.ChartName, expect.ChartName, "chart name")
				testutils.ExpectEqual(t, spec.Namespace, expect.Namespace, "namespace")
				testutils.ExpectEqual(t, spec.RepoURL, expect.RepoURL, "repository URL")
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

					testutils.ExpectEqual(t, actualVal, expectedVal, key)
				}

				return true
			}),
		).
		Return(nil, err)
}

// ExpectUninstall sets up mock expectations for uninstalling a Helm release.
func ExpectUninstall(
	t *testing.T,
	client *MockInterface,
	releaseName, namespace string,
	err error,
) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, releaseName, namespace).
		Return(err)
}
