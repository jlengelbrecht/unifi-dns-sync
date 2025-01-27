FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o unifi-dns-manager ./cmd/main.go

FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

# Copy binary and assets from builder
COPY --from=builder /app/unifi-dns-manager .
COPY web/templates ./web/templates

# Create data directory and set permissions
RUN mkdir -p /app/data && chmod 755 /app/data

# Create volume for persistent data
VOLUME /app/data

# Expose default port
EXPOSE 52638

# Set environment variables
ENV TZ=UTC

# Run the application
CMD ["./unifi-dns-manager", "-port", "52638", "-data-dir", "/app/data"]