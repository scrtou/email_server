# Dockerfile for Go Backend Application

# ---- Build Stage ----
FROM golang:1.23.9-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
# - CGO_ENABLED=0: Disable CGO to build a statically linked binary
# - GOOS=linux: Build for Linux
# - -ldflags="-s -w": Strip debugging information and symbol table to reduce binary size
# - -o /app/email_server_app: Output the binary to /app/email_server_app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/email_server_app ./main.go

# ---- Run Stage ----
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/email_server_app /app/email_server_app

# Expose port 5555 to the outside world.
# This is the default port the application listens on, configured via SERVER_PORT environment variable.
EXPOSE 5555

# Command to run the executable
# The application reads its configuration (e.g., database connection, JWT secret, server port)
# from environment variables. Ensure these are set when running the container.
# Example environment variables:
# - SERVER_PORT (defaults to 5555 if not set)
# - SQLITE_FILE (defaults to ./gorm.db if not set)
# - JWT_SECRET (defaults to "your-super-secret-jwt-key-change-in-production" if not set)
# - JWT_EXPIRES_IN (defaults to 24 if not set)
CMD ["/app/email_server_app"]