//nolint:testpackage // test needs access to unexported builder struct
package oci

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/require"
)

type fakePusher struct {
	ref  name.Reference
	push int
	img  v1.Image
	err  error
}

func (f *fakePusher) Push(_ context.Context, ref name.Reference, img v1.Image) error {
	f.push++
	f.ref = ref
	f.img = img
	return f.err
}

func TestBuilderBuildSuccess(t *testing.T) {
	t.Parallel()

	manifestDir := t.TempDir()
	manifestPath := filepath.Join(manifestDir, "deployment.yaml")
	require.NoError(t, os.WriteFile(manifestPath, []byte("apiVersion: v1"), 0o600))

	pusher := &fakePusher{}
	builder := &builder{pusher: pusher}

	result, err := builder.Build(context.Background(), BuildOptions{
		SourcePath:       manifestDir,
		RegistryEndpoint: "localhost:5000",
		Repository:       "sample/app",
		Version:          "1.2.3",
	})

	require.NoError(t, err)
	require.Equal(t, "sample/app", result.Artifact.Repository)
	require.Equal(t, "1.2.3", result.Artifact.Version)
	require.Equal(t, "1.2.3", result.Artifact.Tag)
	require.Equal(t, "localhost:5000", result.Artifact.RegistryEndpoint)
	require.Equal(t, filepath.Clean(manifestDir), result.Artifact.SourcePath)
	require.NotZero(t, result.Artifact.CreatedAt.Time)
	require.Equal(t, "app", result.Artifact.Name)
	require.Equal(t, 1, pusher.push)
	require.Equal(t, "localhost:5000/sample/app:1.2.3", pusher.ref.Name())
}

func TestBuilderBuildRequiresManifests(t *testing.T) {
	t.Parallel()

	manifestDir := t.TempDir()
	artifact := filepath.Join(manifestDir, "readme.txt")
	require.NoError(t, os.WriteFile(artifact, []byte("hello"), 0o600))

	builder := &builder{pusher: &fakePusher{}}

	_, err := builder.Build(context.Background(), BuildOptions{
		SourcePath:       manifestDir,
		RegistryEndpoint: "localhost:5000",
		Repository:       "sample/app",
		Version:          "1.2.3",
	})

	require.ErrorIs(t, err, ErrNoManifestFiles)
}

func TestBuilderBuildPropagatesPushError(t *testing.T) {
	t.Parallel()

	manifestDir := t.TempDir()
	manifestPath := filepath.Join(manifestDir, "deployment.yaml")
	require.NoError(t, os.WriteFile(manifestPath, []byte("apiVersion: v1"), 0o600))

	//nolint:err113 // test error for push failure
	pushErr := errors.New("push failed")
	pusher := &fakePusher{err: pushErr}
	builder := &builder{pusher: pusher}

	_, err := builder.Build(context.Background(), BuildOptions{
		SourcePath:       manifestDir,
		RegistryEndpoint: "localhost:5000",
		Repository:       "sample/app",
		Version:          "1.2.3",
	})

	require.ErrorIs(t, err, pushErr)
}

func TestBuilderBuildRejectsEmptyManifest(t *testing.T) {
	t.Parallel()

	manifestDir := t.TempDir()
	emptyManifest := filepath.Join(manifestDir, "empty.yaml")
	require.NoError(t, os.WriteFile(emptyManifest, []byte{}, 0o644))

	builder := &builder{pusher: &fakePusher{}}

	_, err := builder.Build(context.Background(), BuildOptions{
		SourcePath:       manifestDir,
		RegistryEndpoint: "localhost:5000",
		Repository:       "sample/app",
		Version:          "1.2.3",
	})

	require.ErrorContains(t, err, "empty")
}
