# AGENTS.md — AI Gateway

Secure, observable middleware for agentic AI systems. OpenAI-compatible proxy with multi-tenancy, LLM routing, guardrails, PII masking, and document security.

---

## Repo Layout

```
ai-gateway/
├── backend/               Go backend (two HTTP servers)
│   ├── cmd/gateway/       Entry point — main.go
│   ├── config/            Config struct + env loading
│   ├── internal/
│   │   ├── admin/         Admin REST API (port 8081)
│   │   ├── auth/          API key generation/validation + JWT
│   │   ├── db/            GORM models, migrations, seeds
│   │   ├── providers/     LLM provider model-list service
│   │   ├── proxy/         Hot path — OpenAI-compatible proxy (port 8080)
│   │   ├── router/        LLM router with circuit breaker + scoring
│   │   └── telemetry/     OTel setup
│   └── deployments/       docker-compose.yml, otel-collector config
├── frontend/              React + Vite admin UI
│   └── src/
│       ├── api/           React Query hooks (client.ts, agents.ts, keys.ts, ...)
│       ├── components/    Sidebar, shadcn UI components
│       └── pages/         Router, Agents, APIKeys, Upstreams, Tenants, Login
└── landing/               Marketing site (not part of gateway runtime)
```

---

## Architecture

Two servers start from a single binary:

| Server | Port | Purpose |
|---|---|---|
| Proxy | 8080 | OpenAI-compatible hot path — forwards requests to upstream LLMs |
| Admin | 8081 | Management API — tenants, agents, keys, upstreams, router metrics |

Both share one `llmrouter.Router` instance (live metrics observable from admin).

**Request flow (proxy):**
```
Client → proxy:8080
  → TenantContext middleware (validates API key via Redis cache → Postgres fallback)
  → OTel span start
  → llmrouter.Router.Select(model, tenantID) → upstream endpoint
  → Reverse proxy / streaming passthrough
  → OTel span end (TTFT, TPS, token counts)
```

**Admin auth flow:**
```
POST /api/v1/auth/login → bcrypt verify AdminUser → JWT (HS256)
All other /api/v1/* → requireAuth middleware → validate JWT
```

---

## Key Packages

### `internal/proxy`
- `server.go` — builds chi router, mounts OTel + TenantContext middleware
- `context.go` — extracts tenant from `Authorization: Bearer sk_*` header; caches in Redis
- `handlers.go` — proxies `/v1/chat/completions` (streaming + non-streaming), records TTFT/TPS spans

### `internal/router`
- `endpoint.go` — `Endpoint` struct with EWMA metrics (TTFT P50/P99, TPS, error rate, circuit breaker)
- `router.go` — `Router.Select(model, tenantID)` scores endpoints; returns best or error; `Add`/`Remove` for live updates

Circuit breaker states: `Closed → Open → Half-Open → Closed`.

Scoring: `w1×(1/TTFT_P99) + w2×avgTPS + w3×(1-errorRate)`

### `internal/auth`
- `keys.go` — `GenerateAPIKey()` → `sk_[32-byte hex]`; `ValidateKey(key, hash)`
- `jwt.go` — `GenerateToken(userID, username, secret, ttl)` / `ValidateToken(token, secret)` → `*Claims`

### `internal/db`
- `models.go` — GORM models: `Tenant`, `Agent`, `APIKey`, `Upstream`, `AdminUser`
- `seed.go` — `SeedDefaultTenant(db, upstreams)`, `SeedAdminUser(db, username, password)` — both idempotent
- Auto-migrate runs on every boot via `db.AutoMigrate(db)`

**Model relationships:**
```
Tenant
  ├── []Agent
  ├── []APIKey   (Tenant-scoped; optionally Agent-scoped via AgentID)
  └── []Upstream (provider credentials)
AdminUser          (separate — admin UI login only, not tenant auth)
```

### `internal/admin`
- `server.go` — all management endpoints; JWT `requireAuth` middleware on protected routes
- Public: `GET /health`, `GET /ready`, `POST /api/v1/auth/login`
- Protected: everything under `/api/v1/*`

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PROXY_PORT` | `8080` | OpenAI proxy port |
| `ADMIN_PORT` | `8081` | Admin API port |
| `POSTGRES_DSN` | `postgres://gateway:gateway@localhost:6432/gateway?sslmode=disable` | Postgres |
| `REDIS_ADDR` | `localhost:6379` | Redis |
| `ADMIN_USERNAME` | `admin` | Admin user seeded on first boot |
| `ADMIN_PASSWORD` | _(empty — login disabled if unset)_ | Admin password (plaintext, bcrypt-hashed in DB) |
| `JWT_SECRET` | _(auto-generated random — tokens invalidated on restart)_ | HS256 JWT signing key |
| `JWT_TTL_HOURS` | `24` | JWT expiry |
| `OTEL_COLLECTOR_ADDR` | `localhost:4317` | OTLP gRPC endpoint |
| `UPSTREAM_N_KEY_ID` | — | Upstream N label (N=0..9) |
| `UPSTREAM_N_PROVIDER` | `openai` | Provider name |
| `UPSTREAM_N_MODEL` | — | Model name |
| `UPSTREAM_N_BASE_URL` | `https://api.openai.com` | Provider base URL |
| `UPSTREAM_N_API_KEY` | — | Provider API key |

