// Package generator provides an interface for generating files from code.
package generator

// Generator is implemented by specific distribution generators (kind, k3d, kustomization).
type Generator[T any] interface {
	Generate(model T) (string, error)
}
