package kube

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ObjMetadata contains metadata information for a Kubernetes object.
type ObjMetadata struct {
	// Namespace of the object
	Namespace string
	// Name of the object
	Name string
	// GroupKind of the object
	GroupKind schema.GroupKind
}

// UnstructuredToObjMetadata extracts the identifying information from an
// Unstructured object and returns it as ObjMetadata object.
func UnstructuredToObjMetadata(obj *unstructured.Unstructured) ObjMetadata {
	return ObjMetadata{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
		GroupKind: obj.GroupVersionKind().GroupKind(),
	}
}

// ID returns the object ID in the format <kind>/<namespace>/<name>.
func (obj *ObjMetadata) ID() string {
	var builder strings.Builder
	builder.WriteString(obj.GroupKind.Kind + "/")
	if obj.Namespace != "" {
		builder.WriteString(obj.Namespace + "/")
	}
	builder.WriteString(obj.Name)
	return builder.String()
}
