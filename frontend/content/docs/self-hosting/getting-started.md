---
title: Getting started
order: 1
---

One image with frontend and backend: `ghcr.io/m-mattia-m/linkshelf:latest`

| Port   | What     |
|--------|----------|
| `3000` | Frontend |
| `8085` | Backend  |

You need:

- PostgreSQL or MySQL
- A random `authentication.jwtSecret` of at least 32 bytes, e.g. from `openssl rand -base64 48`. LinkShelf refuses to
  start without one
- SMTP, unless you turn off email verification and password reset

Deploy with:

- [Docker](/docs/self-hosting/docker)
- [Kubernetes](/docs/self-hosting/kubernetes)

Then:

- [Configuration](/docs/self-hosting/configuration) for all options
- [Custom domains](/docs/self-hosting/custom-domains) to serve a shelf on its own domain
