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

//nolint:gochecknoglobals // static set of valid manifest extensions
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
	if err := remote.Write(ref, img, remote.WithContext(ctx)); err != nil {
		return fmt.Errorf("write image to registry: %w", err)
	}

	return nil
}

//nolint:ireturn // returns interface for dependency injection
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
//nolint:ireturn // returns interface for internal use
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

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if _, ok := manifestExtensions[ext]; !ok {
			return nil
		}

		info, statErr := entry.Info()
		if statErr != nil {
			return fmt.Errorf("get file info for %s: %w", path, statErr)
		}

		if info.Size() == 0 {
			//nolint:err113 // includes dynamic file path for debugging
			return fmt.Errorf("manifest file %s is empty", path)
		}

		manifests = append(manifests, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk directory %s: %w", root, err)
	}

	sort.Strings(manifests)
//nolint:ireturn // returns interface from external library
	return manifests, nil
}

func newManifestLayer(root string, files []string) (v1.Layer, error) {
	buf := bytes.NewBuffer(nil)
	tarWriter := tar.NewWriter(buf)

	for _, path := range files {
		if err := addFileToArchive(tarWriter, root, path); err != nil {
			return nil, err
		}
	}

	//nolint:staticcheck // using deprecated API for compatibility
	if err := tarWriter.Close(); err != nil {
		return nil, fmt.Errorf("close tar writer: %w", err)
	}

	layer, err := tarball.LayerFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("create layer from tar: %w", err)
	}

	return layer, nil
}

func addFileToArchive(tarWriter *tar.Writer, root, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file %s: %w", path, err)
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return fmt.Errorf("get relative path for %s: %w", path, err)
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("create tar header for %s: %w", path, err)
	}

	header.Name = filepath.ToSlash(rel)
	header.Mode = 0o644

	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("write tar header for %s: %w", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file %s: %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(tarWriter, file); err != nil {
		return fmt.Errorf("copy file %s to tar: %w", path, err)
	}
//nolint:ireturn // returns interface from external library

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
		return nil, fmt.Errorf("set config file: %w", err)
	}

	finalImg, err := mutate.AppendLayers(img, layer)
	if err != nil {
		return nil, fmt.Errorf("append layer: %w", err)
	}

	return finalImg, nil
}
