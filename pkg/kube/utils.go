package kube

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	krmyaml "kcl-lang.io/krm-kcl/pkg/yaml"
)

// JsonByteToRawExtension converts a JSON byte array to a runtime.RawExtension object.
func JsonByteToRawExtension(jsonByte []byte) (runtime.RawExtension, error) {
	o, err := JsonByteToUnstructured(jsonByte)
	if err != nil {
		return runtime.RawExtension{}, err
	}
	return UnstructuredToRawExtension(o)
}

// YamlByteToUnstructured converts a Yaml byte array to an unstructured.Unstructured object.
func YamlByteToUnstructured(yamlByte []byte) (*unstructured.Unstructured, error) {
	var data map[string]interface{}
	err := yaml.Unmarshal(yamlByte, &data)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{Object: data}
	return u, nil
}

// YamlByteToUnstructured returns the manifests list from the YAML stream data.
func YamlStreamByteToUnstructuredList(yamlByte []byte) (result []*unstructured.Unstructured, err error) {
	bytes, err := krmyaml.SplitDocuments(string(yamlByte))
	if err != nil {
		return
	}
	for _, b := range bytes {
		var data interface{}
		err = yaml.Unmarshal([]byte(b), &data)
		if err != nil {
			return
		}

		// Convert map[any]any to map[string]any
		normalizedData, err := NormalizeMap(data)
		if err != nil {
			return nil, err
		}

		result = append(result, &unstructured.Unstructured{
			Object: normalizedData.(map[string]interface{}),
		})
	}
	return
}

// JsonByteToUnstructured converts a JSON byte array to an unstructured.Unstructured object.
func JsonByteToUnstructured(jsonByte []byte) (*unstructured.Unstructured, error) {
	var data map[string]interface{}
	err := json.Unmarshal(jsonByte, &data)
	if err != nil {
		return nil, err
	}
	u := &unstructured.Unstructured{Object: data}
	return u, nil
}

// UnstructuredToRawExtension converts an unstructured.Unstructured object to a runtime.RawExtension object.
func UnstructuredToRawExtension(obj *unstructured.Unstructured) (runtime.RawExtension, error) {
	if obj == nil {
		return runtime.RawExtension{}, nil
	}
	raw, err := obj.MarshalJSON()
	if err != nil {
		return runtime.RawExtension{}, err
	}
	return runtime.RawExtension{Raw: raw}, nil
}

// ObjToRawExtension converts an arbitrary object to a runtime.RawExtension object using JSON encoding.
func ObjToRawExtension(obj interface{}) (runtime.RawExtension, error) {
	if obj == nil {
		return runtime.RawExtension{}, nil
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return runtime.RawExtension{}, err
	}
	return runtime.RawExtension{Raw: raw}, nil
}

// HasDrifted checks if the current object has drifted from the existing object.
func HasDrifted(existing, current *unstructured.Unstructured) bool {
	// Check if the current object has an empty resource version, indicating it has drifted.
	if current.GetResourceVersion() == "" {
		return true
	}
	// Perform a semantic equality check on labels and annotations, return true if they differ.
	if !equality.Semantic.DeepEqual(current.GetLabels(), existing.GetLabels()) {
		return true
	}
	if !equality.Semantic.DeepEqual(current.GetAnnotations(), existing.GetAnnotations()) {
		return true
	}
	existingObj := CopyAndRemoveMetadataAndStatus(existing)
	currentObj := CopyAndRemoveMetadataAndStatus(current)
	return !equality.Semantic.DeepEqual(currentObj.Object, existingObj.Object)
}

// RemoveMetadataAndStatus removes the metadata and status fields from the object to prepare for semantic equality check.
func CopyAndRemoveMetadataAndStatus(object *unstructured.Unstructured) *unstructured.Unstructured {
	deepCopy := object.DeepCopy()
	return NormalizeServerSideFields(deepCopy)
}

// UnstructuredID returns the object ID in the format <kind>/<namespace>/<name>.
func UnstructuredID(obj *unstructured.Unstructured) string {
	m := UnstructuredToObjMetadata(obj)
	return m.ID()
}

// NormalizeServerSideFields removes the metadata and status fields from the object to
// prepare for semantic equality check.
func NormalizeServerSideFields(ur *unstructured.Unstructured) *unstructured.Unstructured {
	const metadata = "metadata"
	unstructured.RemoveNestedField(ur.Object, "status")
	unstructured.RemoveNestedField(ur.Object, metadata, "resourceVersion")
	unstructured.RemoveNestedField(ur.Object, metadata, "creationTimestamp")
	unstructured.RemoveNestedField(ur.Object, metadata, "selfLink")
	unstructured.RemoveNestedField(ur.Object, metadata, "uid")
	unstructured.RemoveNestedField(ur.Object, metadata, "generation")
	unstructured.RemoveNestedField(ur.Object, metadata, "managedFields")
	return ur
}

// NormalizeMap converts map[interface{}]interface{} to map[string]interface{}
func NormalizeMap(input interface{}) (interface{}, error) {
	switch in := input.(type) {
	case map[interface{}]interface{}:
		normalized := make(map[string]interface{})
		for key, value := range in {
			strKey, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("found non-string key in the map")
			}
			normalizedValue, err := NormalizeMap(value)
			if err != nil {
				return nil, err
			}
			normalized[strKey] = normalizedValue
		}
		return normalized, nil
	case []interface{}:
		for i, v := range in {
			normalizedValue, err := NormalizeMap(v)
			if err != nil {
				return nil, err
			}
			in[i] = normalizedValue
		}
		return in, nil
	default:
		return input, nil
	}
}
