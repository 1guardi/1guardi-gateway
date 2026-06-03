# AI Gateway

> Secure, observable middleware for agentic AI systems — an OpenAI-compatible proxy with multi-tenancy, LLM routing, guardrails, PII masking, and document security.

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/chaitanyabankanhal/ai-gateway)](https://goreportcard.com/report/github.com/chaitanyabankanhal/ai-gateway)

AI Gateway sits between your applications and upstream LLM providers. It speaks the
OpenAI API, so existing clients work unchanged, while adding tenant isolation,
intelligent routing with circuit breaking, and full OpenTelemetry tracing on the
hot path.

---

## Features

| Feature | Status |
|---|---|
| OpenAI-compatible proxy | ✅ Implemented |
| Multi-tenancy (tenant isolation) | ✅ Implemented |
| LLM router (fallback + circuit breaker) | ✅ Implemented |
| Admin UI (tenants, agents, keys, upstreams) | ✅ Implemented |
| JWT admin auth | ✅ Implemented |
| OTel tracing (TTFT, TPS) | ✅ Implemented |
| Guardrails engine | 🚧 UI stub (backend WIP) |
| PII masking & vault | 🚧 UI stub |
| Document security (ClamAV, VT, inspector) | 📋 Planned |
| Rate limiting | 📋 Planned |

---

## Architecture

A single binary starts two HTTP servers sharing one router instance:

| Server | Port | Purpose |
|---|---|---|
| **Proxy** | `8080` | OpenAI-compatible hot path — forwards requests to upstream LLMs |
| **Admin** | `8081` | Management API — tenants, agents, keys, upstreams, router metrics |

```
Client → proxy:8080
  → TenantContext middleware (API key → Redis cache → Postgres fallback)
  → OTel span start
  → llmrouter.Router.Select(model, tenantID) → upstream endpoint
  → reverse proxy / streaming passthrough
  → OTel span end (TTFT, TPS, token counts)
```

Routing scores endpoints by `w1×(1/TTFT_P99) + w2×avgTPS + w3×(1-errorRate)`
with a per-endpoint circuit breaker (`Closed → Open → Half-Open → Closed`).

---

## Repository Layout

```
ai-gateway/
├── backend/     Go backend — proxy + admin servers (see backend/CLAUDE.md)
├── cli/         Go CLI (aigw)
├── frontend/    React + Vite admin UI
├── landing/     Marketing site (not part of gateway runtime)
└── docs/        Documentation site
```

| Component | Stack |
|---|---|
| backend, cli | Go |
| frontend, landing, docs | React 19 / Vite / TypeScript / pnpm |

---

## Quick Start

### Prerequisites

- Go 1.22+
- Docker + Docker Compose (for Postgres, Redis, OTel collector, Jaeger)
- Node.js 20+ and pnpm (for frontend/docs)

### Run the backend

```bash
cd backend
cp .env.example .env          # then edit values
make tidy                     # one-time: fetch deps
make infra                    # start postgres + redis + otel + jaeger
make dev                      # run both servers locally (loads .env)
```

Proxy listens on `:8080`, admin on `:8081`. Jaeger UI at http://localhost:16686.

Send an OpenAI-compatible request:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk_<your-tenant-key>" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hello"}]}'
```

### Run the admin UI

```bash
cd frontend
pnpm install
pnpm dev                      # Vite dev server on :5173, proxies /api → :8081
```

### Full stack in Docker

```bash
cd backend
make up                       # build + start gateway with all dependencies
make logs                     # tail logs
make down                     # stop (downv to also drop volumes)
```

---

## Configuration

Backend is configured via environment variables (see [`backend/.env.example`](backend/.env.example)):

| Variable | Default | Description |
|---|---|---|
| `PROXY_PORT` | `8080` | OpenAI proxy port |
| `ADMIN_PORT` | `8081` | Admin API port |
| `POSTGRES_DSN` | `postgres://gateway:gateway@localhost:6432/gateway?sslmode=disable` | Postgres DSN |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `ADMIN_USERNAME` | `admin` | Admin user seeded on first boot |
| `ADMIN_PASSWORD` | _(empty — login disabled if unset)_ | Admin password (bcrypt-hashed in DB) |
| `JWT_SECRET` | _(random — tokens invalidated on restart)_ | HS256 JWT signing key |
| `JWT_TTL_HOURS` | `24` | JWT expiry |
| `OTEL_COLLECTOR_ADDR` | `localhost:4317` | OTLP gRPC endpoint |
| `UPSTREAM_N_*` | — | Upstream `N` (0–9): `KEY_ID`, `PROVIDER`, `MODEL`, `BASE_URL`, `API_KEY` |

Set `JWT_SECRET` and `ADMIN_PASSWORD` explicitly in any non-local deployment.

---

## Development

```bash
make test          # go test -race ./... (from root, delegates to backend)
make test-cover    # coverage report
make lint          # go vet / golangci-lint
make build         # compile
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow and coding
conventions, and [SECURITY.md](SECURITY.md) to report vulnerabilities.

---

## Documentation

Full docs live in [`docs/`](docs/). Run locally with `make docs-dev`.

---

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) and our
[Code of Conduct](CODE_OF_CONDUCT.md) before opening a pull request.

## Security

This is security middleware — **do not report vulnerabilities via public issues.**
See [SECURITY.md](SECURITY.md) for the private disclosure process.

## License

Licensed under the [Apache License 2.0](LICENSE).
