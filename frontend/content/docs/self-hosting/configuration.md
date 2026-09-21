---
title: Configuration
order: 6
---

## Configuration Options

All options are available via YAML configuration and can be overwritten via environment variables. Every variable
starts with `APP_`, followed by the key path. For example `app.name` in yaml is overridden by the environment variable
`APP_APP_NAME`, and `database.host` by `APP_DATABASE_HOST`.

```yaml
app:
  name: LinkShelf
  description: LinkShelf is an OpenSource alternative to linktree.
  environment: production
  logo: <base64-encoded-logo-or-path>
  frontendUrl: "http://localhost:3000"
  # When true, a shelf's public URL is /<username>/<path>, so two users can
  # both own /profile. When false (the default) it is /<path>, unique across
  # the whole instance - the simpler choice for a single-user or private
  # instance. Switching back to false fails on startup if two shelves then
  # share a path.
  userBasedPaths: false
  # When true, the API only accepts what belongs to this instance:
  #   - browsers may call it only from the origin of frontendUrl, or from the
  #     domain of a shelf that has one (so a shelf served on its own domain can
  #     still load its content). Requests without an Origin header, such as
  #     the frontend's server-side calls or curl, are not affected.
  #   - it only answers requests whose Host is server.host (plus server.port
  #     when domain.openapi.usePort is true). Anything else gets a 421. The
  #     health endpoints are always exempt so probes that use a pod IP work.
  # Requires frontendUrl and server.host to be set to the real public values.
  # Leave it false while they are still the localhost defaults.
  strictOrigins: false
server:
  scheme: http
  host: localhost
  port: 8085 # Do not change this port since the Containerfile exposes this port. It's just for development purposes.
  trustedProxies:
    - 127.0.0.1
database:
  engine: POSTGRES # MYSQL # POSTGRES
  host: localhost
  port: 15432 # 3306 # 5432
  username: linkshelf
  password: linkshelf
  name: linkshelf
  params: "sslmode=disable" # "charset=utf8mb4&parseTime=true" # "sslmode=disable" # optional
logging:
  level: info
domain:
  openapi:
    usePort: false
themes:
  # Directory of instance-theme YAML files (one file per theme, each with a
  # "name" and a "config"), scanned once at startup. Removing a file removes
  # that theme on the next restart; leave empty to skip instance themes
  # entirely. Mount a host directory here the same way compose.dev.yaml
  # mounts config.default.yaml.
  directory: ""
assets:
  # Directory of static files (e.g. theme background images) an instance
  # admin wants to make available to themes, scanned once at startup and
  # served at basePath (e.g. a file "my-dog.webp" here becomes
  # "{basePath}/my-dog.webp"). Leave directory empty to skip entirely.
  directory: ""
  basePath: "/images"
authentication:
  type: LOCAL # LOCAL # OIDC
  # jwtSecret signs the access tokens this backend issues itself, no matter
  # which auth type is active below. Override this in production via the
  # APP_AUTHENTICATION_JWTSECRET environment variable - never ship the
  # default value.
  jwtSecret: "change-me-to-a-long-random-value-in-production"
  accessTokenExpiryMinutes: 5
  refreshTokenExpiryMinutes: 1440 # 24h
  bootstrapAdmin:
    # Idempotently created/refreshed on every startup with role=admin.
    # Leave email and password empty to skip bootstrapping an admin account.
    email: admin@example.com
    password: "change-me"
    # Applied as configured, without the checks a user-chosen username goes
    # through (so it may be a reserved word such as "admin").
    username: admin
  oidc:
    # issuer and clientId are required when authentication.type is OIDC.
    # Works with any standards-compliant OIDC issuer (Keycloak, Zitadel,
    # Auth0, Google, ...).
    issuer: ""
    clientId: ""
    # clientSecret is optional: the login flow always runs PKCE, so a public
    # client registered without a secret works fine - leave this empty in
    # that case.
    clientSecret: ""
    # Must match the frontend's OIDC callback page, which completes the
    # login by POSTing the code+state here to /v1/auth/oidc/callback.
    redirectUrl: "http://localhost:3000/auth/callback"
  # Set to false to disable public self-registration (POST /v1/users without
  # an admin token). Admins can still create accounts, and OIDC
  # auto-provisioning is unaffected either way.
  registrationEnabled: true
  emailVerification:
    # When true, smtp.host and smtp.from must be set (checked at startup) and
    # local/password accounts must verify their email before they can log in.
    # OIDC accounts are always exempt.
    enabled: true
    tokenExpiryHours: 24
  passwordReset:
    # When true, the sign-in page offers "Forgot password?" and emails a
    # reset link. smtp.host and smtp.from must be set (checked at startup).
    # Set to false on an instance without working email.
    enabled: true
    tokenExpiryMinutes: 60
smtp:
  # Defaults to the Mailpit dev container (see compose.yaml) for local
  # development - matches the pattern of database.host/port also pointing at
  # its published compose port.
  host: localhost
  port: 1025
  username: ""
  password: ""
  from: "no-reply@linkshelf.local"
  # none | starttls | tls - Mailpit speaks plain SMTP with no TLS at all.
  tlsMode: none
```

