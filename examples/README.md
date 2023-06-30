# KRM KCL Examples

[![Go Report Card](https://goreportcard.com/badge/kcl-lang.io/krm-kcl)](https://goreportcard.com/report/kcl-lang.io/krm-kcl)
[![GoDoc](https://godoc.org/kcl-lang.io/krm-kcl?status.svg)](https://godoc.org/kcl-lang.io/krm-kcl)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://kcl-lang.io/krm-kcl/blob/main/LICENSE)

The KCL programming language can be used to:

+ Add labels or annotations based on a condition.
+ Inject a sidecar container in all KRM resources that contain a PodTemplate.
+ Validate all KRM resources using KCL schema.
+ Use an abstract model to generate KRM resources.

The examples are divided into three categories:

+ **Abstraction**: Input KCL params and output KRM list.
+ **Mutation**: Input KCL params and KRM list and output KRM list.
+ **Validation**: Input KCL params and KRM list and output KRM list and the validation result.
