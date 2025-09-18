// Package helpers provides common functionality for config managers to eliminate duplication.
package helpers

import (
"fmt"
"os"
"path/filepath"

"github.com/devantler-tech/ksail-go/pkg/io"
yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
)

//nolint:ireturn // Generic function must return interface type
func LoadConfigFromFile[T any](
configPath string,
createDefault func() T,
createEmpty func() T,
setDefaults func(T) T,
) (T, error) {
// Resolve the config path (traverse up from current dir if relative)
resolvedPath, err := io.FindFile(configPath)
if err != nil {
var zero T

return zero, fmt.Errorf("failed to resolve config path: %w", err)
}

// Check if config file exists
_, err = os.Stat(resolvedPath)
if os.IsNotExist(err) {
// File doesn't exist, return default configuration
return createDefault(), nil
}

// Read file contents safely
// Since we've resolved the path through traversal, we use the directory containing the file as the base
baseDir := filepath.Dir(resolvedPath)

data, err := io.ReadFileSafe(baseDir, resolvedPath)
if err != nil {
var zero T

return zero, fmt.Errorf("failed to read config file %s: %w", resolvedPath, err)
}

// Parse YAML into config
config := createEmpty()
marshaller := yamlmarshaller.YAMLMarshaller[T]{}

err = marshaller.Unmarshal(data, &config)
if err != nil {
var zero T

return zero, fmt.Errorf("failed to unmarshal config from %s: %w", resolvedPath, err)
}

// Apply defaults
config = setDefaults(config)

return config, nil
}
