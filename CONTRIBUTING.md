# Contributing

We welcome contributions from the community! Whether you're fixing a bug, adding a new feature, or improving
documentation, your help is appreciated. Please follow the guidelines below to ensure a smooth contribution process.

## How to Contribute

1. **Fork the Repository**: Start by forking the repository to your own GitHub account.
2. **Clone the Forked Repository**: Clone your forked repository to your local machine using:
   ```bash
   git clone https://github.com/m-mattia-m/LinkShelf
   ```
3. **Create a New Branch**: Create a new branch for your feature or bug fix:
   ```bash
   git checkout -b my-feature-branch
   ```
4. **Make Your Changes**: Implement your changes in the codebase. Ensure your code follows the project's coding standards.
5. **Test Your Changes**: Run existing tests and add new tests if necessary to verify your changes work as expected.
6. **Commit Your Changes**: Commit your changes with a descriptive [Conventional Commit](https://www.conventionalcommits.org) message. Releases are generated from these messages (see [Releases](#releases)):
   ```bash
   git commit -m "feat: add feature X"
   ```
7. **Push to Your Fork**: Push your changes to your forked repository:
   ```bash
   git push origin my-feature-branch
   ```
8. **Create a Pull Request**: Go to the original repository and create a pull request from your forked repository.


## Releases

Versions and release notes are generated automatically by [semantic-release](https://semantic-release.org) from the commit messages on `main`. The repository squash-merges, so **the pull request title is the commit message** and must be a Conventional Commit (checked by the `pr-title` workflow):

| PR title | Release |
| --- | --- |
| `fix: ...`, `perf: ...`, `fix(deps): ...` | patch (`0.1.0` → `0.1.1`) |
| `feat: ...` | minor (`0.1.1` → `0.2.0`) |
| `feat!: ...` or a `BREAKING CHANGE:` footer | minor while the version is below `1.0.0`, major afterwards |
| `docs:`, `chore:`, `ci:`, `test:`, `refactor:`, `style:`, `build:` | no release |

A release is cut when a non-Dependabot commit lands on `main`, every Tuesday at 06:00 UTC, or on demand via the *Release* workflow (*Run workflow*). Dependabot's weekly, grouped dependency updates are merged automatically once CI is green and are collected into the next release. There is no need to create tags or GitHub releases by hand.

## Development Setup

```bash
# start the whole stack (DB + app) in detached mode
docker compose -f compose.yaml -f compose.dev.yaml up -d
# if you only want to start the DBs (for development):
docker compose up -d

# if you want to build a new image:
buildah build . --file ./.container/Containerfile --tag linkshelf --label linkshelf --build-arg IMAGE_NAME=linkshelf --build-arg IMAGE_TAG=linkshelf
```