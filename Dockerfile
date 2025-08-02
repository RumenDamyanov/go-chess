# Build stage
FROM golang:1.22-alpine AS builder

# Install git and ca-certificates for fetching dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chess-server ./examples/api-server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S chess && \
    adduser -u 1001 -S chess -G chess

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/chess-server .

# Copy environment example file
COPY .env.example .env.example

# Change ownership to non-root user
RUN chown -R chess:chess /app

# Switch to non-root user
USER chess

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./chess-server"]
