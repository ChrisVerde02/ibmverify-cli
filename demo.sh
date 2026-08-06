#!/bin/bash
# demo.sh — runs the ibmverify CLI step by step using your real credentials.
# Usage: ./demo.sh

set -e  # stop on first error

# ── Credentials ────────────────────────────────────────────────────────────────
TENANT="https://ChrisVerde02.verify.ibm.com"

STS_CLIENT_ID="9ada0c66-b6e9-4fdb-a6c8-62156f5ae533"
STS_CLIENT_SECRET="TvFt1rBA9gMBHYM8W5Rq"

CERT_MANAGER_CLIENT_ID="8c1e34de-c3ca-4dcb-82f4-5d43ecdcd955"
CERT_MANAGER_CLIENT_SECRET="w6EzRUfpsi5ybCU6IawF"

SUBJECT="bretton"
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
