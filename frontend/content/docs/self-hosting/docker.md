---
title: Docker
order: 4
---

Every config option can be set as an environment variable, see [Configuration](/docs/self-hosting/configuration).

```yaml
services:
  linkshelf:
    image: ghcr.io/m-mattia-m/linkshelf:latest # pin a version instead of latest
    restart: unless-stopped
    ports:
      - "3000:3000" # frontend
      - "8085:8085" # backend
    environment:
      DATABASE_ENGINE: POSTGRES
      DATABASE_HOST: postgres
      DATABASE_PORT: 5432
      DATABASE_PASSWORD: linkshelf # change
      AUTHENTICATION_JWTSECRET: change-me-to-a-long-random-value
      AUTHENTICATION_BOOTSTRAPADMIN_EMAIL: admin@example.com
      AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD: change-me
      # No SMTP? Then turn these off, or new users can't sign in.
      # AUTHENTICATION_EMAILVERIFICATION_ENABLED: false
      # AUTHENTICATION_PASSWORDRESET_ENABLED: false
    depends_on:
      # The backend exits if its first DB ping fails, so wait for postgres.
      postgres:
        condition: service_healthy

  postgres:
    image: postgres
    restart: unless-stopped
    shm_size: 128mb
    environment:
      POSTGRES_USER: linkshelf
      POSTGRES_PASSWORD: linkshelf # change, same as above
      POSTGRES_DB: linkshelf
    volumes:
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: [ "CMD-SHELL", "pg_isready -U linkshelf -d linkshelf" ]
      interval: 2s
      timeout: 3s
      retries: 30

volumes:
  postgres-data:
```

```bash
docker compose up -d
```

Open `http://localhost:3000` and sign in with the bootstrap admin.

## Behind a domain

The browser talks to the backend directly, so it needs a public address of its own:

- `NUXT_PUBLIC_API_BASE`: public URL of the backend, e.g. `https://api.example.com`
- `APP_FRONTENDURL`: public URL of the frontend, used in emails
- `SERVER_HOST`: public host of the backend

Add `APP_STRICTORIGINS=true` to lock the API to those addresses, see
[Strict origins](/docs/self-hosting/configuration#strict-origins). To also reach the instance on another domain, add it
to `APP_ADDITIONALORIGINS` (comma-separated for more than one), see [Reachable on more than one
domain](/docs/self-hosting/configuration#reachable-on-more-than-one-domain).

## Analytics

Optional, cookie-free visitor analytics via [Plausible](https://plausible.io), off by default.

- `NUXT_PUBLIC_PLAUSIBLE_ENABLED`: `true` to turn it on
- `NUXT_PUBLIC_PLAUSIBLE_DOMAIN`: the site identifier registered in Plausible. Falls back to the request's hostname,
  so set it when the instance is reachable on more than one domain
- `NUXT_PUBLIC_PLAUSIBLE_API_HOST`: your Plausible instance, e.g. `https://plausible.example.com`. Defaults to
  `https://plausible.io`
- `NUXT_PUBLIC_PLAUSIBLE_PROXY`: `true` to route the script and API calls through this app's own domain
  (under `/_plausible`), to reduce ad-blocker interference

```yaml
    environment:
      NUXT_PUBLIC_PLAUSIBLE_ENABLED: "true"
      NUXT_PUBLIC_PLAUSIBLE_DOMAIN: links.example.org
      NUXT_PUBLIC_PLAUSIBLE_API_HOST: https://plausible.example.org
```

## Config file

To use a file instead of environment variables, mount it and point to it:

```yaml
    environment:
      CONFIGURATION_FILE_PATH: /app/config.local.yaml
    volumes:
      - ./config.local.yaml:/app/config.local.yaml
```

It overrides the defaults from `config.default.yaml`, environment variables win over both.

## Themes and images

Mount a directory for your own [instance themes and images](/docs/self-hosting/configuration#themes-and-assets).

## Custom domains for shelves

See [Custom domains](/docs/self-hosting/custom-domains). Short version: put a reverse proxy in front of port `3000` that
**passes the original `Host` header (or `X-Forwarded-Host`) on**.
