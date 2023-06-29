resource = option("resource_list")
items = resource.items
params = resource.functionConfig.spec.params
# Use `k = v` to override existing annotations
annotations = {k = v for k, v in params.annotations}
[item | {
    metadata.annotations: annotations
} for item in items]
