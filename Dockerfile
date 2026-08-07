# Multi-stage build for pvecli
# Stage 1: Build the application
FROM golang:1.26.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o pvecli main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 pvecli && \
    adduser -D -u 1000 -G pvecli pvecli

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/pvecli /app/pvecli

# Create config directory
RUN mkdir -p /home/pvecli/.pvecli && \
    chown -R pvecli:pvecli /home/pvecli

# Switch to non-root user
USER pvecli

# Set environment variables
ENV HOME=/home/pvecli

# Default command
ENTRYPOINT ["/app/pvecli"]
CMD ["--help"]
