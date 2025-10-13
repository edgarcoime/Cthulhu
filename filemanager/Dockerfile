# Build stage
FROM golang:1.25.1-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates tzdata

# Copy shared package first
COPY pkg/ ./pkg/

# Copy go mod and sum files
COPY filemanager/go.mod filemanager/go.sum ./filemanager/

# Change to filemanager directory
WORKDIR /app/filemanager

# Download dependencies
RUN go mod download

# Copy source code
COPY filemanager/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/filemanager/main .

# Create app directory with proper permissions
RUN mkdir -p /app/fileDump && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Set dummy environment variables
ENV FILE_FOLDER=/app/fileDump
ENV PORT=5000
ENV CORS_ORIGIN=http://localhost:3000
ENV LOG_LEVEL=info

# Expose port
EXPOSE 5000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5000/health || exit 1

# Run the application
CMD ["./main"]
