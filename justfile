# gqlt - GraphQL CLI Tool
# Justfile for common development and release tasks

# Default recipe - show available commands
default:
    @just --list

# Get current version from VERSION file
get-version:
    @cat VERSION

# Set version in VERSION file
set-version version:
    @echo "{{version}}" > VERSION
    @echo "Version set to {{version}}"

# Bump patch version (0.1.0 -> 0.1.1)
bump-patch:
    @current=$(cat VERSION) && \
    major=$(echo $current | cut -d. -f1) && \
    minor=$(echo $current | cut -d. -f2) && \
    patch=$(echo $current | cut -d. -f3) && \
    new_version="$major.$minor.$((patch + 1))" && \
    echo "$new_version" > VERSION && \
    echo "Version bumped to $new_version"

# Bump minor version (0.1.0 -> 0.2.0)
bump-minor:
    @current=$(cat VERSION) && \
    major=$(echo $current | cut -d. -f1) && \
    minor=$(echo $current | cut -d. -f2) && \
    new_version="$major.$((minor + 1)).0" && \
    echo "$new_version" > VERSION && \
    echo "Version bumped to $new_version"

# Bump major version (0.1.0 -> 1.0.0)
bump-major:
    @current=$(cat VERSION) && \
    major=$(echo $current | cut -d. -f1) && \
    new_version="$((major + 1)).0.0" && \
    echo "$new_version" > VERSION && \
    echo "Version bumped to $new_version"

# Build the binary for current platform
build:
    @version=$(cat VERSION) && \
    echo "Building gqlt v$version..." && \
    mkdir -p dist && \
    CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$version" -o dist/gqlt ./cmd && \
    echo "Built gqlt v$version to dist/gqlt"

# Build for all platforms
build-all:
    @version=$(cat VERSION) && \
    echo "Building gqlt v$version for all platforms..." && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$version" -o dist/gqlt-linux-amd64 ./cmd && \
    CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=$version" -o dist/gqlt-linux-arm64 ./cmd && \
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.version=$version" -o dist/gqlt-darwin-amd64 ./cmd && \
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=$version" -o dist/gqlt-darwin-arm64 ./cmd && \
    CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=$version" -o dist/gqlt-windows-amd64.exe ./cmd && \
    echo "Built binaries for all platforms in dist/"

# Create distribution directory
dist-dir:
    @mkdir -p dist

# Build and create archives
package:
    @just dist-dir
    @just build-all
    @version=$(cat VERSION) && \
    cd dist && \
    tar -czf gqlt-$version-linux-amd64.tar.gz gqlt-linux-amd64 && \
    tar -czf gqlt-$version-linux-arm64.tar.gz gqlt-linux-arm64 && \
    tar -czf gqlt-$version-darwin-amd64.tar.gz gqlt-darwin-amd64 && \
    tar -czf gqlt-$version-darwin-arm64.tar.gz gqlt-darwin-arm64 && \
    zip gqlt-$version-windows-amd64.zip gqlt-windows-amd64.exe && \
    echo "Created distribution archives in dist/"

# Run tests
test:
    @echo "Running tests..."
    @go test -v ./...

# Run tests with coverage
test-coverage:
    @echo "Running tests with coverage..."
    @go test -cover ./...

# Run tests with verbose output
test-verbose:
    @echo "Running tests with verbose output..."
    @go test -v ./...

# Run tests for specific package
test-package package:
    @echo "Running tests for {{package}}..."
    @go test ./{{package}}

# Lint the code
lint:
    @echo "Running linter..."
    @go vet ./...
    @golangci-lint run

# Format the code
fmt:
    @echo "Formatting code..."
    @go fmt ./...

# Clean build artifacts
clean:
    @echo "Cleaning build artifacts..."
    @rm -rf dist/
    @rm -rf bin/
    @rm -f gqlt
    @go clean

# Install dependencies
deps:
    @echo "Installing dependencies..."
    @go mod download
    @go mod tidy

# Update dependencies
update-deps:
    @echo "Updating dependencies..."
    @go get -u ./...
    @go mod tidy

# Run the CLI with help
help:
    @./dist/gqlt --help

# Run the CLI with version
version:
    @./dist/gqlt --version

# Install the binary to $GOPATH/bin
install:
    @version=$(cat VERSION) && \
    echo "Installing gqlt v$version..." && \
    go install -ldflags "-X main.version=$version" ./cmd

# Uninstall the binary
uninstall:
    @echo "Uninstalling gqlt..."
    @go clean -i ./cmd

# Check if version is valid semantic version
check-version:
    @version=$(cat VERSION) && \
    if ! echo "$version" | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$$' > /dev/null; then \
        echo "Error: Version '$version' is not a valid semantic version (major.minor.patch)"; \
        exit 1; \
    fi && \
    echo "Version '$version' is valid"


# Show current git status
git-status:
    @echo "Git status:"
    @git status --short

# Show current git tags
git-tags:
    @echo "Git tags:"
    @git tag --sort=-version:refname


