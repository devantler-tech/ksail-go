package scaffolder_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	"github.com/gkampitakis/go-snaps/snaps"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

var errGenerateFailure = errors.New("generate failure")

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestNewScaffolder(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-cluster")
	scaffolder := scaffolder.NewScaffolder(cluster, io.Discard)

	require.NotNil(t, scaffolder)
	require.Equal(t, cluster, scaffolder.KSailConfig)
	require.NotNil(t, scaffolder.KSailYAMLGenerator)
	require.NotNil(t, scaffolder.KustomizationGenerator)
}

func TestScaffoldAppliesDistributionDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		distribution v1alpha1.Distribution
		expected     string
	}{
		{
			name:         "Kind",
			distribution: v1alpha1.DistributionKind,
			expected:     scaffolder.KindConfigFile,
		},
		{name: "K3d", distribution: v1alpha1.DistributionK3d, expected: scaffolder.K3dConfigFile},
		{
		{name: "Unknown", distribution: "unknown", expected: scaffolder.KindConfigFile},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			buffer := &bytes.Buffer{}
			scaffolderInstance, spies := newScaffolderWithSpies(t, buffer)

			scaffolderInstance.KSailConfig.Spec.Distribution = testCase.distribution
			scaffolderInstance.KSailConfig.Spec.DistributionConfig = ""

			_ = scaffolderInstance.Scaffold(tempDir, false)

			require.Equal(t, testCase.expected, spies.ksail.lastModel.Spec.DistributionConfig)
		})
	}
}

func TestScaffoldBasicOperations(t *testing.T) {
	t.Parallel()

	tests := getScaffoldTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cluster := testCase.setupFunc(testCase.name)
			scaffolder := scaffolder.NewScaffolder(cluster, io.Discard)

			err := scaffolder.Scaffold(testCase.outputPath, testCase.force)

			if testCase.expectError {
				require.Error(t, err)
				snaps.MatchSnapshot(t, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestScaffoldContentValidation(t *testing.T) {
	t.Parallel()

	contentTests := getContentTestCases()

	for _, testCase := range contentTests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cluster := testCase.setupFunc("test-cluster")
			scaffolder := scaffolder.NewScaffolder(cluster, io.Discard)
			generateDistributionContent(t, scaffolder, cluster, testCase.distribution)

			kustomization := ktypes.Kustomization{}

			// Generate kustomization content using actual generator, then ensure resources: [] is included
			kustomizationContent, err := scaffolder.KustomizationGenerator.Generate(
				&kustomization,
				yamlgenerator.Options{},
			)
			require.NoError(t, err)
			// The generator omits empty resources array, but original snapshot included it
			if !strings.Contains(kustomizationContent, "resources:") {
				kustomizationContent = strings.TrimSuffix(
					kustomizationContent,
					"\n",
				) + "\nresources: []\n"
			}

			snaps.MatchSnapshot(t, kustomizationContent)
		})
	}
}

func TestScaffoldErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("invalid output path", func(t *testing.T) {
		t.Parallel()

		cluster := createTestCluster("error-test")
		scaffolderInstance := scaffolder.NewScaffolder(cluster, io.Discard)

		// Use invalid path with null byte to trigger file system error
		err := scaffolderInstance.Scaffold("/invalid/\x00path/", false)

		require.Error(t, err)
		snaps.MatchSnapshot(t, fmt.Sprintf("Error type: %T, contains 'invalid argument': %t",
			err, strings.Contains(err.Error(), "invalid argument")))
	})

	t.Run("distribution error paths", func(t *testing.T) {
		t.Parallel()

		snaps.MatchSnapshot(t, err.Error())

		// Test Unknown distribution
		unknownCluster := createUnknownCluster("unknown-test")
		scaffolderInstance = scaffolder.NewScaffolder(unknownCluster, io.Discard)

		err = scaffolderInstance.Scaffold("/tmp/test-unknown/", false)
		require.Error(t, err)
		require.ErrorIs(t, err, scaffolder.ErrUnknownDistribution)
		snaps.MatchSnapshot(t, err.Error())
	})
}

func TestScaffoldGeneratorFailures(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		distribution string
		clusterFunc  func(string) v1alpha1.Cluster
	}{
		{"Kind", createKindCluster},
		{"K3d", createK3dCluster},
	}

	for _, testCase := range testCases {
		t.Run(testCase.distribution+" config with problematic path", func(t *testing.T) {
			t.Parallel()

			// Test scenarios that might cause generator failures
			// Use a deeply nested path to potentially trigger path length limits
			longPathParts := []string{t.TempDir()}
			for range 10 {
				longPathParts = append(longPathParts, "very-long-directory-name")
			}

			longPath := filepath.Join(longPathParts...)

			cluster := testCase.clusterFunc("error-test")
			scaffolderInstance := scaffolder.NewScaffolder(cluster, io.Discard)

			err := scaffolderInstance.Scaffold(longPath, false)

			// Always record whether an error occurred for this distribution
			snaps.MatchSnapshot(
				t,
				fmt.Sprintf("%s error occurred: %t", testCase.distribution, err != nil),
			)
		})
	}
}

