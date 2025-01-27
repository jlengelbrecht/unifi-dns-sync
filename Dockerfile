FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

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
RUN apk add --no-cache ca-certificates sqlite

# Copy binary and assets from builder
COPY --from=builder /app/unifi-dns-manager .
COPY --from=builder /app/web/templates ./web/templates

# Create volume for persistent data
VOLUME /app/data

# Expose default port
EXPOSE 52638

# Run the application
CMD ["./unifi-dns-manager", "-port", "52638"]