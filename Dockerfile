# ABOUTME: Multi-stage Dockerfile for building minimal Hex container image
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
    -ldflags="-s -w -X github.com/2389-research/hex/internal/core.Version=${VERSION} -X github.com/2389-research/hex/internal/core.Commit=${COMMIT} -X github.com/2389-research/hex/internal/core.Date=${DATE}" \
    -o hex \
    ./cmd/hex

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 hex && \
    adduser -D -u 1000 -G hex hex

# Create config directory
RUN mkdir -p /home/hex/.hex && \
    chown -R hex:hex /home/hex

WORKDIR /home/hex

# Copy binary from builder
COPY --from=builder /build/hex /usr/local/bin/hex

# Switch to non-root user
USER hex

# Set environment variables
ENV HOME=/home/hex

# Default command
ENTRYPOINT ["/usr/local/bin/hex"]
CMD ["--help"]
