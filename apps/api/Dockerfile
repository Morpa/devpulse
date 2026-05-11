# Stage 1: Build
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# -s remove symbol table, -w remove DWARF debug info (~30% menor)
# ./cmd/ aponta para onde está o main.go
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /devpulse-api \
    ./cmd/

# Stage 2: Runtime
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /devpulse-api /devpulse-api

EXPOSE 8080

USER 65534:65534

ENTRYPOINT ["/devpulse-api"]