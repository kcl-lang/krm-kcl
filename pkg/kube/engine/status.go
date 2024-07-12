package engine

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kcl-lang.io/krm-kcl/pkg/kube"
)

// ResourceStatus represents the status of a Kubernetes resource.
type ResourceStatus struct {
	Entries []ResourceStatusEntry
}

// ResourceStatusEntry holds the status information for a resource.
type ResourceStatusEntry struct {
	Object *unstructured.Unstructured
	// ObjMetadata holds the unique identifier of this entry.
	ObjMetadata kube.ObjMetadata
	// GroupVersion holds the API group version of this entry.
	GroupVersion string
	// Subject represents the Object ID in the format 'kind/namespace/name'.
	Subject string
	// Status represents the Status type taken by the reconciler for this object.
	Status Status
}

// Status is a type representing the status of a resource.
type Status string

const (
	// CreatedStatus represents the creation of a new object.
	CreatedStatus Status = "created"
	// ConfiguredStatus represents the update of an existing object.
	ConfiguredStatus Status = "configured"
	// UnchangedStatus represents the absence of any Status to an object.
	UnchangedStatus Status = "unchanged"
	// DeletedStatus represents the deletion of an object.
	DeletedStatus Status = "deleted"
	// SkippedStatus represents the fact that no Status was performed on an object
	// due to the object being excluded from the reconciliation.
	SkippedStatus Status = "skipped"
	// ExistedStatus represents a resource is existed in the cluster.
	ExistedStatus Status = "existed"
	// UnknownStatus represents an unknown Status.
	UnknownStatus Status = "unknown"
)

// EntryFromUnstructuredWithStatus creates a ResourceStatusEntry from an unstructured object along with the specified status.
func EntryFromUnstructuredWithStatus(o *unstructured.Unstructured, status Status) ResourceStatusEntry {
	return ResourceStatusEntry{
		Object:       o,
		ObjMetadata:  kube.UnstructuredToObjMetadata(o),
		GroupVersion: o.GroupVersionKind().Version,
		Subject:      kube.UnstructuredID(o),
		Status:       status,
	}
}
