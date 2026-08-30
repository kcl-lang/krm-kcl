package edit

import (
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

const kclRunFnCfg = `
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: set-annotation
spec:
  source: |
    a = 1
`

func TestWrapResourcesFiltersOutKCLRun(t *testing.T) {
	nodes := []*yaml.RNode{
		mustParse(t, "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx\n"),
		mustParse(t, kclRunFnCfg),
		mustParse(t, "apiVersion: v1\nkind: Service\nmetadata:\n  name: svc\n"),
	}
	out, err := WrapResources(nodes, nil)
	if err != nil {
		t.Fatalf("WrapResources() error = %v", err)
	}
	if got := out.GetApiVersion(); got != kio.ResourceListAPIVersion {
		t.Errorf("apiVersion = %q, want %q", got, kio.ResourceListAPIVersion)
	}
	if got := out.GetKind(); got != kio.ResourceListKind {
		t.Errorf("kind = %q, want %q", got, kio.ResourceListKind)
	}

	items, err := out.Pipe(yaml.Lookup("items"))
	if err != nil {
		t.Fatalf("lookup items: %v", err)
	}
	elems, err := items.Elements()
	if err != nil {
		t.Fatalf("items.Elements(): %v", err)
	}
	// The KCLRun resource is the function config, not an input resource, so it
	// must not be handed back to the KCL program as an item.
	if len(elems) != 2 {
		t.Fatalf("len(items) = %d, want 2 (KCLRun must be filtered out)", len(elems))
	}
	for _, e := range elems {
		if e.GetKind() == "KCLRun" {
			t.Error("KCLRun leaked into items")
		}
	}
}

func TestWrapResourcesSetsFunctionConfig(t *testing.T) {
	nodes := []*yaml.RNode{
		mustParse(t, "apiVersion: v1\nkind: Service\nmetadata:\n  name: svc\n"),
	}
	fc := mustParse(t, kclRunFnCfg)

	out, err := WrapResources(nodes, fc)
	if err != nil {
		t.Fatalf("WrapResources() error = %v", err)
	}
	got, err := out.Pipe(yaml.Lookup("functionConfig", "metadata", "name"))
	if err != nil {
		t.Fatalf("lookup functionConfig: %v", err)
	}
	if got.IsNil() {
		t.Fatal("functionConfig was not set on the resource list")
	}
	if name := yaml.GetValue(got); name != "set-annotation" {
		t.Errorf("functionConfig.metadata.name = %q, want %q", name, "set-annotation")
	}
}

func TestWrapResourcesEmptyInput(t *testing.T) {
	out, err := WrapResources(nil, nil)
	if err != nil {
		t.Fatalf("WrapResources() error = %v", err)
	}
	items, err := out.Pipe(yaml.Lookup("items"))
	if err != nil {
		t.Fatalf("lookup items: %v", err)
	}
	elems, err := items.Elements()
	if err != nil {
		t.Fatalf("items.Elements(): %v", err)
	}
	if len(elems) != 0 {
		t.Errorf("len(items) = %d, want 0", len(elems))
	}
}

func TestUnwrapResourcesEmpty(t *testing.T) {
	outs, fc, err := UnwrapResources(nil)
	if err != nil {
		t.Fatalf("UnwrapResources() error = %v", err)
	}
	if len(outs) != 0 {
		t.Errorf("len(outs) = %d, want 0", len(outs))
	}
	if fc != nil {
		t.Errorf("functionConfig = %v, want nil", fc)
	}
}

func TestUnwrapResourcesBareResource(t *testing.T) {
	// A single node that is not a ResourceList has no `items`, so it is
	// returned as-is rather than being treated as an empty list.
	in := []*yaml.RNode{
		mustParse(t, "apiVersion: v1\nkind: Service\nmetadata:\n  name: svc\n"),
	}
	outs, _, err := UnwrapResources(in)
	if err != nil {
		t.Fatalf("UnwrapResources() error = %v", err)
	}
	if len(outs) != 1 {
		t.Fatalf("len(outs) = %d, want 1", len(outs))
	}
	if got := outs[0].GetKind(); got != "Service" {
		t.Errorf("kind = %q, want Service", got)
	}
}

func TestUnwrapResourcesResourceList(t *testing.T) {
	in := []*yaml.RNode{
		mustParse(t, `
apiVersion: config.kubernetes.io/v1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: nginx
- apiVersion: v1
  kind: Service
  metadata:
    name: svc
functionConfig:
  apiVersion: krm.kcl.dev/v1alpha1
  kind: KCLRun
  metadata:
    name: set-annotation
`),
	}
	outs, fc, err := UnwrapResources(in)
	if err != nil {
		t.Fatalf("UnwrapResources() error = %v", err)
	}
	if len(outs) != 2 {
		t.Fatalf("len(outs) = %d, want 2", len(outs))
	}
	if fc == nil || fc.GetKind() != "KCLRun" {
		t.Errorf("functionConfig = %v, want a KCLRun node", fc)
	}
}

func TestWrapUnwrapRoundTrip(t *testing.T) {
	nodes := []*yaml.RNode{
		mustParse(t, "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx\n"),
		mustParse(t, "apiVersion: v1\nkind: Service\nmetadata:\n  name: svc\n"),
	}
	wrapped, err := WrapResources(nodes, nil)
	if err != nil {
		t.Fatalf("WrapResources() error = %v", err)
	}
	outs, _, err := UnwrapResources([]*yaml.RNode{wrapped})
	if err != nil {
		t.Fatalf("UnwrapResources() error = %v", err)
	}
	if len(outs) != len(nodes) {
		t.Fatalf("round trip changed resource count: got %d, want %d", len(outs), len(nodes))
	}
	for i := range nodes {
		want, err := nodes[i].String()
		if err != nil {
			t.Fatalf("marshalling input %d: %v", i, err)
		}
		got, err := outs[i].String()
		if err != nil {
			t.Fatalf("marshalling output %d: %v", i, err)
		}
		if got != want {
			t.Errorf("round trip changed resource %d:\ngot:\n%s\nwant:\n%s", i, got, want)
		}
	}
}