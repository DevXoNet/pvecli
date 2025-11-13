# Variables
APP_NAME := pvecli
# Project version
VERSION := $(strip 0.0.1)
# Git commit hash
BUILD := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
# Linker flags
LD_FLAGS := -w -X main.version=$(VERSION) -X main.build=$(BUILD)

# Default target
all: build

# Help command
help:
	@echo "pvecli Makefile commands:"
	@echo ""
	@echo "  make setup       - Download and install all required Go modules"
	@echo "  make build       - Build local binary"
	@echo "  make build-all   - Cross-compile for all platforms"
	@echo "  make run         - Build and run locally"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make deps        - Update dependencies (go mod tidy)"
	@echo "  make test        - Run tests"
	@echo "  make rebuild     - Full rebuild (clean + deps + build)"
	@echo "  make release     - Package release with checksums"
	@echo "  make help        - Show this help message"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Build:   $(BUILD)"

# Setup - Download and install all required Go modules
setup:
	@echo "==> Installing golang.org/x/crypto/ssh..."
	go get golang.org/x/crypto/ssh@v0.31.0
	@echo "==> Downloading Go modules..."
	go mod download
	@echo "==> Verifying Go modules..."
	go mod verify
	@echo "==> Tidying Go modules..."
	go mod tidy
	@echo "==> Setup complete! All dependencies installed."
	@echo ""
	@echo "Required modules:"
	@go list -m all | grep -v "^pvecli$$"

# Build local binary
build:
	CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o $(APP_NAME) main.go

# Cross-compilation for multiple platforms
build-all:
	mkdir -p _build

	@echo "==> Building for darwin/amd64"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-darwin-amd64"

	@echo "==> Building for darwin/arm64"
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-darwin-arm64"

	@echo "==> Building for linux/amd64"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-linux-amd64"

	@echo "==> Building for linux/arm"
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-linux-arm"

	@echo "==> Building for linux/arm64"
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-linux-arm64"

	@echo "==> Building for freebsd/amd64"
	GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-freebsd-amd64"

	@echo "==> Building for freebsd/arm64"
	GOOS=freebsd GOARCH=arm64 CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-freebsd-arm64"

	@echo "==> Building for linux/ppc64le"
	GOOS=linux GOARCH=ppc64le CGO_ENABLED=0 go build -tags release -ldflags "$(LD_FLAGS)" -o "_build/$(APP_NAME)-$(VERSION)-linux-ppc64le"

	@echo "==> Checksumming..."
	cd _build && sha256sum * > sha256sums.txt

	@echo "==> Done"

# Run the local binary
run:
	rm -f $(APP_NAME)
	go build -ldflags "$(LD_FLAGS)" -o $(APP_NAME)
	./$(APP_NAME)

# Clean outputs
clean:
	rm -f $(APP_NAME)
	rm -rf _build/
	rm -rf release/

# Install/update dependencies
deps:
	go mod tidy

# Run tests
test:
	go test ./...

# Full rebuild
rebuild: clean deps build

# Package and verify release artifacts
release:
	mkdir -p release
	cp _build/* release/
	cd release && sha256sum --quiet --check sha256sums.txt
	cd release && gh release create v$(VERSION) -d -t v$(VERSION) *
	# cd release && glab release create v$(VERSION) * --notes "Release v$(VERSION). "
	# cd release && glab release create v$(VERSION) * --notes-file ../RELEASE_NOTES.md

.PHONY: all setup build build-all run clean deps test rebuild release help
