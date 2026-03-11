package icloud

import "strings"

// MaxOperationPerRequest specifies the maximum number of operations in a
// request.
const MaxOperationPerRequest = 200

// IsICloudContainer returns true if the given container identifier is a valid
// iCloud container name.
func IsICloudContainer(container string) bool {
	return strings.HasPrefix(container, "iCloud.")
}
