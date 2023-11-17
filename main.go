package main

import (
	"fmt"
	"os"

	"kcl-lang.io/krm-kcl/pkg/options"
)

func main() {
	// Input KRM resource list and KCL function config from os.Stdin
	// YAML Stream Example
	//
	// ```yaml
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata:
	//   name: deployment
	// spec:
	//   replicas: 2
	// ---
	// apiVersion: krm.kcl.dev/v1alpha1
	// kind: KCLRun
	// metadata:
	//   name: set-annotation
	// spec:
	//   source: |
	//     [resource | {if resource.kind == "Deployment": metadata.annotations: {"managed-by" = "krm-kcl"}} for resource in option("resource_list").items]
	// ```
	//
	// Resource List Example
	//
	// ```yaml
	// apiVersion: config.kubernetes.io/v1
	// kind: ResourceList
	// items:
	// - apiVersion: apps/v1
	//   kind: Deployment
	//   metadata:
	//     name: deployment
	//   spec:
	//     replicas: 2
	// functionConfig:
	//   apiVersion: krm.kcl.dev/v1alpha1
	//   kind: KCLRun
	//   metadata:
	//     name: set-annotation
	//   spec:
	//     source: |
	//       [resource | {if resource.kind == "Deployment": metadata.annotations: {"managed-by" = "krm-kcl"}} for resource in option("resource_list").items]
	// ```
	if err := options.NewRunOptions().Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
