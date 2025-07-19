# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the applications with optimized flags (same as in Makefile)
RUN mkdir -p bin && \
    go build -ldflags "-s -w" -o bin/order ./cmd/order && \
    go build -ldflags "-s -w" -o bin/product ./cmd/product

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/bin/ /app/bin/

# Copy configuration files
COPY config/ /app/config/

# Create directory for migrations
COPY migrations/ /app/migrations/

# Make binaries executable
RUN chmod +x /app/bin/order /app/bin/product

# The command will be specified in docker-compose.yml for each service