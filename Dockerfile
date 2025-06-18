# Build stage
FROM golang:1.24-alpine AS builder

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

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pullpoet ./cmd/main.go

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

# Set entrypoint
ENTRYPOINT ["./pullpoet"] 