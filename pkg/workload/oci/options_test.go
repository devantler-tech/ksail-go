package oci_test
//nolint:cyclop,funlen,errcheck // complex validation logic acceptable in tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/workload/oci"
	"github.com/stretchr/testify/require"
)

//nolint:funlen // comprehensive validation test cases
func TestBuildOptionsValidate(t *testing.T) {
	t.Parallel()

	t.Run("requires source path", func(t *testing.T) {
		t.Parallel()

		opts := oci.BuildOptions{}

		_, err := opts.Validate()

		require.ErrorIs(t, err, oci.ErrSourcePathRequired)
	})

	t.Run("fails when source path missing", func(t *testing.T) {
		t.Parallel()

		opts := oci.BuildOptions{SourcePath: filepath.Join(t.TempDir(), "missing"), RegistryEndpoint: "localhost:5000", Version: "1.0.0"}

		_, err := opts.Validate()

		require.ErrorIs(t, err, oci.ErrSourcePathNotFound)
	})

	t.Run("fails when source path is file", func(t *testing.T) {
		t.Parallel()

		file := filepath.Join(t.TempDir(), "manifest.yaml")
		require.NoError(t, os.WriteFile(file, []byte("apiVersion: v1"), 0o600))

		opts := oci.BuildOptions{SourcePath: file, RegistryEndpoint: "localhost:5000", Version: "1.0.0"}

		_, err := opts.Validate()

		require.ErrorIs(t, err, oci.ErrSourcePathNotDirectory)
	})

	t.Run("requires registry endpoint", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		opts := oci.BuildOptions{SourcePath: tempDir, Version: "1.0.0"}

		_, err := opts.Validate()

		require.ErrorIs(t, err, oci.ErrRegistryEndpointRequired)
	})

	t.Run("requires version", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		opts := oci.BuildOptions{SourcePath: tempDir, RegistryEndpoint: "localhost:5000"}

		_, err := opts.Validate()

		require.ErrorIs(t, err, oci.ErrVersionRequired)
	})

	t.Run("requires semantic version", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		opts := oci.BuildOptions{SourcePath: tempDir, RegistryEndpoint: "localhost:5000", Version: "invalid"}

		_, err := opts.Validate()

		require.ErrorIs(t, err, oci.ErrVersionInvalid)
	})

	t.Run("allows latest tag", func(t *testing.T) {
		t.Parallel()

		source := filepath.Join(t.TempDir(), "k8s")
		require.NoError(t, os.MkdirAll(source, 0o750))

		opts := oci.BuildOptions{SourcePath: source, RegistryEndpoint: "localhost:5000", Version: "latest"}

		validated, err := opts.Validate()

		require.NoError(t, err)
		require.Equal(t, "latest", validated.Version)
	})

	t.Run("applies defaults", func(t *testing.T) {
		t.Parallel()

		source := filepath.Join(t.TempDir(), "k8s")
		require.NoError(t, os.MkdirAll(source, 0o750))

		opts := oci.BuildOptions{SourcePath: source, RegistryEndpoint: "localhost:5000", Version: "1.0.0"}

		validated, err := opts.Validate()

		require.NoError(t, err)
		require.Equal(t, filepath.Clean(source), validated.SourcePath)
		require.Equal(t, "localhost:5000", validated.RegistryEndpoint)
		require.Equal(t, "1.0.0", validated.Version)
		require.NotEmpty(t, validated.Repository)
	})

	t.Run("normalizes repository name", func(t *testing.T) {
		t.Parallel()

		source := filepath.Join(t.TempDir(), "my App")
		require.NoError(t, os.MkdirAll(source, 0o750))

		opts := oci.BuildOptions{
			SourcePath:       source,
			RegistryEndpoint: "localhost:5000",
			Version:          "1.0.0",
			Repository:       "  KSail/Workloads/My-App  ",
		}

		validated, err := opts.Validate()

		require.NoError(t, err)
		require.Equal(t, "ksail/workloads/my-app", validated.Repository)
	})
}
