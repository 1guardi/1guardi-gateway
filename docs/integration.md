# Integration Guide

Connect your application to AI Gateway. The gateway exposes an **OpenAI-compatible API** — any tool or SDK that supports custom OpenAI endpoints works without code changes.

::: info Gateway URL
All examples use `{{GATEWAY_URL}}` as a placeholder. Your actual gateway URL will be different —
your administrator can provide it.
:::

---

## Quick Start

Set your base URL to the gateway's proxy URL and use your gateway API key:

```bash
export OPENAI_BASE_URL="{{GATEWAY_URL}}/v1"
export OPENAI_API_KEY="sk_your_gateway_key"
```

Then use the OpenAI SDK as normal — requests flow through the gateway automatically.

---

## Code Examples

### Python

```python
from openai import OpenAI

client = OpenAI(
    base_url="{{GATEWAY_URL}}/v1",
    api_key="sk_your_gateway_key"
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}],
    extra_headers={
        "X-Agent-Id": "my-agent",
        "X-Tenant-Id": "your-tenant-id"
    }
)
```

### Node.js / TypeScript

```ts
import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: '{{GATEWAY_URL}}/v1',
  apiKey: 'sk_your_gateway_key',
});

const response = await openai.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Hello!' }],
}, {
  headers: {
    'X-Agent-Id': 'my-agent',
    'X-Tenant-Id': 'your-tenant-id',
  }
});
```

### cURL

```bash
curl {{GATEWAY_URL}}/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk_your_gateway_key" \
  -H "X-Agent-Id: my-agent" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

---

## Context Headers

Pass metadata via HTTP headers. The gateway uses these for tracing, routing, and tenant isolation.

| Header | Required | Description |
|---|---|---|
| `Authorization: Bearer sk_...` | ✅ Yes | Your gateway API key. Identifies the tenant and agent scope. |
| `X-Agent-Id` | Recommended | Identifies the agent for tracing and policy enforcement. |
| `X-Tenant-Id` | Optional | Override tenant namespace (must match API key's tenant). |
| `X-Thread-Id` | Optional | Persistent conversation ID. Groups multiple turns into one trace. |
| `X-Session-Id` | Optional | User session identifier for vault scoping. |

::: tip API Key Scope
If your API key is scoped to a specific agent, the `X-Agent-Id` header is ignored and the key's agent is used. This prevents one key from impersonating another agent.
:::

---

## Protocol Translation

The gateway translates between OpenAI and Anthropic protocols automatically.

| Scenario | How it works |
|---|---|
| **OpenAI SDK → any model** | Send requests to `/v1/chat/completions`. The gateway routes to the correct upstream regardless of model. |
| **Anthropic SDK / Claude Code** | Point `ANTHROPIC_BASE_URL` to the gateway. Use `/v1/messages` for native Anthropic passthrough. |
| **Custom OpenAI-compatible** | Set `OpenAI`-compatible base URL when configuring your provider. Works with any API matching the OpenAI schema. |

### Native Anthropic Passthrough

```bash
export ANTHROPIC_BASE_URL="{{GATEWAY_URL}}/v1"
export ANTHROPIC_API_KEY="sk_your_gateway_key"
claude
```

The gateway detects the `x-api-key` and `anthropic-version` headers and returns the correct model list format.

---

## CLI & IDE Integration

The gateway works with any tool supporting custom OpenAI endpoints.

| Tool | Setup |
|---|---|
| **Cursor** | Settings → Models → OpenAI API Key → set Base URL to gateway |
| **Continue (VS Code)** | `config.json` → `"apiBase": "{{GATEWAY_URL}}/v1"` |
| **Cline (VS Code)** | API Provider → OpenAI Compatible → set Base URL |
| **aichat** | `OPENAI_BASE_URL={{GATEWAY_URL}}/v1 aichat` |
| **Claude Code** | `ANTHROPIC_BASE_URL={{GATEWAY_URL}}/v1 claude` |

::: warning Model-Specific CLIs
For CLIs tightly coupled to a single provider (e.g., `gemini-cli`), use an OpenAI-compatible adapter like LiteLLM if the CLI doesn't support custom base URLs.
:::

---

## Endpoint Reference

### OpenAI-Compatible Surface

| Endpoint | Method | Description |
|---|---|---|
| `/v1/chat/completions` | POST | Chat completions (streaming + non-streaming) |
| `/v1/completions` | POST | Legacy text completions |
| `/v1/embeddings` | POST | Embedding generation |
| `/v1/models` | GET | List available models |

### Anthropic-Compatible Surface

| Endpoint | Method | Description |
|---|---|---|
| `/v1/messages` | POST | Anthropic Messages API (native passthrough) |
