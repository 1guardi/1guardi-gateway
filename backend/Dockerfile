# ── Build stage ────────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Fetch dependencies first — cached unless go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -trimpath \
    -o /app/gateway ./cmd/gateway

# ── Runtime stage ───────────────────────────────────────────────────────────────
FROM alpine:3.21

# Add ca-certificates for HTTPS and wget for healthchecks
RUN apk add --no-cache ca-certificates tzdata wget

COPY --from=builder /app/gateway /gateway

EXPOSE 8080 8081

ENTRYPOINT ["/gateway"]
