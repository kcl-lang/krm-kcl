package api

import "sigs.k8s.io/kustomize/kyaml/yaml"

// API defines the input API schema as a struct
type API struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Spec              struct {
		// Source is a required field for providing a KCL script inline.
		Source string `json:"source" yaml:"source"`
		// Params are the parameters in key-value pairs format.
		Params map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	} `json:"spec" yaml:"spec"`
}
