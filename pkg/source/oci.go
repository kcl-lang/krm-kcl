package source

import (
	"fmt"
	"net/url"
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

// ParseOCIReference parses an OCI reference (the part after "oci://") and
// returns the cleaned reference URL together with the optional tag.
//
// The tag may appear in either of two places:
//
//   - As the last colon-separated segment of the repository path, e.g.
//     "ghcr.io/kcl-lang/helloworld:0.1.0".
//   - As a "?tag=X" query parameter, e.g. "host:7900/myapp?tag=0.1.0".
//
// When both are present, the query parameter wins.
//
// The port portion of the registry (":port" in "host:port") is preserved on
// the returned reference. A naive split on ":" would mistake it for the tag
// separator and break OCI URLs that target a non-default registry port (see
// kcl-lang/krm-kcl issue #450).
func ParseOCIReference(ref string) (string, string) {
	// 1. Strip and parse any query string. A query tag, when present, takes
	//    precedence over a path-suffix tag.
	var queryTag string
	if idx := strings.Index(ref, "?"); idx != -1 {
		if q, err := url.ParseQuery(ref[idx+1:]); err == nil {
			queryTag = q.Get("tag")
		}
		ref = ref[:idx]
	}

	// 2. Look for a ":tag" suffix on the path. We split at the first slash to
	//    isolate the host (which may legitimately contain a ":port") from the
	//    repository path, then take the last colon in the path as the tag.
	var pathTag string
	if pathIdx := strings.Index(ref, "/"); pathIdx != -1 {
		pathPart := ref[pathIdx+1:]
		if colon := strings.LastIndex(pathPart, ":"); colon != -1 {
			pathTag = pathPart[colon+1:]
			ref = ref[:pathIdx+1+colon]
		}
	}

	tag := pathTag
	if queryTag != "" {
		tag = queryTag
	}
	return OCIPrefix(ref), tag
}
