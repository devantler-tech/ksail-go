package gen_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/workload/gen"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestNewGenCmdRequiresSubcommand(t *testing.T) {
	t.Parallel()

	cmd := gen.NewGenCmd(runtime.NewRuntime())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err, "gen command should show help when run without subcommand")
}

func TestNewHelmReleaseCmdRequiresName(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err, "helm-release command should require a name argument")
}

func TestHelmReleaseCmdRequiresSourceOrChartRef(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"podinfo", "--export"})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "either --source with --chart or --chart-ref must be specified")
}

func TestHelmReleaseCmdGeneratesValidYAML(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"podinfo",
		"--source=HelmRepository/podinfo",
		"--chart=podinfo",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()

	// Check that output contains YAML
	assert.Contains(t, output, "apiVersion: helm.toolkit.fluxcd.io/v2")
	assert.Contains(t, output, "kind: HelmRelease")
	assert.Contains(t, output, "name: podinfo")

	// Validate YAML can be parsed
	var result map[string]interface{}
	lines := strings.Split(output, "\n")
	yamlLines := []string{}
	for _, line := range lines {
		if !strings.HasPrefix(line, "✔") {
			yamlLines = append(yamlLines, line)
		}
	}
	yamlContent := strings.Join(yamlLines, "\n")
	err = yaml.Unmarshal([]byte(yamlContent), &result)
	require.NoError(t, err, "generated YAML should be valid")
}

func TestHelmReleaseCmdWithSourceAndChart(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--namespace=production",
		"--source=HelmRepository/charts.flux-system",
		"--chart=webapp",
		"--chart-version=^1.0.0",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "name: webapp")
	assert.Contains(t, output, "namespace: production")
	assert.Contains(t, output, "chart: webapp")
	assert.Contains(t, output, "version: ^1.0.0")
	assert.Contains(t, output, "kind: HelmRepository")
	assert.Contains(t, output, "name: charts")
	assert.Contains(t, output, "namespace: flux-system")
}

func TestHelmReleaseCmdWithChartRef(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--chart-ref=OCIRepository/webapp.flux-system",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "chartRef:")
	assert.Contains(t, output, "kind: OCIRepository")
	assert.Contains(t, output, "name: webapp")
	assert.Contains(t, output, "namespace: flux-system")
}

func TestHelmReleaseCmdWithDependsOn(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--depends-on=database",
		"--depends-on=production/redis",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "dependsOn:")
	assert.Contains(t, output, "name: database")
	assert.Contains(t, output, "name: redis")
	assert.Contains(t, output, "namespace: production")
}

func TestHelmReleaseCmdWithValuesFile(t *testing.T) {
	t.Parallel()

	// Create a temporary values file
	tmpDir := t.TempDir()
	valuesFile := tmpDir + "/values.yaml"
	valuesContent := `
replicaCount: 3
image:
  tag: v2.0.0
`
	err := os.WriteFile(valuesFile, []byte(valuesContent), 0o600)
	require.NoError(t, err)

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--values=" + valuesFile,
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err = cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "values:")
	assert.Contains(t, output, "replicaCount: 3")
	assert.Contains(t, output, "tag: v2.0.0")
}

func TestHelmReleaseCmdWithValuesFrom(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--values-from=Secret/my-values",
		"--values-from=ConfigMap/common-config",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "valuesFrom:")
	assert.Contains(t, output, "kind: Secret")
	assert.Contains(t, output, "name: my-values")
	assert.Contains(t, output, "kind: ConfigMap")
	assert.Contains(t, output, "name: common-config")
}

func TestHelmReleaseCmdWithAllFlags(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--namespace=production",
		"--source=HelmRepository/charts.flux-system",
		"--chart=webapp",
		"--chart-version=^1.0.0",
		"--target-namespace=apps",
		"--storage-namespace=flux-system",
		"--create-target-namespace",
		"--release-name=webapp-prod",
		"--service-account=webapp-sa",
		"--crds=CreateReplace",
		"--interval=5m",
		"--timeout=10m",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "namespace: production")
	assert.Contains(t, output, "releaseName: webapp-prod")
	assert.Contains(t, output, "targetNamespace: apps")
	assert.Contains(t, output, "storageNamespace: flux-system")
	assert.Contains(t, output, "serviceAccountName: webapp-sa")
	assert.Contains(t, output, "createNamespace: true")
	assert.Contains(t, output, "interval: 5m0s")
	assert.Contains(t, output, "timeout: 10m0s")
	assert.Contains(t, output, "crds: CreateReplace")
}

func TestHelmReleaseCmdInvalidCRDsPolicy(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--crds=InvalidPolicy",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid crds policy")
}

func TestHelmReleaseCmdCannotUseBothSourceAndChartRef(t *testing.T) {
	t.Parallel()

	cmd := gen.NewHelmReleaseCmd(runtime.NewRuntime())
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--chart-ref=OCIRepository/webapp",
		"--export",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot specify both")
}

func TestHelmReleaseCmdAliases(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		alias string
	}{
		{name: "hr", alias: "hr"},
		{name: "helmrelease", alias: "helmrelease"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd := gen.NewGenCmd(runtime.NewRuntime())
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{
				tc.alias,
				"podinfo",
				"--source=HelmRepository/podinfo",
				"--chart=podinfo",
				"--export",
			})

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			t.Cleanup(cancel)
			cmd.SetContext(ctx)

			err := cmd.Execute()
			require.NoError(t, err)

			output := out.String()
			assert.Contains(t, output, "kind: HelmRelease")
		})
	}
}
