# Security Policy

AI Gateway is security middleware that sits on the critical path between
applications and LLM providers. We take vulnerabilities seriously and appreciate
responsible disclosure.

## Reporting a Vulnerability

**Do not report security vulnerabilities through public GitHub issues,
discussions, or pull requests.**

Instead, please report privately using **one** of:

- [GitHub Private Vulnerability Reporting](../../security/advisories/new)
  (preferred) — go to the **Security** tab → **Report a vulnerability**.
- Email **bchaitanya15@gmail.com** with subject `SECURITY: ai-gateway`.

Please include:

- A description of the vulnerability and its impact.
- Steps to reproduce (proof of concept if possible).
- Affected component(s) and version/commit.
- Any suggested remediation.

## What to Expect

| Stage | Target |
|---|---|
| Acknowledgement of report | within 3 business days |
| Initial assessment & severity | within 7 business days |
| Fix or mitigation plan | communicated after assessment |

We will keep you informed throughout, credit you in the advisory (unless you
prefer to remain anonymous), and coordinate a disclosure timeline with you.

## Supported Versions

This project is pre-1.0 and under active development. Security fixes are applied
to the latest `master` and the most recent tagged release. Pin to a tagged
release for production use and upgrade promptly when advisories are published.

| Version | Supported |
|---|---|
| latest `master` | ✅ |
| latest tagged release | ✅ |
| older releases | ❌ |

## Scope

In scope: the gateway proxy, admin API, auth/JWT handling, tenant isolation,
API-key generation/validation, and the CLI.

Out of scope: vulnerabilities in third-party upstream LLM providers, issues
requiring physical access, and findings against the marketing `landing/` site
that do not affect the gateway runtime.

## Hardening Reminders for Operators

- Always set a strong `JWT_SECRET` and `ADMIN_PASSWORD` — never rely on defaults.
- Run behind TLS; terminate at a trusted proxy/load balancer.
- Restrict network access to the admin port (`8081`).
- Rotate tenant API keys and upstream provider credentials regularly.
