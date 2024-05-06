package kio

import (
	"kcl-lang.io/krm-kcl/pkg/api/v1alpha1"
	"kcl-lang.io/krm-kcl/pkg/config"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Filter implements kio.Filter
type Filter struct {
	rw *kio.ByteReadWriter
}

// Filter checks each input and ensures that all containers have cpu and memory
// reservations set, otherwise it returns an error.
func (f Filter) Filter(in []*yaml.RNode) ([]*yaml.RNode, error) {
	// Whether has fnCfg in the `functionConfig` field input resource list
	hasFnCfg := f.rw.FunctionConfig != nil
	configs, idxs, err := f.parseConfigs(in)
	if err != nil {
		return nil, err
	}
	for idx, c := range configs {
		var fnCfg *yaml.RNode
		if hasFnCfg {
			fnCfg = f.rw.FunctionConfig
		} else {
			fnCfg = in[idxs[idx]]
		}
		in, err = c.Transform(in, fnCfg)
		if err != nil {
			return nil, err
		}
	}
	return in, nil
}

// parseConfigs parses the input manifests into an API struct.
func (f *Filter) parseConfigs(in []*yaml.RNode) ([]*config.KCLRun, []int, error) {
	var configs []*config.KCLRun
	var idxs []int
	// If KCLRun is not found in the function config, find it in the input manifests
	if f.rw.FunctionConfig == nil {
		for idx, i := range in {
			if i.GetApiVersion() == v1alpha1.KCLRunAPIVersion && i.GetKind() == v1alpha1.KCLRunKind {
				// Parse the input function config.
				config, err := f.parseConfig(i)
				f.rw.FunctionConfig = i
				if err != nil {
					return nil, nil, err
				}
				idxs = append(idxs, idx)
				configs = append(configs, config)
			}
		}
	} else {
		// Parse the input function config.
		config, err := f.parseConfig(f.rw.FunctionConfig)
		if err != nil {
			return nil, nil, err
		}
		configs = append(configs, config)
	}
	return configs, idxs, nil
}

// parseConfig parses the functionConfig into an API struct.
func (f *Filter) parseConfig(in *yaml.RNode) (*config.KCLRun, error) {
	// Parse the input function config.
	var config config.KCLRun
	if err := yaml.Unmarshal([]byte(in.MustString()), &config); err != nil {
		return nil, err
	}
	return &config, nil
}
