// Package registries contains helpers for managing shared mirror registry state across
// different provisioners.
//
// This package provides functions used by Kind and K3d provisioners to create,
// connect, and clean up Docker registry containers consistently, enabling
// pull-through caching to upstream registries like Docker Hub.
package registries
