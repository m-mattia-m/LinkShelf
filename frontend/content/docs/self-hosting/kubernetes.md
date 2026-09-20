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
