# Git information

GITCOMMIT := env('GITCOMMIT', `git rev-parse HEAD 2> /dev/null || echo "unknown"`)
GITDATE := env('GITDATE', `git show -s --format='%ct' 2> /dev/null || echo "0"`)

# Version detection

VERSION := env('VERSION', "untagged")

# Build variables

GOOS := env('GOOS', `go env GOOS`)
GOARCH := env('GOARCH', `go env GOARCH`)
TARGETOS := env('TARGETOS', GOOS)
TARGETARCH := env('TARGETARCH', GOARCH)

# Build ldflags

LDFLAGS := "-X main.GitCommit=" + GITCOMMIT + " -X main.GitDate=" + GITDATE + " -X main.Version=" + VERSION

# Output binary path

BINARY := "./bin/op-conductor-init"
DOCKER_IMAGE := env('DOCKER_IMAGE', 'op-conductor-init:latest')

# Default recipe - build binary
default: build

# Build op-conductor-init binary
build:
    mkdir -p ./bin
    env GO111MODULE=on GOOS={{ TARGETOS }} GOARCH={{ TARGETARCH }} CGO_ENABLED=0 \
        go build -v -ldflags="{{ LDFLAGS }}" -o {{ BINARY }} ./cmd/op-conductor-init

# Clean build artifacts
clean:
    rm -rf ./bin
    go clean

# Run tests
test:
    go test -v ./...

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    revive -config .revive.toml ./...

# Base docker bake arguments

BAKE_ARGS := '--set "*.args.GIT_COMMIT=' + GITCOMMIT + '" --set "*.args.GIT_DATE=' + GITDATE + '"'

# Internal helper for docker buildx bake commands
_docker-bake TARGET="" VERSION_SUFFIX="" EXTRA_ARGS="" ACTION="":
    docker buildx bake -f build.hcl {{ TARGET }} \
        {{ BAKE_ARGS }} \
        --set "*.args.VERSION={{ VERSION }}{{ VERSION_SUFFIX }}" \
        {{ EXTRA_ARGS }} \
        {{ ACTION }}

# Build Docker image using buildx bake
docker-build:
    @just _docker-bake "" "" "" "--load"

# Build Docker image for development
docker-build-dev:
    @just _docker-bake "dev" "-dev" "" "--load"

# Build and push Docker image for release
docker-push:
    @just _docker-bake "release" "" "" "--push"

# Build multi-platform image (without pushing)
docker-build-multiplatform:
    @just _docker-bake "" "" '--set "*.platform=linux/amd64,linux/arm64"' ""

# Build with custom registry and repository
docker-build-custom REGISTRY REPOSITORY TAG="latest":
    @just _docker-bake "" "" '--set "REGISTRY={{ REGISTRY }}" --set "REPOSITORY={{ REPOSITORY }}" --set "TAG={{ TAG }}"' "--load"

# Show what would be built
docker-bake-print:
    docker buildx bake -f build.hcl --print

# Install binary locally
install: build
    cp {{ BINARY }} $(go env GOPATH)/bin/

# Generate example state for testing
example: build
    {{ BINARY }} generate \
        --nodes=sequencer-1:50050,sequencer-2:50050,sequencer-3:50050 \
        --server-ids=sequencer-1,sequencer-2,sequencer-3 \
        --initial-leader=sequencer-1 \
        --output-dir=./example-state

# Verify example state
verify-example: build
    {{ BINARY }} verify --state-dir=./example-state/sequencer-1

# Run example and verify in one command
test-run: clean build example verify-example

# Validate goreleaser config
goreleaser-check:
    @echo "Checking .goreleaser.yaml syntax..."
    @goreleaser check

# Create a snapshot release (without publishing)
goreleaser-snapshot:
    @goreleasers release --snapshot --clean

# Show help
help:
    @just --list
