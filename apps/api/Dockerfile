# apps/api/Dockerfile — Multi-stage build for the DevPulse Go API.
#
# Stage 1 (builder): compile with full Go toolchain → tiny static binary
# Stage 2 (runtime): copy only the binary into a ~5 MB scratch image
#
# Result: the final Docker image is under 10 MB with no shell or OS overhead.

# ── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

# Install git so 'go mod download' can fetch VCS-based dependencies
RUN apk add --no-cache git

WORKDIR /src

# Copy dependency manifests first — Docker caches this layer separately so
# 'go mod download' only re-runs when go.mod/go.sum change, not on every code edit.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source and build a fully-static binary
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \   
    # -s strips the symbol table
    # -w removes DWARF debug info
    # Together they cut the binary size by ~30%
    -o /devpulse-api .

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM scratch

# Copy TLS certificates so the binary can make HTTPS calls to api.github.com
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /devpulse-api /devpulse-api

# Expose the API port
EXPOSE 8080

# Set a non-root user for security (scratch images support numeric UID directly)
USER 65534:65534

ENTRYPOINT ["/devpulse-api"]
