package helm //nolint:testpackage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	ksailio "github.com/devantler-tech/ksail-go/pkg/io"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	helmtime "helm.sh/helm/v3/pkg/time"
)

func expectEqual[T comparable](t *testing.T, got, want T, description string) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected %s: got %v want %v", description, got, want)
	}
}

func expectDeepEqual[T any](t *testing.T, got, want T, description string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected %s: got %#v want %#v", description, got, want)
	}
}

func expectNoError(t *testing.T, err error, description string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: unexpected error: %v", description, err)
	}
}

func expectErrorContains(t *testing.T, err error, substr, description string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: expected error containing %q but got nil", description, substr)
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("%s: expected error to contain %q, got %q", description, substr, err.Error())
	}
}

var errOperationFailed = errors.New("operation failed")

func TestNewClient(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		kubeConfig  string
		kubeContext string
		wantErr     bool
	}{
		{
			name:        "empty config and context",
			kubeConfig:  "",
			kubeContext: "",
			wantErr:     false,
		},
		{
			name:        "with kubeconfig path",
			kubeConfig:  "/path/to/kubeconfig",
			kubeContext: "",
			wantErr:     false,
		},
		{
			name:        "with kubeconfig and context",
			kubeConfig:  "/path/to/kubeconfig",
			kubeContext: "test-context",
			wantErr:     false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewClient(testCase.kubeConfig, testCase.kubeContext)

			if testCase.wantErr {
				if err == nil {
					t.Fatalf("%s: expected error but got nil", testCase.name)
				}

				if client != nil {
					t.Fatalf("%s: expected client to be nil", testCase.name)
				}

				return
			}

			if err != nil {
				t.Fatalf("%s: unexpected error: %v", testCase.name, err)
			}

			if client == nil {
				t.Fatalf("%s: expected client instance", testCase.name)
			}
		})
	}
}

func TestNewClientWithDebug(t *testing.T) {
	t.Parallel()

	debugCalled := false
	debugFunc := func(string, ...interface{}) {
		debugCalled = true
	}

	client, err := NewClientWithDebug("", "", debugFunc)
	expectNoError(t, err, "NewClientWithDebug")

	if client == nil {
		t.Fatal("expected client instance")
	}

	if debugCalled {
		t.Fatal("expected debug function not to be invoked during construction")
	}
}

func TestChartSpec_DefaultValues(t *testing.T) {
	t.Parallel()

	spec := &ChartSpec{
		ReleaseName: "test-release",
		ChartName:   "test/chart",
		Namespace:   "test-namespace",
	}

	expectEqual(t, spec.ReleaseName, "test-release", "ReleaseName")
	expectEqual(t, spec.ChartName, "test/chart", "ChartName")
	expectEqual(t, spec.Namespace, "test-namespace", "Namespace")

	expectEqual(t, spec.CreateNamespace, false, "CreateNamespace")
	expectEqual(t, spec.Wait, false, "Wait")
	expectEqual(t, spec.Timeout, time.Duration(0), "Timeout")
}

func TestRepositoryEntry_Fields(t *testing.T) {
	t.Parallel()

	entry := &RepositoryEntry{
		Name:                  "test-repo",
		URL:                   "https://charts.example.com",
		Username:              "user",
		Password:              "pass",
		CertFile:              "/path/to/cert",
		KeyFile:               "/path/to/key",
		CaFile:                "/path/to/ca",
		InsecureSkipTLSverify: true,
		PlainHTTP:             true,
	}

	expectEqual(t, entry.Name, "test-repo", "Name")
	expectEqual(t, entry.URL, "https://charts.example.com", "URL")
	expectEqual(t, entry.Username, "user", "Username")
	expectEqual(t, entry.Password, "pass", "Password")
	expectEqual(t, entry.CertFile, "/path/to/cert", "CertFile")
	expectEqual(t, entry.KeyFile, "/path/to/key", "KeyFile")
	expectEqual(t, entry.CaFile, "/path/to/ca", "CaFile")
	expectEqual(t, entry.InsecureSkipTLSverify, true, "InsecureSkipTLSverify")
	expectEqual(t, entry.PlainHTTP, true, "PlainHTTP")
}

