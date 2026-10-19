---
title: Automating custom domains
description: "The small, generic API contract an external controller can use to automate routing and certificates for LinkShelf custom domains."
order: 8
---

LinkShelf only decides what a domain shows. Routing, certificates and proof of ownership are infrastructure, and
LinkShelf knows nothing about Kubernetes, Ingress or cert-manager. This page describes the small, generic
contract an external **controller** (a script, an operator, a Caddy or nginx updater) can use to automate them.

## Spec, status and a nudge

| Piece | Owner | What it is |
| --- | --- | --- |
| Spec: "shelf X wants domain D" | LinkShelf | `GET /v1/service/domains` |
| Status: "D is pending, ready or in error, and these DNS records are needed" | the controller | `PUT /v1/service/domains/{domain}/status`, shown in the shelf form |
| Nudge: "something changed, look again" | LinkShelf | an optional [webhook](#webhooks) |

Write the controller **level-triggered**: on every run, read the full list and make the world match it. A webhook only
shortens the wait. **A webhook is a hint, the API is the truth.** Keep a periodic resync (every minute without
webhooks, every ten with them).

## The service token

Give the [service token](/docs/self-hosting/configuration#service-token) the scopes `domains:read` and
`domains:write-status`. Service endpoints accept only this token, a user token (even an admin's) gets `401`, and the
token is rejected on every user route.

## `GET /v1/service/domains`

Scope `domains:read`. Cursor paginated (`?limit=` 1 to 500, default 100, `?cursor=` the previous `next_cursor`) and
ordered by shelf id. Send `If-None-Match` with the `ETag` of an earlier page to get `304` when it did not change.

```json
{
  "items": [
    {
      "domain": "profile.example.com",
      "shelf_id": "...",
      "owner_id": "...",
      "created_at": "...",
      "updated_at": "...",
      "status": { "state": "pending", "reason": "txt_missing", "message": "", "records": [], "observed_domain": "profile.example.com" }
    }
  ],
  "next_cursor": null
}
```

`state` is empty until a controller reported one. Two shelves can list the same domain, because a saved domain is only
a claim, see [`requireReady`](#requireready). Only domains and ids are listed, never shelf content.

## `PUT /v1/service/domains/{domain}/status`

Scope `domains:write-status`.

```json
{
  "shelf_id": "...",
  "state": "pending",
  "reason": "txt_missing",
  "message": "optional free text fallback",
  "records": [
    { "type": "TXT", "name": "_linkshelf-challenge.profile.example.com", "value": "linkshelf-verification=..." }
  ]
}
```

- `state` is `pending`, `ready` or `error`. `reason` and `message` are free text up to 64 and 1024 characters, `records`
  up to 20 entries and 4 KB.
- The shelf form shows `state` as a badge, `records` as "set these DNS records", and `reason` through its own
  translations (known: `txt_missing`, `lookup_failed`, `limit_reached`, `reserved`, `taken`), falling back to `message`.
  LinkShelf attaches no other meaning to the values.
- `shelf_id` is the shelf the controller read the domain from. The answer is `409` when that shelf no longer has this
  domain, so a stale controller can not label a new one. Changing or clearing a domain resets its status.
- `ready` answers `409` while another shelf is already ready on the same domain.

## `requireReady`

`customDomains.requireReady` (`CUSTOMDOMAINS_REQUIREREADY`) is `false` by default: every saved domain is served, and
the status is informational. Self-hosters notice nothing.

With `true`, LinkShelf serves a domain only while its state is `ready`, and "one domain, one shelf" applies only among
ready domains. Without it a user could squat a domain by saving it unproven, and an unproven domain would be served as
soon as DNS points at your cluster. Turn it on only together with a controller that writes `ready`, otherwise no domain
is served. Switching it on also makes existing domains wait for their first `ready`.

## Webhooks

Off by default. Add endpoints under `webhooks`:

```yaml
webhooks:
  enabled: true
  endpoints:
    - url: https://controller.example.com/hooks/linkshelf
      secret: ${WEBHOOKS_CONTROLLER_SECRET}   # read from the environment, at least 32 bytes
      events: ["shelf.domain.changed"]         # or ["*"]
```

Every delivery is a `POST` with a JSON envelope. The payload holds ids and domains only:

```json
{ "id": "evt_...", "type": "shelf.domain.changed", "occurred_at": "...",
  "data": { "shelf_id": "...", "owner_id": "...", "domain": "new.example.com", "previous_domain": "" } }
```

`shelf.domain.changed` is sent when a domain is set, changed or cleared, and when its shelf or owner is deleted.

Verify the header `X-LinkShelf-Signature: t=<unix>,v1=<hex>`, where `v1` is the HMAC-SHA256 of `t + "." + body` with
your secret. Reject timestamps older than a few minutes. `X-LinkShelf-Event-Id` identifies a delivery, so handle
duplicates idempotently.

Delivery is **best effort**: asynchronous, a short timeout, three retries with backoff, then dropped with a log line.
Nothing is persisted, and events are lost on a restart. So answer `2xx` quickly and resync from the API.
