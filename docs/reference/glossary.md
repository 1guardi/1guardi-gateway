# Glossary

Common terms used throughout AI Gateway.

| Term | Definition |
|---|---|
| **Tenant** | An isolated namespace — typically a team, project, or customer. All resources (agents, keys, rules, traces) belong to a tenant. |
| **Agent** | An AI agent within a tenant. Can have its own API keys, guardrail rules, and trace data. |
| **API Key** | A `sk_`-prefixed key used to authenticate requests to the gateway proxy. Scoped to a tenant and optionally to an agent or user. |
| **Upstream / Provider Key** | Credentials the gateway uses to forward requests to an LLM provider (OpenAI, Anthropic, etc.). Never exposed to your application. |
| **Endpoint** | A specific provider key + model combination tracked by the router. Each endpoint has health metrics and a circuit breaker. |
| **Router** | Selects the healthiest endpoint for each request. Falls back between keys if one fails. |
| **Circuit Breaker** | A safety mechanism that stops routing to a failing endpoint. States: Closed (healthy), Open (failing), Half-Open (testing recovery). |
| **Guardrail** | A safety rule that inspects request inputs or LLM outputs. Can block, log, rewrite, or tag content. |
| **Managed Rule** | A built-in guardrail provided by the gateway (e.g., prompt injection detection). Updated centrally. |
| **Custom Rule** | A guardrail defined by the tenant using regex patterns. |
| **Trace** | A single user request and its LLM response, tracked end-to-end. Includes all spans. |
| **Span** | An atomic operation within a trace (LLM call, guardrail check, document scan). |
| **TTFT** | Time To First Token — latency from request to the first response token. |
| **TPS** | Tokens Per Second — throughput rate for streaming responses. |
| **PII** | Personally Identifiable Information. The gateway can detect and mask PII in requests and responses. |
