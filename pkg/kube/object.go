package kube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// KubeObject represents a Kubernetes object with associated functions.
type KubeObject struct {
	node *yaml.RNode
}

// Node returns the underlying RNode of the Kubernetes object.
func (o *KubeObject) Node() *yaml.RNode {
	return o.node
}

// IsNilOrEmpty checks if the Kubernetes object is either nil or empty.
func (o *KubeObject) IsNilOrEmpty() bool {
	return o.node.IsNilOrEmpty()
}

// GetAPIVersion retrieves the APIVersion of the Kubernetes object.
func (o *KubeObject) GetAPIVersion() string {
	return o.node.GetApiVersion()
}

// GetName retrieves the name of the Kubernetes object.
func (o *KubeObject) GetName() string {
	return o.node.GetName()
}

// GetNamespace retrieves the namespace of the Kubernetes object.
func (o *KubeObject) GetNamespace() string {
	return o.node.GetNamespace()
}

// GetKind retrieves the kind of the Kubernetes object.
func (o *KubeObject) GetKind() string {
	return o.node.GetKind()
}

// As unmarshals the Kubernetes object into a Go struct pointed by ptr.
func (o *KubeObject) As(ptr interface{}) error {
	if ptr == nil || reflect.ValueOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("ptr must be a pointer to an object")
	}
	j, err := o.node.MarshalJSON()
	if err != nil {
		return err
	}
	err = json.Unmarshal(j, ptr)
	if err != nil {
		return err
	}
	return nil
}

// GetNestedMap retrieves a nested field from a Kubernetes object
// as a KubeObject. The field must be a map type.
func (o *KubeObject) GetNestedMap(field string) (*KubeObject, error) {
	node, err := o.node.Pipe(yaml.Lookup(field))
	if err != nil {
		return nil, err
	}
	return &KubeObject{node}, nil
}

// GetNestedSlice retrieves a nested field from Kubernetes object
// as a slice of KubeObjects. The field must be a slice type.
func (o *KubeObject) GetNestedSlice(field string) (KubeObjects, error) {
	node, err := o.node.Pipe(yaml.Lookup(field))
	if err != nil {
		return nil, err
	}
	nodes, err := node.Elements()
	if err != nil {
		return nil, err
	}
	var kubeObjects []*KubeObject
	for _, node := range nodes {
		kubeObjects = append(kubeObjects, &KubeObject{node})
	}
	return kubeObjects, nil
}

// MustString returns the YAML string representation of the Kubernetes object.
// It panics if any error occurs while encoding to YAML.
func (o *KubeObject) MustString() string {
	return o.node.MustString()
}

// ParseKubeObjects parses a YAML byte slice into a slice of KubeObjects.
func ParseKubeObjects(in []byte) ([]*KubeObject, error) {
	reader := kio.ByteReader{Reader: bytes.NewReader(in)}
	nodes, err := reader.Read()
	if err != nil {
		return nil, err
	}
	var kubeObjects []*KubeObject
	for _, node := range nodes {
		kubeObjects = append(kubeObjects, &KubeObject{node})
	}
	return kubeObjects, nil
}

// ParseKubeObject parses a single YAML object from a byte slice into a KubeObject.
func ParseKubeObject(in []byte) (*KubeObject, error) {
	node, err := yaml.Parse(string(in))
	if err != nil {
		return nil, err
	}
	return &KubeObject{node}, nil
}

// KubeObjects defines a slice of KubeObject types with convenience methods.
type KubeObjects []*KubeObject

// Len returns the number of KubeObjects.
func (o KubeObjects) Len() int { return len(o) }

// Swap switches the positions of two KubeObjects in the slice.
func (o KubeObjects) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

// MustString concatenates all KubeObjects into a single YAML string,
// separated by '---' (YAML document separator).
func (o KubeObjects) MustString() string {
	var elems []string
	for _, obj := range o {
		elems = append(elems, strings.TrimSpace(obj.MustString()))
	}
	return strings.Join(elems, "\n---\n")
}
