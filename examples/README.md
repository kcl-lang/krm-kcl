# KRM KCL Examples

[![Go Report Card](https://goreportcard.com/badge/github.com/KusionStack/krm-kcl)](https://goreportcard.com/report/github.com/KusionStack/krm-kcl)
[![GoDoc](https://godoc.org/github.com/KusionStack/krm-kcl?status.svg)](https://godoc.org/github.com/KusionStack/krm-kcl)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/KusionStack/krm-kcl/blob/main/LICENSE)

The KCL programming language can be used to:

+ Add labels or annotations based on a condition.
+ Inject a sidecar container in all KRM resources that contain a PodTemplate.
+ Validate all KRM resources using KCL schema.
+ Use an abstract model to generate KRM resources.

The examples are divided into three categories:

+ **Abstraction**: Input KCL params and output KRM list.
+ **Mutation**: Input KCL params and KRM list and output KRM list.
+ **Validation**: Input KCL params and KRM list and output KRM list and the validation result.
