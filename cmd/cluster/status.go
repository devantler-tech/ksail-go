package cluster

import (
	"time"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"status",
		"Show status of the Kubernetes cluster",
		`Show the current status of the Kubernetes cluster.`,
		helpers.HandleConfigLoadRunE,
		configmanager.StandardContextFieldSelector(),
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			Description:  "Path to kubeconfig file",
			DefaultValue: "~/.kube/config",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for status check operations",
			DefaultValue: metav1.Duration{Duration: defaultStatusTimeout},
		},
	)
}
