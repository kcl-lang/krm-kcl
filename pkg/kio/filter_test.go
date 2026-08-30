package kio

import (
	"bytes"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func mustParse(t *testing.T, s string) *yaml.RNode {
	t.Helper()
	node, err := yaml.Parse(s)
	if err != nil {
		t.Fatalf("parsing yaml: %v", err)
	}
	return node
}

const deployment = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 2
`

func kclRun(t *testing.T, name, source string) *yaml.RNode {
	t.Helper()
	return mustParse(t, `
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: `+name+`
spec:
  source: |
    `+source+`
`)
}

// When the function config is supplied out of band (the `functionConfig` field
// of the ResourceList), it is parsed directly and no index is reported.
func TestParseConfigsFromFunctionConfig(t *testing.T) {
	f := &Filter{rw: &kio.ByteReadWriter{FunctionConfig: kclRun(t, "from-fn-cfg", "a = 1")}}

	configs, idxs, err := f.parseConfigs([]*yaml.RNode{mustParse(t, deployment)})
	if err != nil {
		t.Fatalf("parseConfigs() error = %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("len(configs) = %d, want 1", len(configs))
	}
	if configs[0].Name != "from-fn-cfg" {
		t.Errorf("config name = %q, want %q", configs[0].Name, "from-fn-cfg")
	}
	if strings.TrimSpace(configs[0].Spec.Source) != "a = 1" {
		t.Errorf("config source = %q, want %q", configs[0].Spec.Source, "a = 1")
	}
	// Indices only make sense for configs discovered inside the input stream.
	if len(idxs) != 0 {
		t.Errorf("len(idxs) = %d, want 0 when the config comes from functionConfig", len(idxs))
	}
}

// Without a function config, KCLRun resources are discovered inside the input
// stream and their positions are reported so the caller can pair them up.
func TestParseConfigsFromInputStream(t *testing.T) {
	in := []*yaml.RNode{
		mustParse(t, deployment),
		kclRun(t, "first", "a = 1"),
		mustParse(t, "apiVersion: v1\nkind: Service\nmetadata:\n  name: svc\n"),
		kclRun(t, "second", "b = 2"),
	}
	f := &Filter{rw: &kio.ByteReadWriter{}}

	configs, idxs, err := f.parseConfigs(in)
	if err != nil {
		t.Fatalf("parseConfigs() error = %v", err)
	}
	if len(configs) != 2 {
		t.Fatalf("len(configs) = %d, want 2", len(configs))
	}
	if configs[0].Name != "first" || configs[1].Name != "second" {
		t.Errorf("config names = [%q %q], want [first second]", configs[0].Name, configs[1].Name)
	}
	if len(idxs) != 2 || idxs[0] != 1 || idxs[1] != 3 {
		t.Errorf("idxs = %v, want [1 3]", idxs)
	}
}

func TestParseConfigsNoKCLRun(t *testing.T) {
	f := &Filter{rw: &kio.ByteReadWriter{}}

	configs, idxs, err := f.parseConfigs([]*yaml.RNode{mustParse(t, deployment)})
	if err != nil {
		t.Fatalf("parseConfigs() error = %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("len(configs) = %d, want 0", len(configs))
	}
	if len(idxs) != 0 {
		t.Errorf("len(idxs) = %d, want 0", len(idxs))
	}
}

// A resource that merely looks like a KCLRun (right kind, wrong group) must be
// left alone and treated as a plain input resource.
func TestParseConfigsIgnoresForeignAPIVersion(t *testing.T) {
	in := []*yaml.RNode{
		mustParse(t, "apiVersion: example.com/v1\nkind: KCLRun\nmetadata:\n  name: impostor\n"),
	}
	f := &Filter{rw: &kio.ByteReadWriter{}}

	configs, _, err := f.parseConfigs(in)
	if err != nil {
		t.Fatalf("parseConfigs() error = %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("len(configs) = %d, want 0 for a non krm.kcl.dev KCLRun", len(configs))
	}
}

func TestParseConfigSpecFields(t *testing.T) {
	node := mustParse(t, `
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: full
spec:
  source: |
    a = 1
  params:
    key: value
  config:
    vendor: true
    sortKeys: true
    arguments:
    - env=prod
`)
	f := &Filter{rw: &kio.ByteReadWriter{}}

	cfg, err := f.parseConfig(node)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}
	if !cfg.Spec.Config.Vendor {
		t.Error("spec.config.vendor = false, want true")
	}
	if !cfg.Spec.Config.SortKeys {
		t.Error("spec.config.sortKeys = false, want true")
	}
	if len(cfg.Spec.Config.Arguments) != 1 || cfg.Spec.Config.Arguments[0] != "env=prod" {
		t.Errorf("spec.config.arguments = %v, want [env=prod]", cfg.Spec.Config.Arguments)
	}
	if cfg.Spec.Params["key"] != "value" {
		t.Errorf("spec.params[key] = %v, want value", cfg.Spec.Params["key"])
	}
}

// Filter is a no-op when the input carries no KCLRun at all: resources must
// pass through untouched instead of being dropped.
func TestFilterWithoutConfigIsPassthrough(t *testing.T) {
	in := []*yaml.RNode{
		mustParse(t, deployment),
		mustParse(t, "apiVersion: v1\nkind: Service\nmetadata:\n  name: svc\n"),
	}
	f := Filter{rw: &kio.ByteReadWriter{}}

	out, err := f.Filter(in)
	if err != nil {
		t.Fatalf("Filter() error = %v", err)
	}
	if len(out) != len(in) {
		t.Fatalf("len(out) = %d, want %d", len(out), len(in))
	}
	for i := range in {
		want, err := in[i].String()
		if err != nil {
			t.Fatalf("marshalling input %d: %v", i, err)
		}
		got, err := out[i].String()
		if err != nil {
			t.Fatalf("marshalling output %d: %v", i, err)
		}
		if got != want {
			t.Errorf("resource %d was modified:\ngot:\n%s\nwant:\n%s", i, got, want)
		}
	}
}

func TestFilterAppliesKCLFromFunctionConfig(t *testing.T) {
	fnCfg := mustParse(t, `
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: set-annotation
spec:
  source: |
    [resource | {if resource.kind == "Deployment": metadata.annotations: {"managed-by" = "krm-kcl"}} for resource in option("resource_list").items]
`)
	f := Filter{rw: &kio.ByteReadWriter{FunctionConfig: fnCfg}}

	out, err := f.Filter([]*yaml.RNode{mustParse(t, deployment)})
	if err != nil {
		t.Fatalf("Filter() error = %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}
	ann := out[0].GetAnnotations()
	if ann["managed-by"] != "krm-kcl" {
		t.Errorf("annotations = %v, want managed-by=krm-kcl", ann)
	}
}

func TestFilterPropagatesKCLCompileError(t *testing.T) {
	fnCfg := mustParse(t, `
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: broken
spec:
  source: |
    this is not valid kcl @@@
`)
	f := Filter{rw: &kio.ByteReadWriter{FunctionConfig: fnCfg}}

	if _, err := f.Filter([]*yaml.RNode{mustParse(t, deployment)}); err == nil {
		t.Fatal("Filter() error = nil, want a compile error to be surfaced")
	}
}

func TestNewPipelineEndToEnd(t *testing.T) {
	input := `apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
  spec:
    replicas: 2
functionConfig:
  apiVersion: krm.kcl.dev/v1alpha1
  kind: KCLRun
  metadata:
    name: set-annotation
  spec:
    source: |
      [resource | {if resource.kind == "Deployment": metadata.annotations: {"managed-by" = "krm-kcl"}} for resource in option("resource_list").items]
`
	var out bytes.Buffer
	pipeline := NewPipeline(strings.NewReader(input), &out, false)
	if err := pipeline.Execute(); err != nil {
		t.Fatalf("pipeline.Execute() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "managed-by: krm-kcl") {
		t.Errorf("pipeline output missing the injected annotation:\n%s", got)
	}
	if !strings.Contains(got, "name: nginx") {
		t.Errorf("pipeline output lost the original resource:\n%s", got)
	}
}
