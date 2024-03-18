package source

import (
	"path/filepath"
	"strings"

	kpmreporter "kcl-lang.io/kpm/pkg/reporter"
	kpmrunner "kcl-lang.io/kpm/pkg/runner"
)

// IsLocal determines whether or not a source is to be treated as a local path.
func IsLocal(src string) bool {
	if filepath.IsAbs(src) || strings.HasPrefix(src, ".") {
		return true
	}
	return false
}

// ReadFromOCISource reads source code from an OCI (Open Container Initiative) source.
//
// Parameters:
// - src: a string containing the OCI source URL.
//
// Return:
// A string containing the source code, and an error if any.
func ReadFromLocalSource(src string) (string, error) {
	// Find the mod root

	modpath, kpmerr := kpmrunner.FindModRootFrom(src)
	if kpmerr != nil {
		if kpmerr.Type() != kpmreporter.KclModNotFound {
			return "", kpmerr
		} else {
			modpath = filepath.Dir(src)
		}
	}

	return GetSourceFromDir(modpath)
}
