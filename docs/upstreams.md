# Provider Keys (Upstreams)

Provider keys are the credentials the gateway uses to forward requests to upstream LLM providers (OpenAI, Anthropic, Gemini). You configure them once per tenant — your application never touches these keys directly.

---

## How It Works

```
Your App  →  Gateway (your API key)  →  Upstream LLM (provider key)
```

Your application authenticates to the gateway with a [Gateway API Key](/api-keys). The gateway forwards the request to the upstream using the **provider key** you configure here. Your provider credentials stay on the gateway — your app never sees them.

---

## Supported Providers

| Provider | Base URL | Notes |
|---|---|---|
| **OpenAI** | `https://api.openai.com` | Full chat, completions, embeddings, models |
| **Anthropic** | `https://api.anthropic.com` | Native passthrough via `/v1/messages` |
| **Gemini** | `https://generativelanguage.googleapis.com` | OpenAI-compatible translation |
| **Custom (OpenAI Compatible)** | Any | Any API matching the OpenAI schema (LiteLLM, vLLM, local models) |

---

## Adding a Provider Key

1. Navigate to **Provider Keys** in the sidebar.
2. Click **+ Add Provider Key**.
3. Configure the upstream:

| Field | Required | Description |
|---|---|---|
| **Key Identifier** | ✅ | Unique label (e.g., `openai-primary`, `anthropic-backup`). Used by the router to identify this endpoint. |
| **Provider** | ✅ | Select OpenAI, Anthropic, Gemini, or Custom. |
| **API Key** | ✅ | Your provider API key. Stored encrypted at rest. |
| **Base URL** | Auto-filled | Automatically set based on provider. Override for custom endpoints. |
| **Models** | ✅ | Select which models this key enables. The gateway fetches the available model list from your provider after you enter the API key. |

4. **Select models** — After entering the API key, the gateway fetches available models from your provider. Check the models you want this key to serve.
5. Click **Add Key**.

::: tip Multiple keys for the same model
Add multiple provider keys for the same model (e.g., two OpenAI keys both serving `gpt-4`). The [Router](/router) automatically distributes traffic and fails over between them.
:::

---

## Managing Provider Keys

### Editing

Click the edit icon on any row to change the API key, provider, or models. The key identifier cannot be changed after creation — delete and recreate if needed.

### Deleting

Click the trash icon to remove a provider key. The router immediately stops routing to it. Any in-flight requests complete normally.

::: warning Delete carefully
If this is your only key for a model, requests to that model will fail until you add a replacement key.
:::

---

## Router Integration

Every provider key you add becomes an **endpoint** in the [Router](/router). The router:

- Tracks TTFT P50/P99, TPS, and error rate for each endpoint
- Computes a health score
- Selects the best endpoint per request
- Opens a **circuit breaker** if an endpoint fails repeatedly

You can view endpoint health on the **Router** page.

---

## Security

- Provider API keys are stored **encrypted at rest** in the gateway database.
- Keys are never exposed in API responses or logs.
- The gateway uses keys only to authenticate requests to upstream providers on your behalf.
- Your application never needs access to provider credentials.
