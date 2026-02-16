# Contributing to GitOps-Time-Machine

Thank you for your interest in contributing! Here's how you can help.

## Getting Started

1. **Fork** the repository
2. **Clone** your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/GitOps-Time-Machine.git
   cd GitOps-Time-Machine
   ```
3. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Go 1.22 or later
- A Kubernetes cluster (optional, for integration testing)
- `golangci-lint` (optional, for linting)

### Build & Test

```bash
# Build
make build

# Run all tests
make test

# Format code
make fmt

# Lint
make lint
```

## Making Changes

1. Write your code following existing patterns and conventions
2. Add tests for any new functionality
3. Ensure all tests pass: `make test`
4. Format your code: `make fmt`
5. Update documentation if needed

## Pull Request Process

1. Update the README.md if your change affects the user interface
2. Include a clear description of what your PR does and why
3. Reference any related issues
4. Ensure CI passes

## Coding Guidelines

- Follow standard Go conventions and idioms
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Handle errors explicitly (no silent swallowing)

## Reporting Bugs

Open an issue with:
- Clear title and description
- Steps to reproduce
- Expected vs actual behavior
- Go version and OS

## Feature Requests

Open an issue with:
- Clear use case description
- Why the existing functionality doesn't cover it
- Proposed approach (if any)

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
