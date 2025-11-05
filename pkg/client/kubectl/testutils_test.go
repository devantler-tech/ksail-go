package kubectl_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// newTestClient creates a new kubectl client with test IOStreams.
// It returns the client and the output buffer for verification.
func newTestClient() (*kubectl.Client, *bytes.Buffer) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    outBuf,
		ErrOut: errBuf,
	}

	return kubectl.NewClient(ioStreams), outBuf
}

// assertNamespaceYAML verifies that the output contains expected namespace YAML.
func assertNamespaceYAML(t *testing.T, output, namespaceName string) {
	t.Helper()

	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Namespace")
	require.Contains(t, output, namespaceName)
}
