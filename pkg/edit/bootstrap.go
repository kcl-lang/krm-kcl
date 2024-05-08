package edit

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kcl-lang.io/cli/pkg/options"
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
	return RunKCLWithConfig(name, source, resourceList, nil)
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
func RunKCLWithConfig(name, source string, resourceList *yaml.RNode, config *api.ConfigSpec) ([]*yaml.RNode, error) {
	// 1. Construct KCL code from source.
	entry, err := SourceToTempEntry(source)
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
	opts.Entries = []string{entry}
	opts.Writer = result
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
func SourceToTempEntry(src string) (string, error) {
	if source.IsOCI(src) {
		// Read code from a OCI source.
		return src, nil
	} else if source.IsLocal(src) {
		return src, nil
	} else if source.IsRemoteUrl(src) || source.IsGit(src) || source.IsVCSDomain(src) {
		// Read code from local path or a remote url.
		return source.ReadThroughGetter(src)
	} else {
		// May be a inline code source.
		tmpDir, err := os.MkdirTemp("", "sandbox")
		if err != nil {
			return "", fmt.Errorf("error creating temp directory: %v", err)
		}
		// Write kcl code in the temp file.
		file := filepath.Join(tmpDir, "prog.k")
		err = os.WriteFile(file, []byte(src), 0666)
		if err != nil {
			return "", errors.Wrap(err)
		}
		return file, nil
	}
}

func constructOptions(resourceList *yaml.RNode, config *api.ConfigSpec) (*options.RunOptions, error) {
	resourceListOptionKCLValue, err := ToKCLValueString(resourceList, emptyConfig)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	v, err := resourceList.Pipe(yaml.Lookup("items"))
	if err != nil {
		return nil, errors.Wrap(err)
	}
	itemsOptionKCLValue, err := ToKCLValueString(v, emptyList)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	v, err = resourceList.Pipe(yaml.Lookup("functionConfig", "spec", "params"))
	if err != nil {
		return nil, errors.Wrap(err)
	}
	paramsOptionKCLValue, err := ToKCLValueString(v, emptyConfig)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	// 4. Read environment variables.
	pathOptionKCLValue := os.Getenv("PATH")

	// 5. Read Env map
	envMapOptionKCLValue, err := getEnvMapOptionKCLValue()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	opts := options.NewRunOptions()
	opts.NoStyle = true
	if config != nil {
		opts.Debug = config.Debug
		opts.DisableNone = config.DisableNone
		opts.Overrides = config.Overrides
		opts.PathSelectors = config.PathSelectors
		opts.Settings = config.Settings
		opts.ShowHidden = config.ShowHidden
		opts.SortKeys = config.SortKeys
		opts.StrictRangeCheck = config.StrictRangeCheck
		opts.Vendor = config.Vendor
		opts.Arguments = config.Arguments
	}
	opts.Arguments = append(opts.Arguments,
		// resource_list
		fmt.Sprintf("%s=%s", resourceListOptionName, resourceListOptionKCLValue),
		// resource.items
		fmt.Sprintf("%s=%s", itemsOptionName, itemsOptionKCLValue),
		// resource.functionConfig.spec.params
		fmt.Sprintf("%s=%s", paramsOptionName, paramsOptionKCLValue),
		// environment variable example (PATH)
		fmt.Sprintf("PATH=%s", pathOptionKCLValue),
		// environment map example (option("env"))
		fmt.Sprintf("env=%s", envMapOptionKCLValue),
	)
	return opts, nil
}

// getEnvMapOptionKCLValue retrieves the environment map from the KCL 'option("env")' function.
func getEnvMapOptionKCLValue() (string, error) {
	envMap := make(map[string]string)
	env := os.Environ()
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		envMap[pair[0]] = pair[1]
	}

	envMapInterface := make(map[string]interface{})
	for k, v := range envMap {
		envMapInterface[k] = v
	}

	v, err := yaml.FromMap(envMapInterface)
	if err != nil {
		return "", errors.Wrap(err)
	}

	// 4. Convert the YAML RNode to its KCL value string representation.
	envMapOptionKCLValue, err := ToKCLValueString(v, emptyConfig)
	if err != nil {
		return "", errors.Wrap(err)
	}

	return envMapOptionKCLValue, nil
}
