// Package gen provides interfaces and implementations for generating Kubernetes resource manifests.
//
// The package exposes a Generator interface that can be implemented to support different
// manifest generation strategies. The primary implementation is in the kubernetes subpackage,
// which wraps kubectl create commands with forced --dry-run=client -o yaml flags.
//
// # Usage
//
// Create a generator for a specific resource type and use it to generate a command:
//
//	import "github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
//
//	generator := kubernetes.NewNamespaceGenerator()
//	namespaceCmd := generator.Generate()
//
//	// Or for other resources:
//	deploymentGenerator := kubernetes.NewDeploymentGenerator()
//	deploymentCmd := deploymentGenerator.Generate()
//
// The generated commands can be integrated into CLI applications or used standalone.
// No kubeconfig is required since --dry-run=client doesn't need cluster access.
//
// # Generator Interface
//
// The Generator interface allows for custom implementations:
//
//	type Generator interface {
//	    Generate() *cobra.Command
//	}
//
// # Supported Resource Types
//
// Each resource type has its own generator constructor:
//   - NewClusterRoleGenerator(), NewClusterRoleBindingGenerator()
//   - NewConfigMapGenerator()
//   - NewCronJobGenerator()
//   - NewDeploymentGenerator()
//   - NewIngressGenerator()
//   - NewJobGenerator()
//   - NewNamespaceGenerator()
//   - NewPodDisruptionBudgetGenerator()
//   - NewPriorityClassGenerator()
//   - NewQuotaGenerator()
//   - NewRoleGenerator(), NewRoleBindingGenerator()
//   - NewSecretGenerator() (with subcommands: generic, tls, docker-registry)
//   - NewServiceGenerator() (with subcommands: clusterip, nodeport, loadbalancer, externalname)
//   - NewServiceAccountGenerator()
package gen
