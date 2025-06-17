# Multi-stage build for Content Service
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY src/ ./src/

# Build the content service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o risk-service ./src/cmd/risk-service

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/risk-service .

# Expose service port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8087/health || exit 1

# Run the binary
CMD ["./risk-service"]
