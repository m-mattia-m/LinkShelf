-- Replaces the placeholder About, Contact, Imprint, Terms of Use and Privacy
-- Policy pages seeded by 0002 (and 0014) with fuller defaults. The texts are
-- written for any self-hosted instance: they point to the instance owner for
-- questions about the instance and to the LinkShelf repository (LICENSING.md)
-- for questions about the software, so no contact details are shipped here.
-- Only rows still holding the untouched seed are replaced, so an operator who
-- already rewrote a page keeps their own wording. English only; other
-- languages fall back to English until an operator adds a translation.

UPDATE `setting`
SET `value` = '
# About

LinkShelf is a free, open-source and self-hosted alternative to Linktree. Collect your links on a page (a *shelf*), style it and share it with anyone.

## What you can do

- Create as many shelves, sections and links as you like and order them by drag and drop
- Add icons and colors to your links
- Pick one of the built-in themes or create your own
- Share a shelf by URL, QR code or on its own domain
- Hide a shelf from search engines if you want to

## Who runs this website

LinkShelf is software that anyone can install on their own server. This website is one of those installations (an *instance*) and is operated by its own owner, not by the LinkShelf project. See the [Imprint](/imprint) for details.

## Who builds LinkShelf

LinkShelf is developed by [m-mattia-m](https://github.com/m-mattia-m) and the open-source community. It is licensed under the [GNU AGPLv3](https://github.com/m-mattia-m/LinkShelf/blob/main/LICENSE).

To learn more, report a bug or contribute, visit the [LinkShelf repository on GitHub](https://github.com/m-mattia-m/LinkShelf).
'
WHERE `key` = 'about' AND `language` = 'en'
  AND `value` = '
# About

LinkShelf is an open-source bookmark manager designed to help you organize and access your favorite websites easily. It allows you to create shelves, sections, and links, providing a structured way to manage your bookmarks.
';

UPDATE `setting`
SET `value` = '
# Contact

This page is about the LinkShelf software. For questions about this website, a specific shelf or your account here, please contact the owner of this instance (see the [Imprint](/imprint)). The LinkShelf project cannot help with the content, accounts or settings of individual instances.

## LinkShelf project

- **Bugs and feature requests:** open an [issue on GitHub](https://github.com/m-mattia-m/LinkShelf/issues)
- **Community and questions:** join us on [Discord](https://discord.com/linkshelf)
- **Email:** see the contact address in [LICENSING.md](https://github.com/m-mattia-m/LinkShelf/blob/main/LICENSING.md#questions) and start the subject with "LinkShelf"

Please do not report security issues publicly. See the [security policy](https://github.com/m-mattia-m/LinkShelf/blob/main/SECURITY.md) instead.
'
WHERE `key` = 'contact' AND `language` = 'en'
  AND `value` = '
# Contact

If you have any questions, suggestions, or need support, feel free to reach out to us:
- [Discord](https://discord.com/linkshelf)
';

UPDATE `setting`
SET `value` = '
# Imprint

This website runs [LinkShelf](https://github.com/m-mattia-m/LinkShelf), open-source software that anyone can host themselves. Every LinkShelf instance is operated independently by its own owner. The LinkShelf project does not host, operate or control this instance.

## Operator of this instance

*The owner of this instance has not added their details yet.*

> **Instance owners:** replace this section with your name and contact address, and any other information required in your country. You can edit this page in the admin area.

## Responsibility for content

- The **instance owner** is responsible for operating this website and for deciding who may use it.
- The **person who created a shelf** is responsible for its content and for the links they share.

The LinkShelf project has no access to this instance and cannot review, edit or remove content on it. To report content or ask about an account, contact the instance owner.

## About LinkShelf

The source code is available in the [LinkShelf repository on GitHub](https://github.com/m-mattia-m/LinkShelf). If you have a question about how LinkShelf itself works, you can contact the maintainers using the address in [LICENSING.md](https://github.com/m-mattia-m/LinkShelf/blob/main/LICENSING.md#questions). For everything about this website, please contact the instance owner.
'
WHERE `key` = 'imprint' AND `language` = 'en'
  AND `value` = '
# Imprint

LinkShelf is developed and maintained by the LinkShelf Team.
For more information, visit our [GitHub repository](https://github.com/m-mattia-m/LinkShelf).
';

UPDATE `setting`
SET `value` = '
# Terms of Use

This website runs LinkShelf and is operated by the instance owner named in the [Imprint](/imprint). By using it, you agree to these terms.

## Your account

- Provide accurate information and keep your password safe.
- You are responsible for everything done with your account.

## Your content

- You keep all rights to the shelves, links and themes you create.
- You allow the instance owner to store your content and show it to others as needed to run the service.
- Shelves are public. Anyone with the link can see them, so do not add anything you want to keep private.
- You are responsible for the content you publish and the websites you link to.

## Acceptable use

Do not use this website to:

- publish or link to illegal content, malware or phishing
- send spam or mislead people, for example by impersonating others
- infringe the rights of others, such as copyright or trademarks
- disrupt or attack the service or other users

## Moderation

The instance owner may remove content or suspend accounts that break these terms, and may change or stop the service at any time.

## No warranty

This service is provided "as is", without any warranty. The LinkShelf software is published under the [GNU AGPLv3](https://github.com/m-mattia-m/LinkShelf/blob/main/LICENSE), which also comes without warranty. The LinkShelf project is not a party to these terms.

## Changes

The instance owner may update these terms. Continuing to use the website after a change means you accept the new terms.
'
WHERE `key` = 'terms_of_use' AND `language` = 'en'
  AND `value` = '
# Terms of Use

By using LinkShelf, you agree to comply with our terms of use. Please read them carefully before using the service. For more details, visit our [GitHub repository](https://github.com/m-mattia-m/LinkShelf)
';

UPDATE `setting`
SET `value` = '
# Privacy Policy

This website runs LinkShelf and is operated by the instance owner named in the [Imprint](/imprint), who is responsible for your data. The LinkShelf project does not receive any data from this instance.

## Data we store

- **Account:** first and last name, email address, username and a hashed password. If you sign in with an external provider (single sign-on), we store the provider''s user ID instead of a password.
- **Content:** your shelves, sections, links and themes.
- **Technical:** login tokens and one-time tokens for email verification, password reset and invitations.

## Visitors of shelves

You can view shelves without an account. LinkShelf does not set cookies. Depending on how this instance is hosted, the server may log standard request data such as your IP address and browser for security and troubleshooting.

## Browser storage

When you sign in, your login tokens are stored in your browser''s local storage so you stay signed in. They are removed when you sign out.

## Emails

We only send emails that are needed for your account, such as email verification, password reset and invitations.

## Sharing

Your data is not sold. It is only shared where needed to run the service, for example with the hosting provider, the email provider or your single sign-on provider.

## Your rights

Your data is kept as long as your account exists. To access, correct or delete your data, contact the instance owner.


## Analytics

This website may use [Plausible Analytics](https://plausible.io), a privacy-friendly, cookie-free analytics tool, if enabled by the site operator. Plausible does not use cookies and does not collect personally identifiable information. See the [Plausible data policy](https://plausible.io/data-policy) for details.
'
WHERE `key` = 'privacy_policy' AND `language` = 'en'
  AND `value` = '
# Privacy Policy

LinkShelf is committed to protecting your privacy. We do not collect personal data beyond what is necessary for the functionality of the service. For more details, visit our [GitHub repository](https://github.com/m-mattia-m/LinkShelf)


## Analytics

This website may use [Plausible Analytics](https://plausible.io), a privacy-friendly, cookie-free analytics tool, if enabled by the site operator. Plausible does not use cookies and does not collect personally identifiable information. See the [Plausible data policy](https://plausible.io/data-policy) for details.
';
