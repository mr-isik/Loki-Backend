# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies including swag for generating swagger docs
RUN apk add --no-cache git ca-certificates && \
    go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate swagger documentation
RUN swag init -g cmd/main.go -o docs

# Build the application from cmd directory
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose application port
EXPOSE 3000

# Run the application
CMD ["./main"]
