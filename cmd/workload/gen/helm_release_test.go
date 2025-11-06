package gen //nolint:testpackage // Tests need access to unexported newTestRuntime

import (
	"bytes"
	"os"
	"strings"
	"testing"

	testutils "github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

// TestGenHelmReleaseSimple tests generating a simple HelmRelease manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenHelmReleaseSimple(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewHelmReleaseCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"podinfo",
		"--source=HelmRepository/podinfo",
		"--chart=podinfo",
		"--export",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen helmrelease to succeed, got %v", err)
	}

	// Remove timing information from output for consistent snapshots
	output := removeTimingInfo(buffer.String())
	snaps.MatchSnapshot(t, output)
}

// TestGenHelmReleaseWithVersion tests generating a HelmRelease with a specific chart version.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenHelmReleaseWithVersion(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewHelmReleaseCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"webapp",
		"--namespace=production",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--chart-version=^1.0.0",
		"--export",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen helmrelease with version to succeed, got %v", err)
	}

	output := removeTimingInfo(buffer.String())
	snaps.MatchSnapshot(t, output)
}

// TestGenHelmReleaseWithChartRef tests generating a HelmRelease using a chart reference.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenHelmReleaseWithChartRef(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewHelmReleaseCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"webapp",
		"--chart-ref=OCIRepository/webapp.flux-system",
		"--export",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen helmrelease with chart-ref to succeed, got %v", err)
	}

	output := removeTimingInfo(buffer.String())
	snaps.MatchSnapshot(t, output)
}

// TestGenHelmReleaseWithDependencies tests generating a HelmRelease with dependencies.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenHelmReleaseWithDependencies(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewHelmReleaseCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--depends-on=database",
		"--depends-on=production/redis",
		"--export",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen helmrelease with dependencies to succeed, got %v", err)
	}

	output := removeTimingInfo(buffer.String())
	snaps.MatchSnapshot(t, output)
}

// TestGenHelmReleaseWithValuesFile tests generating a HelmRelease with values from a file.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenHelmReleaseWithValuesFile(t *testing.T) {
	// Create a temporary values file
	tmpDir := t.TempDir()
	valuesFile := tmpDir + "/values.yaml"
	valuesContent := `replicaCount: 3
image:
  tag: v2.0.0
`
	err := os.WriteFile(valuesFile, []byte(valuesContent), 0o600)
	if err != nil {
		t.Fatalf("failed to create values file: %v", err)
	}

	rt := newTestRuntime()
	cmd := NewHelmReleaseCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--values=" + valuesFile,
		"--export",
	})

	err = cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen helmrelease with values file to succeed, got %v", err)
	}

	output := removeTimingInfo(buffer.String())
	snaps.MatchSnapshot(t, output)
}

// TestGenHelmReleaseWithValuesFrom tests generating a HelmRelease with values from ConfigMap/Secret.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenHelmReleaseWithValuesFrom(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewHelmReleaseCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{
		"webapp",
		"--source=HelmRepository/charts",
		"--chart=webapp",
		"--values-from=Secret/my-values",
		"--values-from=ConfigMap/common-config",
		"--export",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen helmrelease with values-from to succeed, got %v", err)
	}

	output := removeTimingInfo(buffer.String())
	snaps.MatchSnapshot(t, output)
}

// TestGenHelmReleaseWithAllFlags tests generating a HelmRelease with multiple flags.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenHelmReleaseWithAllFlags(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewHelmReleaseCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
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

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen helmrelease with all flags to succeed, got %v", err)
	}

	output := removeTimingInfo(buffer.String())
	snaps.MatchSnapshot(t, output)
}

// removeTimingInfo removes timing output from the command output for consistent snapshots.
func removeTimingInfo(output string) string {
	lines := strings.Split(output, "\n")
	var filtered []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "✔") {
			filtered = append(filtered, line)
		}
	}
	return strings.Join(filtered, "\n")
}
