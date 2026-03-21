# Goober Build Makefile

# Binary name
BINARY_NAME=gbr

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all build build-linux clean help test deps

# Build for all platforms
build: build-linux-amd64 build-linux-arm64 build-freebsd-amd64 build-freebsd-arm64

# Build for Linux amd64
build-linux-amd64:
	@echo "Building for linux/amd64..."
	@mkdir -p dist/linux-amd64
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "-X github.com/dknr/goober/cmd/version.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/linux-amd64/$(BINARY_NAME) .

# Build for Linux arm64
build-linux-arm64:
	@echo "Building for linux/arm64..."
	@mkdir -p dist/linux-arm64
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags "-X github.com/dknr/goober/cmd/version.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/linux-arm64/$(BINARY_NAME) .

# Build for FreeBSD amd64
build-freebsd-amd64:
	@echo "Building for freebsd/amd64..."
	@mkdir -p dist/freebsd-amd64
	GOOS=freebsd GOARCH=amd64 $(GOBUILD) -ldflags "-X github.com/dknr/goober/cmd/version.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/freebsd-amd64/$(BINARY_NAME) .

# Build for FreeBSD arm64
build-freebsd-arm64:
	@echo "Building for freebsd/arm64..."
	@mkdir -p dist/freebsd-arm64
	GOOS=freebsd GOARCH=arm64 $(GOBUILD) -ldflags "-X github.com/dknr/goober/cmd/version.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/freebsd-arm64/$(BINARY_NAME) .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf dist/
	@rm -f $(BINARY_NAME)

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Show help
help:
	@echo "Goober Build Targets:"
	@echo "  all           - Build for Linux (amd64, arm64)"
	@echo "  build-linux-amd64   - Build for Linux amd64"
	@echo "  build-linux-arm64   - Build for Linux arm64"
	@echo "  clean         - Remove dist/ folder and local binary"
	@echo "  test          - Run tests"
	@echo "  deps          - Download dependencies"
	@echo ""
	@echo "Note: FreeBSD builds require cross-compilation toolchain"
	@echo "      Build for FreeBSD on a FreeBSD system directly"