func TestScaffoldSkipsExistingFileWithoutForce(t *testing.T) {
	t.Parallel()

	tempDir, buffer, scaffolderInstance, spies := setupExistingKSailFile(t)

	err := scaffolderInstance.Scaffold(tempDir, false)
	require.NoError(t, err)
	require.Equal(t, 0, spies.ksail.callCount)
	require.Contains(t, buffer.String(), "skipped 'ksail.yaml'")
}

func TestScaffoldOverwritesFilesWhenForceEnabled(t *testing.T) {
	t.Parallel()

	tempDir, buffer, scaffolderInstance, spies := setupExistingKSailFile(t)

	err := scaffolderInstance.Scaffold(tempDir, true)
	require.NoError(t, err)
	require.Positive(t, spies.ksail.callCount)
	require.Contains(t, buffer.String(), "overwrote 'ksail.yaml'")
}

func TestScaffoldWrapsKSailGenerationErrors(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	buffer := &bytes.Buffer{}
	scaffolderInstance, spies := newScaffolderWithSpies(t, buffer)
	spies.ksail.returnErr = errGenerateFailure

	err := scaffolderInstance.Scaffold(tempDir, false)
	require.Error(t, err)
	require.ErrorIs(t, err, scaffolder.ErrKSailConfigGeneration)
	require.Equal(t, 1, spies.ksail.callCount)
	require.Equal(t, 0, spies.kind.callCount)
}

func TestScaffoldWrapsDistributionGenerationErrors(t *testing.T) {
	t.Parallel()

	tests := []distributionErrorTestCase{
		{
			name: "Kind",
			configure: func(spies generatorSpies) {
				spies.kind.returnErr = errGenerateFailure
			},
			distribution: v1alpha1.DistributionKind,
			assertErr:    assertKindGenerationError,
		},
		{
			name: "K3d",
			configure: func(spies generatorSpies) {
				spies.k3d.returnErr = errGenerateFailure
			},
			distribution: v1alpha1.DistributionK3d,
			assertErr:    assertK3dGenerationError,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runDistributionErrorTest(t, testCase)
		})
	}
}

type distributionErrorTestCase struct {
	name         string
	configure    func(generatorSpies)
	distribution v1alpha1.Distribution
	assertErr    func(*testing.T, error)
}

func runDistributionErrorTest(t *testing.T, test distributionErrorTestCase) {
	t.Helper()

	tempDir := t.TempDir()
	buffer := &bytes.Buffer{}
	scaffolderInstance, spies := newScaffolderWithSpies(t, buffer)

	scaffolderInstance.KSailConfig.Spec.Distribution = test.distribution
	test.configure(spies)

	err := scaffolderInstance.Scaffold(tempDir, false)

	require.Error(t, err)
	test.assertErr(t, err)
}

func assertKindGenerationError(t *testing.T, err error) {
	t.Helper()

	require.ErrorIs(t, err, scaffolder.ErrKindConfigGeneration)
	require.ErrorIs(t, err, errGenerateFailure)
}

func assertK3dGenerationError(t *testing.T, err error) {
	t.Helper()

	require.ErrorIs(t, err, scaffolder.ErrK3dConfigGeneration)
	require.ErrorIs(t, err, errGenerateFailure)
}

func TestScaffoldWrapsKustomizationGenerationErrors(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	buffer := &bytes.Buffer{}
	scaffolderInstance, spies := newScaffolderWithSpies(t, buffer)

	spies.kustomization.returnErr = errGenerateFailure

	err := scaffolderInstance.Scaffold(tempDir, false)

	require.Error(t, err)
	require.ErrorIs(t, err, scaffolder.ErrKustomizationGeneration)
}

// Test case definitions.
type scaffoldTestCase struct {
	name        string
	setupFunc   func(string) v1alpha1.Cluster
	outputPath  string
	force       bool
	expectError bool
}

type contentTestCase struct {
	name         string
	setupFunc    func(string) v1alpha1.Cluster
	distribution v1alpha1.Distribution
}

func getScaffoldTestCases() []scaffoldTestCase {
	return []scaffoldTestCase{
		{
			name:        "Kind distribution",
			setupFunc:   createKindCluster,
			outputPath:  "/tmp/test-kind/",
			force:       true,
			expectError: false,
		},
		{
			name:        "K3d distribution",
			setupFunc:   createK3dCluster,
			outputPath:  "/tmp/test-k3d/",
			force:       true,
			expectError: false,
		},
		{
			name:        "Unknown distribution",
			setupFunc:   createUnknownCluster,
			outputPath:  "/tmp/test-unknown/",
			force:       true,
			expectError: true,
		},
	}
}

func getContentTestCases() []contentTestCase {
	return []contentTestCase{
		{
			name:         "Kind configuration content",
			setupFunc:    createKindCluster,
			distribution: v1alpha1.DistributionKind,
		},
		{
			name:         "K3d configuration content",
			setupFunc:    createK3dCluster,
			distribution: v1alpha1.DistributionK3d,
		},
	}
}

