package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"kcl-lang.io/kpm/pkg/settings"
	"kcl-lang.io/krm-kcl/pkg/api/v1alpha1"
	"kcl-lang.io/krm-kcl/pkg/edit"
)

const (
	// ConfigMapAPIVersion represents the API version for the ConfigMap resource.
	ConfigMapAPIVersion = "v1"

	// ConfigMapKind represents the kind of resource for the ConfigMap resource.
	ConfigMapKind = "ConfigMap"

	// DefaultProgramName is the default name for the KCL function program.
	DefaultProgramName = "kcl-function-run"

	// AnnotationAllowInSecureSource represents the annotation key for allowing insecure sources in KCLRun.
	AnnotationAllowInSecureSource = "krm.kcl.dev/allow-insecure-source"
)

// KCLRun is a custom resource to provider KPT `functionConfig`, KCL source and params.
type KCLRun struct {
	yaml.ResourceMeta `json:",inline" yaml:",inline"`
	// Spec is the KCLRun spec.
	Spec struct {
		// Source is a required field for providing a KCL script inline.
		Source string `json:"source" yaml:"source"`
		// Params are the parameters in key-value pairs format.
		Params map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
		// MatchConstraints defines the resource matching rules.
		MatchConstraints MatchConstraints `json:"matchConstraints,omitempty" yaml:"matchConstraints,omitempty"`
	} `json:"spec" yaml:"spec"`
}

// MatchConstraints defines the resource matching rules.
type MatchConstraints struct {
	ResourceRules []ResourceRule `json:"resourceRules,omitempty" yaml:"resourceRules,omitempty"`
}

// ResourceRule defines a rule for matching resources.
type ResourceRule struct {
	APIGroups   []string `json:"apiGroups,omitempty" yaml:"apiGroups,omitempty"`
	APIVersions []string `json:"apiVersions,omitempty" yaml:"apiVersions,omitempty"`
	Resources   []string `json:"resources,omitempty" yaml:"resources,omitempty"`
}

// Config is used to configure the KCLRun instance based on the given FunctionConfig.
// It converts ConfigMap to KCLRun or assigns values directly from KCLRun.
// If an error occurs during the configuration process, an error message will be returned.
func (r *KCLRun) Config(fnCfg *fn.KubeObject) error {
	fnCfgKind := fnCfg.GetKind()
	fnCfgAPIVersion := fnCfg.GetAPIVersion()
	switch {
	case fnCfg.IsEmpty():
		return fmt.Errorf("FunctionConfig is missing. Expect `ConfigMap` or `KCLRun`")
	case fnCfgAPIVersion == ConfigMapAPIVersion && fnCfgKind == ConfigMapKind:
		cm := &corev1.ConfigMap{}
		if err := fnCfg.As(cm); err != nil {
			return err
		}
		// Convert ConfigMap to KCLRun
		r.Name = cm.Name
		r.Namespace = cm.Namespace
		r.Spec.Params = map[string]interface{}{}
		for k, v := range cm.Data {
			if k == v1alpha1.SourceKey {
				r.Spec.Source = v
			}
			r.Spec.Params[k] = v
		}
	case fnCfgAPIVersion == v1alpha1.KCLRunAPIVersion && fnCfgKind == v1alpha1.KCLRunKind:
		if err := fnCfg.As(r); err != nil {
			return err
		}
	default:
		return fmt.Errorf("`functionConfig` must be either %v or %v, but we got: %v",
			schema.FromAPIVersionAndKind(ConfigMapAPIVersion, ConfigMapKind).String(),
			schema.FromAPIVersionAndKind(v1alpha1.KCLRunAPIVersion, v1alpha1.KCLRunKind).String(),
			schema.FromAPIVersionAndKind(fnCfg.GetAPIVersion(), fnCfg.GetKind()).String())
	}

	// Defaulting
	if r.Name == "" {
		r.Name = DefaultProgramName
	}
	// Validation
	if r.Spec.Source == "" {
		return fmt.Errorf("`source` must not be empty")
	}
	return nil
}

// Transform is used to transform the ResourceList with the KCLRun instance.
// It parses the FunctionConfig and each object in the ResourceList, transforms them according to the KCLRun configuration,
// and updates the ResourceList with the transformed objects.
// If an error occurs during the transformation process, an error message will be returned.
func (r *KCLRun) Transform(rl *fn.ResourceList) error {
	var transformedObjects []*fn.KubeObject
	var nodes []*yaml.RNode

	r.DealAnnotations()

	fcRN, err := yaml.Parse(rl.FunctionConfig.String())
	if err != nil {
		return err
	}
	for _, obj := range rl.Items {
		objRN, err := yaml.Parse(obj.String())
		if err != nil {
			return err
		}
		nodes = append(nodes, objRN)
	}

	st := &edit.SimpleTransformer{
		Name:           r.Name,
		Source:         r.Spec.Source,
		FunctionConfig: fcRN,
	}
	transformedNodes, err := st.Transform(nodes)
	if err != nil {
		return err
	}
	for _, n := range transformedNodes {
		obj, err := fn.ParseKubeObject([]byte(n.MustString()))
		if err != nil {
			return err
		}

		// Check if the transformed object matches the resource rules
		if r.MatchResourceRules(obj) {
			transformedObjects = append(transformedObjects, obj)
		}
	}
	rl.Items = transformedObjects
	return nil
}

// MatchResourceRules checks if the given Kubernetes object matches the resource rules specified in KCLRun.
func (r *KCLRun) MatchResourceRules(obj *fn.KubeObject) bool {
	// check if MatchConstraints is set
	if isEmptyStruct(r.Spec.MatchConstraints) {
		return true
	}
	// if MatchConstraints.ResourceRules is not set (nil or empty), return true by default
	if r.Spec.MatchConstraints.ResourceRules == nil || len(r.Spec.MatchConstraints.ResourceRules) == 0 {
		return true
	}
	// iterate through each resource rule
	for _, rule := range r.Spec.MatchConstraints.ResourceRules {
		if containsString(rule.APIGroups, obj.GroupKind().Group) &&
			containsString(rule.APIVersions, obj.GetAPIVersion()) &&
			containsString(rule.Resources, obj.GetKind()) {
			return true
		}
	}
	// if no match is found, return false
	return false
}

// DealAnnotations handles annotations, e.g., allow-insecure-source.
func (r *KCLRun) DealAnnotations() {
	// Deal the allow-insecure-source annotation
	if v, ok := r.ObjectMeta.Annotations[AnnotationAllowInSecureSource]; ok && isOk(v) {
		os.Setenv(settings.DEFAULT_OCI_PLAIN_HTTP_ENV, settings.ON)
	}
}

// isOk checks if a given string is in the list of "OK" values.
func isOk(value string) bool {
	okValues := []string{"ok", "yes", "true", "1", "on"}
	for _, v := range okValues {
		if strings.EqualFold(strings.ToLower(value), strings.ToLower(v)) {
			return true
		}
	}
	return false
}

// containsString checks if a slice contains a string.
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// isEmptyStruct checks if a struct is empty (i.e., all fields are zero or nil).
func isEmptyStruct(mc MatchConstraints) bool {
	v := reflect.ValueOf(mc)
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !reflect.DeepEqual(v.Field(i).Interface(), reflect.Zero(v.Field(i).Type()).Interface()) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
