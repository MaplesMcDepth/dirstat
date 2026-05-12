# Contributing

Thank you for your interest in contributing!

## Development Setup

```bash
git clone https://github.com/MaplesMcDepth/'$repo'.git
cd '$repo'
go build ./...
```

## Running Tests

```bash
go test ./...
```

## Code Style

- Follow standard Go conventions (`gofmt`)
- Keep functions small and focused
- Add comments for exported functions
- Update README if behavior changes

## Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Ensure `go vet` passes
6. Submit a pull request

## Reporting Issues

Please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
