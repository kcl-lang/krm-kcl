package edit

import (
	"kcl-lang.io/krm-kcl/pkg/api"
	"kcl-lang.io/krm-kcl/pkg/api/v1alpha1"
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// WrapResources wraps resources and an optional functionConfig in a resourceList
func WrapResources(nodes []*yaml.RNode, fc *yaml.RNode) (*yaml.RNode, error) {
	var ynodes []*yaml.Node
	for _, rnode := range nodes {
		// Filter KCLRun resources
		if rnode.GetApiVersion() == v1alpha1.KCLRunAPIVersion && rnode.GetKind() == api.KCLRunKind {
			continue
		}
		ynodes = append(ynodes, rnode.YNode())
	}
	m := map[string]interface{}{
		"apiVersion": kio.ResourceListAPIVersion,
		"kind":       kio.ResourceListKind,
		"items":      []interface{}{},
	}
	out, err := yaml.FromMap(m)
	if err != nil {
		return nil, err
	}
	_, err = out.Pipe(
		yaml.Lookup("items"),
		yaml.Append(ynodes...))
	if err != nil {
		return nil, err
	}
	if fc != nil {
		_, err = out.Pipe(
			yaml.SetField("functionConfig", fc))
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

// UnwrapResources unwraps the resources and the functionConfig from a resourceList
func UnwrapResources(nodes []*yaml.RNode) ([]*yaml.RNode, *yaml.RNode, error) {
	var in *yaml.RNode
	if len(nodes) == 0 {
		return []*yaml.RNode{}, nil, nil
	} else if len(nodes) == 1 {
		in = nodes[0]
	} else {
		out, err := WrapResources(nodes, nil)
		if err != nil {
			return nil, nil, errors.Wrap(err)
		}
		in = out
	}
	// Find items
	items, err := in.Pipe(yaml.Lookup("items"))
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	var outs []*yaml.RNode
	// If the items field does not exist, regard the input resource as the output resource.
	if items.IsNil() && !in.IsNil() {
		outs = []*yaml.RNode{in}
	} else {
		outs, err = items.Elements()
		if err != nil {
			return nil, nil, errors.Wrap(err)
		}
	}
	fc, err := in.Pipe(yaml.Lookup("functionConfig"))
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	return outs, fc, nil
}
