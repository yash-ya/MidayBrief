# Stage 1: Build
FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o midaybrief ./cmd

# Stage 2: Use distroless image with certs
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /app/midaybrief .

# Expose port
EXPOSE 8080

# Run binary
CMD ["/app/midaybrief"]
