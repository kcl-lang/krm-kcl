# KRM KCL

[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/kpt-kcl-sdk)](https://goreportcard.com/report/github.com/KusionStack/kpt-kcl-sdk)
[![GoDoc](https://godoc.org/github.com/KusionStack/kpt-kcl-sdk?status.svg)](https://godoc.org/github.com/KusionStack/kpt-kcl-sdk)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/KusionStack/kpt-kcl-sdk/blob/main/LICENSE)

The KRM KCL function SDK contains a KRM KCL spec and a interpreter to run KCL codes to generate, mutate or validate Kubernetes resources.

The KCL programming language can be used to:

+ Add annotations based on a condition.
+ Inject a sidecar container in all KRM resources that contain a PodTemplate.
+ Validate all KRM resources using KCL schema.

## Test the KRM function

You need to put your KCL script source in the functionConfig of kind KCLRun and then the function will run the KCL script that you provide.

This function can be used both declaratively and imperatively.

```bash
make test
```

## FunctionConfig

There are 2 kinds of `functionConfig` supported by this function:

+ ConfigMap
+ A custom resource of kind `KCLRun`

To use a ConfigMap as the functionConfig, the KCL script source must be specified in the data.source field. Additional parameters can be specified in the data field.

Here's an example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: set-replicas
data:
  replicas: "5"
  source: |
    resources = option("resource_list")
    setReplicas = lambda items, replicas {
        [item | {
            if item.kind == "Deployment": 
                spec.replicas = int(replicas)
        } for item in items]
    }
    setReplicas(resources.items or [], resources.functionConfig.data.replicas)
```

In the example above, the script accesses the replicas parameters using `option("resource_list").functionConfig.data.replicas`.

To use a KCLRun as the functionConfig, the KCL source must be specified in the source field. Additional parameters can be specified in the params field. The params field supports any complex data structure as long as it can be represented in YAML.

```yaml
apiVersion: krm.kcl.dev/v1alpha1
kind: KCLRun
metadata:
  name: conditionally-add-annotations
spec:
  params:
    toMatch:
      config.kubernetes.io/local-config: "true"
    toAdd:
      configmanagement.gke.io/managed: disabled
  source: |
    resource = option("resource_list")
    items = resource.items
    params = resource.functionConfig.spec.params
    toMatch = params.toMatch
    toAdd = params.toAdd
    [item | {
       # If all annotations are matched, patch more annotations
       if all key, value in toMatch {
          item.metadata.annotations[key] == value
       }:
           metadata.annotations: toAdd
    } for item in items]
```

In the example above, the script accesses the `toMatch` parameters using `option("resource_list").functionConfig.spec.params.toMatch`.

## Guides for Developing KCL

Here's what you can do in the KCL script:

+ Read resources from `option("resource_list")`. The `option("resource_list")` complies with the [KRM Functions Specification](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md#krm-functions-specification). You can read the input resources from `option("resource_list")["items"]` and the `functionConfig` from `option("resource_list")["functionConfig"]`.
+ Return a KRM list for output resources.
+ Return an error using `assert {condition}, {error_message}`.
+ Read the environment variables. e.g. `option("PATH")` (Not yet implemented).
+ Read the OpenAPI schema. e.g. `option("open_api")["definitions"]["io.k8s.api.apps.v1.Deployment"]` (Not yet implemented).

## Library

You can directly use [KCL standard libraries](https://kcl-lang.io/docs/reference/model/overview) without importing them, such as `regex.match`, `math.log`.
