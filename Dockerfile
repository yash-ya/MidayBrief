# Stage 1: Build the Go binary
FROM golang:1.21 as builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the app
COPY . .

# 🔧 Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o midaybrief ./cmd

# Stage 2: Minimal image
FROM scratch

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/midaybrief .

# Expose the app port
EXPOSE 8080

# Run
CMD ["./midaybrief"]
