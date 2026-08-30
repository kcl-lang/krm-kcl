package kio

import (
	"bytes"
	"strings"
	"testing"
)

// TestFilter_MultipleKCLRun_Regression reproduces the exact shape from
// kcl-lang/kubectl-kcl#9: a YAML stream that interleaves a Service
// with two KCLRun objects. Older versions of the kio.Filter panicked
// with `runtime error: index out of range [2] with length 1` at
// `pkg/kio/filter.go:31` because the per-KCLRun loop looked up the
// function config by *index* into the input slice (`in[idxs[idx]]`)
// and the SimpleTransformer was mutating `in` in-place, dropping
// processed KCLRun objects so the next iteration's index became
// stale.
//
// The fix rewrote the per-KCLRun loop so it never indexes `in` after
// invoking a transformer on it. This test pins the new behavior by
// streaming the exact YAML from #9 through `kio.NewPipeline` (the
// production code path) and asserting that the call completes without
// panicking.
func TestFilter_MultipleKCLRun_Regression(t *testing.T) {
	const stream = `apiVersion: v1
kind: Service
metadata:
  name: default-domain-service
  namespace: knative-serving
spec:
  clusterIP: None
  selector:
    app: default-domain
  ports:
    - name: http
      port: 80
      targetPort: 8080
  type: ClusterIP
---
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: change-service-type
spec:
  source: |
    [item | {} for item in option("items")]
---
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: add-labels
spec:
  source: |
    [item | {} for item in option("items")]
`

	var out bytes.Buffer
	p := NewPipeline(strings.NewReader(stream), &out, false)
	// The old code panicked here for #9's input. A non-nil error is
	// fine — the bug was the *panic*, and a graceful error is much
	// better UX. We just want to ensure control returns.
	if err := p.Execute(); err != nil {
		t.Logf("pipeline.Execute() returned error (acceptable): %v", err)
	}
}
