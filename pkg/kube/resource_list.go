package kube

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ResourceList represents a list of Kubernetes objects typically used for configuration purposes.
type ResourceList struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	Items             KubeObjects `yaml:"items" json:"items"`                                       // Items is a slice of Kubernetes objects.
	FunctionConfig    *KubeObject `yaml:"functionConfig,omitempty" json:"functionConfig,omitempty"` // FunctionConfig is an optional KubeObject that describes the function configuration.
}

// ParseResourceList takes a byte slice of a YAML ResourceList and parses it into a ResourceList structure.
func ParseResourceList(in []byte) (*ResourceList, error) {
	rl := &ResourceList{}
	o, err := ParseKubeObject(in)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input bytes: %w", err)
	}
	// Ensure that the parsed object is of kind ResourceList.
	if o.GetKind() != kio.ResourceListKind {
		return nil, fmt.Errorf("input was of unexpected kind %q; expected ResourceList", o.GetKind())
	}
	// Parse FunctionConfig if present.
	if fc, err := o.GetNestedMap("functionConfig"); err == nil {
		rl.FunctionConfig = fc
	}
	// Parse Items. Items can be empty (e.g., an input ResourceList for a generator function may not have items).
	if items, err := o.GetNestedSlice("items"); err == nil {
		rl.Items = items
	}
	return rl, nil
}
