# ABOUTME: Multi-stage Dockerfile for building minimal Clem container image
# ABOUTME: Produces a small, secure image with only the binary and minimal runtime dependencies

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build binary with version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/harper/clem/internal/core.Version=${VERSION} -X github.com/harper/clem/internal/core.Commit=${COMMIT} -X github.com/harper/clem/internal/core.Date=${DATE}" \
    -o clem \
    ./cmd/clem

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 clem && \
    adduser -D -u 1000 -G clem clem

# Create config directory
RUN mkdir -p /home/clem/.clem && \
    chown -R clem:clem /home/clem

WORKDIR /home/clem

# Copy binary from builder
COPY --from=builder /build/clem /usr/local/bin/clem

# Switch to non-root user
USER clem

# Set environment variables
ENV HOME=/home/clem

# Default command
ENTRYPOINT ["/usr/local/bin/clem"]
CMD ["--help"]
