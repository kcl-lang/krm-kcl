package edit

import (
	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// WrapResources wraps resources and an optional functionConfig in a resourceList
func WrapResources(nodes []*yaml.RNode, fc *yaml.RNode) (*yaml.RNode, error) {
	var ynodes []*yaml.Node
	for _, rnode := range nodes {
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
func UnwrapResources(in *yaml.RNode) ([]*yaml.RNode, *yaml.RNode, error) {
	items, err := in.Pipe(yaml.Lookup("items"))
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	nodes, err := items.Elements()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	fc, err := in.Pipe(yaml.Lookup("functionConfig"))
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}
	return nodes, fc, nil
}
