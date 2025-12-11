package yamlgenerator_test

import (
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"

	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]any]()

	createCluster := func(name string) map[string]any {
		return map[string]any{"name": name}
	}

	assertContent := func(t *testing.T, result, _ string) {
		t.Helper()
		snaps.MatchSnapshot(t, result)
	}

	generatortestutils.RunStandardGeneratorTests(
		t,
		gen,
		createCluster,
		"output.yaml",
		assertContent,
	)
}

func TestGenerateWithComplexModel(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]any]()

	// Test with complex nested structure
	complexModel := map[string]any{
		"metadata": map[string]any{
			"name":      "test-cluster",
			"namespace": "default",
			"labels": map[string]string{
				"app": "ksail",
				"env": "test",
			},
		},
		"spec": map[string]any{
			"replicas": 3,
			"ports":    []int{8080, 9090},
			"config": map[string]any{
				"enabled": true,
				"timeout": "30s",
			},
		},
	}

	result, err := gen.Generate(complexModel, generator.Options{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	// Verify YAML structure is valid by checking it contains expected keys
	expectedKeys := []string{"metadata:", "spec:", "name:", "replicas:"}
	for _, key := range expectedKeys {
		if !strings.Contains(result, key) {
			t.Errorf("expected result to contain %q, but it didn't. Result: %s", key, result)
		}
	}
}

func TestGenerateWithEmptyModel(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]any]()

	// Test with empty model
	emptyModel := map[string]any{}

	result, err := gen.Generate(emptyModel, generator.Options{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "{}\n"
	if result != expected {
		t.Fatalf("expected result %q, got %q", expected, result)
	}
}

func TestGenerateWithOutputPath(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]any]()
	tempDir := t.TempDir()

	model := map[string]any{
		"test": "value",
	}

	outputPath := tempDir + "/test-output.yaml"

	result, err := gen.Generate(model, generator.Options{
		Output: outputPath,
		Force:  false,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result when writing to file")
	}

	// Verify the result contains our test data
	if !strings.Contains(result, "test: value") {
		t.Errorf("expected result to contain 'test: value', got: %s", result)
	}
}

func TestGenerateWithInvalidOutputDirectory(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]any]()

	model := map[string]any{
		"test": "value",
	}

	// Use invalid path that should cause write error
	invalidPath := "/invalid/path/that/does/not/exist/output.yaml"

	_, err := gen.Generate(model, generator.Options{
		Output: invalidPath,
		Force:  false,
	})
	if err == nil {
		t.Fatal("expected error for invalid output path, got none")
	}

	expectedErrorSubstring := "failed to write YAML to file"
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("expected error to contain %q, got: %v", expectedErrorSubstring, err)
	}
}
