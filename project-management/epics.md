# AI Gateway MVP: Epics

Based on the actual state of the backend (Guardrails engine is fully functional and wired in), we have restructured the remaining MVP scope into 5 distinct Epics. They are ordered by priority, focusing on shipping the highest-leverage AI-specific features first.

---

## Epic 0: Identity Foundation (SSO / IdP Integration)
**Goal:** No user passwords ever stored in the gateway. All human identity sourced from external IdP via OIDC (primary) or SAML (enterprise). Prerequisite for Epics 5, 9, 11.
**Status:** Backlog / Prerequisite

**Approach:** Phase 1 — `coreos/go-oidc` v3 + `golang.org/x/oauth2` for OIDC. Covers Google Workspace, Microsoft Entra/Azure AD, Okta-OIDC, Auth0, GitHub, generic OIDC providers. Phase 2 — `crewjam/saml` if enterprise customers demand SAML. Re-evaluate WorkOS / Zitadel only if SCIM + many SAML connections needed.

**Package layout:**
```
internal/auth/
  oidc/        # provider registry, discovery, JWKS, ID-token verification
  saml/        # (phase 2) SP middleware, assertion parsing
  session/     # redis-backed browser sessions (httpOnly, SameSite=strict)
  middleware/  # http.Handler wrappers: requires-auth, requires-role
  jit/         # JIT user provisioning from IdP claims
  tenant/      # per-tenant IdP config CRUD
```

**Stories:**
- [ ] **E0.S1: Tenant IdP Config:** Admin UI + API to register OIDC IdP per tenant (issuer, client_id, client_secret, redirect_uri, scopes, claim mappings). Validate via discovery doc on save.
- [ ] **E0.S2: OIDC Auth-Code + PKCE Flow:** `/auth/login` initiates flow w/ state + PKCE. `/auth/callback` verifies ID token via `coreos/go-oidc` verifier, mints browser session. State/nonce in Redis w/ 10min TTL.
- [ ] **E0.S3: JIT User Provisioning:** First successful SSO creates user record. Roles derived from IdP group claims via configurable mapping (e.g., `groups: ["aigw-admins"] → role: admin`).
- [ ] **E0.S4: Redis Session Management:** httpOnly, SameSite=strict cookies. Idle (30min) + absolute (12h) timeouts. `/auth/logout` revokes session + best-effort IdP RP-initiated logout.
- [ ] **E0.S5: Auth Middleware:** `RequiresAuth` + `RequiresRole(...)` Go middleware. Injects user context into request. Standard 401/403 responses.
- [ ] **E0.S6: Bootstrap Magic-Link Admin:** Onboarding flow for a new tenant's first admin — one-time signed link emailed, no password. Auto-disabled once SSO is configured.
- [ ] **E0.S7: Break-Glass WebAuthn Admin:** Per-tenant local admin secured by WebAuthn / passkey only (no password). For IdP outages. Library: `go-webauthn/webauthn`. Heavily audited.
- [ ] **E0.S8: Audit Log:** `auth.login`, `auth.logout`, `auth.idp.config.change`, `auth.breakglass.use` spans + persisted trail. SOC2-grade.
- [ ] **E0.S9: SAML Support (Phase 2, Stretch):** `crewjam/saml` SP integration for enterprise IdPs (ADFS, OneLogin, Okta-SAML). Per-tenant SP metadata. Skip until customer demands.
- [ ] **E0.S10: SCIM 2.0 Sync (Phase 2, Stretch):** Inbound SCIM endpoint for IdP-pushed user/group lifecycle. De-provision on offboarding. Skip until enterprise needs it.

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

---

## Epic 9: User Authentication & CLI Key Rotation
**Goal:** Eliminate long-lived user API keys. Humans authenticate via SSO/OAuth through a CLI. Short-TTL tokens minted, stored in OS keychain, auto-refreshed transparently. Root signing key lives in KMS — never in app memory.
**Status:** Backlog

