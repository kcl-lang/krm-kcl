items = lambda {
    resource = option("resource_list") or {}
    items = resource.items or []
    params = option("params") or {}
    # Use `k = v` to override existing annotations
    annotations = {k = v for k, v in params.annotations or {}}
    [item | {
        metadata.annotations: annotations
    } for item in items]
}()
