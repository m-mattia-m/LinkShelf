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
6. **Commit Your Changes**: Commit your changes with a descriptive commit message:
   ```bash
   git commit -m "Add feature X"
   ```
7. **Push to Your Fork**: Push your changes to your forked repository:
   ```bash
   git push origin my-feature-branch
   ```
8. **Create a Pull Request**: Go to the original repository and create a pull request from your forked repository.


## Development Setup

```bash
# start the whole stack (DB + app) in detached mode
docker compose -f docker-compose.yaml -f docker-compose.dev.yaml up -d
# if you only want to start the DBs (for development):
docker compose up -d

# if you want to build a new image:
docker build . --file ./.docker/Containerfile --tag linkshelf --label linkshelf --build-arg IMAGE_NAME=linkshelf --build-arg IMAGE_TAG=linkshelf
```