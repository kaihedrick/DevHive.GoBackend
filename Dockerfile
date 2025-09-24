# Build arguments for version pinning
ARG GO_VERSION=1.25
ARG ALPINE_VERSION=3.20

# Build stage - pinned for faster pulls
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache build-base ca-certificates tzdata git

# Install sqlc
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.27.0

# Set working directory
WORKDIR /src

# Cache modules first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Generate sqlc code
RUN sqlc generate

# Build the application with optimizations
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/devhive-api

# Runtime stage - pinned for faster pulls  
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /out/app /usr/local/bin/app

# Copy migrations
COPY --from=builder /src/cmd/devhive-api/migrations ./migrations

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Run the application
ENTRYPOINT ["/usr/local/bin/app"]