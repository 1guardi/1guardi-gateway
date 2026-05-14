# ── Build stage ──────────────────────────────────────────────────────────────
FROM node:22-alpine AS builder

WORKDIR /app
RUN corepack enable

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

# ── Serve stage ──────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine

COPY --from=builder /app/.vitepress/dist /usr/share/nginx/html

# Healthcheck endpoint
RUN echo "OK" > /usr/share/nginx/html/health

COPY nginx.docs.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
