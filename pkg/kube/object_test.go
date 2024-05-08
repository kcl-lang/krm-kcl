package kube

import "testing"

// TestParseKubeObject tests the ParseKubeObject function.
func TestParseKubeObject(t *testing.T) {
	// Define a YAML string of a kube object for testing purposes.
	testYAML := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
`
	// Call ParseKubeObject to parse the YAML string into a KubeObject.
	kubeObj, err := ParseKubeObject([]byte(testYAML))
	if err != nil {
		t.Errorf("ParseKubeObject failed with error: %s", err)
	}

	// Verify the parsed object meets expectations.
	if kubeObj.GetName() != "test-pod" {
		t.Errorf("Expected Pod name to be 'test-pod', but got '%s'", kubeObj.GetName())
	}
}

// TestKubeObjectGetName tests the GetName method.
func TestKubeObjectGetName(t *testing.T) {
	// Directly construct a KubeObject for testing the GetName method.
	testYAML := `
apiVersion: v1
kind: Pod
metadata:
  name: another-pod
  namespace: test-namespace
`
	kubeObj, err := ParseKubeObject([]byte(testYAML))
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	expectedName := "another-pod"
	if name := kubeObj.GetName(); name != expectedName {
		t.Errorf("Expected name to be '%s', but got '%s'", expectedName, name)
	}
}