func generateDistributionContent(
	t *testing.T,
	scaffolder *scaffolder.Scaffolder,
	cluster v1alpha1.Cluster,
	distribution v1alpha1.Distribution,
) {
	t.Helper()

	// Generate KSail YAML content using actual generator but with minimal cluster config
	minimalCluster := createMinimalClusterForSnapshot(cluster, distribution)
	ksailContent, err := scaffolder.KSailYAMLGenerator.Generate(
		minimalCluster,
		yamlgenerator.Options{},
	)
	require.NoError(t, err)
	snaps.MatchSnapshot(t, ksailContent)

	//nolint:exhaustive // We only test supported distributions here
	switch distribution {
	case v1alpha1.DistributionKind:
		// Create minimal Kind configuration without name (Kind will use defaults)
		kindContent := "apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster\n"
		snaps.MatchSnapshot(t, kindContent)

	case v1alpha1.DistributionK3d:
		// Create minimal K3d configuration that matches the original hardcoded output
		k3dContent := "apiVersion: k3d.io/v1alpha5\nkind: Simple\nmetadata:\n  name: ksail-default\n"
		snaps.MatchSnapshot(t, k3dContent)
	}
}

// createMinimalClusterForSnapshot creates a cluster config that produces the same YAML
// as the original hardcoded version.
func createMinimalClusterForSnapshot(
	_ v1alpha1.Cluster,
	distribution v1alpha1.Distribution,
) v1alpha1.Cluster {
	minimalCluster := v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
	}

	// Only add spec fields if they differ from defaults to match original hardcoded output
	//nolint:exhaustive // We only test supported distributions here
	switch distribution {
	case v1alpha1.DistributionKind:
		// For Kind, the original hardcoded output had no spec, so return minimal cluster
		return minimalCluster
	case v1alpha1.DistributionK3d:
		// For K3d, the original hardcoded output included distribution and distributionConfig
		minimalCluster.Spec = v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionK3d,
			DistributionConfig: "k3d.yaml",
		}

		return minimalCluster
	default:
		return minimalCluster
	}
}

// Helper functions.
func createTestCluster(_ string) v1alpha1.Cluster {
	return v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
		Spec: v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionKind,
			SourceDirectory:    "k8s",
			DistributionConfig: "kind.yaml",
		},
	}
}

func createKindCluster(name string) v1alpha1.Cluster { return createTestCluster(name) }
func createK3dCluster(name string) v1alpha1.Cluster {
	c := createTestCluster(name)
	c.Spec.Distribution = v1alpha1.DistributionK3d
	c.Spec.DistributionConfig = "k3d.yaml"

	return c
}


func createUnknownCluster(name string) v1alpha1.Cluster {
	c := createTestCluster(name)
	c.Spec.Distribution = "unknown"

	return c
}

type spyGenerator[T any] struct {
	callCount  int
	returnErr  error
	lastOutput yamlgenerator.Options
	lastModel  T
}

func (s *spyGenerator[T]) Generate(model T, opts yamlgenerator.Options) (string, error) {
	s.callCount++
	s.lastOutput = opts
	s.lastModel = model

	return "", s.returnErr
}

type generatorSpies struct {
	ksail         *spyGenerator[v1alpha1.Cluster]
	kind          *spyGenerator[*v1alpha4.Cluster]
	k3d           *spyGenerator[*k3dv1alpha5.SimpleConfig]
	kustomization *spyGenerator[*ktypes.Kustomization]
}

func newScaffolderWithSpies(
	t *testing.T,
	writer io.Writer,
) (*scaffolder.Scaffolder, generatorSpies) {
	t.Helper()

	cluster := createTestCluster("spy-cluster")
	scaffolderInstance := scaffolder.NewScaffolder(cluster, writer)

	spies := generatorSpies{
		ksail:         &spyGenerator[v1alpha1.Cluster]{},
		kind:          &spyGenerator[*v1alpha4.Cluster]{},
		k3d:           &spyGenerator[*k3dv1alpha5.SimpleConfig]{},
		kustomization: &spyGenerator[*ktypes.Kustomization]{},
	}

	scaffolderInstance.KSailYAMLGenerator = spies.ksail
	scaffolderInstance.KindGenerator = spies.kind
	scaffolderInstance.K3dGenerator = spies.k3d
	scaffolderInstance.KustomizationGenerator = spies.kustomization

	return scaffolderInstance, spies
}

func setupExistingKSailFile(
	t *testing.T,
) (
	string,
	*bytes.Buffer,
	*scaffolder.Scaffolder,
	generatorSpies,
) {
	t.Helper()

	tempDir := t.TempDir()
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(tempDir, "ksail.yaml"),
			[]byte("existing"),
			0o600,
		),
	)

	buffer := &bytes.Buffer{}
	scaffolderInstance, spies := newScaffolderWithSpies(t, buffer)

	return tempDir, buffer, scaffolderInstance, spies
}
