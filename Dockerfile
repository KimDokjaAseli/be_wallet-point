# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Set timezone (optional, can be overridden)
ENV TZ=Asia/Jakarta

# Copy the binary from builder
COPY --from=builder /app/server .

# Create uploads directory
RUN mkdir -p uploads

# Expose the port
EXPOSE 8080

# Command to run the executable
CMD ["./server"]
