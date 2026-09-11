# LinkShelf

[![CI](https://github.com/m-mattia-m/LinkShelf/actions/workflows/ci.yaml/badge.svg)](https://github.com/m-mattia-m/LinkShelf/actions/workflows/ci.yaml) ![coverage](https://raw.githubusercontent.com/m-mattia-m/LinkShelf/refs/heads/main/.badges/main/coverage.svg) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m-mattia-m/LinkShelf?filename=backend%2Fgo.mod) ![Release](https://img.shields.io/github/v/release/m-mattia-m/LinkShelf)

LinkShelf is a web application designed to help users organize, manage, and share their favorite links and bookmarks in
a user-friendly interface. Whether you're a student, professional, or casual internet user, LinkShelf provides an
efficient way to keep track of important web resources.

## Features

- **Unlimited collections**: Create unlimited collections of links without any extra effort.
- **Accounts**: Manage your links and collections across multiple devices with user accounts.
- **Own Domains**: Use your own custom domain for one or several of your collections.
- **Theming**: Choose from multiple themes to personalize the look and feel of your LinkShelf.
- **Customization**: Customize the appearance and layout of your collections to suit your preferences.
- **Responsive Design**: Optimized for both desktop and mobile devices for seamless access on the go.
- **: Markdown Support**: Add descriptions and notes to your links using Markdown formatting.
- **Self-hosting**: Self-hosting option available for users who want complete control over their data.
- **Kubernetes**: Easily deploy LinkShelf on Kubernetes for scalable and reliable hosting.
- **Database**: Supports multiple database backends including PostgreSQL and MySQL.
- **Admin Dashboard**: Comprehensive admin dashboard for managing users, links, and settings.
- **OIDC**: Supports OpenID Connect (OIDC) for secure and flexible authentication.
- **Open Source**: LinkShelf is open source, allowing users to contribute to its development and customize it as needed.

## Contributing and Development

See [CONTRIBUTING.md](CONTRIBUTING.md)

### Todo:

- [ ] add validations
    - [ ] does user exist on shelfCreate/Update
    - [ ] does shelf exist on sectionCreate/Update
    - [ ] does sectiotion exist on linkCreate/Update
    - [ ] are all required fields given
        - [ ] ShelfPath, ShelfName, ...
    - [ ] handle default values like for themes
- [ ] Add No-Index Option for Search Engine like Google, ... -> Not crawlable


---

### Plain Text Notes

- add link-site (non-app) the public viewable one
- Add backend validation option for the api
- add the stats on the /app root path in the UI
- delete package-lock in the root
- delete .DS_store from the root and add these kind of file in the .gitignore and .containerignore
- add branding directory in the root with the different logos and credits for the SVG-Repo autor for my link → also check the license which is in the SVG-Repo for this asset: https://www.svgrepo.com/svg/525982/link-circle
- rename docker-compose... to compose...
- add the auth layer → don't use the mocked User-UUID in the frontend
- add OIDC/OAuth2.0 providers/functionallity → map them to local users
- add icon picker for new shelf/link/section
- missing validation for Link-Titel on submit
- test OIDC login
- delete Login providers (google, github) from multiselect in ui settings page → or check if they work
- increase test-coverage
- check if there are tests against a postgres and mysql database in the pipeline
- fix the latest vunerabilites → can not be patched since there is no one
- the image is about 225MB shrink it down to a much smaller image
- if i save the settings for each element a single request is made → group them as one request and create a new endpoint which can handle many settings update at once. also add tests (unit and integration) for this.

---

- add SMTP to send emails to verify the users email-address
- add option to turn off the user registration for example for self-hosted instances (admin should still be able to create a user which then can finisht the registration via the email-link)

- add functionallity to order links in a section and sections in a shelf
- add themeing functioanllity (color, icon, theme, custom css, ...)

- add frontend test incl. converage for readme

- add dependabot automerge for dependency-updates once a week if all pipelines pass
- rename .dockerignore into .containerignore → and check if this works → this forces me to use buildah since docker does not accept non-.dockeringore files, ..

---

- create a collage of different pages in the shelf-app which can be presented as one image in the front starter (for these screenshots i need better test-data in my linkshelf instance)
- update the docmentation in the UI
- create a discord channel or delete discord stuff from github/ui
- add paging (is this needed? It would only be for shelves)


- fix branding → there are color schemas in the root README.md file → logo should be aligned with the branding colors which should NOT be the default Nuxt colors
- think about the .badges directory which creates different sub-directories for each branch
- marketing: fully european software and hosted in europe
- post on reddit, and producthunt, ...