func TestReleaseInfo_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	info := &ReleaseInfo{
		Name:       "test-release",
		Namespace:  "test-namespace",
		Revision:   2,
		Status:     "deployed",
		Chart:      "nginx-1.0.0",
		AppVersion: "1.20.1",
		Updated:    now,
		Notes:      "Release notes here",
	}

	expectEqual(t, info.Name, "test-release", "Name")
	expectEqual(t, info.Namespace, "test-namespace", "Namespace")
	expectEqual(t, info.Revision, 2, "Revision")
	expectEqual(t, info.Status, "deployed", "Status")
	expectEqual(t, info.Chart, "nginx-1.0.0", "Chart")
	expectEqual(t, info.AppVersion, "1.20.1", "AppVersion")
	expectEqual(t, info.Updated, now, "Updated")
	expectEqual(t, info.Notes, "Release notes here", "Notes")
}

func TestChartSpec_WithTimeout(t *testing.T) {
	t.Parallel()

	spec := &ChartSpec{
		ReleaseName: "test-release",
		ChartName:   "test/chart",
		Namespace:   "test-namespace",
		Timeout:     5 * time.Minute,
	}

	expectEqual(t, spec.Timeout, 5*time.Minute, "Timeout")
}

func TestChartSpec_WithValues(t *testing.T) {
	t.Parallel()

	spec := &ChartSpec{
		ReleaseName: "test-release",
		ChartName:   "test/chart",
		Namespace:   "test-namespace",
		ValuesYaml:  "replicas: 3\nimage: nginx:latest",
		SetValues: map[string]string{
			"service.type": "LoadBalancer",
			"image.tag":    "1.20",
		},
		ValueFiles: []string{"/path/to/values.yaml", "/path/to/override.yaml"},
	}

	expectEqual(t, spec.ValuesYaml, "replicas: 3\nimage: nginx:latest", "ValuesYaml")
	expectEqual(t, spec.SetValues["service.type"], "LoadBalancer", "SetValues service.type")
	expectEqual(t, spec.SetValues["image.tag"], "1.20", "SetValues image.tag")
	expectEqual(t, len(spec.ValueFiles), 2, "ValueFiles length")

	if !contains(spec.ValueFiles, "/path/to/values.yaml") {
		t.Fatalf("ValueFiles missing /path/to/values.yaml: %#v", spec.ValueFiles)
	}

	if !contains(spec.ValueFiles, "/path/to/override.yaml") {
		t.Fatalf("ValueFiles missing /path/to/override.yaml: %#v", spec.ValueFiles)
	}
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}

	return false
}

func TestChartSpec_InstallOptions(t *testing.T) {
	t.Parallel()

	spec := &ChartSpec{
		ReleaseName:     "test-release",
		ChartName:       "test/chart",
		Namespace:       "test-namespace",
		CreateNamespace: true,
		UpgradeCRDs:     true,
		Atomic:          true,
		Wait:            true,
		WaitForJobs:     true,
		Silent:          true,
	}

	expectEqual(t, spec.CreateNamespace, true, "CreateNamespace")
	expectEqual(t, spec.UpgradeCRDs, true, "UpgradeCRDs")
	expectEqual(t, spec.Atomic, true, "Atomic")
	expectEqual(t, spec.Wait, true, "Wait")
	expectEqual(t, spec.WaitForJobs, true, "WaitForJobs")
	expectEqual(t, spec.Silent, true, "Silent")
}

func TestChartSpec_RepositoryOptions(t *testing.T) {
	t.Parallel()

	spec := &ChartSpec{
		ReleaseName:           "test-release",
		ChartName:             "test/chart",
		Namespace:             "test-namespace",
		Version:               "1.2.3",
		RepoURL:               "https://charts.example.com",
		Username:              "user",
		Password:              "pass",
		CertFile:              "/path/to/cert",
		KeyFile:               "/path/to/key",
		CaFile:                "/path/to/ca",
		InsecureSkipTLSverify: true,
	}

	expectEqual(t, spec.Version, "1.2.3", "Version")
	expectEqual(t, spec.RepoURL, "https://charts.example.com", "RepoURL")
	expectEqual(t, spec.Username, "user", "Username")
	expectEqual(t, spec.Password, "pass", "Password")
	expectEqual(t, spec.CertFile, "/path/to/cert", "CertFile")
	expectEqual(t, spec.KeyFile, "/path/to/key", "KeyFile")
	expectEqual(t, spec.CaFile, "/path/to/ca", "CaFile")
	expectEqual(t, spec.InsecureSkipTLSverify, true, "InsecureSkipTLSverify")
}

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()

	expectEqual(t, DefaultTimeout, 5*time.Minute, "DefaultTimeout")
}

