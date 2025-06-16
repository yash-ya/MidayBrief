# Stage 1: Build the Go binary
FROM golang:1.21 AS builder

WORKDIR /app

# Copy go.mod and go.sum first (for caching deps)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -o midaybrief ./cmd

# Stage 2: Create a minimal image with just the binary
FROM debian:bullseye-slim

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/midaybrief .

# Expose the app port
EXPOSE 8080

# Run the binary
CMD ["./midaybrief"]
