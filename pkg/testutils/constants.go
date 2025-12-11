package testutils

import (
	"context"
	"errors"
	"io/fs"
	"time"
)

// Suppress unused warnings for shared utilities that may not be used in all test files.
var (
	_ = context.Background
	_ = time.Second
)

// File permissions.
const (
	// filePermUserReadWrite is the permission for user read/write only (0o600).
	filePermUserReadWrite fs.FileMode = 0o600

	// testDirectoryPerm is the permission for test directories (0o750).
	testDirectoryPerm fs.FileMode = 0o750

	// testFilePerm is the permission for test files (0o600).
	testFilePerm fs.FileMode = 0o600
)

// Default configuration content for tests.
const (
	defaultKsailConfigContent = `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  sourceDirectory: k8s
`

	defaultKindConfigContent = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: kind
`
)

// Common test errors used across installer tests.
var (
	ErrInstallFailed   = errors.New("install failed")
	ErrAddRepoFailed   = errors.New("add repo failed")
	ErrUninstallFailed = errors.New("uninstall failed")
	ErrDaemonSetBoom   = errors.New("boom")
	ErrDeploymentFail  = errors.New("fail")
	ErrPollBoom        = errors.New("boom")
)
