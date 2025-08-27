# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o devhive ./cmd/main.go

# Final runtime image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata \
  && addgroup -g 1001 -S appgroup \
  && adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the built Go binary
COPY --from=builder /app/devhive .

# Copy static assets
COPY --from=builder /app/static ./static

# Copy DB schema (optional)
COPY --from=builder /app/db/schema.sql ./db/schema.sql

# Ensure required folders exist
RUN mkdir -p /app/static/avatars /app/static/uploads \
  && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./devhive"]
