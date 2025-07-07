# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-alpine3.20 AS builder

# Build arguments for versioning
ARG VERSION=v0.0.0
ARG GIT_COMMIT
ARG GIT_DATE
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Install build dependencies
RUN apk add --no-cache gcc musl-dev linux-headers git jq bash

# Set up workspace
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Remove the replace directive that points to parent directory
RUN go mod edit -dropreplace github.com/ethereum-optimism/optimism

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source code
COPY . .

# Build the binary with cache mount
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p bin && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -v -ldflags="-X main.GitCommit=$GIT_COMMIT -X main.GitDate=$GIT_DATE -X main.Version=$VERSION" \
        -o ./bin/op-conductor-init ./cmd/op-conductor-init

# Final stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Copy the binary
COPY --from=builder /app/bin/op-conductor-init /usr/local/bin/

# Default to op-conductor-init
ENTRYPOINT ["op-conductor-init"]