**Stories:**
- [ ] **E9.S1: Root Signing Key in KMS:** Asymmetric JWT signing key (RS256 or EdDSA) provisioned in cloud KMS / Vault Transit. Gateway signs via KMS API; private key never leaves HSM. JWKS endpoint publishes public key. KMS-native rotation with overlapping key versions.
- [ ] **E9.S2: Token Mint Service:** `/v1/auth/token` endpoint supporting `authorization_code` + `refresh_token` grants. JWT w/ jti, 15min default access TTL, 30-day rotating refresh.
- [ ] **E9.S3: Revocation & JTI Store:** Redis-backed jti tracking + revocation list. Proxy hot-path checks revocation per request. Refresh-token reuse detection triggers family revoke.
- [ ] **E9.S4: SSO / IdP Integration:** OAuth2 / OIDC integration with Google, Okta, Azure AD, GitHub. IdP assertion is the root of trust — no shared user secret stored in gateway.
- [ ] **E9.S5: CLI `aigw login`:** Browser-based SSO flow (PKCE). Refresh token stored in OS keychain (macOS Keychain, Windows Credential Manager, libsecret on Linux). Fallback to encrypted file w/ passphrase.
- [ ] **E9.S6: CLI Auto-Refresh & UX:** Background refresh before TTL expiry. Transparent retry on 401. Commands: `aigw whoami`, `aigw logout`, `aigw token` (print short-lived token for piping).
- [ ] **E9.S7: Per-User Rotation Policy:** Admin-configurable access TTL, refresh max-age, idle revoke, device binding. Per-tenant defaults.
- [ ] **E9.S8: Legacy Key Deprecation Path:** Existing long-lived user keys allowed but flagged in audit log + email warning. Force-migrate deadline per tenant.
- [ ] **E9.S9: Audit Log:** OTel spans `auth.user.mint`, `auth.user.refresh`, `auth.user.revoke` + persisted audit trail for SOC2.
- [ ] **E9.S10: Device-Bound Keys (Hardware POP):** CLI generates non-exportable keypair in Secure Enclave (macOS) / TPM (Windows/Linux) / Android Keystore at login. Public key registered w/ gateway, bound to user + device_id. Private key never leaves hardware.
- [ ] **E9.S11: DPoP Token Binding (RFC 9449):** Tokens include `cnf` thumbprint claim. Every mint/refresh and (configurable) proxy request requires DPoP proof signed by device key. Stolen token alone is useless without device.
- [ ] **E9.S12: Device Management:** `aigw devices list / revoke / rename` CLI commands. Admin UI to view + revoke devices per user. Audit log on register/revoke. Lost-device recovery via re-login from another device.
- [ ] **E9.S13: Software-Key Fallback:** Encrypted keypair on disk for environments without hardware enclave. Protected by OS keyring passphrase. Flagged as lower-assurance in audit; admin policy can disallow.

---

## Epic 10: Agent Authentication & Workload Identity
**Goal:** Eliminate long-lived API keys in agent / service deployments. SDK obtains short-TTL tokens via workload identity (cloud IAM, k8s, CI/CD OIDC) — zero secrets in code, env vars, or container images.
**Status:** Backlog

