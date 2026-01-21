# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the example
RUN CGO_ENABLED=0 GOOS=linux go build -o ai-gateway-example .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder and set permissions
COPY --from=builder /app/ai-gateway-example .
RUN chmod +x ai-gateway-example

# Expose port
EXPOSE 8083

# Run the binary
CMD ["./ai-gateway-example"]
