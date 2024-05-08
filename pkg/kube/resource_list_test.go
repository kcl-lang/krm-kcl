package kube

import (
	"testing"
)

// TestParseResourceList tests the ParseResourceList function.
func TestParseResourceList(t *testing.T) {
	// Define a YAML string that represents a ResourceList for testing purposes.
	testYAML := `
kind: ResourceList
items:
  - apiVersion: v1
    kind: Pod
    metadata:
      name: test-pod
functionConfig:
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: test-config
`
	// Call ParseResourceList to parse the YAML string into a ResourceList.
	resourceList, err := ParseResourceList([]byte(testYAML))
	if err != nil {
		t.Fatalf("ParseResourceList failed with an error: %v", err)
	}

	// Assert that the ResourceList has one item and a function config.
	expectedItemsLen := 1
	if len(resourceList.Items) != expectedItemsLen {
		t.Errorf("Expected item length to be '%d', but got '%d'", expectedItemsLen, len(resourceList.Items))
	}
	if resourceList.FunctionConfig == nil {
		t.Error("Expected a function config but got nil")
	}
}
