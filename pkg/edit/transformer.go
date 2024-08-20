package edit

import (
	"fmt"

	"github.com/hashicorp/go-getter"
	"kcl-lang.io/krm-kcl/pkg/api"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// The SimpleTransformer type implements the Transformer interface.
var _ Transformer = &SimpleTransformer{}

// Transformer is an interface that defines the transformer operations for
// YAML values.
type Transformer interface {
	// Transform YAML nodes and return error if any error occurs.
	Transform(nodes []*yaml.RNode) ([]*yaml.RNode, error)
}

// SimpleTransformer transforms a set of resources through the provided KCL
// program. It doesn't touch the id annotation. It doesn't copy comments.
type SimpleTransformer struct {
	// Name of the KCL program
	Name string
	// Source is a KCL script which will be run against the resources
	Source string
	// Dependencies are the external dependencies for the KCL code.
	Dependencies []string
	// FunctionConfig is the functionConfig for the function.
	FunctionConfig *yaml.RNode
	// Config is the compile config.
	Config *api.ConfigSpec
	// Getter options
	GetterOptions []getter.ClientOption
}

// Format transformer using the name and source.
func (st *SimpleTransformer) String() string {
	return fmt.Sprintf(
		"name: %v source: %v", st.Name, st.Source)
}

// Transform YAML nodes and return error if any error occurs.
func (st *SimpleTransformer) Transform(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	// 1. Wrap KCLRun resource to KRM the function spec.
	in, err := WrapResources(nodes, st.FunctionConfig)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 2. Run code
	out, err := RunKCLWithConfig(st.Name, st.Source, st.Dependencies, in, st.Config, st.GetterOptions...)

	if err != nil {
		return nil, errors.Wrap(err)
	}

	// 3. Unwrap KRM function spec to KCLRun resource.
	updatedNodes, _, err := UnwrapResources(out)
	return updatedNodes, err
}
