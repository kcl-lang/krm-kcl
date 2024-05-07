schema ResourceList:
    apiVersion: "config.kubernetes.io/v1" = "config.kubernetes.io/v1"
    kind: "ResourceList" = "ResourceList"
    metadata?: {str:}
    items?: []
    functionConfig: KCLRun

schema KCLRun:
    apiVersion: "krm.kcl.dev/v1alpha1" = "krm.kcl.dev/v1alpha1"
    kind: "KCLRun" = "KCLRun"
    metadata?: {str:}
    spec: KCLRunSpec

schema KCLRunSpec:
    params?: {str:}
    source: str

resource: ResourceList = option("resource_list")
# Use `k = v` to override existing annotations
annotations = {"${k}" = v for k, v in option("params")?.annotations or {}}
items = [item | {
    metadata.annotations: annotations
} for item in option("items") or []]
