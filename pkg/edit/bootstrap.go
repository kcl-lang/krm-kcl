package edit

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	src "kcl-lang.io/krm-kcl/pkg/source"

	"github.com/Masterminds/sprig/v3"
	"github.com/acarl005/stripansi"
	"kusionstack.io/kclvm-go"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

//go:embed _code.tmpl
var codeTemplateString string

var codeTemplate = template.Must(template.New("code.tmpl").Funcs(sprig.TxtFuncMap()).Parse(codeTemplateString))

const resourceListOptionName = "resource_list"

// RunKCL runs a KCL program specified by the given source code or url,
// with the given resource list as input, and returns the resulting KRM resource list.
//
// Parameters:
// - name: a string that represents the name of the KCL program. Not used in the function.
// - source: a string that represents the source code of the KCL program.
// - resourceList: a pointer to a yaml.RNode object that represents the input KRM resource list.
//
// Return:
// A pointer to a yaml.RNode object that represents the output YAML objects of the KCL program, and an error if any.
func RunKCL(name, source string, resourceList *yaml.RNode) (*yaml.RNode, error) {
	// 1. Construct KCL code from source.
	file, err := SourceToTempFile(source)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer os.RemoveAll(file)

	// 2. Construct option list.
	resourceListOptionKCLValue, err := ToKCLValueString(resourceList)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 3. Run the KCL code.
	r, err := kclvm.Run(file, kclvm.WithOptions(fmt.Sprintf("%s=%s", resourceListOptionName, resourceListOptionKCLValue)))
	if err != nil {
		return nil, errors.Wrap(stripansi.Strip(err.Error()))
	}

	// 4. Parse YAML objects.
	rn, err := yaml.Parse(r.GetRawYamlResult())
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return rn, nil
}

// ToKCLValueString converts YAML value to KCL top level argument json value.
func ToKCLValueString(value *yaml.RNode) (string, error) {
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
	buffer := new(bytes.Buffer)
	codeTemplate.Execute(buffer, &struct{ Source string }{localeSource})
	// Create temp files.
	tmpDir, err := os.MkdirTemp("", "sandbox")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v", err)
	}
	// Write kcl code in the temp file.
	file := filepath.Join(tmpDir, "prog.k")
	err = os.WriteFile(file, buffer.Bytes(), 0666)
	if err != nil {
		return "", errors.Wrap(err)
	}
	return file, nil
}