## User-based paths

By default a shelf is available at `/<path>`, and every path is unique on the instance. That fits a private instance.

Set `app.userBasedPaths` to `true` (or `APP_APP_USERBASEDPATHS=true`) and shelves live at `/<username>/<path>` instead,
so two users can both have `/profile`.

- Every account has a username: lowercase letters, numbers and hyphens, 3 to 30 characters. Words LinkShelf uses itself,
  like `docs`, `app` or `api`, and a few generic ones like `admin` are not allowed.
- Accounts that exist without one get a username derived from their email address on the next start, so
  `jane.doe@example.com` becomes `jane-doe`. Users who sign in through OIDC get their `preferred_username` claim, or the
  part of the email before the `@`. The bootstrap admin uses `authentication.bootstrapAdmin.username`.
- Renaming a user changes the URL of all their shelves. Links that were shared with the old username stop working.
- Switching the setting on breaks existing `/<path>` links. Switching it off again needs unique paths, so LinkShelf
  refuses to start and lists the shelves that share one.
- A shelf remembers which way the setting was when it was created. If the setting has changed since, its edit page
  says so and shows the URL it has now. Shelves created under the current setting show no notice.
- While it is off, words like `app`, `auth` and `docs` can't be used as a shelf path, because those URLs belong to
  LinkShelf's own pages.

## Strict origins

By default the API accepts calls from any website. Set `app.strictOrigins` to `true`
(`APP_APP_STRICTORIGINS=true`) to lock it to your instance:

- **Browsers.** A page may only call the API if it was loaded from `app.frontendUrl`, or from the domain of a shelf (see
  [Custom domains](/docs/self-hosting/custom-domains)). Any other website gets a `403`. A `http://` shelf domain is only
  accepted when `frontendUrl` itself starts with `http://`. A domain that was just added or removed takes up to a minute
  to count.
- **Host.** The API only answers requests addressed to `server.host`. Anything else gets a `421 Misdirected Request`
  without a body. `server.port` is only compared when `domain.openapi.usePort` is `true`, because behind a proxy the
  public address usually has no port. If a trusted proxy from `server.trustedProxies` sends `X-Forwarded-Host`, that is
  used instead of `Host`.
- **Not affected.** Requests without an `Origin` header, like the frontend's server-side calls, `curl` or scripts, and
  the two health endpoints, so Kubernetes probes that address the pod by its IP keep working.

This limits which *websites* can use the API from a visitor's browser. It is not authentication, every endpoint still
checks its own token.

LinkShelf refuses to start with the setting on unless `app.frontendUrl` is a full `http(s)` URL and `server.host` is set,
and it logs what it allows. Set both to your real public addresses first. In the Helm chart it is switched on
automatically once the ingress is enabled, see [Kubernetes](/docs/self-hosting/kubernetes#custom-domains).

## Password reset

While `authentication.passwordReset.enabled` is `true` (the default), the sign-in page offers "Forgot password?". It
emails a link that lets the user choose a new password. That needs working email, so LinkShelf refuses to start with
the setting on and no `smtp.host` or `smtp.from`. On an instance without email, set it to `false`
(`APP_AUTHENTICATION_PASSWORDRESET_ENABLED=false`).

- The link is valid for `authentication.passwordReset.tokenExpiryMinutes` (60 by default) and works once.
- Changing the password signs the user out on every device.
- The page answers the same way whether or not the address has an account, and an account gets at most one email a
  minute.
- The bootstrap admin's password is set from `authentication.bootstrapAdmin.password` on every start, so it overwrites a
  reset one. Change it in the config instead.
