package oci

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var manifestExtensions = map[string]struct{}{
	".yaml": {},
	".yml":  {},
	".json": {},
}

type imagePusher interface {
	Push(ctx context.Context, ref name.Reference, img v1.Image) error
}

// remoteImagePusher pushes OCI images using the go-containerregistry remote helpers.
type remoteImagePusher struct{}

func (remoteImagePusher) Push(ctx context.Context, ref name.Reference, img v1.Image) error {
	return remote.Write(ref, img, remote.WithContext(ctx))
}

// NewWorkloadArtifactBuilder returns a concrete implementation backed by go-containerregistry.
func NewWorkloadArtifactBuilder() WorkloadArtifactBuilder {
	return &builder{pusher: remoteImagePusher{}}
}

type builder struct {
	pusher imagePusher
}

// Build collects manifests from the source path, packages them into an OCI artifact, and pushes it to the registry.
func (b *builder) Build(ctx context.Context, opts BuildOptions) (BuildResult, error) {
	validated, err := opts.Validate()
	if err != nil {
		return BuildResult{}, err
	}

	manifestFiles, err := collectManifestFiles(validated.SourcePath)
	if err != nil {
		return BuildResult{}, fmt.Errorf("discover manifests: %w", err)
	}

	if len(manifestFiles) == 0 {
		return BuildResult{}, ErrNoManifestFiles
	}

	layer, err := newManifestLayer(validated.SourcePath, manifestFiles)
	if err != nil {
		return BuildResult{}, fmt.Errorf("package manifests: %w", err)
	}

	img, err := buildImage(layer, validated)
	if err != nil {
		return BuildResult{}, fmt.Errorf("build image: %w", err)
	}

	ref, err := name.ParseReference(
		fmt.Sprintf("%s/%s:%s", validated.RegistryEndpoint, validated.Repository, validated.Version),
		name.WeakValidation,
		name.Insecure,
	)
	if err != nil {
		return BuildResult{}, fmt.Errorf("parse reference: %w", err)
	}

	pusher := b.ensurePusher()
	if err := pusher.Push(ctx, ref, img); err != nil {
		return BuildResult{}, fmt.Errorf("push artifact: %w", err)
	}

	artifact := v1alpha1.OCIArtifact{
		Name:             validated.Name,
		Version:          validated.Version,
		RegistryEndpoint: validated.RegistryEndpoint,
		Repository:       validated.Repository,
		Tag:              validated.Version,
		SourcePath:       validated.SourcePath,
		CreatedAt:        metav1.NewTime(time.Now().UTC()),
	}

	return BuildResult{Artifact: artifact}, nil
}

func (b *builder) ensurePusher() imagePusher {
	if b.pusher != nil {
		return b.pusher
	}

	b.pusher = remoteImagePusher{}
	return b.pusher
}

func collectManifestFiles(root string) ([]string, error) {
	var manifests []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, ok := manifestExtensions[ext]; !ok {
			return nil
		}

		info, statErr := d.Info()
		if statErr != nil {
			return statErr
		}

		if info.Size() == 0 {
			return fmt.Errorf("manifest file %s is empty", path)
		}

		manifests = append(manifests, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Strings(manifests)
	return manifests, nil
}

func newManifestLayer(root string, files []string) (v1.Layer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)

	for _, path := range files {
		if err := addFileToArchive(tw, root, path); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return tarball.LayerFromReader(bytes.NewReader(buf.Bytes()))
}

func addFileToArchive(tw *tar.Writer, root, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	header.Name = filepath.ToSlash(rel)
	header.Mode = 0o644

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}

func buildImage(layer v1.Layer, opts ValidatedBuildOptions) (v1.Image, error) {
	cfg := &v1.ConfigFile{
		Architecture: runtime.GOARCH,
		OS:           runtime.GOOS,
		Created:      v1.Time{Time: time.Now().UTC()},
		Config: v1.Config{
			Labels: map[string]string{
				"org.opencontainers.image.title":        opts.Name,
				"org.opencontainers.image.version":      opts.Version,
				"org.opencontainers.image.source":       opts.SourcePath,
				"devantler.tech/ksail/repository":       opts.Repository,
				"devantler.tech/ksail/registryEndpoint": opts.RegistryEndpoint,
			},
		},
	}

	img, err := mutate.ConfigFile(empty.Image, cfg)
	if err != nil {
		return nil, err
	}

	return mutate.AppendLayers(img, layer)
}
