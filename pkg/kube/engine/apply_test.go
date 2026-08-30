package engine

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kcl-lang.io/krm-kcl/pkg/kube"
)

// newEngineForTest returns an Engine wired to whatever cluster the ambient
// kubeconfig points at, skipping the test when no cluster is reachable.
//
// Availability is probed with a real (cheap) API call rather than inferred from
// the host OS: a kubeconfig can be absent, stale, or point at a cluster that is
// down on any platform, and in those cases the test has nothing to assert.
func newEngineForTest(t *testing.T) *Engine {
	t.Helper()

	engine, err := NewDefaultEngine()
	if err != nil {
		t.Skipf("skipping: no usable kubeconfig (%v)", err)
	}
	gvr := schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}
	if _, err := engine.Client().Resource(gvr).List(context.Background(), metav1.ListOptions{Limit: 1}); err != nil {
		t.Skipf("skipping: cluster is not reachable (%v)", err)
	}
	return engine
}

func TestEngineApplyAll(t *testing.T) {
	ctx := context.TODO()
	objects := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "my-deployment",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"app": "my-app",
						},
					},
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app": "my-app",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{
								map[string]interface{}{
									"name":  "my-container",
									"image": "nginx:latest",
									"ports": []interface{}{
										map[string]interface{}{
											"containerPort": int64(80),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":      "my-service",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"selector": map[string]interface{}{
						"app": "my-app",
					},
					"ports": []interface{}{
						map[string]interface{}{
							"protocol":   "TCP",
							"port":       int64(80),
							"targetPort": int64(8080),
						},
					},
				},
			},
		},
	}
	crd, err := kube.YamlByteToUnstructured([]byte(`
{
	"apiVersion": "apiextensions.k8s.io/v1",
	"kind": "CustomResourceDefinition",
	"metadata": {
		"name": "mycrds.example.com"
	},
	"spec": {
		"group": "example.com",
		"names": {
			"kind": "MyCR",
			"listKind": "MyCRList",
			"plural": "mycrds",
			"singular": "mycrd"
		},
		"scope": "Namespaced",
		"versions": [
			{
				"name": "v1",
				"served": true,
				"storage": true,
				"schema": {
					"openAPIV3Schema": {
						"type": "object",
						"properties": {
							"spec": {
								"type": "string"
							}
						},
						"required": [
							"spec"
						]
					}
				}
			}
		]
	}
}
`))
	if err != nil {
		t.Fatalf("Generate CRD error: %v", err)
	}
	objects = append(objects, crd)

	engine := newEngineForTest(t)

	// Execute
	status, err := engine.ApplyAll(ctx, objects, &ApplyOptions{})
	// Verify
	if err != nil {
		t.Fatalf("Apply returned unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("Apply returned nil status, want one entry per object")
	}
	if len(status.Entries) != len(objects) {
		t.Errorf("Apply returned unexpected number of status entries: got %d, want %d", len(status.Entries), len(objects))
	}
	// Double Execute, applying the same objects again must be idempotent.
	status, err = engine.ApplyAll(ctx, objects, &ApplyOptions{})
	// Double Verify
	if err != nil {
		t.Fatalf("Apply returned unexpected error: %v", err)
	}
	if status == nil {
		t.Fatal("Re-apply returned nil status, want one entry per object")
	}
	if len(status.Entries) != len(objects) {
		t.Errorf("Apply returned unexpected number of status entries: got %d, want %d", len(status.Entries), len(objects))
	}
}
