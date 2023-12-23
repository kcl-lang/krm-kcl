package source

import (
	"fmt"
	"strings"
)

const (
	// OCIScheme is the URL scheme for OCI-based requests
	OCIScheme = "oci"
	// TarPattern is the wildcard pattern for tar files.
	TarPattern = "*.tar"
)

// IsOCI determines whether or not a URL is to be treated as an OCI URL
func IsOCI(src string) bool {
	return strings.HasPrefix(src, fmt.Sprintf("%s://", OCIScheme))
}

// ReadFromOCISource reads source code from an OCI (Open Container Initiative) source.
//
// Parameters:
// - src: a string containing the OCI source URL.
//
// Return:
// A string containing the source code, and an error if any.
func ReadFromOCISource(src string) (string, error) {
	return src, nil
}
