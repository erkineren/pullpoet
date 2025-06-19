# Build stage
FROM golang:1.24-alpine AS builder

# Build argument for version (can be passed during docker build)
ARG VERSION=""

# Install git and ca-certificates for git operations
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with version
# Since .git is excluded from Docker context, we rely on build ARG
RUN VERSION_TO_USE=${VERSION:-dev} && \
    echo "Using version: ${VERSION_TO_USE}" && \
    CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-s -w -X main.version=${VERSION_TO_USE}" \
    -o pullpoet ./cmd/main.go

# Runtime stage
FROM alpine:latest

# Install git and ca-certificates for runtime
RUN apk --no-cache add git ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/pullpoet .

# Change ownership to non-root user
RUN chown appuser:appgroup /app/pullpoet

# Switch to non-root user
USER appuser

# Add helpful volume mount point
VOLUME ["/workspace"]

# Set default working directory to workspace
WORKDIR /workspace

# Set entrypoint
ENTRYPOINT ["/app/pullpoet"] 