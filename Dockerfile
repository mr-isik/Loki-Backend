# Development stage
FROM golang:1.25-alpine AS development

RUN apk add --no-cache git ca-certificates

# Install Air for hot reload
RUN go install github.com/air-verse/air@latest

# Install Swag for API documentation
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate swagger docs
RUN swag init -g cmd/main.go -o docs

EXPOSE 3000

# Use Air for hot reload
CMD ["air", "-c", ".air.toml"]

# Builder stage for production
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates && \
    go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN swag init -g cmd/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/main ./cmd

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 3000

CMD ["./main"]
