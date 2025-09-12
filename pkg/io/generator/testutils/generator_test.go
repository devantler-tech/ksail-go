package testutils_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
)

// Generator that can have its marshaller replaced for testing
type testGeneratorWithMarshaller struct {
	Marshaller marshaller.Marshaller[testConfig]
}

func (g *testGeneratorWithMarshaller) Generate(
	config testConfig,
	opts yamlgenerator.Options,
) (string, error) {
	content, err := g.Marshaller.Marshal(config)
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

func TestTestGeneratorMarshalError(t *testing.T) {
	t.Parallel()

	t.Run("generates_marshal_error", func(t *testing.T) {
		t.Parallel()

		gen := &testGeneratorWithMarshaller{
			Marshaller: testutils.MarshalFailer[testConfig]{
				Marshaller: yamlmarshaller.NewMarshaller[testConfig](),
			},
		}
		config := testConfig{Name: "test-cluster"}

		// This tests the utility function itself
		testutils.TestGeneratorMarshalError[testConfig, testConfig](
			t,
			gen,
			config,
			"marshal failed", // expected error contains
		)
	})
}
