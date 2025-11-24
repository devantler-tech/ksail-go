package registry

// InitPortAllocation prepares the used ports set and the next available port based on provided base mapping.
func InitPortAllocation(baseUsedPorts map[int]struct{}) (map[int]struct{}, int) {
	usedPorts := make(map[int]struct{}, len(baseUsedPorts))
	nextPort := DefaultRegistryPort

	for port := range baseUsedPorts {
		usedPorts[port] = struct{}{}

		if port >= nextPort {
			nextPort = port + 1
		}
	}

	return usedPorts, nextPort
}
