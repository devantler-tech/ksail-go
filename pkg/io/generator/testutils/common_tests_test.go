package testutils_test

import (
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
)

// Simple test struct for testing
type testConfig struct {
	Name string `yaml:"name"`
}

// Simple generator for testing
type testGenerator struct {
	marshaller marshaller.Marshaller[testConfig]
}

func (g *testGenerator) Generate(config testConfig, opts yamlgenerator.Options) (string, error) {
	content, err := g.marshaller.Marshal(config)
	if err != nil {
		return "", err
	}

	if opts.Output != "" {
		_, err := io.TryWriteFile(content, opts.Output, opts.Force)
		if err != nil {
			return "", err
		}
	}

	return content, nil
}

func assertTestContent(t *testing.T, content, expectedName string) {
	t.Helper()
	if !strings.Contains(content, expectedName) {
		t.Errorf("Expected content to contain %s, got: %s", expectedName, content)
	}
}

func TestTestExistingFile(t *testing.T) {
	t.Parallel()

	t.Run("with_force_true", func(t *testing.T) {
		t.Parallel()

		gen := &testGenerator{
			marshaller: yamlmarshaller.NewMarshaller[testConfig](),
		}
		config := testConfig{Name: "test-cluster"}

		// This tests the utility function itself
		testutils.TestExistingFile(
			t,
			gen,
			config,
			"test-config.yaml",
			assertTestContent,
			"test-cluster",
			true, // force = true
		)
	})

	t.Run("with_force_false", func(t *testing.T) {
		t.Parallel()

		gen := &testGenerator{
			marshaller: yamlmarshaller.NewMarshaller[testConfig](),
		}
		config := testConfig{Name: "test-cluster"}

		// This tests the utility function itself
		testutils.TestExistingFile(
			t,
			gen,
			config,
			"test-config.yaml",
			assertTestContent,
			"test-cluster",
			false, // force = false
		)
	})
}

func TestTestFileWriteError(t *testing.T) {
	t.Parallel()

	t.Run("generates_write_error", func(t *testing.T) {
		t.Parallel()

		gen := &testGenerator{
			marshaller: yamlmarshaller.NewMarshaller[testConfig](),
		}
		config := testConfig{Name: "test-cluster"}

		// This tests the utility function itself
		testutils.TestFileWriteError(
			t,
			gen,
			config,
			"test-config.yaml",
			"directory", // expected error contains (directory creation error)
		)
	})
}
