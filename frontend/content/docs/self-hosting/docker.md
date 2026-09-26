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
      APP_DATABASE_ENGINE: POSTGRES
      APP_DATABASE_HOST: postgres
      APP_DATABASE_PORT: 5432
      APP_DATABASE_PASSWORD: linkshelf # change
      APP_AUTHENTICATION_JWTSECRET: change-me-to-a-long-random-value
      APP_AUTHENTICATION_BOOTSTRAPADMIN_EMAIL: admin@example.com
      APP_AUTHENTICATION_BOOTSTRAPADMIN_PASSWORD: change-me
      # No SMTP? Then turn these off, or new users can't sign in.
      # APP_AUTHENTICATION_EMAILVERIFICATION_ENABLED: false
      # APP_AUTHENTICATION_PASSWORDRESET_ENABLED: false
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
- `APP_APP_FRONTENDURL`: public URL of the frontend, used in emails
- `APP_SERVER_HOST`: public host of the backend

Add `APP_APP_STRICTORIGINS=true` to lock the API to those addresses, see
[Strict origins](/docs/self-hosting/configuration#strict-origins).

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
