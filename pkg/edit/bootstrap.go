package edit

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

// RunKCL runs the KCL script and modify resourceList.
func RunKCL(name, source string, resourceList *yaml.RNode) (*yaml.RNode, error) {
	resourceListOptionKCLValue, err := ToKCLValueString(resourceList)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	buffer := new(bytes.Buffer)
	codeTemplate.Execute(buffer, &struct{ Source string }{source})
	// Create temp files.
	tmpDir, err := ioutil.TempDir("", "sandbox")
	if err != nil {
		return nil, fmt.Errorf("error creating temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	// Write kcl code in the temp file.
	kFile := filepath.Join(tmpDir, "prog.k")
	err = os.WriteFile(kFile, buffer.Bytes(), 0666)
	if err != nil {
		return nil, err
	}

	r, err := kclvm.Run(kFile, kclvm.WithOptions(fmt.Sprintf("%s=%s", resourceListOptionName, resourceListOptionKCLValue)))
	if err != nil {
		return nil, errors.Wrap(stripansi.Strip(err.Error()))
	}

	rn, err := yaml.Parse(r.GetRawYamlResult())
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return rn, nil
}

// ToKCLValueString converts YAML value to KCL top level argument json value.
func ToKCLValueString(resourceList *yaml.RNode) (string, error) {
	jsonString, err := resourceList.MarshalJSON()
	if err != nil {
		return "", errors.Wrap(err)
	}
	// In KCL, `true`, `false` and `null` are denoted by `True`, `False` and `None`.
	result := strings.Replace(string(jsonString), ": true", ": True", -1)
	result = strings.Replace(result, ": false", ": False", -1)
	result = strings.Replace(result, ": null", ": None", -1)
	return result, nil
}
