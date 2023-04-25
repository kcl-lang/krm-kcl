package kio

import (
	"github.com/KusionStack/krm-kcl/pkg/api"
	"github.com/KusionStack/krm-kcl/pkg/edit"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// filter implements kio.Filter
type Filter struct {
	rw *kio.ByteReadWriter
}

// Filter checks each input and ensures that all containers have cpu and memory
// reservations set, otherwise it returns an error.
func (f Filter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	api, err := f.parseAPI()
	if err != nil {
		return nil, err
	}
	st := &edit.SimpleTransformer{
		Name:           "kcl-function-run",
		Source:         api.Spec.Source,
		FunctionConfig: f.rw.FunctionConfig,
	}
	return st.Transform(in)
}

// parseAPI parses the functionConfig into an API struct.
func (f *Filter) parseAPI() (*api.API, error) {
	// Parse the input function config.
	var api api.API
	if err := yaml.Unmarshal([]byte(f.rw.FunctionConfig.MustString()), &api); err != nil {
		return nil, err
	}
	return &api, nil
}
