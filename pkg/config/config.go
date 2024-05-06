package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"kcl-lang.io/kpm/pkg/client"
	"kcl-lang.io/kpm/pkg/settings"
	"kcl-lang.io/krm-kcl/pkg/api/v1alpha1"
	"kcl-lang.io/krm-kcl/pkg/edit"
	src "kcl-lang.io/krm-kcl/pkg/source"
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
		// Credentials for remote locations
		Credentials CredSpec `json:"credentials" yaml:"credentials"`
		// Params are the parameters in key-value pairs format.
		Params map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
		// MatchConstraints defines the resource matching rules.
		MatchConstraints MatchConstraints `json:"matchConstraints,omitempty" yaml:"matchConstraints,omitempty"`
	} `json:"spec" yaml:"spec"`
}

// CredSpec defines authentication credentials for remote locations
type CredSpec struct {
	Url      string `json:"url" yaml:"url"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// MatchConstraints defines the resource matching rules.
type MatchConstraints struct {
	ResourceRules []ResourceRule `json:"resourceRules,omitempty" yaml:"resourceRules,omitempty"`
}

// ResourceRule defines a rule for matching resources.
type ResourceRule struct {
	APIGroups   []string `json:"apiGroups,omitempty" yaml:"apiGroups,omitempty"`
	APIVersions []string `json:"apiVersions,omitempty" yaml:"apiVersions,omitempty"`
	Kinds       []string `json:"kinds,omitempty" yaml:"kinds,omitempty"`
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

// TransformResourceList is used to transform the ResourceList with the KCLRun instance.
// It parses the FunctionConfig and each object in the ResourceList, transforms them according to the KCLRun configuration,
// and updates the ResourceList with the transformed objects.
// If an error occurs during the transformation process, an error message will be returned.
func (r *KCLRun) TransformResourceList(rl *fn.ResourceList) error {
	var transformedObjects []*fn.KubeObject
	var nodes []*yaml.RNode

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
	transformedNodes, err := r.Transform(nodes, fcRN)
	if err != nil {
		return err
	}
	for _, n := range transformedNodes {
		obj, err := fn.ParseKubeObject([]byte(n.MustString()))
		if err != nil {
			return err
		}
		transformedObjects = append(transformedObjects, obj)
	}
	rl.Items = transformedObjects
	return nil
}

// Transform is used to transform the input nodes with the KCLRun instance and function config.
func (c *KCLRun) Transform(in []*yaml.RNode, fnCfg *yaml.RNode) ([]*yaml.RNode, error) {
	var filterNodes []*yaml.RNode
	for _, n := range in {
		obj, err := fn.ParseKubeObject([]byte(n.MustString()))
		if err != nil {
			return nil, err
		}
		// Check if the transformed object matches the resource rules
		if MatchResourceRules(obj, &c.Spec.MatchConstraints) {
			filterNodes = append(filterNodes, n)
		}
	}
	c.DealAnnotations()

	// Authenticate with credentials to remote source
	if os.Getenv("KCL_SRC_URL") != "" {
		c.Spec.Credentials.Url = os.Getenv("KCL_SRC_URL")
	}
	if os.Getenv("KCL_SRC_USERNAME") != "" {
		c.Spec.Credentials.Username = os.Getenv("KCL_SRC_USERNAME")
	}
	if os.Getenv("KCL_SRC_PASSWORD") != "" {
		c.Spec.Credentials.Password = os.Getenv("KCL_SRC_PASSWORD")
	}
	if src.IsOCI(c.Spec.Source) && c.Spec.Credentials.Url != "" {
		cli, err := client.NewKpmClient()
		if err != nil {
			return nil, err
		}
		if err := cli.LoginOci(c.Spec.Credentials.Url, c.Spec.Credentials.Username, c.Spec.Credentials.Password); err != nil {
			return nil, err
		}
	}

	st := &edit.SimpleTransformer{
		Name:           DefaultProgramName,
		Source:         c.Spec.Source,
		FunctionConfig: fnCfg,
	}
	return st.Transform(filterNodes)
}

// MatchResourceRules checks if the given Kubernetes object matches the resource rules specified in KCLRun.
func MatchResourceRules(obj *fn.KubeObject, MatchConstraints *MatchConstraints) bool {
	// if MatchConstraints.ResourceRules is not set (nil or empty), return true by default
	if len(MatchConstraints.ResourceRules) == 0 {
		return true
	}
	// iterate through each resource rule
	for _, rule := range MatchConstraints.ResourceRules {
		if containsString(rule.APIGroups, obj.GroupKind().Group) &&
			containsString(rule.APIVersions, obj.GetAPIVersion()) &&
			containsString(rule.Kinds, obj.GetKind()) {
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

// containsString checks if a slice contains a string or "*"
func containsString(slice []string, str string) bool {
	if len(slice) == 0 {
		return true
	}
	for _, s := range slice {
		if s == "*" || s == str {
			return true
		}
	}
	return false
}
