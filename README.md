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

## Commands

### `run` — full flow in one command

Generates a certificate, uploads it, signs a JWT, exchanges it for an access token, and introspects it.

```bash
ibmverify run \
  --tenant               https://example.verify.ibm.com \
  --sts-client-id        <sts-client-id> \
  --sts-client-secret    <sts-client-secret> \
  --cert-manager-client-id     <cert-manager-client-id> \
  --cert-manager-client-secret <cert-manager-client-secret> \
  --subject   myusername \
  --issuer    https://demo.ibm.com \
  --label     DemoTokenSigner
```

### `token get` — JWT token exchange

Same flow as `run` but prints only the raw access token to stdout (pipeable).

```bash
TOKEN=$(ibmverify token get \
  --tenant               https://example.verify.ibm.com \
  --sts-client-id        <sts-client-id> \
  --sts-client-secret    <sts-client-secret> \
  --cert-manager-client-id     <cert-manager-client-id> \
  --cert-manager-client-secret <cert-manager-client-secret> \
  --subject   myusername \
  --issuer    https://demo.ibm.com \
  --label     DemoTokenSigner)
```

Optional flags: `--organization` (default `IBM`), `--country` (default `US`), `--validity-days` (default `365`), `--key-size` (default `4096`), `--subject-token-type`.

### `token introspect` — introspect an access token

```bash
ibmverify token introspect \
  --tenant        https://example.verify.ibm.com \
  --client-id     <client-id> \
  --client-secret <client-secret> \
  --token         <access-token>
```

### `cert upload` — upload a signer certificate

```bash
ibmverify cert upload \
  --tenant        https://example.verify.ibm.com \
  --client-id     <cert-manager-client-id> \
  --client-secret <cert-manager-client-secret> \
  --cert-file     ./cert.pem \
  --label         DemoTokenSigner
```

### `cert list` — get a signer certificate by label

```bash
ibmverify cert list \
  --tenant        https://example.verify.ibm.com \
  --client-id     <cert-manager-client-id> \
  --client-secret <cert-manager-client-secret> \
  --label         DemoTokenSigner
```

### `cert delete` — delete a signer certificate

```bash
ibmverify cert delete \
  --tenant        https://example.verify.ibm.com \
  --client-id     <cert-manager-client-id> \
  --client-secret <cert-manager-client-secret> \
  --label         DemoTokenSigner
```

## Demo script

[`demo.sh`](demo.sh) runs a full end-to-end walkthrough against a real tenant using four chained CLI calls:

| Step | Command | What it does |
|---|---|---|
| 1 | `cert delete` | Cleans up any existing cert with the same label (safe to skip if none exists) |
| 2 | `token get` | Runs the full cert → upload → JWT → exchange flow; captures the token |
| 3 | `token introspect` | Validates the token and prints subject, username, scope, and expiry |
| 4 | *(print)* | Echoes the full token for use in Postman or `curl` |

**Run it:**

```bash
# 1. Build the binary first
go build -o ibmverify .

# 2. Edit the credentials at the top of the script, then:
chmod +x demo.sh
./demo.sh
```

The script uses `set -e`, so it stops immediately if any step fails. The `cert delete` step at the start is intentionally non-fatal — it prints `(none found — skipping)` if no cert exists yet.

## IBM Verify setup

Two API clients are required in your IBM Verify tenant:

| Client | Required entitlement | Used for |
|---|---|---|
| STS client | Token exchange | Exchanging a JWT for an access token, introspection |
| Cert-manager client | `manageCerts` | Uploading signer certificates |

## Requirements

- Go 1.21+

## Copy-paste demo script

Fill in your credentials and save this as `demo.sh`, then run `chmod +x demo.sh && ./demo.sh`:

```bash
#!/bin/bash
# demo.sh — runs the ibmverify CLI step by step using your real credentials.
# Usage: ./demo.sh

set -e  # stop on first error

# ── Credentials ────────────────────────────────────────────────────────────────
TENANT=""               # e.g. https://example.verify.ibm.com

STS_CLIENT_ID=""
STS_CLIENT_SECRET=""

CERT_MANAGER_CLIENT_ID=""
CERT_MANAGER_CLIENT_SECRET=""

SUBJECT=""              # IBM Verify username
ISSUER="https://demo.ibm.com"
LABEL="demotokensigner"

# ── Binary path ────────────────────────────────────────────────────────────────
# Assumes you built it with: go build -o ibmverify .
BIN="./ibmverify"

# ──────────────────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════"
echo " IBM Verify CLI — step-by-step demo"
echo "══════════════════════════════════════════"
echo ""

# ── Step 1: Clean up any existing cert with the same label ────────────────────
echo "Step 1 — Delete existing signer cert (if any)"
$BIN cert delete \
  --tenant "$TENANT" \
  --client-id "$CERT_MANAGER_CLIENT_ID" \
  --client-secret "$CERT_MANAGER_CLIENT_SECRET" \
  --label "$LABEL" \
  && echo "  (deleted)" \
  || echo "  (none found — skipping)"

echo ""

# ── Step 2: Get an access token (gen cert → upload → sign JWT → exchange) ─────
echo "Step 2 — Get access token (full flow)"
TOKEN=$($BIN token get \
  --tenant "$TENANT" \
  --sts-client-id "$STS_CLIENT_ID" \
  --sts-client-secret "$STS_CLIENT_SECRET" \
  --cert-manager-client-id "$CERT_MANAGER_CLIENT_ID" \
  --cert-manager-client-secret "$CERT_MANAGER_CLIENT_SECRET" \
  --subject "$SUBJECT" \
  --issuer "$ISSUER" \
  --label "$LABEL")

echo "  ✓ Token received: ${TOKEN:0:40}..."
echo ""

# ── Step 3: Introspect the token ──────────────────────────────────────────────
echo "Step 3 — Introspect the token"
$BIN token introspect \
  --tenant "$TENANT" \
  --client-id "$STS_CLIENT_ID" \
  --client-secret "$STS_CLIENT_SECRET" \
  --token "$TOKEN"

echo ""

# ── Step 4: Print the full token (for use in scripts / Postman) ───────────────
echo "Step 4 — Full access token (copy into Postman or use in curl)"
echo ""
echo "$TOKEN"
echo ""

# ── Example: how to use the token in a curl call ──────────────────────────────
echo "──────────────────────────────────────────"
echo " Example curl (replace the path as needed)"
echo "──────────────────────────────────────────"
echo "curl -s -H \"Authorization: Bearer \$TOKEN\" \\"
echo "  $TENANT/some-api-endpoint"
echo ""
```
