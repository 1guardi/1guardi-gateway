# API Keys

API keys authenticate requests to the gateway proxy. Each key is scoped to a tenant and optionally to a specific agent or user.

---

## Key Concepts

| Concept | Description |
|---|---|
| **Tenant key** | Generic key for a tenant. Can be used by any agent in that tenant. |
| **Agent-scoped key** | Bound to a specific agent. Requests using this key inherit the agent's policy. |
| **User-scoped key** | Bound to a specific user. Used for spend tracking and per-user rate limits. |
| **Prefix / Suffix** | Keys start with `sk_` and the suffix (last 4 chars) is shown in the UI for identification. The full key is shown **only once** at creation time. |

---

## Creating a Key

1. Navigate to **API Keys** in the sidebar.
2. Click **+ New Key**.
3. Fill in:

| Field | Required | Description |
|---|---|---|
| **Name** | ✅ | A label to identify the key (e.g., "Production", "my-agent-key") |
| **Agent** | Optional | Scope this key to a specific agent |
| **User** | Optional | Scope this key to a specific user (for spend tracking) |

4. Click **Create Key**.

::: danger Copy the key immediately
The full API key is displayed **only once** after creation. Copy it and store it securely. Once you dismiss the dialog, the full key cannot be retrieved — only the prefix and suffix remain visible.
:::

---

## Managing Keys

The API Keys page shows all keys for the current tenant:

| Column | What it shows |
|---|---|
| **Name** | The label you gave the key |
| **Prefix** | `sk_` — standard for all keys |
| **Suffix** | Last 4 characters for identification |
| **Scoped To** | Agent or user this key is bound to (blank = tenant-wide) |
| **Last Used** | When the key was last used for a proxy request |
| **Status** | Active or Revoked |

### Revoking a Key

Click the trash icon on any key row to revoke it. Revoked keys immediately stop working — all requests using that key return `401 Unauthorized`.

Revocation is irreversible. If you need a replacement, create a new key.

---

## How Keys Work in Requests

Every request to the gateway proxy must include an API key:

```bash
Authorization: Bearer sk_your_key_here
```

The gateway validates the key, identifies the tenant, and enforces any agent or user scope:

- If the key is **agent-scoped**, the `X-Agent-Id` header is ignored — the key's agent is always used.
- If the key is **tenant-wide**, you can pass any `X-Agent-Id` for the tenant's agents.

::: tip Key rotation
To rotate keys: create a new key, update your applications, then revoke the old key. There's no downtime if both keys are active during the switch.
:::
