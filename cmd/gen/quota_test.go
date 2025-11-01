package gen

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

// TestGenQuota tests generating a quota manifest.
func TestGenQuota(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewQuotaCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-quota", "--hard=cpu=1,memory=1Gi,pods=10"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen quota to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
