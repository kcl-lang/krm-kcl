package config

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"kcl-lang.io/kpm/pkg/client"
	"kcl-lang.io/kpm/pkg/settings"
	"kcl-lang.io/krm-kcl/pkg/api"
	"kcl-lang.io/krm-kcl/pkg/api/v1alpha1"
	"kcl-lang.io/krm-kcl/pkg/edit"
	"kcl-lang.io/krm-kcl/pkg/kube"
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
		// Config is the compile config.
		Config api.ConfigSpec `json:"config" yaml:"config"`
		// Credentials for remote locations
		Credentials api.CredSpec `json:"credentials" yaml:"credentials"`
		// Params are the parameters in key-value pairs format.
		Params map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
		// MatchConstraints defines the resource matching rules.
		MatchConstraints api.MatchConstraintsSpec `json:"matchConstraints,omitempty" yaml:"matchConstraints,omitempty"`
	} `json:"spec" yaml:"spec"`
}

// Config is used to configure the KCLRun instance based on the given FunctionConfig.
// It converts ConfigMap to KCLRun or assigns values directly from KCLRun.
// If an error occurs during the configuration process, an error message will be returned.
func (r *KCLRun) Config(o *kube.KubeObject) error {
	if o == nil {
		return fmt.Errorf("FunctionConfig is missing. Expect `ConfigMap` or `KCLRun`")
	}
	kind := o.GetKind()
	apiVersion := o.GetAPIVersion()
	switch {
	case o.IsNilOrEmpty():
		return fmt.Errorf("FunctionConfig is missing. Expect `ConfigMap` or `KCLRun`")
	case apiVersion == v1alpha1.KCLRunAPIVersion && kind == api.KCLRunKind:
		if err := o.As(r); err != nil {
			return err
		}
	default:
		return fmt.Errorf("`functionConfig` must be either %v, but we got: %v",
			schema.FromAPIVersionAndKind(v1alpha1.KCLRunAPIVersion, api.KCLRunKind).String(),
			schema.FromAPIVersionAndKind(apiVersion, kind).String())
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

// Run is used to output the YAML list with the KCLRun instance.
func (r *KCLRun) Run() ([]*yaml.RNode, error) {
	c, err := yaml.Marshal(r)
	if err != nil {
		return nil, err
	}
	fnCfg, err := yaml.Parse(string(c))
	if err != nil {
		return nil, err
	}
	return r.Transform(nil, fnCfg)
}

// TransformResourceList is used to transform the ResourceList with the KCLRun instance.
// It parses the FunctionConfig and each object in the ResourceList, transforms them according to the KCLRun configuration,
// and updates the ResourceList with the transformed objects.
// If an error occurs during the transformation process, an error message will be returned.
func (r *KCLRun) TransformResourceList(rl *kube.ResourceList) error {
	var transformedObjects []*kube.KubeObject
	var nodes []*yaml.RNode

	fcRN, err := yaml.Parse(rl.FunctionConfig.MustString())
	if err != nil {
		return err
	}
	for _, obj := range rl.Items {
		objRN, err := yaml.Parse(obj.MustString())
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
		obj, err := kube.ParseKubeObject([]byte(n.MustString()))
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
		obj, err := kube.ParseKubeObject([]byte(n.MustString()))
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
	if os.Getenv(SrcUrlEnvVar) != "" {
		c.Spec.Credentials.Url = os.Getenv(SrcUrlEnvVar)
	}
	if os.Getenv(SrcUrlUsernameEnvVar) != "" {
		c.Spec.Credentials.Username = os.Getenv(SrcUrlUsernameEnvVar)
	}
	if os.Getenv(SrcUrlPasswordEnvVar) != "" {
		c.Spec.Credentials.Password = os.Getenv(SrcUrlPasswordEnvVar)
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
		Config:         &c.Spec.Config,
	}
	return st.Transform(filterNodes)
}

// DealAnnotations handles annotations, e.g., allow-insecure-source.
func (r *KCLRun) DealAnnotations() {
	// Deal the allow-insecure-source annotation
	if v, ok := r.ObjectMeta.Annotations[AnnotationAllowInSecureSource]; ok && isOk(v) {
		os.Setenv(settings.DEFAULT_OCI_PLAIN_HTTP_ENV, settings.ON)
	}
}
