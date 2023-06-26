package source

import (
	"path/filepath"
	"strings"
)

// IsLocal determines whether or not a source is to be treated as a local path.
func IsLocal(src string) bool {
	if filepath.IsAbs(src) || strings.HasPrefix(src, ".") {
		return true
	}
	return false
}
