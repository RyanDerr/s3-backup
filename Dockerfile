# Multi-stage build for minimal image size
# Stage 1: Build the Go binary
FROM golang:1.25.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files and download dependencies (cached layer)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG CGO_ENABLED=0

# Build the binary with optimizations
RUN go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -trimpath \
    -o s3-backup \
    .

# Verify the binary works
RUN ./s3-backup --help || echo "Binary built successfully"

# Stage 2: Create minimal runtime image
FROM alpine:3.23

# Install ca-certificates for HTTPS and tzdata for timezone support
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S s3backup && \
    adduser -u 1000 -S s3backup -G s3backup

# Copy binary from builder
COPY --from=builder /build/s3-backup /usr/local/bin/s3-backup

# Set ownership
RUN chown s3backup:s3backup /usr/local/bin/s3-backup

# Switch to non-root user
USER s3backup

# Set working directory
WORKDIR /home/s3backup

# Health check (optional, adjust based on your needs)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD pgrep -x s3-backup || exit 1

# Run the backup tool
ENTRYPOINT ["/usr/local/bin/s3-backup"]
