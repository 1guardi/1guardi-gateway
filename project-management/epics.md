# AI Gateway MVP: Epics

Based on the actual state of the backend (Guardrails engine is fully functional and wired in), we have restructured the remaining MVP scope into 5 distinct Epics. They are ordered by priority, focusing on shipping the highest-leverage AI-specific features first.

---

## Epic 1: Performance Benchmarking & De-risking
**Goal:** Ensure the proxy hot-path can handle payload inspection without unacceptable latency overhead, especially for streaming responses.
**Status:** Ready for Development

**Stories:**
- [ ] **E1.S1: Synthetic Benchmark Script:** Build a script that pumps streaming traffic through the proxy and measures P50/P99 TTFT and TPS.
- [ ] **E1.S2: Proxy Latency Budget:** Instrument `handlers.go` with synthetic CPU load to determine the maximum millisecond budget the Guardrails/PII engine has per chunk.
- [ ] **E1.S3: Guardrails Optimization:** (If necessary based on S1) Optimize the current regex-based guardrails engine for streaming chunks.

---

## Epic 2: Dedicated Security LLM Pipeline
**Goal:** Implement a fast, fine-tuned security LLM to perform real-time semantic analysis for prompt injection, data poisoning, and complex issue patterns that regex cannot catch.
**Status:** High Priority / Ready for Development

**Stories:**
- [ ] **E2.S1: Security LLM Orchestration:** Integrate the dedicated security LLM call into the `internal/proxy` hot-path.
- [ ] **E2.S2: Semantic Threat Detection:** Fine-tune or configure the system prompt to detect advanced prompt injections and jailbreaks.
- [ ] **E2.S3: Data Poisoning Detection:** Implement analysis of inbound RAG context/tool outputs to detect data poisoning attempts.
- [ ] **E2.S4: Caching Layer:** Implement strict caching (e.g., exact hash match) for security LLM verdicts to minimize latency.

---

## Epic 3: Bidirectional PII Masking & Session Vault
**Goal:** Implement the core data privacy feature: detecting, tokenizing, and unmasking PII in real-time.
**Status:** Ready for Development

**Stories:**
- [ ] **E3.S1: Redis Session Vault:** Implement the `internal/vault` package to store and retrieve ephemeral mapping tokens (e.g., `[PII_EMAIL_1]` -> `actual@email.com`) with a 24h TTL.
- [ ] **E3.S2: Inbound Masking Pipeline (Regex/NER):** Intercept requests in `internal/proxy`, detect entities, store them in the vault, and rewrite the payload before it hits the LLM.
- [ ] **E3.S3: Outbound Unmasking (Non-Streaming):** Scan complete LLM responses, identify vault tokens, dereference them, and rewrite the final response.
- [ ] **E3.S4: Outbound Unmasking (Streaming):** Implement a chunk-aware scanner in `handlers.go` that can identify tokens split across stream frames and unmask them smoothly.
- [ ] **E3.S5: Token Index API:** Ensure the API can return the `pii_map` for client-side rendering when `pii_response_mode: tokens` is set.

---

## Epic 4: Document Security Pipeline
**Goal:** Ensure all files/URLs processed by agents are free of malware and semantic threats.
**Status:** Backlog

**Stories:**
- [ ] **E4.S1: Pipeline Orchestration:** Create the `internal/docsec` package to route files through the 3 stages before content extraction.
- [ ] **E4.S2: ClamAV Integration (Stage 1):** Add the ClamAV sidecar to Docker Compose and implement the streaming signature scan.
- [ ] **E4.S3: VirusTotal Integration (Stage 2):** Implement the VT API check for external URLs/hashes.
- [ ] **E4.S4: LLM Payload Inspector (Stage 3):** Implement the prompt injection and semantic threat detection call to the dedicated inspector LLM model, including caching.
- [ ] **E4.S5: OTel Tracing:** Emit `doc.scan.*`, `doc.vt.*`, and `doc.inspector.*` spans.

---

## Epic 5: Access Management & HITL Authentication
**Goal:** Provide secure user administration and human-in-the-loop approvals for agent tools.
**Status:** Backlog

**Stories:**
- [ ] **E5.S1: Self-Service API Keys & Rate Limiting:** Implement rate limiting (TPS, TPM, RPM, RPD) logic on the proxy path using Redis.
- [ ] **E5.S2: MFA & Re-Captcha:** Add Re-Captcha verification to login and support TOTP/Email OTP for admin actions.
- [ ] **E5.S3: HITL Tool Approvals:** Intercept sensitive tool calls and hold the LLM trace open while waiting for human MFA verification via the admin UI.
- [ ] **E5.S4: HRMS Sync (Stretch):** Basic polling ingestion of users from BambooHR/Workday.

---

## Epic 6: UI/UX Integration & Polish
**Goal:** Wire up all new backend APIs to the React frontend.
**Status:** Blocked (Depends on Epics 3, 4, 5)

**Stories:**
- [ ] **E6.S1: PII Vault Dashboard:** Connect the React UI to manage PII entities and view Session Vault mappings (if permitted by policy).
- [ ] **E6.S2: Document Security Policy UI:** Add UI settings to configure ClamAV bypasses, VT thresholds, and Inspector fail-open/fail-closed states.
- [ ] **E6.S3: HITL Approval Queue:** Create a dashboard view where humans can review and approve pending tool calls.

---

## Epic 7: Security Testing & Compliance
**Goal:** Ensure the AI Gateway meets enterprise security standards, including SOC2 and HIPAA compliance requirements.
**Status:** Backlog

**Stories:**
- [ ] **E7.S1: Penetration Testing:** Conduct automated and manual security testing on proxy and admin endpoints.
- [ ] **E7.S2: SOC2 Audit Preparation:** Implement comprehensive audit logging for all access control and configuration changes.
- [ ] **E7.S3: HIPAA Compliance Controls:** Ensure BAA (Business Associate Agreement) prerequisites are met, including end-to-end encryption and strict PHI (Protected Health Information) data handling within the vault.
- [ ] **E7.S4: Vulnerability Scanning:** Integrate SAST and DAST tools into the CI/CD pipeline.

---

## Epic 8: Advanced Provider & Routing Policies
**Goal:** Implement granular control over provider-specific features and complex routing strategies like provisioned throughput and priority-based failover.
**Status:** Backlog

**Stories:**
- [ ] **E8.S1: Thought Signature Passthrough:** Ensure `thought_signature` and other provider-specific response headers/metadata are correctly preserved and passed to the client.
- [ ] **E8.S2: Provisioned Throughput Support:** Add support for routing to provisioned throughput endpoints (e.g., Azure OpenAI PTU or AWS Bedrock Provisioned Throughput) with specific rate-limit logic.
- [ ] **E8.S3: Priority Routing (Pay-as-you-go vs. Provisioned):** Implement routing policies that prioritize provisioned throughput and fallback to pay-as-you-go based on latency or cost settings.
- [ ] **E8.S4: Provider Metadata API:** Extend the Admin API to allow configuring advanced provider-specific metadata per endpoint.
