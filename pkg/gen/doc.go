// Package gen provides interfaces and implementations for generating Kubernetes resource manifests.
//
// The package exposes a Generator interface that can be implemented to support different
// manifest generation strategies. The primary implementation is KubectlGenerator, which
// wraps kubectl create commands with forced --dry-run=client -o yaml flags.
//
// # Usage
//
// Create a generator and use it to generate commands:
//
//	generator := gen.NewKubectlGenerator("/path/to/kubeconfig")
//	namespaceCmd := generator.GenerateCommand("namespace")
//	deploymentCmd := generator.GenerateCommand("deployment")
//
// The generated commands can be integrated into CLI applications or used standalone.
//
// # Generator Interface
//
// The Generator interface allows for custom implementations:
//
//	type Generator interface {
//	    GenerateCommand(resourceType string) *cobra.Command
//	}
//
// # Supported Resource Types
//
// All kubectl create resource types are supported:
//   - clusterrole, clusterrolebinding
//   - configmap
//   - cronjob
//   - deployment
//   - ingress
//   - job
//   - namespace
//   - poddisruptionbudget
//   - priorityclass
//   - quota
//   - role, rolebinding
//   - secret (with subcommands: generic, tls, docker-registry)
//   - service (with subcommands: clusterip, nodeport, loadbalancer, externalname)
//   - serviceaccount
package gen
