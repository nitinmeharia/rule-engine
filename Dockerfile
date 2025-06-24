# Build stage
FROM golang:1.24.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rule-engine ./cmd/api

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S ruleengine && \
    adduser -u 1001 -S ruleengine -G ruleengine

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/rule-engine .

# Copy config directory
COPY --from=builder /app/configs ./configs

# Change ownership to non-root user
RUN chown -R ruleengine:ruleengine /app

# Switch to non-root user
USER ruleengine

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./rule-engine"] 