func TestHelmClientInterface(t *testing.T) {
	t.Parallel()

	var _ Interface = (*Client)(nil)
}

func TestHelmClientContextSupport(t *testing.T) {
	t.Parallel()

	client, err := NewClient("", "")
	expectNoError(t, err, "NewClient")

	ctx := context.Background()
	spec := &ChartSpec{ReleaseName: "test", ChartName: "test/chart", Namespace: "default"}

	_, _ = client.InstallChart(ctx, spec)
	_, _ = client.InstallOrUpgradeChart(ctx, spec)
	_ = client.UninstallRelease(ctx, "test", "default")

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = client.AddRepository(
		canceledCtx,
		&RepositoryEntry{Name: "test", URL: "https://example.com"},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled error, got %v", err)
	}
}

//nolint:paralleltest // uses process-wide environment variables.
func TestClientAddRepositorySuccess(t *testing.T) {
	repoCache, repoConfig := setupHelmRepoEnv(t)

	server := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path == "/index.yaml" {
				_, _ = writer.Write([]byte("apiVersion: v1\nentries: {}\n"))

				return
			}

			http.NotFound(writer, request)
		}),
	)
	defer server.Close()

	client, err := NewClient("", "")
	expectNoError(t, err, "NewClient")

	entry := &RepositoryEntry{Name: "cilium", URL: server.URL}
	addErr := client.AddRepository(context.Background(), entry)
	expectNoError(t, addErr, "AddRepository")

	indexPath := filepath.Join(repoCache, "cilium-index.yaml")

	_, err = os.Stat(indexPath)
	if err != nil {
		t.Fatalf("expected repository index at %s: %v", indexPath, err)
	}

	configData, err := ksailio.ReadFileSafe(filepath.Dir(repoConfig), repoConfig)
	expectNoError(t, err, "ReadFileSafe")

	if !strings.Contains(string(configData), server.URL) {
		t.Fatalf("repository config missing server URL: %s", server.URL)
	}
}

//nolint:paralleltest // uses process-wide environment variables.
func TestClientAddRepositoryDownloadFailure(t *testing.T) {
	setupHelmRepoEnv(t)

	server := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			http.Error(writer, "error", http.StatusInternalServerError)
		}),
	)
	defer server.Close()

	client, err := NewClient("", "")

	expectNoError(t, err, "NewClient")

	err = client.AddRepository(
		context.Background(),
		&RepositoryEntry{Name: "cilium", URL: server.URL},
	)

	expectErrorContains(t, err, "failed to download repository index file", "AddRepository failure")
}

func TestParseChartRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		ref           string
		expectedRepo  string
		expectedChart string
	}{
		{name: "ChartOnly", ref: "nginx", expectedRepo: "", expectedChart: "nginx"},
		{
			name:          "RepositoryAndChart",
			ref:           "stable/nginx",
			expectedRepo:  "stable",
			expectedChart: "nginx",
		},
		{
			name:          "OnlySplitsFirstSlash",
			ref:           "stable/nested",
			expectedRepo:  "stable",
			expectedChart: "nested",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			repo, chart := parseChartRef(testCase.ref)

			expectEqual(t, repo, testCase.expectedRepo, "repo")
			expectEqual(t, chart, testCase.expectedChart, "chart")
		})
	}
}

func TestReleaseToInfo(t *testing.T) {
	t.Parallel()

	t.Run("NilRelease", func(t *testing.T) {
		t.Parallel()

		if releaseToInfo(nil) != nil {
			t.Fatal("expected nil result for nil release")
		}
	})

	t.Run("PopulatedRelease", func(t *testing.T) {
		t.Parallel()

		timestamp := time.Now()
		rel := &release.Release{
			Name:      "demo",
			Namespace: "default",
			Version:   3,
			Chart: &chart.Chart{Metadata: &chart.Metadata{
				Name:       "demo-chart",
				AppVersion: "1.2.3",
			}},
			Info: &release.Info{
				Status:       release.StatusDeployed,
				LastDeployed: helmtime.Time{Time: timestamp},
				Notes:        "deployment notes",
			},
		}

		info := releaseToInfo(rel)

		if info == nil {
			t.Fatal("expected release info for populated release")
		}

		expectEqual(t, info.Name, "demo", "Name")
		expectEqual(t, info.Namespace, "default", "Namespace")
		expectEqual(t, info.Revision, 3, "Revision")
		expectEqual(t, info.Status, release.StatusDeployed.String(), "Status")
		expectEqual(t, info.Chart, "demo-chart", "Chart")
		expectEqual(t, info.AppVersion, "1.2.3", "AppVersion")
		expectEqual(t, info.Updated, timestamp, "Updated")
		expectEqual(t, info.Notes, "deployment notes", "Notes")
	})
}

