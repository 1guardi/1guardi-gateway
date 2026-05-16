# aigw CLI

Cross-platform CLI for AI Gateway: browser SSO login and API key management.

## Build

```sh
make build      # -> bin/aigw
make install    # -> $GOBIN/aigw
```

## Usage

```sh
aigw config set endpoint http://localhost:8081   # admin API base URL
aigw login [--provider google|microsoft]         # browser SSO via loopback listener
aigw whoami                                       # show logged-in identity
aigw token                                        # print JWT (for curl/scripts)
aigw logout                                        # clear stored session

aigw tenants list [--json]

aigw config set tenant 1                          # default tenant for key commands
aigw keys list   [--tenant <id>] [--json]
aigw keys create [--tenant <id>] --name <name> [--agent <id>] [--user <id>]
aigw keys revoke [--tenant <id>] --id <keyID>

aigw config set proxy http://localhost:8080       # proxy base URL SDKs hit
eval "$(aigw env openai)"                          # export OPENAI_* into shell
aigw env anthropic --write .env                    # upsert ANTHROPIC_* into .env
aigw run openai -- python app.py                   # run a command with the vars set
```

## Pointing an SDK at the gateway

`aigw env` / `aigw run` make an SDK route through the gateway proxy with no
code changes — they set the SDK's API-key + base-URL env vars to a gateway key
and the proxy URL.

Supported providers: `openai`, `anthropic`, `generic`.

| Provider    | Vars set                              | Base URL          |
|-------------|---------------------------------------|-------------------|
| `openai`    | `OPENAI_API_KEY`, `OPENAI_BASE_URL`   | `<proxy>/v1`      |
| `anthropic` | `ANTHROPIC_API_KEY`, `ANTHROPIC_BASE_URL` | `<proxy>`     |
| `generic`   | `AIGW_API_KEY`, `AIGW_BASE_URL`       | `<proxy>/v1`      |

```sh
eval "$(aigw env openai)"          # mutate the current shell
aigw run anthropic -- claude       # scope vars to a child process only
aigw env generic --write .env      # write vars to a dotenv file
aigw env openai --refresh          # mint a fresh gateway key
```

On first use the CLI mints a gateway API key (named `aigw-cli`) and caches its
plaintext in the OS keyring per tenant, reusing it on later runs. `--refresh`
forces a new key. `aigw logout` drops the cached key.

## Login flow

`aigw login` starts a local loopback HTTP listener, opens the browser at the
gateway's OIDC login URL with `cli_redirect` pointed back at that listener, and
captures the minted JWT once SSO completes. The gateway only accepts loopback
`cli_redirect` URLs, preventing token exfiltration.

## Configuration

`~/.config/aigw/config.yaml` — keys: `endpoint`, `proxy`, `tenant`. Environment
overrides: `AIGW_ENDPOINT`, `AIGW_PROXY`, `AIGW_TENANT`.

The session JWT is stored in the OS keyring (macOS Keychain / Windows Credential
Manager / libsecret). On systems without one it falls back to an encrypted file
backend; set `AIGW_KEYRING_PASSPHRASE` for non-interactive use.
