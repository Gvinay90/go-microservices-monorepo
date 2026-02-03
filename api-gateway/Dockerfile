# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files from root
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the gateway
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./api-gateway/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary
COPY --from=builder /app/api-gateway .

# Copy config
COPY --from=builder /app/config ./config

EXPOSE 8080

CMD ["./api-gateway"]
