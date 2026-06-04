# Contributing to OpenSustain

Thank you for your interest in contributing to OpenSustain! This document provides guidelines and instructions for contributing to the project.

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Git
- Docker (optional, for containerized builds)

### Setting Up the Development Environment

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/OpenSustain.git
   cd OpenSustain
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build -o OpenSustain ./cmd/OpenSustain
   ```

## Development Workflow

### Making Changes

1. Create a new branch for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bugfix-name
   ```

2. Make your changes following the project's coding standards

3. Write tests for your changes:
   ```bash
   go test ./...
   ```

4. Ensure all tests pass:
   ```bash
   go test -v ./...
   ```

5. Build the project to ensure it compiles:
   ```bash
   go build -o OpenSustain ./cmd/OpenSustain
   ```

### Commit Guidelines

- Use clear, descriptive commit messages
- Follow conventional commit format:
  - `feat:` for new features
  - `fix:` for bug fixes
  - `docs:` for documentation changes
  - `test:` for test changes
  - `refactor:` for code refactoring
  - `chore:` for maintenance tasks

Example:
```
feat: add support for custom scoring thresholds
```

### Pull Request Process

1. Push your changes to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a pull request against the `main` branch

3. Fill out the PR template with:
   - Description of changes
   - Related issues
   - Testing performed
   - Screenshots (if applicable)

4. Wait for code review and address any feedback

5. Once approved, maintainers will merge your PR

## Code Style

- Follow Go standard formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and small
- Write tests for new functionality

## Testing

- Write unit tests for all new functions
- Ensure test coverage is maintained or improved
- Run tests before committing:
  ```bash
  go test ./...
  ```

## Reporting Issues

If you find a bug or have a feature request:

1. Check existing issues to avoid duplicates
2. Create a new issue with:
   - Clear title and description
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Environment details (OS, Go version)
   - Relevant logs or screenshots

## Questions

For questions about usage or development:
- Open a GitHub Discussion
- Check existing documentation
- Review existing issues and discussions

## License

By contributing to OpenSustain, you agree that your contributions will be licensed under the MIT License.
