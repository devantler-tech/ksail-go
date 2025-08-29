// Package testutils provides generic test utilities.
package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneratorMarshalError runs a generic test pattern for generator marshal errors.
func TestGeneratorMarshalError[T, M any](
	t *testing.T,
	gen generator.Generator[T, yamlgenerator.Options],
	cluster T,
	expectedErrorContains string,
) {
	t.Helper()

	// Note: gen.Marshaller should already be set to a failing marshaller before calling this function
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedErrorContains)
	assert.Empty(t, result)
}