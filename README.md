# ibmverify-cli

Command-line tool for [IBM Verify](https://www.ibm.com/products/verify-identity). Perform the full JWT token-exchange flow, manage signer certificates, and introspect access tokens — all from the terminal.

Built on top of [`ibmverify-go`](https://github.com/ChrisVerde02/ibmverify-go).

## Installation

```bash
go install github.com/ChrisVerde02/ibmverify-cli@latest
```

Or clone and build:

```bash
git clone https://github.com/ChrisVerde02/ibmverify-cli
cd ibmverify-cli
go build -o ibmverify .
```

## Quick start

```bash
# 1. Check version
./ibmverify --version

# 2. Run the full flow (credentials via flags)
./ibmverify run \
  --tenant               https://example.verify.ibm.com \
  --sts-client-id        <sts-client-id> \
  --sts-client-secret    <sts-client-secret> \
  --cert-manager-client-id     <cert-manager-client-id> \
  --cert-manager-client-secret <cert-manager-client-secret> \
  --subject   myusername \
  --issuer    https://demo.ibm.com \
  --label     DemoTokenSigner
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

Create `~/.ibmverify/config.yaml`:

```yaml
tenant: https://example.verify.ibm.com
sts-client-id: ...
sts-client-secret: ...
cert-manager-client-id: ...
cert-manager-client-secret: ...
subject: myusername
issuer: https://demo.ibm.com
label: DemoTokenSigner
```

**Precedence:** flag > env var > config file > default.

## Demo script

[`demo.sh`](demo.sh) runs a full end-to-end walkthrough using your `.env` credentials:

```bash
# 1. Set up credentials
cp demo.env.example .env
# edit .env with your real values

# 2. Build and run
go build -o ibmverify .
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

### `run` — full flow in one command

Generates a certificate, uploads it, signs a JWT, exchanges it for a token, and introspects it. Progress goes to stderr; the access token goes to stdout.

```bash
./ibmverify run --subject myusername --issuer https://demo.ibm.com --label DemoTokenSigner
```

Optional flags: `--organization` (default `IBM`), `--country` (default `US`), `--validity-days` (default `365`), `--key-size` (default `4096`), `--subject-token-type`.

### `token get` — JWT token exchange (pipeable)

Same flow as `run` but prints only the raw token to stdout.

```bash
TOKEN=$(./ibmverify token get --subject myusername --issuer https://demo.ibm.com --label DemoTokenSigner)
```

### `token introspect` — inspect a token

```bash
./ibmverify token introspect --token <access-token>
```

### `cert upload` — upload a signer certificate

```bash
./ibmverify cert upload --cert-file ./cert.pem --label DemoTokenSigner
```

### `cert list` — get a signer certificate by label

```bash
./ibmverify cert list --label DemoTokenSigner
```

### `cert delete` — delete a signer certificate

```bash
./ibmverify cert delete --label DemoTokenSigner
```

## Global flags

| Flag | Description |
|---|---|
| `-o`, `--output` | Output format: `text` (default), `json`, `yaml` |
| `--debug` | Print verbose HTTP details to stderr (secrets redacted) |
| `--version` | Print the CLI version |

```bash
# Get token as JSON — pipe into jq
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

Two API clients are required in your IBM Verify tenant:

| Client | Required entitlement | Used for |
|---|---|---|
| STS client | Token exchange | Exchanging a JWT for an access token, introspection |
| Cert-manager client | `manageCerts` | Uploading and deleting signer certificates |

## Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Vet
go vet ./...

# Build with version
go build -ldflags "-X main.version=v1.2.0" -o ibmverify .
```

## Requirements

- Go 1.23+
