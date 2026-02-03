# Fast Docker build using standard Go (not Bazel)
# This is much faster than building with Bazel inside Docker
# Use Bazel for local development, Go build for Docker production

# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build services with static linking
# PostgreSQL driver is pure Go, doesn't need CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o user_service \
    ./services/user

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o expense_service \
    ./services/expense

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o client \
    ./client

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o api_gateway \
    ./api-gateway

# User Service Image
FROM alpine:latest AS user_service
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/user_service ./user_service
COPY config ./config
EXPOSE 50051
CMD ["./user_service"]

# Expense Service Image
FROM alpine:latest AS expense_service
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/expense_service ./expense_service
COPY config ./config
EXPOSE 50052
CMD ["./expense_service"]

# Client Image
FROM alpine:latest AS client
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/client ./client
COPY config ./config
CMD ["./client"]

# API Gateway Image
FROM alpine:latest AS api_gateway
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/api_gateway ./api_gateway
COPY config ./config
EXPOSE 8080
CMD ["./api_gateway"]
