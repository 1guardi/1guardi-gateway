# Single Sign-On (SSO)

AI Gateway supports SSO login via OIDC. Members sign in with their existing
**Google Workspace** or **Microsoft Entra ID (Azure AD)** account — no separate
gateway password.

A provider is **enabled only when both its Client ID and Client Secret are set**.
This guide walks through creating those credentials.

---

## Before you start

You need the gateway's **redirect (callback) URL**. It is built from
`OIDC_REDIRECT_BASE_URL` plus a fixed path:

```
{OIDC_REDIRECT_BASE_URL}/api/v1/auth/oidc/{provider}/callback
```

| Environment | Example callback URL |
|---|---|
| Local dev   | `http://localhost:8081/api/v1/auth/oidc/google/callback` |
| Production  | `{{GATEWAY_URL}}/api/v1/auth/oidc/google/callback` |

Replace `google` with `microsoft` for the Microsoft provider. Register the
**exact** URL with the provider — a mismatch causes a `redirect_uri_mismatch`
error at login.

---

## Google Workspace

### Create the OAuth credentials

1. Open the [Google Cloud Console](https://console.cloud.google.com/) and select
   (or create) a project.
2. Go to **APIs & Services → OAuth consent screen**.
   - Choose **Internal** (Workspace-only) or **External**.
   - Fill in app name, support email, and developer contact. Save.
3. Go to **APIs & Services → Credentials**.
4. Click **Create Credentials → OAuth client ID**.
5. **Application type:** Web application.
6. Under **Authorized redirect URIs**, add the callback URL(s) from
   [Before you start](#before-you-start) — add both the local and production
   URLs if you need both.
7. Click **Create**.
8. Copy the **Client ID** and **Client Secret** shown in the dialog. The secret
   is retrievable later, but treat it as sensitive.

### Configure the gateway

Set these environment variables:

```bash
OIDC_GOOGLE_CLIENT_ID=<your-client-id>
OIDC_GOOGLE_CLIENT_SECRET=<your-client-secret>
```

---

## Microsoft Entra ID (Azure AD)

### Register the application

1. Open the [Azure Portal](https://portal.azure.com/) → **Microsoft Entra ID**.
2. Go to **App registrations → New registration**.
3. Enter a **Name** (e.g. `AI Gateway`).
4. **Supported account types:** pick based on who should sign in
   (single-tenant, multi-tenant, or personal accounts too).
5. **Redirect URI:** select platform **Web** and enter the callback URL from
   [Before you start](#before-you-start) (use the `microsoft` provider path).
6. Click **Register**.
7. On the app **Overview** page, copy the **Application (client) ID** and the
   **Directory (tenant) ID**.

### Create a client secret

1. In the app, go to **Certificates & secrets → Client secrets**.
2. Click **New client secret**, set a description and expiry, then **Add**.
3. Copy the secret **Value** immediately — it is shown only once.

### Configure the gateway

```bash
OIDC_MICROSOFT_CLIENT_ID=<application-client-id>
OIDC_MICROSOFT_CLIENT_SECRET=<client-secret-value>
OIDC_MICROSOFT_TENANT=<directory-tenant-id>
```

`OIDC_MICROSOFT_TENANT` defaults to `common`, which allows any Entra tenant plus
personal Microsoft accounts. Set it to a specific Directory (tenant) ID to
restrict logins to your organization.

---

## Shared OIDC settings

| Variable | Default | Purpose |
|---|---|---|
| `OIDC_REDIRECT_BASE_URL` | `http://localhost:8081` | Base URL the provider redirects back to. Must be publicly reachable in production. |
| `OIDC_FRONTEND_URL` | `http://localhost:5173` | Admin UI URL the browser lands on after a successful login. |

---

## Verify

1. Restart the gateway after setting the variables.
2. Open the admin UI login page — enabled providers show a **Sign in with…**
   button.
3. Complete the provider login. You are redirected back and signed in.

If login fails:

- **`redirect_uri_mismatch`** — the callback URL registered with the provider
  does not exactly match `OIDC_REDIRECT_BASE_URL` + the callback path.
- **No SSO button** — one or both of `CLIENT_ID` / `CLIENT_SECRET` is unset for
  that provider.
- **Secret rejected** — the Microsoft secret expired, or you copied the secret
  ID instead of its **Value**.
