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

Create the Secrets first. The database Secret needs the keys `host`, `port`, `username`, `password` and `database`,
and `jwt-secret` signs the login tokens. The key names are configurable.

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
secrets:
  existingSecret:
    name: linkshelf-secrets
database:
  existingSecret:
    name: linkshelf-db
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

**Settings.** Anything in the backend config file can be set as an environment variable through `env`, for example
`smtp.host` becomes `APP_SMTP_HOST`. The keys are listed in
[config.default.yaml](https://github.com/m-mattia-m/LinkShelf/blob/main/backend/config.default.yaml). Email
verification is on by default, so either set up SMTP or turn it off:

```yaml
env:
  APP_SMTP_HOST: smtp.example.com
  APP_SMTP_FROM: no-reply@example.com
  # APP_AUTHENTICATION_EMAILVERIFICATION_ENABLED: "false"
```

**Secrets.** `secrets.existingSecret.name` and its `jwtSecret` key are required. The bootstrap admin from the image is
switched off. To get an admin, set `APP_AUTHENTICATION_BOOTSTRAPADMIN_EMAIL` in `env` and name the key of the password
in `secrets.existingSecret.keys.bootstrapAdminPassword`. That account does not need email verification, so it works
before SMTP is set up. `smtpPassword` and `oidcClientSecret` work the same way, leave a key empty if you don't use it.

**Database.** Point `database.existingSecret.name` to a Secret with the connection details (key names are configurable
under `database.existingSecret.keys`). Set `database.engine` to `MYSQL` for MySQL and `database.params` for extra
connection parameters such as `sslmode=require`.

**Ingress.** The browser talks to the backend directly, so the frontend and the backend need their own host. The
chart derives `frontendUrl`, `apiUrl` and `oidcRedirectUrl` from these hosts and uses `https` when a host is listed
in `ingress.tls`. Set the three values yourself if you expose LinkShelf another way.

**Themes and assets.** Mount the directories with `extraVolumes` and `extraVolumeMounts` and point
`APP_THEMES_DIRECTORY` and `APP_ASSETS_DIRECTORY` at them.

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
combined with `database.existingSecret`.

## OpenShift

The default security context sets no user or group ID, so the restricted SCC can assign its own. The image runs with an
arbitrary user ID and a read-only root filesystem.

- The bundled Postgres does not run on OpenShift. Bring your own database.
- An Ingress is converted to a Route by OpenShift. If you need a native `Route`, add it with `additionalResources`.
