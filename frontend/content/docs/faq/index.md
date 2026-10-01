---
title: FAQ
icon: i-lucide-circle-help
order: 5
navigation: true
---

## Day-2 operations

### How do I upgrade LinkShelf?

Bump the image tag, or the chart version if you use Helm. With GitOps tools like ArgoCD this means changing the
`targetRevision` in Git. The new pod applies pending database migrations on startup, so there is nothing to run
manually.

### How are database migrations handled?

Migrations are SQL files managed with [golang-migrate](https://github.com/golang-migrate/migrate). They run
automatically on startup, in this order:

1. Validate the config
2. Connect to the database
3. Apply pending migrations
4. Create the bootstrap admin, if none exists
5. Set up OIDC and mail, if enabled
6. Start the API

### Can I roll back to an older version?

Rolling back the image or chart version rolls back the pods, but not the database. A rollback only works if the older
version can still run on the migrated schema. Rollbacks are not tested yet, so take a database backup before you
upgrade.

### How do I back up and restore LinkShelf?

All state is in the database, plus the optional themes and assets directories if you mount them. Backups are left to
the database on purpose: use regular SQL dumps or the tooling of your platform, e.g. an operator like
[CloudNativePG](https://cloudnative-pg.io) with WAL archiving. To restore, restore the database and start LinkShelf
against it.

### Can I run multiple replicas?

Yes. LinkShelf is stateless, so you can set `replicaCount` higher than 1 in the Helm chart. It has been run with 3 pods,
but this is not tested thoroughly yet.
