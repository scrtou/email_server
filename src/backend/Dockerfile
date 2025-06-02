# Dockerfile for Go Backend Application

# ---- Build Stage ----
FROM golang:1.23.9-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download
# Install C Compiler for CGO
RUN apk add --no-cache gcc musl-dev

# Copy the source code into the container
COPY . .

# Build the Go app
# - CGO_ENABLED=0: Disable CGO to build a statically linked binary
# - GOOS=linux: Build for Linux
# - -ldflags="-s -w": Strip debugging information and symbol table to reduce binary size
# - -o /app/email_server_app: Output the binary to /app/email_server_app
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/email_server_app ./main.go

# ---- Run Stage ----
FROM alpine:latest

# 安装必要的工具用于健康检查
RUN apk add --no-cache wget ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/email_server_app /app/email_server_app

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 创建数据目录并设置权限
RUN mkdir -p /data && chown -R appuser:appgroup /data /app

# 切换到非root用户
USER appuser

# Expose port 5555 to the outside world.
# This is the default port the application listens on, configured via SERVER_PORT environment variable.
EXPOSE 5555

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5555/api/v1/health || exit 1

# Command to run the executable
# The application reads its configuration (e.g., database connection, JWT secret, server port)
# from environment variables. Ensure these are set when running the container.
# Example environment variables:
# - SERVER_PORT (defaults to 5555 if not set)
# - SQLITE_FILE (defaults to ./gorm.db if not set)
# - JWT_SECRET (defaults to "your-super-secret-jwt-key-change-in-production" if not set)
# - JWT_EXPIRES_IN (defaults to 24 if not set)
CMD ["/app/email_server_app"]