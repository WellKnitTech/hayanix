# Contributing to Hayanix

Thank you for your interest in contributing to Hayanix! This document provides guidelines and information for contributors.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your feature or bugfix
4. Make your changes
5. Test your changes
6. Submit a pull request

## Development Setup

### Prerequisites
- Go 1.21 or later
- Make (optional, for using Makefile)

### Building
```bash
# Clone the repository
git clone https://github.com/wellknittech/hayanix.git
cd hayanix

# Install dependencies
make deps

# Build the binary
make build

# Run tests
make test
```

## Code Style

- Follow Go conventions and best practices
- Use `gofmt` to format your code
- Add comments for exported functions and types
- Keep functions small and focused
- Use meaningful variable and function names

## Testing

- Write tests for new functionality
- Ensure all existing tests pass
- Aim for good test coverage
- Use table-driven tests where appropriate

## Pull Request Process

1. Ensure your code follows the project's style guidelines
2. Add tests for any new functionality
3. Update documentation as needed
4. Ensure all tests pass
5. Submit a pull request with a clear description

### Pull Request Guidelines

- Use a clear and descriptive title
- Provide a detailed description of changes
- Reference any related issues
- Include screenshots for UI changes
- Ensure the PR is up to date with the main branch

## Reporting Issues

When reporting issues, please include:

- Operating system and version
- Go version
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any relevant log output

## Adding Sigma Rules

When adding new sigma rules:

1. Follow the standard Sigma rule format
2. Include appropriate tags and metadata
3. Test rules with sample log data
4. Document false positives
5. Place rules in the appropriate directory structure

## License

By contributing to Hayanix, you agree that your contributions will be licensed under the MIT License.
