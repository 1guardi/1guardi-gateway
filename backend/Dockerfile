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
# distroless/static has no shell — minimal attack surface for a gateway
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/gateway /gateway
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080 8081

ENTRYPOINT ["/gateway"]
