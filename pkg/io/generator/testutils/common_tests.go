// Package testutils provides generator-specific test utilities.
package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// TestExistingFile runs a common test pattern for generators with existing files.
func TestExistingFile[T any](
	t *testing.T,
	gen generator.Generator[T, yamlgenerator.Options],
	cluster T,
	filename string,
	assertContent func(*testing.T, string, string),
	clusterName string,
	force bool,
) {
	t.Helper()

	// Arrange
	tempDir, outputPath, existingContent := testutils.SetupExistingFile(t, filename)
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  force,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertContent(t, result, clusterName)

	if force {
		// Verify file was overwritten
		testutils.AssertFileEquals(t, tempDir, outputPath, result)

		// Additional check: ensure old content was replaced
		fileContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
		require.NoError(t, err, "File should exist")
		assert.NotEqual(t, existingContent, string(fileContent), "Old content should be replaced")
	} else {
		// Verify file was NOT overwritten
		testutils.AssertFileEquals(t, tempDir, outputPath, existingContent)
	}
}

// TestK3dMarshalError runs a common test pattern for K3d generator marshal errors.
func TestK3dMarshalError(
	t *testing.T,
	createCluster func(string) *v1alpha1.Cluster,
	expectedErrorContains string,
) {
	t.Helper()

	// Arrange
	gen := k3dgenerator.NewK3dGenerator()
	gen.Marshaller = testutils.MarshalFailer[*v1alpha5.SimpleConfig]{
		Marshaller: nil,
	}
	cluster := createCluster("marshal-error-cluster")
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

// TestKindMarshalError runs a common test pattern for Kind generator marshal errors.
func TestKindMarshalError(
	t *testing.T,
	createCluster func(string) *v1alpha4.Cluster,
	expectedErrorContains string,
) {
	t.Helper()

	// Arrange
	gen := kindgenerator.NewKindGenerator()
	gen.Marshaller = testutils.MarshalFailer[*v1alpha4.Cluster]{
		Marshaller: nil,
	}
	cluster := createCluster("marshal-error-cluster")
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

