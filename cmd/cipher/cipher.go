// Package cipher provides the cipher command for integrating with SOPS.
package cipher

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/sops"
	"github.com/spf13/cobra"
)

// NewCipherCmd creates the cipher command that integrates with SOPS.
func NewCipherCmd(_ *runtime.Runtime) *cobra.Command {
	client := sops.NewClient()

	return client.CreateCipherCommand()
}
