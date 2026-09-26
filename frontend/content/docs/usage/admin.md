---
title: Admin
order: 4
---

Users with the `admin` role get **Settings** in the sidebar. The first admin comes from
`authentication.bootstrapAdmin` in the [config](/docs/self-hosting/configuration).

## General

Content of the pages around the app. Markdown, one text per language (English, Deutsch, Schwiizerdütsch):

- About
- Contact
- Imprint
- Terms of use
- Privacy policy

Each page has **Show on public site**. There is also **Redirect to dashboard**, which sends visitors of the homepage
straight to the dashboard.

## Users

- Create a user with a password, or leave it empty to **invite** them by email
- Edit, change password, delete
- Status: `Active`, `Pending verification`, `Invited`
- Resend the verification email, or mark a user as verified by hand

Changing a username changes the URL of their shelves if [user-based paths](/docs/self-hosting/configuration#user-based-paths)
are on.

## Themes

Lists every theme users created. Delete the ones that don't belong.

## Email

Verification, invites and password resets need SMTP. **Settings** shows the host and from-address in use, read-only.
Change them in the [config](/docs/self-hosting/configuration).
