---
title: Docker
order: 4
---

You can start using LinkShelf with Docker. You can use this Docker Compose file 

A docker compose file can look for example like this:

```yaml
services:
  linkshelf:
    image: ghcr.io/m-mattia-m/linkshelf:latest # consider to use a specific version and not latest
    restart: unless-stopped
    ports:
      - "3000:3000" # Frontend port
      - "8085:8085" # Backend port
    # You can overwrite every attribute from the config file like app.name -> APP_NAME
    environment:
      # You don't have to set this variable, the value below is similar to the default value but you can overwrite it
      - CONFIGURATION_FILE_PATH="./config.local.yaml"
      - DATABASE_PASSWORD="linkshelf" # IMPORTANT: should be changed
    volumes:
      - ./backend/config.default.yaml:/app/config.default.yaml
    depends_on:
      # Wait until postgres accepts connections: the backend exits if its first
      # DB ping fails, which happens on a fresh volume without this gate.
      postgres:
        condition: service_healthy

  postgres:
    image: postgres
    restart: unless-stopped
    ports:
      - "5432:5432"
    shm_size: 128mb
    environment:
      POSTGRES_USER: linkshelf
      POSTGRES_PASSWORD: linkshelf # IMPORTANT: change this to a better one
      POSTGRES_DB: linkshelf
    healthcheck:
      test: [ "CMD-SHELL", "pg_isready -U linkshelf -d linkshelf" ]
      interval: 2s
      timeout: 3s
      retries: 30
```

## Custom domains for shelves

A shelf can be served on a domain of its own, like `profile.example.com`, see [Custom domains](/docs/self-hosting/custom-domains).
For that to work, put a reverse proxy in front of port `3000` that sends the domain to LinkShelf, and make sure it
**passes the original `Host` header (or `X-Forwarded-Host`) on and doesn't overwrite it**. If the proxy replaces it with
its own address, every domain looks like your normal LinkShelf address, and shelves only work through their path.

Set `APP_APP_FRONTENDURL` and `APP_SERVER_HOST` to your real public addresses, and add `APP_APP_STRICTORIGINS=true` to lock
the API to them.

