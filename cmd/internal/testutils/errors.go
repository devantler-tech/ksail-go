// Package testutils provides testing utilities and shared test errors.
package testutils

import "errors"

// Static test errors to comply with err113.
var ErrTestConfigLoadError = errors.New("test config load error")
