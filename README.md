# ☸️ pod-lens

Lightweight pod information viewer for Kubernetes, similar to [traefik/whoami](https://github.com/traefik/whoami).

## ✨ Features

- 📦 Display pod information (hostname, IPs, Go version, OS, architecture)
- 🏷️ Show Kubernetes labels
- 📝 Display request headers (with sensitive data masking)
- 🔧 Show safe environment variables
- 🌓 Light & dark theme support
- 🐳 Docker multi-tag support (:light, :dark, :latest)
- 🔒 Automatic sensitive header masking (Authorization, Cookie, X-Api-Key, etc)
- ⚡ Graceful shutdown (Ctrl+C or press 'c')

## 🚀 Quick Start

### Local

```bash
# Build
make build

# Run (default port 8080)
make run

# With custom port and theme
PORT=3000 THEME=dark make run
```

### Docker

```bash
# Build with light theme (default)
make docker-build

# Build with dark theme
make docker-build-dark

# Run container
make docker-run              # light theme
make docker-run theme=dark   # dark theme
```

### Docker Compose

```bash
docker-compose up -d
# Access at http://localhost:8080
```

## 🔧 Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `THEME` | `light` | UI theme (light/dark) |
| `POD_LABELS` | - | Pod labels as `key=value,key=value` |

Example:
```bash
PORT=3000 THEME=dark POD_LABELS="app=myapp,version=1.0" make run
```

## 📝 Available Commands

```bash
make help              # Show all available commands

# Building & Running
make build             # Build binary
make run               # Run application
make stop              # Stop running server

# Testing
make test              # Run all tests with formatter
make test-coverage     # Run tests with coverage report
make test-match case="TestName"  # Run specific test

# Code Quality
make fmt               # Format code
make lint              # Run linter (if available)

# Docker
make docker-build      # Build with light theme (default/latest)
make docker-build-light # Build light theme variant
make docker-build-dark  # Build dark theme variant
make docker-run        # Run container interactively
make docker-stop       # Stop container
make docker-push       # Build & push to willsenabr/pod-lens
```

## 🐳 Docker Images

Available on Docker Hub: `willsenabr/pod-lens`

```bash
docker run -p 8080:8080 willsenabr/pod-lens:light
docker run -p 8080:8080 willsenabr/pod-lens:dark
docker run -p 8080:8080 willsenabr/pod-lens:latest
```

## 📦 Project Structure

```
.
├── main.go                 # Entry point
├── internal/
│   ├── server/            # HTTP server & pod info logic
│   └── settings/          # Configuration management
├── pages/
│   └── index.html         # HTML template
├── static/
│   └── styles/            # CSS themes
├── Containerfile          # Docker build spec
├── docker-compose.yml     # Compose config
└── Makefile              # Build automation
```

## 🔐 Security

- ✅ Sensitive headers are automatically masked
- ✅ Safe environment variable filtering
- ✅ Non-root container user
- ✅ Minimal Alpine base image
- ✅ Secret detection with Gitleaks (configured in `.gitleaks.toml`)

### Gitleaks Configuration

This project uses [Gitleaks](https://github.com/gitleaks/gitleaks) to detect and prevent secrets from being committed. The configuration:

- Runs on every commit in GitHub Actions
- Scans repository history for credential patterns
- Ignores test files and known false positives

To run locally:
```bash
gitleaks detect --verbose
```

## 📊 Test Coverage

Run `make test-coverage` for detailed report.

## 📄 License

MIT
