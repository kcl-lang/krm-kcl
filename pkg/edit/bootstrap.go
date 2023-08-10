package edit

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kcl "kcl-lang.io/kcl-go"
	src "kcl-lang.io/krm-kcl/pkg/source"

	"github.com/acarl005/stripansi"
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
	// 1. Construct KCL code from source.
	file, err := SourceToTempFile(source)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer os.RemoveAll(file)

	// 2. Construct option list.
	opts, err := constructOptions(resourceList)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 3. Run the KCL code.
	r, err := kcl.Run(file, opts...)
	if err != nil {
		return nil, errors.Wrap(stripansi.Strip(err.Error()))
	}

	// 4. Parse YAML objects.
	reader := kio.ByteReader{
		Reader:                strings.NewReader(r.GetRawYamlResult()),
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
	// In KCL, `true`, `false` and `null` are denoted by `True`, `False` and `None`.
	result := strings.Replace(string(jsonString), ": true", ": True", -1)
	result = strings.Replace(result, ": false", ": False", -1)
	result = strings.Replace(result, ": null", ": None", -1)
	return result, nil
}

// SourceToTempFile convert source to a temp KCL file.
func SourceToTempFile(source string) (string, error) {
	// 1. Construct KCL code from source.
	localeSource, err := src.LocaleSource(source)
	if err != nil {
		return "", errors.Wrap(err)
	}
	// Create temp files.
	tmpDir, err := os.MkdirTemp("", "sandbox")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v", err)
	}
	// Write kcl code in the temp file.
	file := filepath.Join(tmpDir, "prog.k")
	err = os.WriteFile(file, []byte(localeSource), 0666)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return file, nil
}

func constructOptions(resourceList *yaml.RNode) ([]kcl.Option, error) {
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
	paramsptionKCLValue, err := ToKCLValueString(v, emptyConfig)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	opts := []kcl.Option{
		kcl.WithOptions(fmt.Sprintf("%s=%s", resourceListOptionName, resourceListOptionKCLValue)),
		// resource.items
		kcl.WithOptions(fmt.Sprintf("%s=%s", itemsOptionName, itemsOptionKCLValue)),
		// resource.functionConfig.spec.params
		kcl.WithOptions(fmt.Sprintf("%s=%s", paramsOptionName, paramsptionKCLValue)),
	}
	return opts, nil
}
