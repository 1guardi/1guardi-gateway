# Getting Started

Welcome to AI Gateway. This guide covers logging in and navigating the admin UI.

---

## Logging In

1. Open the admin UI in your browser (your admin will provide the URL).
2. Enter your **email** and **password**.
3. Click **Login**.

On first login, you may be asked to enrol in **Multi-Factor Authentication (MFA)**. Open your authenticator app, scan the QR code, and enter the 6-digit code.

---

## The Sidebar

After logging in, the left sidebar is your navigation hub.

### Top Bar

- **Tenant selector** — Dropdown at the very top. Switch between tenants (projects/teams) you have access to. Everything you see below is scoped to the selected tenant.
- **Agent selector** — Filters views to a specific agent. Set to "all" to see everything.

### Navigation Sections

| Section | Pages | What you do there |
|---|---|---|
| **Operations** | Router, Traces, Overview | Monitor endpoint health, view request traces, cost dashboards |
| **Safety** | Guardrails, PII Vault | Configure safety rules, manage PII masking |
| **Configuration** | Agents, API Keys, Provider Keys, Members, Tenants | Set up your team, keys, and LLM connections |

---

## First Steps for New Users

A typical onboarding flow:

1. **Get an API key** — Go to **API Keys** → create your first key. Copy it — it's shown only once.
2. **Connect a provider** — Go to **Provider Keys** → add your OpenAI or Anthropic key. The gateway proxies requests through this.
3. **Integrate your app** — Point your OpenAI SDK's `base_url` at the gateway. See the [Integration Guide](/integration).
4. **Send a test request** — Your request flows through the gateway, gets routed to your provider, and shows up under **Traces**.

---

## Roles and Permissions

Your access depends on your role:

| Role | Can do |
|---|---|
| **Super Admin** | Manage all tenants, users, and settings. Full cross-tenant visibility. |
| **Admin** | Manage all resources within assigned tenants. Create agents, keys, upstreams, rules. |
| **Viewer** | Read-only access to dashboards, traces, and configuration. Cannot make changes. |

Contact your administrator if you need elevated permissions.
