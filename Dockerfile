# ============================================
# Multi-Stage Dockerfile for Git Manage Service
# ============================================
# Stage 1: Frontend Builder (Node.js)
# Stage 2: Backend Builder (Go)
# Stage 3: Runtime (Alpine)
#
# Build with version info:
# docker build \
#   --build-arg VERSION=$(git describe --tags --always) \
#   --build-arg BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S') \
#   --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#   -t git-manage-service:latest .
# ============================================

# Frontend Build Stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm install

# Copy frontend source code
COPY frontend/ ./

# Build frontend (output: dist/)
RUN npm run build

# Backend Build Stage
FROM golang:1.24-alpine AS backend-builder

# Set environment variables with China Proxy
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct

# Replace Alpine mirrors with Aliyun mirror
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install build dependencies (git, gcc, musl-dev required for CGO/SQLite)
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy frontend build output to public directory
COPY --from=frontend-builder /app/frontend/dist ./public

# Build arguments for version injection
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Build the application with version info
RUN go build -ldflags "-X 'main.Version=${VERSION}' -X 'main.BuildTime=${BUILD_TIME}' -X 'main.GitCommit=${GIT_COMMIT}'" \
    -o git-manage-service main.go

# Runtime Stage
FROM alpine:latest

# Replace Alpine mirrors with Aliyun mirror
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# Install runtime dependencies
# git: for git operations
# openssh-client: for ssh git access
# ca-certificates: for https git access
# tzdata: for timezone setting
RUN apk add --no-cache \
    git \
    openssh-client \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/git-manage-service .

# Copy frontend assets (already integrated in backend builder)
COPY --from=backend-builder /app/public ./public

# Copy swagger docs
COPY --from=backend-builder /app/docs ./docs

# Copy default config
COPY --from=backend-builder /app/conf/config.yaml ./conf/config.yaml

# Set environment variables
ENV GIN_MODE=release \
    PORT=8080 \
    DB_PATH=/app/data/git_sync.db

# Expose port
EXPOSE 8080
EXPOSE 8888

# Create volume directories
RUN mkdir -p /app/data /root/.ssh && chmod 700 /root/.ssh

# Start command
CMD ["./git-manage-service"]
