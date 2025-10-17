package helm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	ksailio "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
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

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client, err := helm.NewClient(testCase.kubeConfig, testCase.kubeContext)

			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestNewClientWithDebug(t *testing.T) {
	t.Parallel()

	debugCalled := false
	debugFunc := func(_ string, _ ...interface{}) {
		debugCalled = true
	}

	client, err := helm.NewClientWithDebug("", "", debugFunc)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Debug function should be set but not called during creation
	assert.False(t, debugCalled)
}

func TestChartSpec_DefaultValues(t *testing.T) {
	t.Parallel()

	spec := &helm.ChartSpec{
		ReleaseName: "test-release",
		ChartName:   "test/chart",
		Namespace:   "test-namespace",
	}

	// Test default values
	assert.Equal(t, "test-release", spec.ReleaseName)
	assert.Equal(t, "test/chart", spec.ChartName)
	assert.Equal(t, "test-namespace", spec.Namespace)
	assert.False(t, spec.CreateNamespace)
	assert.False(t, spec.Wait)
	assert.Equal(t, time.Duration(0), spec.Timeout)
}

func TestRepositoryEntry_Fields(t *testing.T) {
	t.Parallel()

	entry := &helm.RepositoryEntry{
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

	assert.Equal(t, "test-repo", entry.Name)
	assert.Equal(t, "https://charts.example.com", entry.URL)
	assert.Equal(t, "user", entry.Username)
	assert.Equal(t, "pass", entry.Password)
	assert.Equal(t, "/path/to/cert", entry.CertFile)
	assert.Equal(t, "/path/to/key", entry.KeyFile)
	assert.Equal(t, "/path/to/ca", entry.CaFile)
	assert.True(t, entry.InsecureSkipTLSverify)
	assert.True(t, entry.PlainHTTP)
}

func TestReleaseInfo_Fields(t *testing.T) {
	t.Parallel()

	now := time.Now()
	info := &helm.ReleaseInfo{
		Name:       "test-release",
		Namespace:  "test-namespace",
		Revision:   2,
		Status:     "deployed",
		Chart:      "nginx-1.0.0",
		AppVersion: "1.20.1",
		Updated:    now,
		Notes:      "Release notes here",
	}

	assert.Equal(t, "test-release", info.Name)
	assert.Equal(t, "test-namespace", info.Namespace)
	assert.Equal(t, 2, info.Revision)
	assert.Equal(t, "deployed", info.Status)
	assert.Equal(t, "nginx-1.0.0", info.Chart)
	assert.Equal(t, "1.20.1", info.AppVersion)
	assert.Equal(t, now, info.Updated)
	assert.Equal(t, "Release notes here", info.Notes)
}

// Integration-style tests would require a real Kubernetes cluster
// For unit tests, we focus on the interface contracts and basic functionality

func TestChartSpec_WithTimeout(t *testing.T) {
	t.Parallel()

	spec := &helm.ChartSpec{
		ReleaseName: "test-release",
		ChartName:   "test/chart",
		Namespace:   "test-namespace",
		Timeout:     5 * time.Minute,
	}

	assert.Equal(t, 5*time.Minute, spec.Timeout)
}

func TestChartSpec_WithValues(t *testing.T) {
	t.Parallel()

	spec := &helm.ChartSpec{
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

	assert.Equal(t, "replicas: 3\nimage: nginx:latest", spec.ValuesYaml)
	assert.Equal(t, "LoadBalancer", spec.SetValues["service.type"])
	assert.Equal(t, "1.20", spec.SetValues["image.tag"])
	assert.Len(t, spec.ValueFiles, 2)
	assert.Contains(t, spec.ValueFiles, "/path/to/values.yaml")
	assert.Contains(t, spec.ValueFiles, "/path/to/override.yaml")
}

func TestChartSpec_InstallOptions(t *testing.T) {
	t.Parallel()

	spec := &helm.ChartSpec{
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

	assert.True(t, spec.CreateNamespace)
	assert.True(t, spec.UpgradeCRDs)
	assert.True(t, spec.Atomic)
	assert.True(t, spec.Wait)
	assert.True(t, spec.WaitForJobs)
	assert.True(t, spec.Silent)
}

func TestChartSpec_RepositoryOptions(t *testing.T) {
	t.Parallel()

	spec := &helm.ChartSpec{
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

	assert.Equal(t, "1.2.3", spec.Version)
	assert.Equal(t, "https://charts.example.com", spec.RepoURL)
	assert.Equal(t, "user", spec.Username)
	assert.Equal(t, "pass", spec.Password)
	assert.Equal(t, "/path/to/cert", spec.CertFile)
	assert.Equal(t, "/path/to/key", spec.KeyFile)
	assert.Equal(t, "/path/to/ca", spec.CaFile)
	assert.True(t, spec.InsecureSkipTLSverify)
}

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 5*time.Minute, helm.DefaultTimeout)
}

// Test interface compliance.
func TestHelmClientInterface(t *testing.T) {
	t.Parallel()

	// This test ensures that our Client type implements the helm.Interface interface
	var _ helm.Interface = (*helm.Client)(nil)
}

// Test context support in interface methods.
func TestHelmClientContextSupport(t *testing.T) {
	t.Parallel()

	// Create a client (won't be used for actual operations)
	client, err := helm.NewClient("", "")
	require.NoError(t, err)

	// Test that methods accept context (interface compliance)
	ctx := context.Background()
	spec := &helm.ChartSpec{
		ReleaseName: "test",
		ChartName:   "test/chart",
		Namespace:   "default",
	}

	// These would fail in actual execution without a Kubernetes cluster,
	// but they verify the interface contract
	_, _ = client.InstallChart(ctx, spec)
	_, _ = client.InstallOrUpgradeChart(ctx, spec)
	_ = client.UninstallRelease(ctx, "test", "default")

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = client.AddRepository(
		canceledCtx,
		&helm.RepositoryEntry{Name: "test", URL: "https://example.com"},
	)
	assert.ErrorIs(t, err, context.Canceled)
}

//nolint:paralleltest // uses process-wide env variables
func TestClientAddRepositorySuccess(
	t *testing.T,
) {
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

	client, err := helm.NewClient("", "")
	require.NoError(t, err)

	entry := &helm.RepositoryEntry{Name: "cilium", URL: server.URL}
	err = client.AddRepository(context.Background(), entry)
	require.NoError(t, err)

	indexPath := filepath.Join(repoCache, "cilium-index.yaml")
	_, err = os.Stat(indexPath)
	require.NoError(t, err)

	configData, err := ksailio.ReadFileSafe(filepath.Dir(repoConfig), repoConfig)
	require.NoError(t, err)
	assert.Contains(t, string(configData), server.URL)
}

//nolint:paralleltest // uses process-wide env variables
func TestClientAddRepositoryDownloadFailure(
	t *testing.T,
) {
	setupHelmRepoEnv(t)

	server := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			http.Error(writer, "error", http.StatusInternalServerError)
		}),
	)
	defer server.Close()

	client, err := helm.NewClient("", "")
	require.NoError(t, err)

	err = client.AddRepository(
		context.Background(),
		&helm.RepositoryEntry{Name: "cilium", URL: server.URL},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download repository index file")
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
