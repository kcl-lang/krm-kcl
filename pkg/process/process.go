package process

import (
	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"kcl-lang.io/krm-kcl/pkg/config"
)

// Process is a function that takes a pointer to a ResourceList and processes
// it using the KCL function. It returns a boolean indicating whether the
// processing was successful, and an error (if any).
func Process(resourceList *fn.ResourceList) (bool, error) {
	err := func() error {
		r := &config.KCLRun{}
		if err := r.Config(resourceList.FunctionConfig); err != nil {
			return err
		}
		return r.TransformResourceList(resourceList)
	}()
	if err != nil {
		resourceList.Results = []*fn.Result{
			{
				Message:  err.Error(),
				Severity: fn.Error,
			},
		}
		return false, nil
	}
	return true, nil
}
