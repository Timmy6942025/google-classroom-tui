.PHONY: build build-all build-linux build-darwin build-windows test lint clean help install uninstall

# Binary name and version
BINARY_NAME := google-classroom
VERSION := 0.1.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)

# Build for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/google-classroom

# Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "All platforms built successfully"

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 ./cmd/google-classroom
	@echo "Linux binary: $(BINARY_NAME)-linux-amd64"

# Build for macOS
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-amd64 ./cmd/google-classroom
	@echo "macOS binary: $(BINARY_NAME)-darwin-amd64"

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-windows-amd64.exe ./cmd/google-classroom
	@echo "Windows binary: $(BINARY_NAME)-windows-amd64.exe"

# Run tests
test:
	go test ./... -v -cover -coverprofile=coverage.out

# Run linter
lint:
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out
	go clean

# Install binary to system
install:
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@if [ -f $(BINARY_NAME) ]; then \
		sudo cp $(BINARY_NAME) /usr/local/bin/; \
		sudo chmod +x /usr/local/bin/$(BINARY_NAME); \
		echo "Installed successfully!"; \
	else \
		echo "Binary not found. Run 'make build' first."; \
	fi

# Uninstall binary from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from /usr/local/bin..."
	@if [ -f /usr/local/bin/$(BINARY_NAME) ]; then \
		sudo rm /usr/local/bin/$(BINARY_NAME); \
		echo "Uninstalled successfully!"; \
	else \
		echo "Binary not found at /usr/local/bin/$(BINARY_NAME)"; \
	fi

# Show help
help:
	@echo "Google Classroom TUI - Build and Development Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  build        - Build binary for current platform"
	@echo "  build-all    - Build binaries for all platforms"
	@echo "  build-linux  - Build Linux binary"
	@echo "  build-darwin - Build macOS binary"
	@echo "  build-windows- Build Windows binary"
	@echo ""
	@echo "Development Commands:"
	@echo "  test         - Run all tests with coverage"
	@echo "  lint         - Run linter (requires golangci-lint)"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Install binary to /usr/local/bin"
	@echo "  uninstall    - Remove binary from /usr/local/bin"
	@echo ""
	@echo "Other:"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
