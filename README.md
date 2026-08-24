# ibmverify-cli

Command-line tool for [IBM Verify](https://www.ibm.com/products/verify-identity). Manage applications, users, signer certificates, and access tokens — all from the terminal.

Built on top of [`ibmverify-go`](https://github.com/ChrisVerde02/ibmverify-go).

## Installation

```bash
go install github.com/ChrisVerde02/ibmverify-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/ChrisVerde02/ibmverify-cli
cd ibmverify-cli
go build -o ibmverify ./cmd/ibmverify
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
  --tenant      https://example.verify.ibm.com \
  --client-id   <client-id> \
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

## Demo script

[`demo.sh`](demo.sh) runs a full end-to-end walkthrough using your `.env` credentials:

```bash
# 1. Set up credentials
cp demo.env.example .env
# edit .env with your real values

# 2. Build and run
go build -o ibmverify ./cmd/ibmverify
chmod +x demo.sh
./demo.sh
```

The script runs four steps in order:

| Step | Command | What it does |
|---|---|---|
| 1 | `cert delete` | Cleans up any existing cert with the same label |
| 2 | `token get` | Full cert → upload → JWT → exchange flow; captures the token |
| 3 | `token introspect` | Validates the token and prints subject, username, scope, expiry |
| 4 | *(print)* | Echoes the full token for Postman or `curl` |

## Commands

### `app` — manage IBM Verify applications

#### `app list` — list all applications

```bash
./ibmverify app list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret>
```

#### `app get` — get an application by ID

```bash
./ibmverify app get \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <application-id>
```

#### `app create` — create an application

```bash
./ibmverify app create \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --name          "My Application" \
  --template-id   <template-id>
```

#### `app delete` — delete an application

```bash
./ibmverify app delete \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --id            <application-id>
```

All `app` commands support `--output json|yaml|text`.

---

### `run` — full token-exchange flow in one command

Generates a certificate, uploads it, signs a JWT, exchanges it for a token, and introspects it. Progress goes to stderr; the access token goes to stdout.

```bash
./ibmverify run --subject myusername --issuer https://demo.ibm.com --label DemoTokenSigner
```

Optional flags: `--organization` (default `IBM`), `--country` (default `US`), `--validity-days` (default `365`), `--key-size` (default `4096`), `--jwt-expires-in` (default `300`), `--subject-token-type`.

---

### `token` — OAuth token operations

#### `token get` — JWT token exchange (pipeable)

Same flow as `run` but prints only the raw token to stdout.

```bash
TOKEN=$(./ibmverify token get --subject myusername --issuer https://demo.ibm.com --label DemoTokenSigner)
```

Optional: `--jwt-expires-in <seconds>` to control how long the signed JWT is valid before exchange.

#### `token introspect` — inspect an access token

```bash
./ibmverify token introspect --token <access-token>
```

---

### `cert` — manage signer certificates

#### `cert upload` — upload a signer certificate

```bash
./ibmverify cert upload --cert-file ./cert.pem --label DemoTokenSigner
```

#### `cert get` — fetch a signer certificate by label

```bash
./ibmverify cert get --label DemoTokenSigner
```

#### `cert delete` — delete a signer certificate

```bash
./ibmverify cert delete --label DemoTokenSigner
```

---

### `config` — manage the config file

#### `config init` — create the config file

```bash
./ibmverify config init
```

Creates `~/.ibmverify/config.yaml` with empty placeholders. Safe to run — will not overwrite an existing file.

#### `config get` — display current config

```bash
./ibmverify config get
```

#### `config set` — set a config value

```bash
./ibmverify config set <key> <value>
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

# Get token as JSON
./ibmverify token get -o json | jq .access_token
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

Two API clients are required in your IBM Verify tenant for the token-exchange flow:

| Client | Required entitlement | Used for |
|---|---|---|
| STS client | Token exchange | Exchanging a JWT for an access token, introspection |
| Cert-manager client | `manageCerts` | Uploading and deleting signer certificates |

Application management (`app` commands) requires an API client with application management entitlements.

## Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Vet
go vet ./...

# Build with version
go build -ldflags "-X main.version=v1.4.0" -o ibmverify ./cmd/ibmverify
```

## Requirements

- Go 1.23+
