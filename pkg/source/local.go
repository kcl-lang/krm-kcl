package source

import (
	"path/filepath"
	"strings"

	"kcl-lang.io/kpm/pkg/reporter"
	"kcl-lang.io/kpm/pkg/runner"
)

// IsLocal determines whether or not a source is to be treated as a local path.
func IsLocal(src string) bool {
	if filepath.IsAbs(src) || strings.HasPrefix(src, ".") {
		return true
	}
	return false
}

// ReadFromLocalSource reads source code from a local file system path source.
//
// Parameters:
// - src: a local file system path.
//
// Return:
// A string containing the source code, and an error if any.
func ReadFromLocalSource(src string) (string, error) {
	// Find the mod root

	modpath, kpmerr := runner.FindModRootFrom(src)
	if kpmerr != nil {
		if kpmerr.Type() != reporter.KclModNotFound {
			return "", kpmerr
		} else {
			modpath = filepath.Dir(src)
		}
	}

	return GetSourceFromDir(modpath)
}
