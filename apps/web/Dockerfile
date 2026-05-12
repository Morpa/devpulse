# apps/web/Dockerfile — Multi-stage build for the DevPulse Astro frontend.
#
# Stage 1 (builder): install deps and run 'astro build' → /dist
# Stage 2 (runtime): serve the static output with Caddy (a tiny Go web server)

# ── Stage 1: Build ───────────────────────────────────────────────────────────
FROM node:24-alpine AS builder

WORKDIR /app

RUN npm install -g pnpm

# Copy manifests first for layer caching
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --ignore-scripts

# Copy source and build
COPY . .

# VITE_API_URL must point to the deployed API at build time so fetch() calls
# in the client bundle know where to go. Override with --build-arg.
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=${VITE_API_URL}

RUN pnpm run build
# Output is in /app/dist

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM caddy:2-alpine

# Copy static files from the builder
COPY --from=builder /app/dist /srv

# Minimal Caddyfile: serve static files and forward /api/* to the API
COPY Caddyfile /etc/caddy/Caddyfile

EXPOSE 80
