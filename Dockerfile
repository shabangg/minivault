# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /minivault-api ./main.go

# Final stage
FROM alpine:latest

# Install curl for healthchecks
RUN apk add --no-cache curl

WORKDIR /app

# Copy binary from builder
COPY --from=builder /minivault-api .

# Create logs directory
RUN mkdir -p logs

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/swagger/index.html || exit 1

# Run the application
CMD ["./minivault-api"] 