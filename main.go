package main

import (
	"fmt"
	"os"

	"github.com/KusionStack/krm-kcl/pkg/options"
)

func main() {
	// Input KRM resource list and KCL function config from os.Stdin
	//
	// Example
	//
	// ```yaml
	// apiVersion: config.kubernetes.io/v1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   spec:
	//     replicas: 2
	// - kind: Service
	// functionConfig:
	//   apiVersion: krm.kcl.dev/v1alpha1
	//   kind: KCLRun
	//   metadata:
	//   spec:
	//     source: |
	//       [resource | {if resource.kind == "Deployment": metadata.annotations: {"managed-by" = "krm-kcl"}} for resource in option("resource_list").items]
	// ```
	if err := options.NewRunOptions().Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
