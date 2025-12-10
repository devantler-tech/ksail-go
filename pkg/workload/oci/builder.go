package oci

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

const (
	defaultRepositoryName = "ksail-workloads"
	defaultArtifactName   = "ksail-workload"
)

// WorkloadArtifactBuilder packages Kubernetes manifests into OCI artifacts and pushes them to a registry.
//
//go:generate mockery --name WorkloadArtifactBuilder --output ../../testutils/mocks --outpkg mocks --case underscore
type WorkloadArtifactBuilder interface {
	// Build validates the supplied options, constructs an OCI artifact, and pushes it to the registry.
	Build(ctx context.Context, opts BuildOptions) (BuildResult, error)
}

// BuildOptions capture user-supplied inputs for building an OCI artifact from manifest directories.
type BuildOptions struct {
	Name             string
	SourcePath       string
	RegistryEndpoint string
	Repository       string
	Version          string
}

// ValidatedBuildOptions represents sanitized inputs ready for use by the builder implementation.
type ValidatedBuildOptions struct {
	Name             string
	SourcePath       string
	RegistryEndpoint string
	Repository       string
	Version          string
}

// BuildResult describes the outcome of a successful artifact build.
type BuildResult struct {
	Artifact v1alpha1.OCIArtifact
}

var (
	// ErrSourcePathRequired indicates that no source path was provided in build options.
	ErrSourcePathRequired = errors.New("source path is required")
	// ErrSourcePathNotFound indicates that the provided source path does not exist.
	ErrSourcePathNotFound = errors.New("source path does not exist")
	// ErrSourcePathNotDirectory indicates that the provided source path is not a directory.
	ErrSourcePathNotDirectory = errors.New("source path must be a directory")
	// ErrRegistryEndpointRequired indicates that the registry endpoint is missing.
	ErrRegistryEndpointRequired = errors.New("registry endpoint is required")
	// ErrVersionRequired indicates that no semantic version was provided.
	ErrVersionRequired = errors.New("version is required")
	// ErrVersionInvalid indicates that the supplied version does not follow semantic versioning.
	ErrVersionInvalid = errors.New("version must follow semantic versioning")
	// ErrNoManifestFiles indicates that the source directory does not contain manifest files.
	ErrNoManifestFiles = errors.New("no manifest files found in source directory")
)

// Validate normalizes and verifies the build options before artifact construction.
func (o BuildOptions) Validate() (ValidatedBuildOptions, error) {
	trimmedSource := strings.TrimSpace(o.SourcePath)
	if trimmedSource == "" {
		return ValidatedBuildOptions{}, ErrSourcePathRequired
	}

	absSource, err := filepath.Abs(trimmedSource)
	if err != nil {
		return ValidatedBuildOptions{}, fmt.Errorf("resolve source path: %w", err)
	}

	info, statErr := os.Stat(absSource)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			return ValidatedBuildOptions{}, ErrSourcePathNotFound
		}

		return ValidatedBuildOptions{}, fmt.Errorf("stat source path: %w", statErr)
	}

	if !info.IsDir() {
		return ValidatedBuildOptions{}, ErrSourcePathNotDirectory
	}

	endpoint, err := normalizeRegistryEndpoint(o.RegistryEndpoint)
	if err != nil {
		return ValidatedBuildOptions{}, err
	}

	version, err := normalizeVersion(o.Version)
	if err != nil {
		return ValidatedBuildOptions{}, err
	}

	repository := normalizeRepositoryName(o.Repository, absSource)
	name := normalizeArtifactName(o.Name, repository)

	return ValidatedBuildOptions{
		Name:             name,
		SourcePath:       absSource,
		RegistryEndpoint: endpoint,
		Repository:       repository,
		Version:          version,
	}, nil
}

func normalizeRegistryEndpoint(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "oci://")
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.TrimSuffix(trimmed, "/")
	trimmed = strings.TrimSpace(trimmed)

	if trimmed == "" {
		return "", ErrRegistryEndpointRequired
	}

	if idx := strings.Index(trimmed, "/"); idx > 0 {
		trimmed = trimmed[:idx]
	}

	return trimmed, nil
}

func normalizeVersion(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrVersionRequired
	}

	if strings.EqualFold(trimmed, "latest") {
		return "latest", nil
	}

	trimmed = strings.TrimPrefix(trimmed, "v")

	if trimmed == "" {
		return "", ErrVersionRequired
	}

	if _, err := semver.NewVersion(trimmed); err != nil {
		return "", fmt.Errorf("%w: %w", ErrVersionInvalid, err)
	}

	return trimmed, nil
}

func normalizeRepositoryName(candidate, sourcePath string) string {
	pathCandidate := strings.TrimSpace(candidate)
	if pathCandidate == "" {
		pathCandidate = filepath.Base(sourcePath)
	}

	pathCandidate = filepath.ToSlash(pathCandidate)

	pathCandidate = strings.Trim(pathCandidate, "/")
	if pathCandidate == "" {
		pathCandidate = defaultRepositoryName
	}

	segments := strings.Split(pathCandidate, "/")

	normalized := make([]string, 0, len(segments))
	for _, segment := range segments {
		sanitized := sanitizeSegment(segment)
		if sanitized == "" {
			continue
		}

		normalized = append(normalized, sanitized)
	}

	if len(normalized) == 0 {
		return defaultRepositoryName
	}

	return strings.Join(normalized, "/")
}

//nolint:cyclop // segment sanitization requires character-by-character validation
func sanitizeSegment(segment string) string {
	trimmed := strings.TrimSpace(segment)
	if trimmed == "" {
		return ""
	}

	trimmed = strings.ToLower(trimmed)

	var builder strings.Builder
	prevHyphen := false

	for _, char := range trimmed {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
			prevHyphen = false
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
			prevHyphen = false
		case char == '-':
			if !prevHyphen {
				builder.WriteRune('-')
				prevHyphen = true
			}
		default:
			if !prevHyphen {
				builder.WriteRune('-')
				prevHyphen = true
			}
		}
	}

	return strings.Trim(builder.String(), "-")
}

func normalizeArtifactName(candidate, repository string) string {
	trimmed := strings.TrimSpace(candidate)
	if trimmed == "" {
		parts := strings.Split(repository, "/")
		trimmed = parts[len(parts)-1]
	}

	normalized := sanitizeSegment(trimmed)
	if normalized == "" {
		return defaultArtifactName
	}

	return normalized
}
