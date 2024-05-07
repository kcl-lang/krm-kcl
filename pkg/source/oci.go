package source

import (
	"fmt"
	"strings"
)

const (
	// OCIScheme is the URL scheme for OCI-based requests
	OCIScheme = "oci"
)

// IsOCI determines whether or not a URL is to be treated as an OCI URL
func IsOCI(src string) bool {
	return strings.HasPrefix(src, fmt.Sprintf("%s://", OCIScheme))
}
