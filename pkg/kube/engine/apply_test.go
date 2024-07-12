package engine

import (
	"context"
	"runtime"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kcl-lang.io/krm-kcl/pkg/kube"
)

func TestEngineApplyAll(t *testing.T) {
	if runtime.GOOS == "linux" {
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
			t.Errorf("Generate CRD error: %v", err)
		}
		objects = append(objects, crd)
		engine, err := NewDefaultEngine()
		if err != nil {
			t.Errorf("New default engine error: %v", err)
		}
		// Execute
		status, err := engine.ApplyAll(ctx, objects, &ApplyOptions{})
		// Verify
		if err != nil {
			t.Errorf("Apply returned unexpected error: %v", err)
		}
		if status != nil && len(status.Entries) != len(objects) {
			t.Errorf("Apply returned unexpected number of status entries: got %d, want %d", len(status.Entries), len(objects))
		}
		// Double Execute
		status, err = engine.ApplyAll(ctx, objects, &ApplyOptions{})
		// Double Verify
		if err != nil {
			t.Errorf("Apply returned unexpected error: %v", err)
		}
		if status != nil && len(status.Entries) != len(objects) {
			t.Errorf("Apply returned unexpected number of status entries: got %d, want %d", len(status.Entries), len(objects))
		}
	}
}
