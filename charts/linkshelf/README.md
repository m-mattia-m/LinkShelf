# LinkShelf Helm chart

Deploys [LinkShelf](https://github.com/m-mattia-m/LinkShelf) with a Deployment, a Service and an optional Ingress.
You need a Postgres or MySQL database, either your own or the optional Postgres that comes with the chart.

The chart never creates a Secret. All credentials come from Secrets you provide, for example through External Secrets,
Sealed Secrets or `kubectl`.

The chart version follows the app version. Chart `1.4.0` deploys the image of the release `1.4.0`, unless you set
`image.tag`.

## Install

```bash
helm repo add linkshelf https://m-mattia-m.github.io/LinkShelf
helm repo update
```

Create the Secrets first. There is no dedicated "secrets" section - a secret is just an `env` entry that uses
`valueFrom` instead of `value`, plain Kubernetes syntax, so point it at whatever Secret you (or External Secrets)
already created.

```bash
kubectl create secret generic linkshelf-db \
  --from-literal=host=postgres.example.svc \
  --from-literal=port=5432 \
  --from-literal=username=linkshelf \
  --from-literal=password=change-me \
  --from-literal=database=linkshelf

kubectl create secret generic linkshelf-secrets \
  --from-literal=jwt-secret="$(openssl rand -hex 32)"
```

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

Without an ingress the release notes show how to reach it with `kubectl port-forward`.

## Configuration

Every option is documented in [values.yaml](values.yaml), and `values.schema.json` rejects wrong values on install.

**Settings.** Anything in the backend config file can be set as an environment variable in `env`, native Kubernetes
`env` syntax (`value` or `valueFrom`), for example `smtp.host` becomes `SMTP_HOST`. The keys are listed in
[config.default.yaml](https://github.com/m-mattia-m/LinkShelf/blob/main/backend/config.default.yaml). Email
verification is on by default, so either set up SMTP or turn it off:

```yaml
env:
  - name: SMTP_HOST
    value: smtp.example.com
  - name: SMTP_FROM
    value: no-reply@example.com
  # - name: AUTHENTICATION_EMAILVERIFICATION_ENABLED
  #   value: "false"
```

**Secrets.** There is no separate "secrets" section - use `valueFrom.secretKeyRef` on the `env` entry instead of
`value`, exactly like a Pod spec. `AUTHENTICATION_JWTSECRET` is required. The bootstrap admin from the image is
switched off; to get an admin, add `AUTHENTICATION_BOOTSTRAPADMIN_EMAIL` and `AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD`
to `env`. Its username is `admin`, change it with `AUTHENTICATION_BOOTSTRAPADMIN_USERNAME`. That account does not need
email verification, so it works before SMTP is set up. `SMTP_PASSWORD` and `AUTHENTICATION_OIDC_CLIENTSECRET` work the
same way, leave them out if you don't use them.

**Database.** Add `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USERNAME`, `DATABASE_PASSWORD` and `DATABASE_NAME` to
`env` (all 5 are required, each can come from a different Secret). Set `database.engine` to `MYSQL` for MySQL and
`database.params` for extra connection parameters such as `sslmode=require`.

**Ingress.** The browser talks to the backend directly, so the frontend and the backend need their own host. The
chart derives `frontendUrl`, `apiUrl` and `oidcRedirectUrl` from these hosts and uses `https` when a host is listed
in `ingress.tls`. Set the three values yourself if you expose LinkShelf another way.

**Custom domains.** A shelf can be served on a domain of its own. List the domains in `ingress.extraHosts` (a wildcard
like `*.shelves.example.com` works too) and they are sent to the frontend. Add a certificate for them to `ingress.tls`.
The ingress controller has to pass the original `Host` header (or `X-Forwarded-Host`) on, and no annotation or load
balancer in front of it may overwrite it, otherwise shelves only work through their path.

**Strict origins.** With the ingress enabled, `strictOrigins` (default `true`) locks the API to the instance: it only
answers requests for the `apiUrl` host, and only browsers from `frontendUrl` or from a shelf's domain may call it. The
chart sets `APP_STRICTORIGINS`, `SERVER_HOST` and `SERVER_SCHEME` for that. Set it to `false` to keep the
API open. Without an ingress it has no effect.

**A second domain for the instance itself.** To make the instance (not a shelf) reachable on another domain too, e.g.
`links.example.org` besides `frontendUrl`, add it to `ingress.extraHosts` and `ingress.tls` like a shelf's custom
domain, and also allowlist it for CORS with `APP_ADDITIONALORIGINS` in `env` (comma-separated for more than one):

```yaml
ingress:
  extraHosts:
    - links.example.org
  tls:
    - secretName: linkshelf-tls
      hosts:
        - linkshelf.example.com
        - api.linkshelf.example.com
        - links.example.org
env:
  - name: APP_ADDITIONALORIGINS
    value: "https://links.example.org"
```

**Themes and assets.** Mount the directories with `extraVolumes` and `extraVolumeMounts` and point
`THEMES_DIRECTORY` and `ASSETS_DIRECTORY` at them.

**Config file.** `env` covers most settings, but some are awkward as a single flattened variable, e.g.
`server.trustedProxies` (a list) or a full `oidc` block. `configFile` takes the same nested YAML shape as
[config.default.yaml](https://github.com/m-mattia-m/LinkShelf/blob/main/backend/config.default.yaml) and is rendered
into a ConfigMap. When it's non-empty the chart mounts it and points `CONFIGURATION_FILE_PATH` at it automatically -
no manual wiring needed, and `env` entries still win over the same key here. Never put secrets in it, a ConfigMap is
plain text - use `env` with `valueFrom.secretKeyRef` for those instead.

```yaml
configFile:
  server:
    trustedProxies:
      - 10.0.0.0/8
  logging:
    level: debug
```

**Resource usage.** The chart's `resources` defaults (100m CPU / 192Mi memory requested, 512Mi memory limit) are
generous headroom, not a sizing guide. A basic instance with little traffic has been observed running at around
15m CPU and 40MB memory - tune `resources` down from there if you're running many small instances and want to pack
them tighter.

**Anything else.** `additionalResources` takes a list of manifests, as maps or as strings. Both are rendered with
`tpl`, so `{{ .Release.Name }}` works.

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

`postgresql.enabled: true` deploys a single Postgres instance next to LinkShelf. It uses the
[groundhog2k/postgres](https://github.com/groundhog2k/helm-charts/tree/master/charts/postgres) chart as a dependency,
pinned in `Chart.yaml`. That chart runs the official `postgres` image, and everything under `postgresql` in the values
is passed on to it.

It needs two Secrets. `settings.existingSecret` holds the superuser password under `POSTGRES_PASSWORD`.
`userDatabase.existingSecret` holds `POSTGRES_DB`, `USERDB_USER` and `USERDB_PASSWORD`, and LinkShelf reads the same
Secret. The key names can be changed with the `secretKey` fields, see the groundhog2k chart.

```yaml
postgresql:
  enabled: true
  settings:
    existingSecret: linkshelf-postgres-admin
  userDatabase:
    existingSecret: linkshelf-postgres
```

There is no backup, replication or upgrade handling. Use your own database for anything you care about. It cannot be
combined with setting `DATABASE_HOST`/`DATABASE_PORT`/`DATABASE_USERNAME`/`DATABASE_PASSWORD`/`DATABASE_NAME` in `env`
yourself.

## OpenShift

The default security context sets no user or group ID, so the restricted SCC can assign its own. The image runs with an
arbitrary user ID and a read-only root filesystem.

- The bundled Postgres does not run on OpenShift. Bring your own database.
- An Ingress is converted to a Route by OpenShift. If you need a native `Route`, add it with `additionalResources`.
