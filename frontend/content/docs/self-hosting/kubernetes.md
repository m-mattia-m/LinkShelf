---
title: Kubernetes
order: 5
---

LinkShelf comes with a Helm chart. It needs a Postgres or MySQL database and a secret for the login tokens.

```bash
helm repo add linkshelf https://m-mattia-m.github.io/LinkShelf

kubectl create secret generic linkshelf-secrets \
  --from-literal=jwt-secret="$(openssl rand -hex 32)"

kubectl create secret generic linkshelf-db \
  --from-literal=host=postgres.example.svc \
  --from-literal=port=5432 \
  --from-literal=username=linkshelf \
  --from-literal=password=change-me \
  --from-literal=database=linkshelf

helm install linkshelf linkshelf/linkshelf \
  --set secrets.existingSecret.name=linkshelf-secrets \
  --set database.existingSecret.name=linkshelf-db
```

Then open it with `kubectl port-forward svc/linkshelf 3000:3000 8085:8085`, or enable the ingress. The frontend and the
backend need their own host, because the browser talks to the backend directly.

Email verification is on by default. Configure SMTP or turn it off with
`--set env.APP_AUTHENTICATION_EMAILVERIFICATION_ENABLED=false`, otherwise new accounts can't log in.

The chart can also deploy a Postgres for testing, and it works on OpenShift with your own database. Ingress, SMTP,
secrets and everything else are described in the
[chart README](https://github.com/m-mattia-m/LinkShelf/tree/main/charts/linkshelf).

## Custom domains

Shelves can be served on a domain of their own, see [Custom domains](/docs/self-hosting/custom-domains). The ingress has
to send those domains to the frontend, so list them in `ingress.extraHosts`, or use a wildcard for all of them, and add
a certificate for them to `ingress.tls`:

```yaml
ingress:
  enabled: true
  frontend:
    host: links.example.com
  backend:
    host: api.example.com
  extraHosts:
    - profile.example.com
    - "*.shelves.example.com"
  tls:
    - secretName: linkshelf-tls
      hosts:
        - links.example.com
        - api.example.com
        - profile.example.com
```

The ingress controller has to **pass the original `Host` header (or `X-Forwarded-Host`) on, and no annotation or
load balancer in front of it may overwrite it**. Otherwise LinkShelf can't tell which domain was asked for, and shelves
only work through their path.

With the ingress enabled the chart also turns on `app.strictOrigins`, so the API only answers on the `apiUrl` host and only
to browsers from the frontend or from a shelf's domain, see [Strict origins](/docs/self-hosting/configuration#strict-origins).
Set `strictOrigins: false` in the values to keep the API open.

