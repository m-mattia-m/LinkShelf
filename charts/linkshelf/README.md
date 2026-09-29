# LinkShelf Helm chart

Deploys [LinkShelf](https://github.com/m-mattia-m/LinkShelf) with a Deployment, a Service and an optional Ingress.

```bash
helm repo add linkshelf https://m-mattia-m.github.io/LinkShelf
helm install linkshelf linkshelf/linkshelf -f values.yaml
```

Installation, secrets, database, ingress, custom domains and every other option are described in the
[Kubernetes docs](https://github.com/m-mattia-m/LinkShelf/blob/main/frontend/content/docs/self-hosting/kubernetes.md).
All values are documented in [values.yaml](values.yaml).
