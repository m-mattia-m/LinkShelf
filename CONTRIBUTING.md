# Contributing

1. Fork the repository and clone your fork:
   ```bash
   git clone https://github.com/<your-username>/LinkShelf
   ```
2. Create a branch prefixed with `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `style/` or `test/`. CI only runs on
   these prefixes, so other branch names never get the required checks:
   ```bash
   git checkout -b feat/my-feature
   ```
3. Make your changes, and add or update tests.
4. Commit with a [Conventional Commit](https://www.conventionalcommits.org) message (see [Releases](#releases)):
   ```bash
   git commit -m "feat: add feature X"
   ```
5. Push the branch to your fork and open a pull request against `main`.

## License

By contributing, you agree that your contributions are licensed under the Apache License 2.0, which allows them to be
included in LinkShelf under the AGPL-3.0. See [LICENSING.md](LICENSING.md#contributions).

## Releases

Versions and release notes are generated automatically by [semantic-release](https://semantic-release.org) from the
commit messages on `main`. The repository squash-merges, so **the pull request title is the commit message** and must be
a Conventional Commit (checked by the `pr-title` workflow):

| PR title                                                           | Release                                                    |
|--------------------------------------------------------------------|------------------------------------------------------------|
| `fix: ...`, `perf: ...`, `fix(deps): ...`                          | patch (`0.1.0` → `0.1.1`)                                  |
| `feat: ...`                                                        | minor (`0.1.1` → `0.2.0`)                                  |
| `feat!: ...` or a `BREAKING CHANGE:` footer                        | minor while the version is below `1.0.0`, major afterwards |
| `docs:`, `chore:`, `ci:`, `test:`, `refactor:`, `style:`, `build:` | no release                                                 |

A release is cut when a non-Dependabot commit lands on `main`, every Tuesday at 06:00 UTC, or on demand via the
*Release* workflow (*Run workflow*). Dependabot's weekly, grouped dependency updates are merged automatically once CI is
green and are collected into the next release. There is no need to create tags or GitHub releases by hand.

## Development Setup

```bash
# start the whole stack (DB + app) in detached mode
docker compose -f compose.yaml -f compose.dev.yaml up -d
# if you only want to start the DBs (for development):
docker compose up -d

# if you want to build a new image:
buildah build . --file ./.container/Containerfile --tag linkshelf --label linkshelf --build-arg IMAGE_NAME=linkshelf --build-arg IMAGE_TAG=linkshelf
```