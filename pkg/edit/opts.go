package edit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kcl-lang.io/cli/pkg/options"
	"kcl-lang.io/kpm/pkg/client"
	"kcl-lang.io/krm-kcl/pkg/api"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// LoadDepsFrom parses the kcl external package option from a path.
// It will find `kcl.mod` recursively from the path, resolve deps
// in the `kcl.mod` and return the option. If not found, return the
// empty option.
func LoadDepListFromConfig(cli *client.KpmClient, dependencies string) ([]string, error) {
	if cli == nil {
		return nil, nil
	}
	cli.SetLogWriter(nil)
	modData := fmt.Sprintf("[package]\n\n[dependencies]\n%s", dependencies)
	// May be a inline code source.
	tmpDir, err := os.MkdirTemp("", "sandbox")
	defer os.Remove(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory: %v", err)
	}
	// Write kcl code in the temp file.
	tempFile := filepath.Join(tmpDir, "kcl.mod")
	err = os.WriteFile(tempFile, []byte(modData), 0666)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	pkg, err := cli.LoadPkgFromPath(tmpDir)
	if err != nil {
		return nil, err
	}
	depsMap, err := cli.ResolveDepsIntoMap(pkg)
	if err != nil {
		return nil, err
	}
	deps := []string{}
	for depName, depPath := range depsMap {
		deps = append(deps, fmt.Sprintf("%s=%s", depName, depPath))
	}
	return deps, nil
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
