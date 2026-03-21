# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make

# Copy the entire mono-repo workspace
# This is necessary because of the local replacements in go.mod
COPY . .

# The target service to build
ARG SERVICE_NAME

# Move to the service directory
WORKDIR /app/services/${SERVICE_NAME}

# Build the Go application
# Assuming entrypoints are at cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/server ./cmd/app

# ==========================================
# Final Stage
# ==========================================
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache tzdata ca-certificates curl

WORKDIR /app

ARG SERVICE_NAME

# Copy the binary
COPY --from=builder /app/bin/server .

# Try to copy any configuration files from the service directory.
# (If a service doesn't have these, it won't fail because we use wildcards or just conditionally copy).
# However, Docker COPY requires at least one matched file for wildcards, so we can use a trick:
COPY --from=builder /app/services/${SERVICE_NAME}/config*.yml /app/
# If there are config directories or other resources, they would go here.

# Expose common ports (Account: 8082, Catalog: 8083, Gateway: 8080)
EXPOSE 8080 8082 8083

# Run the binary
ENTRYPOINT ["/app/server"]
