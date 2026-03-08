# syntax=docker/dockerfile:1
# Enables BuildKit features: cache mounts, multi-platform, etc.

FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy BOTH go.mod and go.sum before downloading:
# go.sum is required for checksum verification
# This layer only rebuilds when dependencies change
COPY go.mod go.sum ./

# BuildKit cache mounts persist between builds on the same machine
# Makes rebuilds dramatically faster — modules aren't re-downloaded
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /api-gateway \
    ./cmd
# -w: strip DWARF debug info
# -s: strip symbol table
# Result: smaller binary, no debugging symbols in production

# distroless/static: no shell, no package manager, has CA certs for HTTPS
# :nonroot tag: runs as uid 65532 — never root in production
FROM gcr.io/distroless/static-debian12:nonroot AS release

# Document ports — doesn't publish them, but tools and humans use this
# 8080 HTTP / grpc-gateway
# 7002 gRPC
# 8081 admin (healthcheck, pprof)
EXPOSE 8080   
EXPOSE 7002   
EXPOSE 8081   

COPY --from=builder /api-gateway /api-gateway

# Config is NOT baked into the image:
# - Inject via CONFIG_PATH env var pointing to a mounted file
# - Or configure entirely via env vars (HTTP_PORT, GRPC_PORT, etc.)
# - Same image runs in dev, staging, and prod — only config differs
ENTRYPOINT ["/api-gateway"]