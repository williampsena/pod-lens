.PHONY: help install-deps build run stop test test-ci test-match test-coverage fmt lint clean install-local docker-build docker-build-dark docker-build-light docker-run docker-stop docker-push docker-compose-up docker-compose-down leaks

# Variables
BINARY_NAME=pod-lens
BIN_DIR=bin
GO=go
DESTDIR=$(HOME)/.local/bin
ENV_FILE=.env
PID_FILE=.$(BINARY_NAME).pid
DOCKER_USERNAME=willsenabr
DOCKER_IMAGE=$(DOCKER_USERNAME)/$(BINARY_NAME)

# Help target
help:
	@echo "Available targets:"
	@echo ""
	@echo "  make install-deps       Install global dependencies (gotestfmt)"
	@echo "  make build              Build the application"
	@echo "  make run                Run the application (use: make run opts=\"--flag=value\")"
	@echo "  make stop               Stop the running application"
	@echo "  make test               Run all tests with formatter"
	@echo "  make test-ci            Run tests for CI environment"
	@echo "  make test-match         Run specific tests (use: make test-match case=\"TestName\")"
	@echo "  make test-coverage      Run tests with coverage report"
	@echo "  make fmt                Format code"
	@echo "  make lint               Run golangci-lint (if available)"
	@echo "  make leaks              Scan for secrets (last commit only, via Docker)"
	@echo "  make leaks-history      Scan entire repository for secrets (via Docker)"
	@echo "  make clean              Clean build artifacts"
	@echo "  make install-local      Build and install locally to $(DESTDIR)/$(BINARY_NAME)"
	@echo ""
	@echo "  Docker targets:"
	@echo "  make docker-build       Build Docker image (light theme as default/latest)"
	@echo "  make docker-build-light Build Docker image with light theme (:light tag)"
	@echo "  make docker-build-dark  Build Docker image with dark theme (:dark tag)"
	@echo "  make docker-run         Run Docker container (use: make docker-run theme=dark/light, default: light)"
	@echo "  make docker-stop        Stop running Docker container"
	@echo "  make docker-push        Build and push all themed images (:dark, :light, :latest)"
	@echo "  make docker-compose-up  Start services with docker-compose"
	@echo "  make docker-compose-down Stop docker-compose services"
	@echo ""

# Install global dependencies
install-deps:
	@echo "Installing global dependencies..."
	@command -v gotestfmt > /dev/null 2>&1 || { \
		echo "Installing gotestfmt..."; \
		$(GO) install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest; \
	}
	@echo "✓ Dependencies installed successfully"

# Build the binary
build: 
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/$(BINARY_NAME) .
	@echo "✓ Build complete: $(BIN_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@if [ -f $(PID_FILE) ]; then \
		echo "🛑 Stopping previous instance..."; \
		PID=$$(cat $(PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID 2>/dev/null || true; \
			sleep 1; \
		fi; \
		rm -f $(PID_FILE); \
	fi
	@if [ -f $(ENV_FILE) ]; then \
		echo "Loading $(ENV_FILE)..."; \
		set -a && source $(ENV_FILE) && set +a; \
	else \
		echo "Warning: $(ENV_FILE) not found"; \
	fi
	@echo "🚀 Starting $(BINARY_NAME)..."
	@set -a && source $(ENV_FILE) 2>/dev/null || true && set +a && $(GO) run . $(opts) & \
	echo $$! > $(PID_FILE); \
	echo "✓ Server running with PID: $$(cat $(PID_FILE))"; \
	wait $$(cat $(PID_FILE)) 2>/dev/null; \
	rm -f $(PID_FILE)

# Stop the application
stop:
	@if [ -f $(PID_FILE) ]; then \
		PID=$$(cat $(PID_FILE)); \
		if kill -0 $$PID 2>/dev/null; then \
			echo "🛑 Stopping $(BINARY_NAME) (PID: $$PID)..."; \
			kill $$PID 2>/dev/null || true; \
			sleep 1; \
			if kill -0 $$PID 2>/dev/null; then \
				echo "Force killing process..."; \
				kill -9 $$PID 2>/dev/null || true; \
			fi; \
			rm -f $(PID_FILE); \
			echo "✓ Server stopped"; \
		else \
			echo "⚠ Process not running, cleaning up PID file..."; \
			rm -f $(PID_FILE); \
		fi; \
	else \
		echo "ℹ No running instance found"; \
	fi

# Run all tests with formatter
test: install-deps
	@if [ -f $(ENV_FILE) ]; then \
		echo "Loading $(ENV_FILE)..."; \
		set -a && source $(ENV_FILE) && set +a; \
	else \
		echo "Warning: $(ENV_FILE) not found"; \
	fi
	@echo "Running tests..."
	@set -a && source $(ENV_FILE) 2>/dev/null || true && set +a && GO_ENV=test $(GO) test -json -skip /pkg/test -v ./... $(args) 2>&1 | gotestfmt

# Run specific test
test-match: install-deps
	@if [ -z "$(case)" ]; then \
		echo "Error: Please provide test case name"; \
		echo "Usage: make test-match case=\"TestName\""; \
		exit 1; \
	fi
	@if [ -f $(ENV_FILE) ]; then \
		echo "Loading $(ENV_FILE)..."; \
		set -a && source $(ENV_FILE) && set +a; \
	else \
		echo "Warning: $(ENV_FILE) not found"; \
	fi
	@echo "Running test: $(case)"
	@$(MAKE) test args="-run $(case)"

