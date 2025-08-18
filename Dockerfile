# Build stage
FROM golang:1.23-alpine AS builder

# Install git for version information
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Get version information
RUN VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "docker") && \
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S') && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}'" \
    -o wutp .

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests (if needed)
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh wutp

WORKDIR /home/wutp

# Copy binary from builder stage
COPY --from=builder /app/wutp /usr/local/bin/wutp

# Make sure the binary is executable
RUN chmod +x /usr/local/bin/wutp

# Switch to non-root user
USER wutp

# Set entrypoint
ENTRYPOINT ["wutp"]

# Default command (show help)
CMD ["--help"]

# Metadata
LABEL maintainer="heartwilltell" \
    description="Who Use This Port - A fast port usage analyzer" \
    version="latest"
