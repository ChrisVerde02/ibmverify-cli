# ibmverify-cli

Command-line tool for [IBM Verify](https://www.ibm.com/products/verify-identity). Manage applications, users, API clients, signer certificates, and access tokens — all from the terminal.

Built on top of [`ibmverify-go`](https://github.com/ChrisVerde02/ibmverify-go).

## Installation

```bash
go install github.com/ChrisVerde02/ibmverify-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/ChrisVerde02/ibmverify-cli
cd ibmverify-cli
go build -ldflags "-X main.version=v1.7.0" -o ibmverify ./cmd/ibmverify
```

## Quick start

```bash
# 1. Check version
./ibmverify --version

# 2. Run the full token-exchange flow
./ibmverify run \
  --tenant                     https://example.verify.ibm.com \
  --sts-client-id              <sts-client-id> \
  --sts-client-secret          <sts-client-secret> \
  --cert-manager-client-id     <cert-manager-client-id> \
  --cert-manager-client-secret <cert-manager-client-secret> \
  --subject   myusername \
  --issuer    https://demo.ibm.com \
  --label     DemoTokenSigner

# 3. List applications
./ibmverify app list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret>

# 4. List users
./ibmverify user list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret>

# 5. List API clients
./ibmverify apiclient list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret>
```

Progress is printed to stderr. Stdout is the access token only — so this works:

```bash
TOKEN=$(./ibmverify run --tenant ... --subject myusername --issuer ... --label ...)
```

## Credentials — three ways to provide them

Credentials are never required as flags. You can use any of these instead:

### Option 1 — `.env` file (recommended for local dev)

```bash
cp demo.env.example .env   # copy the template
# edit .env and fill in your values
```

`.env` is gitignored and loaded automatically by the CLI on every run.

### Option 2 — Environment variables (recommended for CI / scripts)

```bash
export VERIFY_TENANT=https://example.verify.ibm.com
export VERIFY_STS_CLIENT_ID=...
export VERIFY_STS_CLIENT_SECRET=...
export VERIFY_CERT_CLIENT_ID=...
export VERIFY_CERT_CLIENT_SECRET=...
export VERIFY_SUBJECT=myusername
export VERIFY_ISSUER=https://demo.ibm.com
export VERIFY_LABEL=DemoTokenSigner

./ibmverify run
```

### Option 3 — Config file

Use the built-in commands to create and fill the config file once:

```bash
ibmverify config init                                                 # creates ~/.ibmverify/config.yaml
ibmverify config set tenant                      https://example.verify.ibm.com
ibmverify config set sts-client-id               <sts-client-id>
ibmverify config set sts-client-secret           <sts-client-secret>
ibmverify config set cert-manager-client-id      <cert-manager-client-id>
ibmverify config set cert-manager-client-secret  <cert-manager-client-secret>
ibmverify config set subject  myusername
ibmverify config set issuer   https://demo.ibm.com
ibmverify config set label    DemoTokenSigner
```

After that every command works with no flags at all:

```bash
./ibmverify run
./ibmverify token get
./ibmverify cert get --label DemoTokenSigner
```

**Precedence:** flag > env var > config file > default.

## Demo scripts

All three scripts auto-build the binary if `./ibmverify` is not present.

### `demo2.sh` — token flow (cert → JWT → exchange → introspect)

End-to-end token exchange pipeline, 9 sections:

```bash
cp demo.env.example .env   # fill in credentials
chmod +x demo2.sh && ./demo2.sh
```

| Section | Command | What it shows |
|---|---|---|
| 1 | `config init` | Create `~/.ibmverify/config.yaml` |
| 2 | `config set` | Write tenant / subject / issuer / label |
| 3 | `cert delete` | Pre-clean any leftover certs |
| 4 | `cert upload` | Upload a temporary self-signed cert |
| 5 | `cert get` | Fetch and display cert label / subject / issuer |
| 6 | `token get` | Full flow with `--jwt-expires-in`, `--validity-days`, `--key-size` |
| 7 | `token introspect` | Validate the token, print subject / scope / expiry |
| 8 | `run` | All-in-one flow (defaults) |
| 9 | `run` | All-in-one flow with all expiration / cert flags explicitly set |

### `demo3.sh` — apps / users / API clients lifecycle

Full create → list → get → delete lifecycle for all three management domains:

```bash
chmod +x demo3.sh && ./demo3.sh
```

| Section | Domain | Commands |
|---|---|---|
| 1–4 | Applications | `app list`, `app create`, `app get`, `app delete` |
| 5–8 | Users (SCIM v2) | `user list`, `user create`, `user get`, `user delete` |
| 9 | API Clients (DCR) | `apiclient list`, `apiclient create`, `apiclient get`, `apiclient delete` |

### `demo-errors.sh` — error handling walkthrough

Shows how the four-layer error stack works end-to-end — SDK typed errors,
automatic retry, exit code mapping, and what each exit code means for CI:

```bash
chmod +x demo-errors.sh && ./demo-errors.sh
```

| Section | What it triggers | Exit code shown |
|---|---|---|
| 1 | Wrong credentials (deliberate) | `3` — auth failure |
| 2 | Cert label that does not exist | `4` — not found |
| 3 | Missing required flag | `2` — usage error |
| 4 | The `APIError` struct explained | — |
| 5 | Full exit code reference + CI pattern | — |

## Commands

### `app` — manage IBM Verify applications

#### `app list`

```bash
./ibmverify app list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret>
```

#### `app get`

```bash
./ibmverify app get \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <application-id>
```

#### `app create`

```bash
./ibmverify app create \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --name          "My Application" \
  --template-id   <template-id>
```

Get a template ID from: `./ibmverify app list -o json | jq -r '.[0].templateId'`

#### `app delete`

```bash
./ibmverify app delete \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <application-id>
```

All `app` commands support `-o json|yaml|text`.

---

### `user` — manage IBM Verify users (SCIM v2)

#### `user list`

```bash
./ibmverify user list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret>

# With SCIM filter
./ibmverify user list ... --filter 'userName eq "john"'
```

#### `user get`

```bash
./ibmverify user get \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <user-id>
```

#### `user create`

```bash
./ibmverify user create \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --username      john.doe \
  --password      "S3cure@Pass!" \
  --display-name  "John Doe"
```

#### `user delete`

```bash
./ibmverify user delete \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <user-id>
```

All `user` commands support `-o json|yaml|text`.

---

### `apiclient` — manage IBM Verify API clients (DCR)

#### `apiclient list`

```bash
./ibmverify apiclient list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret>
```

#### `apiclient get`

```bash
./ibmverify apiclient get \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <api-client-id>
```

#### `apiclient create`

```bash
./ibmverify apiclient create \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --name          "My Automation Client" \
  --entitlements  "manageUserGroups,readUsers" \
  --enabled
```

> **The `clientSecret` is only returned once at creation time.** Capture it immediately — IBM Verify never returns it again from `get` or `list`.

#### `apiclient delete`

```bash
./ibmverify apiclient delete \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <api-client-id>
```

All `apiclient` commands support `-o json|yaml|text`.

---

### `run` — full token-exchange flow in one command

Generates a certificate, uploads it, signs a JWT, exchanges it for a token, and introspects it. Progress goes to stderr; the access token goes to stdout.

```bash
./ibmverify run \
  --subject    myusername \
  --issuer     https://demo.ibm.com \
  --label      DemoTokenSigner

# With expiration / cert overrides
./ibmverify run \
  --subject            myusername \
  --issuer             https://demo.ibm.com \
  --label              DemoTokenSigner \
  --jwt-expires-in     5m \
  --validity-days      90 \
  --key-size           2048 \
  --subject-token-type urn:demo:token-type:user-jwt
```

| Flag | Default | Description |
|---|---|---|
| `--jwt-expires-in` | `15m` | JWT lifetime before exchange (e.g. `5m`, `1h`) |
| `--validity-days` | `365` | X.509 certificate validity window in days |
| `--key-size` | `4096` | RSA key size: `2048`, `3072`, or `4096` |
| `--organization` | `IBM` | Certificate O field |
| `--country` | `US` | Certificate C field (2-letter) |
| `--subject-token-type` | `urn:demo:token-type:user-jwt` | Token-exchange URN |

---

### `token` — OAuth token operations

#### `token get` — JWT token exchange (pipeable)

Same flow as `run` but prints only the raw access token to stdout. Accepts the same `--jwt-expires-in`, `--validity-days`, `--key-size`, and `--subject-token-type` flags as `run`.

```bash
TOKEN=$(./ibmverify token get \
  --subject myusername --issuer https://demo.ibm.com --label DemoTokenSigner)
```

#### `token introspect` — inspect an access token

```bash
./ibmverify token introspect \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --token         <access-token>
```

---

### `cert` — manage signer certificates

#### `cert upload`

```bash
./ibmverify cert upload \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --cert-file     ./cert.pem \
  --label         DemoTokenSigner
```

#### `cert get`

```bash
./ibmverify cert get \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --label         DemoTokenSigner
```

#### `cert delete`

```bash
./ibmverify cert delete \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --label         DemoTokenSigner
```

---

### `config` — manage the config file

```bash
./ibmverify config init               # creates ~/.ibmverify/config.yaml
./ibmverify config get                # display current values
./ibmverify config set <key> <value>  # set a single value
```

Valid keys: `tenant`, `sts-client-id`, `sts-client-secret`, `cert-manager-client-id`, `cert-manager-client-secret`, `subject`, `issuer`, `label`.

---

## Global flags

| Flag | Description |
|---|---|
| `-o`, `--output` | Output format: `text` (default), `json`, `yaml` |
| `--debug` | Print verbose HTTP details to stderr (secrets redacted) |
| `--version` | Print the CLI version |

```bash
# List applications as JSON — pipe into jq
./ibmverify app list --tenant ... --client-id ... --client-secret ... -o json | jq '.[].name'

# Get token as JSON (includes expires_in, subject, username)
./ibmverify run -o json | jq .access_token
```

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Unknown error |
| `2` | Usage / flag error |
| `3` | Authentication failure (check client ID and secret) |
| `4` | Resource not found |
| `5` | Rate limited |
| `6` | IBM Verify server error |
| `7` | Validation error |

## IBM Verify setup

The cert-manager API client is the most capable — a single client can cover all domains if it has the right entitlements:

| Entitlement | Used by |
|---|---|
| `manageCerts` | `cert upload/delete`, `token get`, `run` |
| `manageAppAccessAdmin` | `app create/delete` |
| `manageUserGroups` | `user create/delete` |
| `manageAPIClients` | `apiclient create/delete` |
| `readUsers` | `user list/get` |
| `readAPIClients` | `apiclient list/get` |
| `readAppConfig` | `app list/get` |

## Architecture

```
ibmverify-go (SDK)
  client/        ← top-level Client — Token, Certs, Apps, Users, APIClients
  apps/          ← raw HTTP wrappers (IBM spec mismatches)
  users/         ← SCIM v2, raw HTTP for Create (pwdChangedTime type bug)
  apiclients/    ← raw HTTP (201+Location→GET pattern for Create)
  crypto/        ← local JWT signing and RSA cert generation
  generated/     ← Fern-generated DO NOT EDIT

ibmverify-cli (this repo)
  cmd/run.go           ← full 5-step token-exchange flow
  cmd/token/           ← token get / introspect
  cmd/cert/            ← cert upload / get / delete
  cmd/app/             ← app list / get / create / delete
  cmd/user/            ← user list / get / create / delete (SCIM v2)
  cmd/apiclient/       ← apiclient list / get / create / delete (DCR)
  cmd/config/          ← config init / get / set
  internal/errkind/    ← APIError → exit code mapping
  internal/retry/      ← exponential backoff (3 attempts, 1s→2s→4s)
  internal/output/     ← text | json | yaml formatter
```

**Architecture rule:** the CLI never imports `generated/` directly. All IBM
Verify calls go through `client.Client` → `c.Apps`, `c.Users`, `c.APIClients`.

## Development

```bash
# Run all tests
go test ./...

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Lint
go vet ./...

# Build with version stamp
go build -ldflags "-X main.version=v1.8.0" -o ibmverify ./cmd/ibmverify
```

## Versioning

| Change | Bump |
|---|---|
| Breaking flag / output change | Major |
| New command or domain | Minor |
| Bug fix, demo update, doc update | Patch |

Current: **v1.8.0**

## Requirements

- Go 1.23+
- `openssl` on `$PATH` (only needed by `demo2.sh` to generate a temporary cert)
