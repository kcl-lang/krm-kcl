package source

// LocaleSource is used to get the source code content of different types of sources.
// The parameter src represents the path or content of the source code.
// If the source code is an OCI source, the code content read from the OCI source will be returned.
// If the source code is a local path, a remote URL, or a Git source, the code content will be obtained through the Getter from the source.
// If the source code is a pure code source, the source code content will be returned directly.
// If an error occurs during the acquisition process, an error message will be returned.
func LocaleSource(src string) (string, error) {
	if IsOCI(src) {
		// Read code from a OCI source.
		return ReadFromOCISource(src)
	} else if IsLocal(src) || IsRemoteUrl(src) || IsGit(src) || IsVCSDomain(src) {
		// Read code from local path or a remote url.
		return ReadThroughGetter(src)
	} else {
		// May be a pure code source.
		return src, nil
	}
}
