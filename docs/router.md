# Router

The Router page shows live health metrics for all LLM endpoints.

::: info Documentation coming soon
This page is being written. Check back shortly.
:::

The router tracks every upstream provider key as an endpoint. For each endpoint you can see:

- **TTFT P50/P99** — Time to first token latency
- **Average TPS** — Tokens per second throughput
- **Error rate** — Percentage of failed requests
- **Quota** — Token usage vs. configured limits
- **Circuit breaker state** — Closed, Open, or Half-Open
- **Score** — Composite health score used for routing decisions

When multiple keys serve the same model, the router automatically selects the healthiest endpoint. If an endpoint fails, the circuit breaker opens and traffic fails over to the next best option.
