package edit

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-getter"
	"kcl-lang.io/krm-kcl/pkg/api"
	"kcl-lang.io/krm-kcl/pkg/source"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	resourceListOptionName = "resource_list"
	itemsOptionName        = "items"
	paramsOptionName       = "params"
	emptyConfig            = "{}"
	emptyList              = "[]"
)

// Origin of the KCL Source code to be processed
type KCLEntryOrigin struct {
	// KCL Source (file, directory, url, ...)
	source string
	// Temp directory where the kcl code (if any) is allocated. Optional field only retrieved for cleaning purposes.
	tmpDir string
	// Whether the source is a directory or not.
	// isDir bool
}

// Remove any KCLEntryOrigin's temporary folder
func KCLEntryOriginTmpDirCleanup(entry *KCLEntryOrigin) {
	if entry.tmpDir != "" {
		if _, err := os.Stat(entry.tmpDir); err == nil {
			_ = os.RemoveAll(entry.tmpDir)
		}
	}
}

// RunKCL runs a KCL program specified by the given source code or url,
// with the given resource list as input, and returns the resulting KRM resource list.
//
// Parameters:
// - name: a string that represents the name of the KCL program. Not used in the function.
// - source: a string that represents the source code of the KCL program.
// - resourceList: a pointer to a yaml.RNode object that represents the input KRM resource list.
//
// Return:
// A pointer to []*yaml.RNode objects that represent the output YAML objects of the KCL program.
func RunKCL(name, source string, resourceList *yaml.RNode) ([]*yaml.RNode, error) {
	return RunKCLWithConfig(name, source, []string{}, resourceList, nil)
}

// RunKCLWithConfig runs a KCL program specified by the given source code or url,
// with the given resource list as input, and returns the resulting KRM resource list.
//
// Parameters:
// - name: a string that represents the name of the KCL program. Not used in the function.
// - source: a string that represents the source code of the KCL program.
// - resourceList: a pointer to a yaml.RNode object that represents the input KRM resource list.
// - config: a pointer to a ConfigSpec that represents the compile config.
//
// Return:
// A pointer to []*yaml.RNode objects that represent the output YAML objects of the KCL program.
func RunKCLWithConfig(name, source string, dependencies []string, resourceList *yaml.RNode, config *api.ConfigSpec, getterOptions ...getter.ClientOption) ([]*yaml.RNode, error) {
	// 1. Construct KCL code from source.
	entry, err := SourceToTempEntry(source, getterOptions...)
	defer KCLEntryOriginTmpDirCleanup(entry)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	// 2. Construct option list.
	opts, err := constructOptions(resourceList, config)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	// 3. Run the KCL code.
	result := bytes.NewBuffer([]byte{})
	opts.Entries = []string{entry.source}
	opts.Writer = result
	if len(dependencies) > 0 {
		opts.ExternalPackages = dependencies
	}
	err = opts.Run()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	// 4. Parse YAML objects.
	reader := kio.ByteReader{
		Reader:                result,
		OmitReaderAnnotations: true,
	}
	rn, err := reader.Read()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return rn, nil
}

// ToKCLValueString converts YAML value to KCL top level argument json value.
func ToKCLValueString(value *yaml.RNode, defaultValue string) (string, error) {
	if value.IsNil() {
		return defaultValue, nil
	}
	jsonString, err := value.MarshalJSON()
	if err != nil {
		return "", errors.Wrap(err)
	}
	return string(jsonString), nil
}

// SourceToTempEntry convert source to a temp KCL file.
func SourceToTempEntry(src string, opts ...getter.ClientOption) (*KCLEntryOrigin, error) {
	if source.IsOCI(src) {
		// Read code from a OCI source.
		return &KCLEntryOrigin{src, ""}, nil
	} else if source.IsLocal(src) {
		return &KCLEntryOrigin{src, ""}, nil
	} else if source.IsRemoteUrl(src) || source.IsGit(src) || source.IsVCSDomain(src) {
		// Read code from local path or a remote url.
		src, tmpDir, err := source.ReadThroughGetter(src, opts...)
		return &KCLEntryOrigin{src, tmpDir}, err
	} else {
		// May be a inline code source.
		tmpDir, err := os.MkdirTemp("", "kcl-sandbox")
		if err != nil {
			return &KCLEntryOrigin{"", ""}, fmt.Errorf("error creating temp directory: %v", err)
		}
		// Write kcl code in the temp file.
		file := filepath.Join(tmpDir, "prog.k")
		err = os.WriteFile(file, []byte(src), 0666)
		if err != nil {
			return &KCLEntryOrigin{file, tmpDir}, errors.Wrap(err)
		}
		return &KCLEntryOrigin{file, tmpDir}, nil
	}
}
