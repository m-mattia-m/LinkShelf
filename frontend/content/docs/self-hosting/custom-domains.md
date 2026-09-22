---
title: Custom domains
order: 7
---

A shelf can be served on a domain of its own. If LinkShelf runs at `linkshelf.example.com` and a shelf has the domain
`profile.example.com`, then `https://profile.example.com` shows that shelf, without a path in the URL.

This works the same on an instance with and without [user-based paths](/docs/self-hosting/configuration#user-based-paths).

## The rules

- A shelf is reached through **either a path or a domain**, never both. In the shelf form, the tab that is open when
  you save decides which one is kept, and the other one is cleared.
- A domain belongs to one shelf, and a shelf has one domain. Two shelves can't share a domain, even on different
  accounts.
- It has to be a proper domain name: at least two labels like `profile.example.com`, letters, numbers and hyphens, and a
  top-level domain made of letters. IP addresses, `localhost`, wildcards, and anything with `https://` or a path are
  rejected.
- A port is allowed, for example `profile.example.com:9443`. It has to be the port people actually type in the address
  bar. `:80` and `:443` are removed, because browsers leave them out. `profile.example.com:9443` and
  `profile.example.com` are two different domains.
- LinkShelf tidies the input: it trims it, lowercases it and removes a trailing dot or slash. `Profile.Example.com./`
  is saved as `profile.example.com`.
- A shelf can't use the address of LinkShelf itself, that is the host of `app.frontendUrl` and `server.host`.

On such a domain only `/` shows the shelf. Every other page on that host, like `/app` or `/docs`, still works as usual,
but the visitors of a shelf never need them.

## Set it up

LinkShelf only decides what a domain shows. Getting the traffic to LinkShelf is up to your infrastructure, and it takes
three steps for each domain:

1. **DNS.** Point the domain (or a wildcard such as `*.shelves.example.com`) at the same place as your LinkShelf
   frontend, with an `A`, `AAAA` or `CNAME` record.
2. **Routing.** Send the domain to the frontend, port `3000` of the image, not to the backend.
3. **TLS.** Get a certificate for the domain. LinkShelf always shows `https://` in front of a domain.

Then open the shelf in LinkShelf, choose the **Domain** tab, and enter the domain. The shelf's "Open" button now goes to
it. A domain that was just added or removed can take up to a minute to take effect.

### The Host header has to arrive unchanged

LinkShelf recognizes a shelf by the address the visitor typed. It reads the `X-Forwarded-Host` header, and the `Host`
header when there is none.

**Your reverse proxy or load balancer must pass one of them on, with the original domain, and must not overwrite it.**
A proxy that replaces `Host` with its own address, for example `localhost:3000` or a service name, makes every domain
look like your normal LinkShelf address. Shelves then still work through their path, but never through their domain.

For nginx this means:

```nginx
server {
    server_name profile.example.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
    }
}
```

Caddy and Traefik pass the original host on by default. In Kubernetes see [the chart's
`ingress.extraHosts`](/docs/self-hosting/kubernetes#custom-domains).

## Locking the API to your instance

With `app.strictOrigins` the API only accepts browsers that come from your frontend or from the domain of a shelf, and
only answers requests addressed to `server.host`. Domains are picked up automatically, there is nothing to list. See
[Strict origins](/docs/self-hosting/configuration#strict-origins).

## Trying it on your machine

Most browsers (Chrome, Firefox and Edge) resolve every `*.localhost` name to your own computer, so no DNS is needed. Start LinkShelf with the frontend
on port 3000, create a shelf with the domain `profile.localhost:3000` and open `http://profile.localhost:3000`.