Copy `backend/.env.example` → `backend/.env` before running.

---

## Commands

All from `backend/`:

```bash
make dev          # run both servers locally (loads .env)
make build        # compile to bin/gateway
make test         # go test -race ./...
make test-cover   # coverage report
make lint         # golangci-lint
make tidy         # go mod tidy

make infra        # start postgres + redis + otel + jaeger (Docker)
make up           # full stack including gateway container
make logs         # tail all service logs
make down         # stop + remove volumes
```

Root-level `make test` / `make build` delegate to backend.

Frontend (from `frontend/`):
```bash
npm run dev       # Vite dev server on :5173 (proxies /api → :8081)
npm run build     # production build
```

---

## Admin API Endpoints

All protected by `Authorization: Bearer <jwt>` except login and health probes.

```
POST   /api/v1/auth/login

GET    /api/v1/tenants
POST   /api/v1/tenants
GET    /api/v1/tenants/:id/
PATCH  /api/v1/tenants/:id/
DELETE /api/v1/tenants/:id/          (cascades agents, keys, upstreams)

GET    /api/v1/tenants/:id/agents
POST   /api/v1/tenants/:id/agents
GET    /api/v1/tenants/:id/keys
POST   /api/v1/tenants/:id/keys
DELETE /api/v1/tenants/:id/keys/:keyID   (revoke)

GET    /api/v1/tenants/:id/upstreams
POST   /api/v1/tenants/:id/upstreams
PUT    /api/v1/tenants/:id/upstreams/:keyID
DELETE /api/v1/tenants/:id/upstreams/:keyID

GET    /api/v1/router/endpoints
GET    /api/v1/providers/:provider/models
```

---

## Frontend

React 19 + Vite + Tailwind v4 + shadcn/ui + TanStack React Query.

**Routing:** State-based (no React Router). `App.tsx` owns `page: Page` state.

**Auth guard:** `isAuthenticated` from `localStorage.getItem('admin_token')`. Unauthenticated → `<Login />`. Logout → clear token, re-render.

**Page types:** `'overview' | 'traces' | 'guardrails' | 'pii-vault' | 'router' | 'agents' | 'api-keys' | 'upstreams' | 'tenants'`

**API client:** `src/api/client.ts` — Axios instance, base `${VITE_API_URL}/api/v1`, injects `Authorization: Bearer <token>` from localStorage.

**Coming-soon gate:** `VITE_COMING_SOON=false` to unlock overview/traces/guardrails/pii-vault pages.

---

## Data Flow — Adding a New Upstream

1. Frontend `POST /api/v1/tenants/:id/upstreams` with `{key_id, provider, models[], base_url, api_key}`
2. Admin handler writes `db.Upstream` row
3. Handler calls `llmRouter.Add(config.UpstreamConfig{...})` per model — live, no restart needed
4. Router immediately considers new endpoint in scoring

---

## Coding Conventions

- **Go:** `gofmt` + `goimports`. `log/slog` JSON for all logging. `errors.Is`/`errors.As` for error checking. Wrap with context (`fmt.Errorf("...: %w", err)`).
- **HTTP routing:** `github.com/go-chi/chi/v5`. Middleware for cross-cutting concerns (auth, tenant context).
- **Concurrency:** `golang.org/x/sync/errgroup` for goroutine groups.
- **Tests:** TDD — write failing test first. Use `-race` flag always. Tests live alongside source (`_test.go`).
- **Frontend:** React Query for all server state. `cn()` from `src/lib/utils.ts` for class merging. Monospace font + existing color tokens — match existing page aesthetics.
- **No new abstractions** unless 3+ concrete uses exist.

---

## Infrastructure (Docker Compose)

Services split into two profiles:

- `infra` — postgres, redis, otel-collector, jaeger
- `backend` — gateway (depends on infra)

Jaeger UI: `http://localhost:16686`
Postgres: `localhost:6432` (mapped from 5432)
Redis: `localhost:6379`

---

## MVP Feature Status

| Feature | Status |
|---|---|
| OpenAI-compatible proxy | Implemented |
| Multi-tenancy (tenant isolation) | Implemented |
| LLM router (fallback + circuit breaker) | Implemented |
| Admin UI (tenants, agents, keys, upstreams) | Implemented |
| JWT admin auth | Implemented |
| OTel tracing (TTFT, TPS) | Implemented |
| Guardrails engine | UI stub only (backend not implemented) |
| PII masking & vault | UI stub only |
| Document security (ClamAV, VT, inspector) | Not started |
| Rate limiting | Not started |
| HRMS user sync | Not started |
