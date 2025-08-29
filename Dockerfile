# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations for production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static'" \
    -o devhive \
    ./cmd/main.go

# Final runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && addgroup -g 1001 -S appgroup \
    && adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the built Go binary
COPY --from=builder /app/devhive .

# Copy essential files for all endpoints
COPY --from=builder /app/db/schema.sql ./db/schema.sql
COPY --from=builder /app/config/ ./config/
COPY --from=builder /app/static/ ./static/
COPY --from=builder /app/docs/ ./docs/

# Ensure required folders exist and set proper permissions
RUN mkdir -p /app/db /app/config /app/static /app/docs \
    && chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the port your application runs on
EXPOSE 8080

# Health check to ensure all endpoints are accessible
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment variables for production
ENV GIN_MODE=release
ENV PORT=8080

# Run the application
ENTRYPOINT ["./devhive"]
