import schemas as s

# Convert the `App` model into Kubernetes Deployment and Service Manifests
kubernetesRender = lambda a: s.App {
    # Construct the deployment manifest.
    deployment = {
        apiVersion = "apps/v1"
        kind = "Deployment"
        metadata.name = a.name
        metadata.labels = a.labels
        spec = {
            replicas = a.replicas
            selector.matchLabels = a.labels
            template.metadata.labels = a.labels
            template.spec.containers = [
                {
                    name = name
                    image = c.image
                    command = c.command
                    args = c.args
                    env = c.env
                    volumeMounts = c.volumes
                    resources: c.resources
                    ports = c.ports
                } for name, c in a.containers
            ]
        }
    }
    # Construct the service manifest.
    service = {
        apiVersion = "v1"
        kind = "Service"
        metadata.name = a.name
        metadata.labels = a.labels
        spec = {
            type = a.service?.$type
            selector = a.labels
            ports = a.service?.ports
        }
    }
    # Returns Kubernetes manifests
    [deployment, if a.service: service]
}

params: s.App = option("params")
items = kubernetesRender(params)
