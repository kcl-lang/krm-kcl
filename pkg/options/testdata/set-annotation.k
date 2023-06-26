[resource | {if resource.kind == "Deployment": metadata.annotations: {"managed-by" = "krm-kcl"}} for resource in option("resource_list").items]
