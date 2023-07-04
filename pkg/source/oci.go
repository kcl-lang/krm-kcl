package source

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kcl-lang.io/kpm/pkg/errors"
	"kcl-lang.io/kpm/pkg/oci"
	"kcl-lang.io/kpm/pkg/opt"
	"kcl-lang.io/kpm/pkg/settings"
	"kcl-lang.io/kpm/pkg/utils"
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
	// 1. Parse the OCI url.
	ociOpts, err := opt.ParseOciUrl(src)

	if err == errors.IsOciRef {
		settings, err := settings.GetSettings()
		if err != nil {
			return src, err
		}

		ociOpts, err = opt.ParseOciRef(filepath.Join(settings.DefaultOciRepo(), src))
		if err != nil {
			return src, err
		}
	} else if err != nil {
		return src, err
	}

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return src, errors.InternalBug
	}
	// clean the temp dir.
	defer os.RemoveAll(tmpDir)

	localPath := ociOpts.AddStoragePathSuffix(tmpDir)

	// 2. Pull the tarball from OCI.
	err = oci.Pull(localPath, ociOpts.Reg, ociOpts.Repo, ociOpts.Tag)

	if err != nil {
		return src, err
	}

	// 3. Get the (*.tar) file path.
	matches, err := filepath.Glob(filepath.Join(localPath, TarPattern))
	if err != nil || len(matches) != 1 {
		return src, errors.FailedPullFromOci
	}
	tarPath := matches[0]

	// 4. Extract the package tarball into a directory with the same name.
	// e.g.
	// 'xxx/xxx/xxx/test.tar' will be extracted to the directory 'xxx/xxx/xxx/test'.
	destDir := strings.TrimSuffix(tarPath, filepath.Ext(tarPath))
	err = utils.UnTarDir(tarPath, destDir)
	if err != nil {
		return src, err
	}

	// 5. Read source from the package entry.
	return GetSourceFromDir(destDir)
}
