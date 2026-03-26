# ☸️ pod-lens

[![Tests](https://github.com/williampsena/pod-lens/actions/workflows/tests.yaml/badge.svg)](https://github.com/williampsena/pod-lens/actions/workflows/tests.yaml)
[![Gitleaks](https://github.com/williampsena/pod-lens/actions/workflows/gitleaks.yaml/badge.svg)](https://github.com/williampsena/pod-lens/actions/workflows/gitleaks.yaml)
[![Trivy Security Scan](https://github.com/williampsena/pod-lens/actions/workflows/trivy.yaml/badge.svg)](https://github.com/williampsena/pod-lens/actions/workflows/trivy.yaml)

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

# Run (default port 80)
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

# Run container via make
make docker-run              # light theme
make docker-run theme=dark   # dark theme

# Run container directly with docker run
docker run --rm -p 3000:80 willsenabr/pod-lens:light
docker run --rm -p 3000:80 willsenabr/pod-lens:dark
docker run --rm -p 3000:80 willsenabr/pod-lens:latest
# Access at http://localhost:3000
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
| `PORT` | `80` | Server port |
| `THEME` | `light` | UI theme (light/dark) |
| `POD_LABELS` | - | Pod labels as `key=value,key=value` (comma-separated format) |
| `POD_LABELS_FILE` | - | Path to file with pod labels from Kubernetes downwardAPI (format: `key="value"`, one per line) |
| `DISABLE_MASKING` | `false` | Set to `true` to disable sensitive data masking |

Example:
```bash
PORT=3000 THEME=dark POD_LABELS="app=myapp,version=1.0" make run

# Disable masking for development
DISABLE_MASKING=true make run
```

**Note:** When both `POD_LABELS_FILE` and `POD_LABELS` are set, the file takes precedence.

### Kubernetes Deployment

To automatically inject pod labels into the application, use `fieldRef` in your deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pod-lens
spec:
  template:
    spec:
      containers:
      - name: pod-lens
        image: willsenabr/pod-lens:latest
        ports:
        - containerPort: 80
        env:
        - name: PORT
          value: "80"
        - name: THEME
          value: "dark"
        - name: POD_LABELS
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels
```

This automatically reads the pod's labels and displays them in the UI without manual configuration.

#### Using `POD_LABELS_FILE` with ConfigMap (Recommended)

For more flexibility or when using a ConfigMap to manage labels, use `POD_LABELS_FILE`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pod-labels-config
data:
  labels.txt: |
    app.kubernetes.io/name="pod-lens"
    app.kubernetes.io/instance="prod"
    Environment="production"
    team="platform"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pod-lens
  labels:
    app.kubernetes.io/name: pod-lens
spec:
  template:
    spec:
      containers:
      - name: pod-lens
        image: willsenabr/pod-lens:latest
        ports:
        - containerPort: 80
        env:
        - name: PORT
          value: "80"
        - name: THEME
          value: "dark"
        - name: POD_LABELS_FILE
          value: /etc/pod-labels/labels.txt
        volumeMounts:
        - name: labels-volume
          mountPath: /etc/pod-labels
      volumes:
      - name: labels-volume
        configMap:
          name: pod-labels-config
```

The label file format is `key="value"` with one label per line. See [examples/k8s-deployment-with-pod-labels-file.yaml](examples/k8s-deployment-with-pod-labels-file.yaml) for a complete working example.

### 🔒 Security & Data Masking

Pod-lens automatically masks sensitive data to prevent exposure of credentials and tokens. Masking is applied to:

**Headers:**
- Authorization, Cookie, Set-Cookie
- X-Api-Key, X-Auth-Token, X-Access-Token
- X-Refresh-Token, X-Csrf-Token
- And many other standard sensitive headers

**Environment Variables & Labels:**
- Variables/labels containing: PASSWORD, TOKEN, SECRET, APIKEY, CREDENTIAL, KEY, PRIVATE, PASSWD
- Matching is case-insensitive and works with underscores and dashes

**Smart Masking Display:**
- Values ≤4 chars: `***`
- Values 5-20 chars: Show 2 first + 2 last chars (e.g., `sk***ef`)
- Values 21-50 chars: Show 4 first + 4 last chars (e.g., `1234...7890`)
- Values 51+ chars: Show 6 first + 6 last chars (e.g., `eyJhbG...pXVCJ9`)

**Disable Masking (Development Only):**
```bash
# Show unmasked values for debugging
DISABLE_MASKING=true make run
```

⚠️ **Warning:** Never use `DISABLE_MASKING=true` in production!

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
make leaks             # Scan for secrets (last commit, via Docker)
make leaks-history     # Scan entire repo for secrets (via Docker)

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
docker run -p 8080:80 willsenabr/pod-lens:light
docker run -p 8080:80 willsenabr/pod-lens:dark
docker run -p 8080:80 willsenabr/pod-lens:latest
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

To run locally via Docker:
```bash
# Scan last commit only
make leaks

# Scan entire repository history
make leaks-history
```

Both commands generate a JSON report at `gitleaks_report.json`.

## 📊 Test Coverage

Run `make test-coverage` for detailed report.

## 📄 License

[MIT](./LICENSE.md)
