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

// Trims the protocol prefix from an OCI URL
func TrimOCIPrefix(src string) string {
	return strings.TrimPrefix(src, fmt.Sprintf("%s://", OCIScheme))
}

func OCIPrefix(src string) string {
	return fmt.Sprintf("%s://%s", OCIScheme, src)
}
