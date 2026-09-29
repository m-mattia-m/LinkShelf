---
title: Kubernetes
order: 5
---

LinkShelf comes with a Helm chart that deploys a Deployment, a Service and an optional Ingress. It needs a Postgres or
MySQL database, either your own or the [bundled Postgres](#postgres-for-evaluation).

The chart version follows the app version: chart `1.4.0` deploys the image of release `1.4.0`, unless you set
`image.tag`. Every option is documented in
[values.yaml](https://github.com/m-mattia-m/LinkShelf/blob/main/charts/linkshelf/values.yaml), and
`values.schema.json` rejects wrong values on install.

## Install

The chart never creates a Secret. Create them yourself, e.g. with `kubectl`, External Secrets or Sealed Secrets:

```bash
helm repo add linkshelf https://m-mattia-m.github.io/LinkShelf
helm repo update

kubectl create secret generic linkshelf-secrets \
  --from-literal=jwt-secret="$(openssl rand -hex 32)"

kubectl create secret generic linkshelf-db \
  --from-literal=host=postgres.example.svc \
  --from-literal=port=5432 \
  --from-literal=username=linkshelf \
  --from-literal=password=change-me \
  --from-literal=database=linkshelf
```

A secret is an `env` entry with `valueFrom` instead of `value`, plain Kubernetes syntax:

```yaml
# values.yaml
env:
  - name: AUTHENTICATION_JWTSECRET
    valueFrom:
      secretKeyRef:
        name: linkshelf-secrets
        key: jwt-secret
  - name: DATABASE_HOST
    valueFrom:
      secretKeyRef:
        name: linkshelf-db
        key: host
  - name: DATABASE_PORT
    valueFrom:
      secretKeyRef:
        name: linkshelf-db
        key: port
  - name: DATABASE_USERNAME
    valueFrom:
      secretKeyRef:
        name: linkshelf-db
        key: username
  - name: DATABASE_PASSWORD
    valueFrom:
      secretKeyRef:
        name: linkshelf-db
        key: password
  - name: DATABASE_NAME
    valueFrom:
      secretKeyRef:
        name: linkshelf-db
        key: database
ingress:
  enabled: true
  className: nginx
  frontend:
    host: links.example.com
  backend:
    host: api.example.com
  tls:
    - secretName: linkshelf-tls
      hosts:
        - links.example.com
        - api.example.com
```

```bash
helm install linkshelf linkshelf/linkshelf -f values.yaml
```

Without an ingress, open it with `kubectl port-forward svc/linkshelf 3000:3000 8085:8085`.

## Settings

Every key of the backend config can be set as an environment variable in `env`, e.g. `smtp.host` becomes `SMTP_HOST`.
See [Configuration](/docs/self-hosting/configuration) for the keys. Email verification is on by default, so set up SMTP
or turn it off, otherwise new accounts can't log in:

```yaml
env:
  - name: SMTP_HOST
    value: smtp.example.com
  - name: SMTP_FROM
    value: no-reply@example.com
  # - name: AUTHENTICATION_EMAILVERIFICATION_ENABLED
  #   value: "false"
```

## Secrets

`AUTHENTICATION_JWTSECRET` is required. The bootstrap admin from the image is switched off; to get an admin, add
`AUTHENTICATION_BOOTSTRAPADMIN_EMAIL` and `AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD` to `env`. Its username is `admin`,
change it with `AUTHENTICATION_BOOTSTRAPADMIN_USERNAME`. That account needs no email verification, so it works before
SMTP is set up. Add `SMTP_PASSWORD` and `AUTHENTICATION_OIDC_CLIENTSECRET` the same way if you use them.

## Database

Set `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USERNAME`, `DATABASE_PASSWORD` and `DATABASE_NAME` in `env` (each can
come from a different Secret). The non-secret ones can go in [`configFile`](#config-file) instead. The chart doesn't
check them; the backend fails at startup if the database is unreachable. Set `database.engine` to `MYSQL` for MySQL and
`database.params` for extra connection parameters such as `sslmode=require`.

## Ingress

The browser talks to the backend directly, so the frontend and the backend need their own host. The chart derives
`frontendUrl`, `apiUrl` and `oidcRedirectUrl` from these hosts and uses `https` when a host is listed in `ingress.tls`.
Set the three values yourself if you expose LinkShelf another way.

## Custom domains

Shelves can be served on a domain of their own, see [Custom domains](/docs/self-hosting/custom-domains). List those
domains in `ingress.extraHosts` (a wildcard works too) and add a certificate for them to `ingress.tls`:

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
load balancer in front of it may overwrite it**. Otherwise shelves only work through their path.

## Strict origins

With the ingress enabled, `strictOrigins` (default `true`) sets `app.strictOrigins`: the API only answers on the
`apiUrl` host and only to browsers from the frontend or a shelf's domain. See
[Strict origins](/docs/self-hosting/configuration#strict-origins). Set `strictOrigins: false` to keep the API open.
Without an ingress it has no effect.

To make the instance itself, not a shelf, reachable on another domain too, add it to `ingress.extraHosts` and
`ingress.tls`, and to `APP_ADDITIONALORIGINS` (comma-separated for more than one). See
[Reachable on more than one domain](/docs/self-hosting/configuration#reachable-on-more-than-one-domain).

```yaml
ingress:
  extraHosts:
    - links.example.org
  tls:
    - secretName: linkshelf-tls
      hosts:
        - links.example.com
        - api.example.com
        - links.example.org
env:
  - name: APP_ADDITIONALORIGINS
    value: "https://links.example.org"
```

## Analytics

[Plausible](/docs/self-hosting/docker#analytics) is off by default. The chart's `plausible.enabled`, `plausible.domain`,
`plausible.apiHost` and `plausible.proxy` values set the matching `NUXT_PUBLIC_PLAUSIBLE_*` variables.

```yaml
plausible:
  enabled: true
  domain: links.example.org
  apiHost: https://plausible.example.org
```

## Themes and assets

Mount the directories with `extraVolumes` and `extraVolumeMounts` and point `THEMES_DIRECTORY` and `ASSETS_DIRECTORY`
at them. See [Themes and assets](/docs/self-hosting/configuration#themes-and-assets).

## Config file

Some settings are awkward as a single variable, e.g. `server.trustedProxies` (a list) or a full `oidc` block.
`configFile` takes the same nested YAML as the backend config and is rendered into a ConfigMap. When it's non-empty the
chart mounts it and sets `CONFIGURATION_FILE_PATH`. `env` still wins over the same key. Never put secrets in it.

```yaml
configFile:
  server:
    trustedProxies:
      - 10.0.0.0/8
  logging:
    level: debug
```

## Resources

The default `resources` (100m CPU / 192Mi memory requested, 512Mi limit) leave plenty of headroom. A small instance
with little traffic runs at around 15m CPU and 40MB memory.

## Additional resources

`additionalResources` takes a list of manifests, as maps or as strings. Both are rendered with `tpl`, so
`{{ .Release.Name }}` works.

```yaml
additionalResources:
  - apiVersion: networking.k8s.io/v1
    kind: NetworkPolicy
    metadata:
      name: "{{ .Release.Name }}-default-deny"
    spec:
      podSelector: {}
      policyTypes: [Ingress]
```

## Postgres for evaluation

`postgresql.enabled: true` deploys a single Postgres next to LinkShelf, using the
[groundhog2k/postgres](https://github.com/groundhog2k/helm-charts/tree/master/charts/postgres) chart. Everything under
`postgresql` is passed on to it.

It needs two Secrets. `settings.existingSecret` holds the superuser password under `POSTGRES_PASSWORD`.
`userDatabase.existingSecret` holds `POSTGRES_DB`, `USERDB_USER` and `USERDB_PASSWORD`, and LinkShelf reads the same
Secret. The key names can be changed with the `secretKey` fields of that chart.

```yaml
postgresql:
  enabled: true
  settings:
    existingSecret: linkshelf-postgres-admin
  userDatabase:
    existingSecret: linkshelf-postgres
```

There is no backup, replication or upgrade handling, so use your own database for anything you care about. It can't be
combined with setting the `DATABASE_*` variables in `env`.

## OpenShift

The default security context sets no user or group ID, so the restricted SCC can assign its own. The image runs with an
arbitrary user ID and a read-only root filesystem.

- The bundled Postgres does not run on OpenShift. Bring your own database.
- OpenShift converts an Ingress to a Route. For a native `Route`, add it with `additionalResources`.
