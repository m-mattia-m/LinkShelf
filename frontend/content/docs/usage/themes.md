---
title: Themes
order: 3
---

A theme changes how a shelf's **public page** looks. It never affects the app itself.

## Pick one

In the shelf settings. Two groups:

- **Instance themes**: provided by the admin, the same for everyone
- **Your themes**: created by you, private

If a shelf's theme was deleted, the edit page says so and asks for a new one.

## Create one

Open **Themes**, press **New theme**, and write one `--property: value;` per line:

```css
--shelf-bg: linear-gradient(160deg, #0f172a, #1e293b);
--shelf-text: #f1f5f9;
--shelf-link-bg: #1e293b;
--shelf-link-text: #f8fafc;
--shelf-link-radius: 1rem;
--shelf-font-family: 'Inter', sans-serif;
```

| Property              | Value                                                        |
|-----------------------|--------------------------------------------------------------|
| `--shelf-bg`          | color, `rgb()`, `hsl()` or `linear-gradient()`/`radial-gradient()` |
| `--shelf-text`        | color                                                        |
| `--shelf-link-bg`     | color                                                        |
| `--shelf-link-text`   | color                                                        |
| `--shelf-link-radius` | length: `px`, `rem`, `em` or `%`                             |
| `--shelf-font-family` | font names                                                   |
| `--shelf-bg-image`    | `https://` URL or `/images/<file>` from the instance         |

Nothing else is accepted, no selectors and no other CSS. That keeps themes safe on public pages.

## Share one

Themes are private. To share:

1. **Themes**, open the menu of a theme, **Export**
2. Send the file
3. The other person uses **Import** and saves it as their own

## Instance themes

Admins add themes and images through config, see [Themes and assets](/docs/self-hosting/configuration#themes-and-assets).
Admins can also delete user themes under **Settings → Themes**.