**Stories:**
- [ ] **E10.S1: Workload Identity Federation Endpoint:** `/v1/auth/token` `client_credentials` + `urn:ietf:params:oauth:grant-type:token-exchange` grants. Accepts external OIDC assertions, returns short-lived gateway JWT.
- [ ] **E10.S2: Cloud IAM Trust Providers:** Configure trust for AWS IAM (sigv4 / IRSA), GCP service accounts (Workload Identity Federation), Azure Managed Identity. Gateway verifies cloud-issued tokens, mints its own.
- [ ] **E10.S3: Kubernetes Projected SA Tokens:** Trust `kubernetes.io` issuer per cluster. Pod's projected service-account token → gateway token. Audience-bound to gateway.
- [ ] **E10.S4: CI/CD OIDC Trust:** Trust GitHub Actions, GitLab CI, CircleCI, Buildkite OIDC issuers. Per-repo / per-workflow / per-branch claim matching. No long-lived CI secrets needed.
- [ ] **E10.S5: mTLS Client Auth:** SPIFFE/SPIRE-compatible mTLS option for on-prem / non-cloud workloads. Cert SAN → identity mapping.
- [ ] **E10.S6: Bootstrap Secret (Fallback):** For environments lacking workload identity: short-lived bootstrap secret issued by admin, single-use exchange for credentials, auto-rotates on use.
- [ ] **E10.S7: SDK Auto-Refresh (Python / Node / Go):** SDK detects environment (k8s, GHA, AWS, GCP, Azure) and auto-selects identity provider. Refreshes in background. Consumer code never handles tokens.
- [ ] **E10.S8: Identity-to-Policy Binding:** Map workload identities (e.g., `repo:org/foo:ref:refs/heads/main`) to rate limits, model allow-lists, guardrail policies.
- [ ] **E10.S9: Audit Log:** OTel spans `auth.agent.mint`, `auth.agent.exchange`, `auth.agent.revoke` w/ source issuer + claims for forensics.
- [ ] **E10.S10: Break-Glass Static Token:** Sealed-envelope long-lived token for incidents / unsupported envs. Requires admin MFA to issue, auto-expires 24h, heavily audited.

---

## Epic 11: `aigw` CLI Implementation
**Goal:** Ship a cross-platform Go CLI (`aigw`) for Windows/macOS/Linux that delivers Epic 9's user auth flow: SSO login, hardware-bound device keys, auto-refresh, token piping, device management. Single static binary, no runtime deps.
**Status:** Backlog (depends on Epic 9)

**Stories:**
- [ ] **E11.S1: CLI Scaffold (Cobra + Viper):** Bootstrap repo layout. Commands: `login`, `logout`, `whoami`, `token`, `devices`, `config`, `version`. Shared HTTP client w/ retries + tracing.
- [ ] **E11.S2: OIDC / PKCE Login Flow:** Browser-based auth via `golang.org/x/oauth2` + `coreos/go-oidc`. Local loopback listener for callback. Handles headless mode (device-code grant) for SSH sessions.
- [ ] **E11.S3: OS Keyring Integration:** Use `99designs/keyring` to store refresh token in macOS Keychain / Windows Credential Manager / libsecret / KWallet. File-backed fallback w/ passphrase prompt.
- [ ] **E11.S4: Hardware Key Generation:**
    - macOS: Secure Enclave via `Security.framework` (cgo) or `facebookincubator/sks`.
    - Windows: TPM 2.0 via `google/go-tpm` + Platform Crypto Provider.
    - Linux: TPM 2.0 via `go-tpm`; fallback to encrypted software key.
- [ ] **E11.S5: DPoP Proof Signing:** Generate DPoP JWT per request signed by device key. JWS via `go-jose/go-jose`. Nonce handling + replay protection.
- [ ] **E11.S6: Auto-Refresh Daemon Mode:** Background refresh before TTL expiry. Optional long-running mode (`aigw agent`) exposing local Unix socket / named pipe for SDK token retrieval.
- [ ] **E11.S7: Token Piping & Shell Integration:** `aigw token` prints short-lived JWT for `curl` / scripts. Shell completions (bash/zsh/fish/pwsh). `eval $(aigw env)` for env-var export.
- [ ] **E11.S8: Device Management Commands:** `aigw devices list / revoke / rename`. Pretty-print table + `--json` output.
- [ ] **E11.S9: Cross-Platform Release Pipeline:** `goreleaser` config. Build matrix: linux/mac/windows × amd64/arm64. macOS code signing + notarization (Apple Developer ID). Windows Authenticode signing. SBOM + checksums.
- [ ] **E11.S10: Distribution Channels:** Homebrew tap, Scoop bucket, `.deb` / `.rpm` packages, direct download from `releases.aigw.io`. Install script (`curl ... | sh`) w/ checksum verify.
- [ ] **E11.S11: Self-Update:** `aigw update` command via `minio/selfupdate`. Signature verification against release public key. Opt-in auto-check on startup.
- [ ] **E11.S12: Config Management:** `~/.config/aigw/config.yaml` for endpoint, default tenant, telemetry opt-out. `aigw config get/set`. Env var overrides (`AIGW_ENDPOINT`).
- [ ] **E11.S13: Diagnostics:** `aigw doctor` checks hardware key support, keyring access, network reachability, clock skew (critical for JWT). `aigw logs` tails local refresh daemon.
- [ ] **E11.S14: Telemetry (Opt-In):** Anonymous usage + crash reporting. OTel-compatible. Off by default; clear prompt on first run.