# Run tests for CI
test-ci: install-deps
	@if [ -f $(ENV_FILE) ]; then \
		echo "Loading $(ENV_FILE)..."; \
		set -a && source $(ENV_FILE) && set +a; \
	else \
		echo "Warning: $(ENV_FILE) not found"; \
	fi
	@echo "Running tests for CI..."
	@set -a && source $(ENV_FILE) 2>/dev/null || true && set +a && $(GO) test -json -skip /pkg/test -v ./... 2>&1 | gotestfmt

# Run tests with coverage
test-coverage: install-deps
	@if [ -f $(ENV_FILE) ]; then \
		echo "Loading $(ENV_FILE)..."; \
		set -a && source $(ENV_FILE) && set +a; \
	else \
		echo "Warning: $(ENV_FILE) not found"; \
	fi
	@echo "Running tests with coverage..."
	@set -a && source $(ENV_FILE) 2>/dev/null || true && set +a && $(GO) test -json -skip /pkg/test -v -cover ./... 2>&1 | gotestfmt
	@set -a && source $(ENV_FILE) 2>/dev/null || true && set +a && $(GO) test -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	@$(GO) tool cover -func=coverage.out | tail -1

# Format code
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "✓ Code formatted"

# Run linter (optional)
lint:
	@command -v golangci-lint > /dev/null 2>&1 && { \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	} || { \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	}

# Scan for secrets with Gitleaks
precommit: leaks

# Scan for secrets - last commit only
leaks:
	@echo "🔍 Scanning for secrets (last commit) with Gitleaks..."
	docker run --rm \
		-v $(PWD):/repo \
		zricethezav/gitleaks:latest \
		detect --log-opts="HEAD^..HEAD" --source /repo --config /repo/.gitleaks.toml --report-format json --report-path /repo/gitleaks_report.json
	@echo "✓ Scan complete. Report saved to gitleaks_report.json"

# Scan for secrets - entire repository history
leaks-history:
	@echo "🔍 Scanning entire repository history for secrets..."
	docker run --rm \
		-v $(PWD):/repo \
		zricethezav/gitleaks:latest \
		detect --source /repo --config /repo/.gitleaks.toml --report-format json --report-path /repo/gitleaks_report.json
	@echo "✓ Scan complete. Report saved to gitleaks_report.json"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GO) clean
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out
	@rm -f $(PID_FILE)
	@echo "✓ Clean complete"

# Install locally to user bin directory
install-local: build
	@echo "Installing $(BINARY_NAME) to $(DESTDIR)/..."
	@mkdir -p $(DESTDIR)
	@cp $(BIN_DIR)/$(BINARY_NAME) $(DESTDIR)/$(BINARY_NAME)
	@chmod +x $(DESTDIR)/$(BINARY_NAME)
	@echo "✓ Installation complete: $(DESTDIR)/$(BINARY_NAME)"
	@echo ""
	@echo "Make sure $(DESTDIR) is in your PATH:"
	@echo "  export PATH=\"$(DESTDIR):\$$PATH\""

# Docker targets
docker-build: docker-build-light
	@echo "✓ Default image built: $(DOCKER_IMAGE):latest"

docker-build-light:
	@echo "🐳 Building Docker image with light theme..."
	@docker build --build-arg THEME=light -t $(DOCKER_IMAGE):light -f Containerfile .
	@docker tag $(DOCKER_IMAGE):light $(DOCKER_IMAGE):latest
	@echo "✓ Docker images built: $(DOCKER_IMAGE):light, $(DOCKER_IMAGE):latest"

docker-build-dark:
	@echo "🐳 Building Docker image with dark theme..."
	@docker build --build-arg THEME=dark -t $(DOCKER_IMAGE):dark -f Containerfile .
	@echo "✓ Docker image built: $(DOCKER_IMAGE):dark"

docker-run:
	@THEME=$${theme:-light}; \
	echo "🐳 Starting Docker container with $$THEME theme..."; \
	docker run -it --rm \
		-p 8080:8080 \
		-e THEME=$$THEME \
		-e PORT=8080 \
		--name $(BINARY_NAME) \
		$(DOCKER_IMAGE):$$THEME

docker-stop:
	@echo "🛑 Stopping Docker container..."
	@docker stop $(BINARY_NAME) 2>/dev/null || echo "Container not running"
	@echo "✓ Container stopped"

docker-compose-up:
	@echo "🐳 Starting with docker-compose..."
	@docker-compose up -d
	@echo "✓ Services started. Access at http://localhost:8080"

docker-compose-down:
	@echo "🛑 Stopping docker-compose services..."
	@docker-compose down
	@echo "✓ Services stopped"

docker-push: docker-build-dark docker-build-light
	@echo "📤 Pushing Docker images to $(DOCKER_IMAGE)..."
	@docker push $(DOCKER_IMAGE):dark
	@docker push $(DOCKER_IMAGE):light
	@docker push $(DOCKER_IMAGE):latest
	@echo "✓ All images pushed: $(DOCKER_IMAGE):dark, $(DOCKER_IMAGE):light, $(DOCKER_IMAGE):latest"