func TestConvertMapToSlice(t *testing.T) {
	t.Parallel()

	t.Run("NilWhenEmpty", func(t *testing.T) {
		t.Parallel()

		if convertMapToSlice(nil) != nil {
			t.Fatal("expected nil for nil map")
		}

		if convertMapToSlice(map[string]string{}) != nil {
			t.Fatal("expected nil for empty map")
		}
	})

	t.Run("SortedKeyValuePairs", func(t *testing.T) {
		t.Parallel()

		values := map[string]string{"beta": "2", "alpha": "1"}
		result := convertMapToSlice(values)

		expectDeepEqual(t, result, []string{"alpha=1", "beta=2"}, "sorted key/value slice")
	})
}

func TestCopyStringSlice(t *testing.T) {
	t.Parallel()

	t.Run("NilWhenEmpty", func(t *testing.T) {
		t.Parallel()

		if copyStringSlice(nil) != nil {
			t.Fatal("expected nil for nil slice")
		}

		if copyStringSlice([]string{}) != nil {
			t.Fatal("expected nil for empty slice")
		}
	})

	t.Run("IndependentCopy", func(t *testing.T) {
		t.Parallel()

		original := []string{"one", "two"}
		clone := copyStringSlice(original)

		expectDeepEqual(t, clone, original, "cloned slice equality")

		clone[0] = "changed"

		expectEqual(t, original[0], "one", "original unchanged")
		expectEqual(t, clone[0], "changed", "clone mutated")
	})
}

func TestRunReleaseWithSilencedStderr(t *testing.T) {
	t.Parallel()

	t.Run("SuccessReturnsRelease", func(t *testing.T) {
		t.Parallel()

		releaseResult := &release.Release{Name: "success"}
		originalStderr := os.Stderr

		result, err := runReleaseWithSilencedStderr(func() (*release.Release, error) {
			fmt.Fprintln(os.Stderr, "ignored log")

			return releaseResult, nil
		})

		expectNoError(t, err, "runReleaseWithSilencedStderr success")

		if result != releaseResult {
			t.Fatalf("expected release result %v, got %v", releaseResult, result)
		}

		if os.Stderr != originalStderr {
			t.Fatalf("expected os.Stderr to be restored")
		}
	})

	t.Run("ErrorIncludesCapturedLogs", func(t *testing.T) {
		t.Parallel()

		originalStderr := os.Stderr

		_, err := runReleaseWithSilencedStderr(func() (*release.Release, error) {
			fmt.Fprintln(os.Stderr, "detailed failure")

			return nil, errOperationFailed
		})

		expectErrorContains(
			t,
			err,
			errOperationFailed.Error(),
			"runReleaseWithSilencedStderr error",
		)

		if !errors.Is(err, errOperationFailed) {
			t.Fatalf("expected wrapped error to match original: %v", err)
		}

		if os.Stderr != originalStderr {
			t.Fatalf("expected os.Stderr to be restored")
		}
	})
}

func setupHelmRepoEnv(t *testing.T) (string, string) {
	t.Helper()

	tempDir := t.TempDir()
	repoCache := filepath.Join(tempDir, "cache")
	repoConfig := filepath.Join(tempDir, "repositories.yaml")

	t.Setenv("HELM_REPOSITORY_CACHE", repoCache)
	t.Setenv("HELM_REPOSITORY_CONFIG", repoConfig)
	t.Setenv("HELM_CACHE_HOME", tempDir)
	t.Setenv("HELM_CONFIG_HOME", tempDir)
	t.Setenv("HELM_DATA_HOME", tempDir)

	return repoCache, repoConfig
}
