# Use a multi-stage build to keep the final image small
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o roomer-web ./cmd/web

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    k6 \
    wrk \
    ca-certificates

# Create a non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/roomer-web .

# Copy templates
COPY --from=builder /app/webui/templates ./webui/templates

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the web UI port
EXPOSE 8080

# Run the application
CMD ["./roomer-web"] 