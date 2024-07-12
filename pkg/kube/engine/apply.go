package engine

import (
	"context"

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/client-go/dynamic"
	"kcl-lang.io/krm-kcl/pkg/api"
	"kcl-lang.io/krm-kcl/pkg/kube"
)

// ApplyOptions contains options for applying Kubernetes objects.
type ApplyOptions struct {
	DryRun bool
}

// Apply applies a set of Kubernetes objects using the specified options, and returns the resource status after the apply operation.
func (m *Engine) ApplyAll(ctx context.Context, objects []*unstructured.Unstructured, opts *ApplyOptions) (*ResourceStatus, error) {
	return m.applyAllDynamic(ctx, objects, opts)
}

// Apply applies a set of Kubernetes objects using the specified options, and returns the resource status after the apply operation.
func (m *Engine) applyAllDynamic(ctx context.Context, objects []*unstructured.Unstructured, opts *ApplyOptions) (*ResourceStatus, error) {
	entries := make([]ResourceStatusEntry, len(objects))
	for i, obj := range objects {
		planObj, resource, err := m.buildKubernetesResourceByObject(obj)
		if err != nil {
			return nil, errors.Wrap(err, "build dynamic interface failed")
		}
		// Modified equals to input content
		modified, err := kube.CopyAndRemoveMetadataAndStatus(obj).MarshalJSON()
		if err != nil {
			return nil, errors.Wrap(err, "get modified resource failed")
		}
		status, err := m.Read(ctx, obj)
		if err != nil {
			return nil, errors.Wrap(err, "read resource failed")
		}
		if opts.DryRun {
			if status != nil && status.Status == ExistedStatus {
				// Try ServerSideDryRun first
				patchOptions := metav1.PatchOptions{
					DryRun: []string{metav1.DryRunAll},
				}
				current, err := kube.NormalizeServerSideFields(status.Object).MarshalJSON()
				if err != nil {
					return nil, errors.Wrap(err, "normalize server side fields failed")
				}
				patchBody, err := jsonmergepatch.CreateThreeWayJSONMergePatch([]byte("{}"), []byte(modified), []byte(current))
				if err != nil {
					return nil, errors.Wrap(err, "three way json merge patch failed")
				}
				patchedObj, err := resource.Patch(ctx, planObj.GetName(), types.MergePatchType, patchBody, patchOptions)
				if err != nil {
					return nil, errors.Wrap(err, "patch resource failed")
				}
				entries[i] = EntryFromUnstructuredWithStatus(patchedObj, ConfiguredStatus)
			} else {
				// Try ServerSideDryRun first
				createOptions := metav1.CreateOptions{
					DryRun: []string{metav1.DryRunAll},
				}
				createdObj, err := resource.Create(ctx, planObj, createOptions)
				if err != nil {
					return nil, errors.Wrap(err, "create resource failed")
				}
				entries[i] = EntryFromUnstructuredWithStatus(createdObj, CreatedStatus)
			}
		} else {
			if status != nil && status.Status == ExistedStatus {
				current, err := kube.NormalizeServerSideFields(status.Object).MarshalJSON()
				if err != nil {
					return nil, errors.Wrap(err, "normalize server side fields failed")
				}
				patchBody, err := jsonmergepatch.CreateThreeWayJSONMergePatch([]byte("{}"), []byte(modified), []byte(current))
				if err != nil {
					return nil, errors.Wrap(err, "three way json merge patch failed")
				}
				patchedObj, err := resource.Patch(ctx, planObj.GetName(), types.MergePatchType, patchBody, metav1.PatchOptions{FieldManager: api.FieldManager})
				if err != nil {
					return nil, errors.Wrap(err, "patch resource failed")
				}
				entries[i] = EntryFromUnstructuredWithStatus(patchedObj, ConfiguredStatus)
			} else {
				createdObj, err := resource.Create(ctx, planObj, metav1.CreateOptions{})
				if err != nil {
					return nil, errors.Wrap(err, "create resource failed")
				}
				entries[i] = EntryFromUnstructuredWithStatus(createdObj, CreatedStatus)
			}
		}
	}
	return &ResourceStatus{
		Entries: entries,
	}, nil
}

// Read kubernetes Resource by client-go
func (m *Engine) Read(ctx context.Context, obj *unstructured.Unstructured) (*ResourceStatusEntry, error) {
	// Get resource by attribute
	obj, resource, err := m.buildKubernetesResourceByObject(obj)
	if err != nil {
		// Ignore no match error, cause target apiVersion or kind is not installed yet
		if meta.IsNoMatchError(err) {
			entry := EntryFromUnstructuredWithStatus(obj, SkippedStatus)
			return &entry, nil
		}
		return nil, errors.Wrap(err, "build resource")
	}

	// Read resource
	v, err := resource.Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			entry := EntryFromUnstructuredWithStatus(obj, SkippedStatus)
			return &entry, nil
		}
		return nil, errors.Wrap(err, "build dynamic interface failed")
	}

	entry := EntryFromUnstructuredWithStatus(v, ExistedStatus)
	return &entry, nil
}

// buildKubernetesResourceByObject get resource by attribute
func (m *Engine) buildKubernetesResourceByObject(obj *unstructured.Unstructured) (*unstructured.Unstructured, dynamic.ResourceInterface, error) {
	gvk := obj.GroupVersionKind()
	// Get resource by unstructured
	var resource dynamic.ResourceInterface
	resource, err := buildDynamicResource(m.client, m.mapper, &gvk, obj.GetNamespace())
	if err != nil {
		return nil, nil, err
	}
	return obj, resource, nil
}

// buildDynamicResource get resource interface by gvk and namespace
func buildDynamicResource(
	dyn dynamic.Interface, mapper meta.RESTMapper,
	gvk *schema.GroupVersionKind, namespace string,
) (dynamic.ResourceInterface, error) {
	// Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	// Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if namespace == "" {
			namespace = "default"
		}
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}
	return dr, nil
}
