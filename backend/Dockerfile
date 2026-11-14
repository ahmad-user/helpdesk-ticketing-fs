# ==============================================================================
# Stage 1: Build stage
# ==============================================================================
FROM golang:1.27-alpine AS builder

# Install build prerequisites and security certificates
RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Leverage Docker cache layer for dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy application source code
COPY . .

# Build statically linked binaries with debug symbols stripped (-s -w)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /src/bin/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /src/bin/migrate ./cmd/migrate && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -o /src/bin/seed ./cmd/seed

# ==============================================================================
# Stage 2: Minimal runtime image
# ==============================================================================
FROM alpine:3.21

# Install runtime utilities (curl for healthcheck, ca-certificates for TLS, tzdata for timestamps)
RUN apk --no-cache add ca-certificates tzdata curl && \
    addgroup -S appgroup && \
    adduser -S -u 10001 -G appgroup appuser

WORKDIR /app

# Copy compiled binaries from builder stage
COPY --from=builder /src/bin/api /app/api
COPY --from=builder /src/bin/migrate /app/migrate
COPY --from=builder /src/bin/seed /app/seed

# Copy database migrations, API documentation, and startup entrypoint
COPY --from=builder /src/migrations /app/migrations
COPY --from=builder /src/docs /app/docs
COPY --from=builder /src/entrypoint.sh /app/entrypoint.sh

# Ensure proper non-root permissions
RUN chmod +x /app/entrypoint.sh /app/api /app/migrate /app/seed && \
    chown -R appuser:appgroup /app

USER appuser

# Expose HTTP API port
EXPOSE 8080

# Configure container healthcheck against the health endpoint
HEALTHCHECK --interval=15s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/api"]
