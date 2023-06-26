package source

import (
	"fmt"
	"strings"
)

const (
	// HttpScheme is the URL scheme for Http-based requests
	HttpScheme = "http"
	// HttpsScheme is the URL scheme for Https-based requests
	HttpsScheme = "https"
)

// IsRemoteUrl determines whether or not a URL is to be treated as an URL
func IsRemoteUrl(src string) bool {
	return strings.HasPrefix(src, fmt.Sprintf("%s://", HttpScheme)) || strings.HasPrefix(src, fmt.Sprintf("%s://", HttpsScheme))
}
