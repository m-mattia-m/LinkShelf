---
title: Shelves
order: 2
---

## Path or domain

A shelf is reached through **either a path or a domain**, never both. Pick the tab in the shelf form.

| Option               | Public URL                      |
|----------------------|---------------------------------|
| Path                 | `/<path>`                       |
| Path, user-based     | `/<username>/<path>`            |
| Domain               | `https://profile.example.com`   |

- Path: letters, numbers and hyphens
- With [user-based paths](/docs/self-hosting/configuration#user-based-paths), two users can both have `/profile`
- Without them, paths are unique on the instance, and words like `app`, `auth` or `docs` are reserved
- Domain setup needs DNS and a reverse proxy, see [Custom domains](/docs/self-hosting/custom-domains)

Changing the username in your profile changes the URL of all your shelves when user-based paths are on. The app asks
before it does.

## Sections and links

- **Section**: a title, and a group of links
- **Link**: title, URL, icon, optional color
- Icons: search the [Lucide](https://lucide.dev) and [Simple Icons](https://simpleicons.org) sets
- Color: hex like `#588157`. Empty means the theme decides.
- Order: drag, then **Save order**

## Footer

Shown at the bottom of the public page. Per shelf, one of:

- Default: "Powered by LinkShelf"
- Custom text, up to 500 characters. Supports `**bold**`, `*italic*` and `[links](https://example.com)`, nothing else.
- Off

## Search engines

Turn on **noindex** to ask search engines to skip this shelf. It doesn't affect other shelves or the instance.

## Deleting

Deleting a shelf deletes its sections and links too. The dialog shows how many.
