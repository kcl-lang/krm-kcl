package config

import (
	"testing"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/stretchr/testify/assert"
)

func TestKCLConfig(t *testing.T) {
	testcases := []struct {
		name         string
		config       string
		expectErrMsg string
	}{
		{
			name: "valid KCLRun",
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
  namespace: foo
source: |
  [item | {metadata.namespace = "baz"} for item in option("resource_list")]
`,
		},
		{
			name: "KCLRun missing Source",
			config: `apiVersion: fn.kpt.dev/v1alpha1
kind: KCLRun
metadata:
  name: my-kcl-fn
`,
			expectErrMsg: "`source` must not be empty",
		},
		{
			name: "valid ConfigMap",
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-kcl-fn
data:
  source: |
    # Set namespace to "baz"
    [item | {metadata.namespace = "baz"} for item in option("resource_list")]
`,
		},
		{
			name: "ConfigMap missing source",
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-kcl-fn
`,
			expectErrMsg: "`source` must not be empty",
		},
		{
			name: "ConfigMap with parameter but missing source",
			config: `apiVersion: v1
kind: ConfigMap
metadata:
  name: my-kcl-fn
data:
  param1: foo
`,
			expectErrMsg: "`source` must not be empty",
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := &KCLRun{}
			ko, err := fn.ParseKubeObject([]byte(tc.config))
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
