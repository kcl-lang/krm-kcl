package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kcl-lang.io/krm-kcl/pkg/kube"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestKCLConfig(t *testing.T) {
	testcases := []struct {
		name         string
		config       string
		expectErrMsg string
	}{
		{
			name: "valid KCLRun",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
spec:
  source: |
    [item | {metadata.namespace = "baz"} for item in option("resource_list")]
  matchConstraints:
    resourceRules:
`,
		},
		{
			name: "KCLRun missing Source",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
`,
			expectErrMsg: "`source` must not be empty",
		},
		{
			name: "KCLRun missing matchConstraints",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
spec:
  source: |
    [item | {metadata.namespace = "baz"} for item in option("resource_list")]
`,
			expectErrMsg: "",
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := &KCLRun{}
			ko, err := kube.ParseKubeObject([]byte(tc.config))
			assert.NoError(t, err)
			err = r.Config(ko)
			if tc.expectErrMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			}
		})
	}
}

func TestKCLRun(t *testing.T) {
	testcases := []struct {
		name         string
		config       string
		expectResult string
		expectErrMsg string
	}{
		{
			name: "KCLRunInlineSource",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
spec:
  source: |
    {
        apiVersion = "v1"
    }
`,
			expectResult: `apiVersion: v1`,
		},
		{
			name: "KCLRunWithParams",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
spec:
  params:
    version: v1
  source: |
    {
        apiVersion = option("params")?.version
    }
`,
			expectResult: `apiVersion: v1`,
		},
		{
			name: "KCLRunWithArgumentsConfig",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
spec:
  config:
    arguments:
    - version=v1
  source: |
    {
        apiVersion = option("version")
    }
`,
			expectResult: `apiVersion: v1`,
		},
		{
			name: "KCLRunWithDisableNoneConfig",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
spec:
  config:
    disableNone: true
  source: |
    {
        a = None
        b = 1
    }
`,
			expectResult: `b: 1`,
		},
		{
			name: "KCLRunWithDependencies",
			config: `apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
spec:
  dependencies:
    helloworld = {oci = "oci://ghcr.io/kcl-lang/helloworld", "tag" = "0.1.0"}
  source: |
    import helloworld

    {
        a = helloworld.The_first_kcl_program
    }
`,
			expectResult: `a: Hello World!`,
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := &KCLRun{}
			ko, err := kube.ParseKubeObject([]byte(tc.config))
			assert.NoError(t, err)
			err = r.Config(ko)
			assert.NoError(t, err)
			result, err := r.Run()
			if tc.expectErrMsg == "" {
				assert.NoError(t, err)
				resultYaml, err := yaml.Parse(tc.expectResult)
				assert.NoError(t, err)
				assert.Equal(t, result[0], resultYaml)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErrMsg)
			}
		})
	}
}
