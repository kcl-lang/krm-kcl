schema ResourceList:
    apiVersion: "config.kubernetes.io/v1" = "config.kubernetes.io/v1"
    kind: "ResourceList" = "ResourceList"
    items?: []
    functionConfig: KCLRun

schema KCLRun:
    spec: KCLRunSpec

schema KCLRunSpec:
    params?: {str:}
    source: str