# Complete release workflow - does everything properly
complete-release type="patch" notes="":
    @echo "Starting complete release workflow..."
    @echo "Step 1: Bumping version..."
    @just check-version
    @just bump-{{type}}
    @version=$(cat VERSION) && \
    echo "Version bumped to $version"
    @echo "Step 2: Building and packaging..."
    @just package
    @echo "Step 3: Committing VERSION file..."
    @git add VERSION && \
    git commit -m "Bump version to $version" && \
    echo "VERSION file committed"
    @echo "Step 4: Creating git tag..."
    @git tag v$version && \
    echo "Git tag v$version created"
    @echo "Step 5: Pushing tag..."
    @git push origin v$version && \
    echo "Tag pushed to origin"
    @echo "Step 6: Creating GitHub release..."
    @if [ -n "{{notes}}" ]; then \
        echo "{{notes}}" > /tmp/release-notes-$version.md; \
    else \
        just release-notes > /tmp/release-notes-$version.md; \
    fi && \
    gh release create v$version \
        --notes-file /tmp/release-notes-$version.md \
        dist/gqlt-$version-*.tar.gz \
        dist/gqlt-$version-*.zip && \
    rm /tmp/release-notes-$version.md && \
    echo "GitHub release v$version created successfully!"
    @echo "Complete release v$version finished!"

# Show release notes template
release-notes:
    @version=$(cat VERSION) && \
    echo "Release Notes Template for v$version:" && \
    echo "========================================" && \
    echo "" && \
    echo "## What's New" && \
    echo "- " && \
    echo "" && \
    echo "## Bug Fixes" && \
    echo "- " && \
    echo "" && \
    echo "## Breaking Changes" && \
    echo "- " && \
    echo "" && \
    echo "## Installation" && \
    echo "Download the appropriate binary for your platform:" && \
    echo "- Linux (amd64): gqlt-$version-linux-amd64.tar.gz" && \
    echo "- Linux (arm64): gqlt-$version-linux-arm64.tar.gz" && \
    echo "- macOS (amd64): gqlt-$version-darwin-amd64.tar.gz" && \
    echo "- macOS (arm64): gqlt-$version-darwin-arm64.tar.gz" && \
    echo "- Windows (amd64): gqlt-$version-windows-amd64.zip"

# Generate comprehensive README.md
readme:
	@just build
	@./generate_readme.sh
	@echo "Comprehensive README.md generated!"

# Generate mock server GraphQL code
generate-mockserver:
	@echo "Generating mock server GraphQL code..."
	@cd internal/mockserver && go run github.com/99designs/gqlgen@latest generate --config gqlgen.yml --verbose
	@echo "Mock server code generated!"

# Show project info
info:
    @echo "gqlt - GraphQL CLI Tool"
    @echo "======================"
    @echo "Version: $(cat VERSION)"
    @echo "Go version: $(go version)"
    @echo "Git commit: $(git rev-parse --short HEAD)"
    @echo "Git branch: $(git branch --show-current)"
    @echo "Build date: $(date)"

# Development setup
dev-setup:
    @echo "Setting up development environment..."
    @just deps
    @just fmt
    @just lint
    @just test
    @echo "Development environment ready!"

# Quick development build and test
dev:
    @just build
    @just test
    @echo "Development build complete!"

# Show help for this justfile
help-just:
    @echo "gqlt Justfile Commands:"
    @echo "======================"
    @echo ""
    @echo "Version Management:"
    @echo "  get-version          - Show current version"
    @echo "  set-version VERSION  - Set version to VERSION"
    @echo "  bump-patch           - Bump patch version (0.1.0 -> 0.1.1)"
    @echo "  bump-minor           - Bump minor version (0.1.0 -> 0.2.0)"
    @echo "  bump-major           - Bump major version (0.1.0 -> 1.0.0)"
    @echo "  check-version        - Validate current version format"
    @echo ""
    @echo "Building:"
    @echo "  build                - Build for current platform"
    @echo "  build-all            - Build for all platforms"
    @echo "  package              - Build and create distribution archives"
    @echo "  install              - Install to $GOPATH/bin"
    @echo ""
    @echo "Testing:"
    @echo "  test                 - Run all tests"
    @echo "  test-coverage        - Run tests with coverage"
    @echo "  test-verbose         - Run tests with verbose output"
    @echo "  test-package PACKAGE - Run tests for specific package"
    @echo ""
    @echo "Code Quality:"
    @echo "  lint                 - Run linter"
    @echo "  fmt                  - Format code"
    @echo "  clean                - Clean build artifacts"
    @echo ""
    @echo "Releases:"
    @echo "  complete-release TYPE [NOTES] - Complete release workflow (bump, build, commit, tag, push, GitHub release)"
    @echo ""
    @echo "Git:"
    @echo "  git-status           - Show git status"
    @echo "  git-tags             - Show git tags"
    @echo "  git-tag              - Create git tag for current version"
    @echo "  git-push-tag         - Push git tag"
    @echo ""
    @echo "Development:"
    @echo "  dev-setup            - Set up development environment"
    @echo "  dev                  - Quick development build and test"
    @echo "  info                 - Show project information"
    @echo "  release-notes        - Show release notes template"
    @echo ""
    @echo "Documentation:"
    @echo "  readme               - Generate comprehensive README.md from all commands"
    @echo ""
    @echo "Note: Use 'gqlt docs' command directly for other documentation formats:"
    @echo "  gqlt docs --format md --tree --output docs/  - Generate markdown tree"
    @echo "  gqlt docs --format man --tree --output man/  - Generate man pages"
