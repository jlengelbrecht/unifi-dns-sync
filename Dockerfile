FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with version information
ARG VERSION=dev
ARG COMMIT=unknown
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-X main.Version=${VERSION} -X main.Commit=${COMMIT}" -o unifi-dns-manager ./cmd/main.go

FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs tzdata curl

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

# Copy binary and assets from builder
COPY --from=builder /app/unifi-dns-manager .
COPY web/templates ./web/templates

# Create data directory and set permissions
RUN mkdir -p /app/data && \
    chown -R app:app /app && \
    chmod 755 /app/data

# Switch to non-root user
USER app

# Create volume for persistent data
VOLUME /app/data

# Expose default port
EXPOSE 52638

# Set environment variables
ENV TZ=UTC \
    PORT=52638 \
    DATA_DIR=/app/data

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:${PORT}/health || exit 1

# Run the application
CMD ["./unifi-dns-manager", "-port", "52638", "-data-dir", "/app/data"]