---

## Epic 12: MCP Admin Server (AI-Driven Administration)
**Goal:** Expose the admin API surface as an MCP (Model Context Protocol) server so operators can run admin activities — tenant config, rules, keys, upstreams, members, trace inspection — through an AI agent in natural language. The agent calls typed MCP tools; every call enforces the same JWT auth + `RequirePermission` RBAC as the HTTP admin API.
**Status:** Backlog (depends on Epic 0 for identity; reuses `internal/admin` handlers)

**Approach:** New `internal/mcp` package. Streamable-HTTP MCP transport (`modelcontextprotocol/go-sdk` or `mark3labs/mcp-go`). MCP tools are thin wrappers over existing `internal/admin` service methods — no business logic duplication. Auth context (tenant + caller identity) flows from the MCP session into each tool call; RBAC checks unchanged. Read tools always on; mutating tools gated by policy + per-call confirmation.

**Package layout:**
```
internal/mcp/
  server/    # MCP transport, session lifecycle, tool registry
  tools/     # one file per tool group (tenants, rules, keys, upstreams, members, traces)
  auth/      # MCP-session -> auth.Claims bridge, RBAC enforcement
  policy/    # which tools are exposed, read-only vs mutating, confirmation gates
```

**Stories:**
- [ ] **E12.S1: MCP Server Scaffold:** Stand up `internal/mcp/server` with streamable-HTTP transport, tool registry, and session lifecycle. Mount under admin server (e.g. `/api/v1/mcp`). Health + capability handshake.
- [ ] **E12.S2: Auth & RBAC Bridge:** Map MCP session credentials (Bearer JWT) to `auth.Claims`; enforce `RequirePermission` per tool call. Reject unauthorized tool calls with structured MCP errors. SuperAdmin bypass preserved.
- [ ] **E12.S3: Read Tools:** Expose `list_tenants`, `get_tenant`, `list_rules`, `list_keys`, `list_upstreams`, `list_members`, `list_agents`, `get_guardrail_events` as MCP tools mapped to existing admin GET handlers.
- [ ] **E12.S4: Trace Inspection Tools:** `list_traces`, `get_trace_spans` MCP tools so the agent can debug guardrail/proxy behavior and answer "why was this request blocked".
- [ ] **E12.S5: Mutating Tools w/ Confirmation:** `create_rule`, `update_rule`, `delete_rule`, `create_key`, `revoke_key`, `create_upstream`, `update_upstream`, `delete_upstream`, `add_member`, `remove_member`, `update_tenant`. Each marked destructive; require explicit confirmation arg before execution.
- [ ] **E12.S6: Tool Policy & Exposure Config:** Per-tenant config to enable/disable the MCP server, allow-list tool groups, and force read-only mode. Mutating tools off by default.
- [ ] **E12.S7: Resources & Prompts:** Expose config schemas + guardrail-rule catalog as MCP resources. Ship prompt templates (e.g. "audit this tenant's rules", "diagnose blocked request") to steer the agent.
- [ ] **E12.S8: Audit Log:** OTel spans `mcp.tool.call`, `mcp.session.start`, `mcp.tool.denied` with caller identity, tool name, args digest, and verdict. SOC2-grade trail for every AI-initiated admin action.
- [ ] **E12.S9: Rate Limiting & Guardrails:** Apply rate limits per MCP session; run mutating tool args through the Guardrails engine to block prompt-injection-driven destructive calls.
- [ ] **E12.S10: Docs & Agent Onboarding:** Document connecting Claude Desktop / Claude Code / generic MCP clients. Example `.mcp.json`. Scoped-token issuance guide for the agent.
