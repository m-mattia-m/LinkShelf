# LinkShelf

[![CI](https://github.com/m-mattia-m/LinkShelf/actions/workflows/ci.yaml/badge.svg)](https://github.com/m-mattia-m/LinkShelf/actions/workflows/ci.yaml) ![coverage](https://raw.githubusercontent.com/m-mattia-m/LinkShelf/refs/heads/test/add-integration-tests/.badges/test/add-integration-tests/coverage.svg) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m-mattia-m/LinkShelf?filename=backend%2Fgo.mod) ![Release](https://img.shields.io/github/v/release/m-mattia-m/LinkShelf)

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
- [ ] Write unit tests
- [ ] Add Renovate