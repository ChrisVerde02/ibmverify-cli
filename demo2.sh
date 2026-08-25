#!/bin/bash
# demo2.sh — full walkthrough of every ibmverify command.
#
# Usage:
#   cp demo.env.example .env   # fill in your credentials
#   chmod +x demo2.sh
#   ./demo2.sh
#
# Commands covered:
#   config init / config set
#   cert delete (pre-clean)
#   cert upload
#   cert get
#   token get  (with --jwt-expires-in, --validity-days, --key-size)
#   token introspect
#   run  (default expiration)
#   run  (custom expiration flags — shows all new options)

set -e  # stop on first error

# ── Colours ────────────────────────────────────────────────────────────────────
BOLD='\033[1m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
RESET='\033[0m'

header() { echo -e "\n${BOLD}${CYAN}━━  $1  ━━${RESET}"; }
step()   { echo -e "${BOLD}▶  $1${RESET}"; }
ok()     { echo -e "${GREEN}✓  $1${RESET}"; }
note()   { echo -e "${YELLOW}ℹ  $1${RESET}"; }

# ── Load credentials from .env ─────────────────────────────────────────────────
if [ ! -f .env ]; then
  echo "Error: .env file not found."
  echo "Run: cp demo.env.example .env  then fill in your credentials."
  exit 1
fi
# shellcheck source=.env
source .env

# ── Validate required vars ─────────────────────────────────────────────────────
: "${VERIFY_TENANT:?          Missing VERIFY_TENANT in .env}"
: "${VERIFY_STS_CLIENT_ID:?   Missing VERIFY_STS_CLIENT_ID in .env}"
: "${VERIFY_STS_CLIENT_SECRET:? Missing VERIFY_STS_CLIENT_SECRET in .env}"
: "${VERIFY_CERT_CLIENT_ID:?  Missing VERIFY_CERT_CLIENT_ID in .env}"
: "${VERIFY_CERT_CLIENT_SECRET:? Missing VERIFY_CERT_CLIENT_SECRET in .env}"
: "${VERIFY_SUBJECT:?         Missing VERIFY_SUBJECT in .env}"
: "${VERIFY_ISSUER:?          Missing VERIFY_ISSUER in .env}"
: "${VERIFY_LABEL:?           Missing VERIFY_LABEL in .env}"

# ── Optional expiration / cert overrides (have safe defaults) ──────────────────
JWT_EXPIRES_IN="${VERIFY_JWT_EXPIRES_IN:-15m}"
VALIDITY_DAYS="${VERIFY_VALIDITY_DAYS:-365}"
KEY_SIZE="${VERIFY_KEY_SIZE:-4096}"
SUBJECT_TOKEN_TYPE="${VERIFY_SUBJECT_TOKEN_TYPE:-urn:demo:token-type:user-jwt}"

# ── Build binary if not present ────────────────────────────────────────────────
BIN="./ibmverify"
if [ ! -f "$BIN" ]; then
  note "Binary not found — building now..."
  go build -ldflags "-X main.version=demo" -o ibmverify ./cmd/ibmverify
  ok "Built ./ibmverify"
fi

VERSION=$($BIN --version 2>&1 | head -1)

echo ""
echo -e "${BOLD}══════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  IBM Verify CLI — full command demo              ${RESET}"
echo -e "${BOLD}  $VERSION                                        ${RESET}"
echo -e "${BOLD}  Tenant: $VERIFY_TENANT                         ${RESET}"
echo -e "${BOLD}══════════════════════════════════════════════════${RESET}"

# ══════════════════════════════════════════════════════════════════════════════
header "1 / 9  — config init"
# ══════════════════════════════════════════════════════════════════════════════
step "Create ~/.ibmverify/config.yaml (skips if already exists)"
$BIN config init
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "2 / 9  — config set"
# ══════════════════════════════════════════════════════════════════════════════
step "Write tenant, subject, issuer, label into the config file"
$BIN config set tenant  "$VERIFY_TENANT"
$BIN config set subject "$VERIFY_SUBJECT"
$BIN config set issuer  "$VERIFY_ISSUER"
$BIN config set label   "$VERIFY_LABEL"
ok "Config keys written — picked up automatically by all commands"
note "Precedence: flag > env var > config file > default"
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "3 / 9  — cert delete  (pre-clean)"
# ══════════════════════════════════════════════════════════════════════════════
step "Remove any existing signer cert with label '$VERIFY_LABEL'"
$BIN cert delete \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --label         "$VERIFY_LABEL" \
  && ok "Deleted existing cert" \
  || note "No existing cert found — skipping"
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "4 / 9  — cert upload"
# ══════════════════════════════════════════════════════════════════════════════
step "Generate a temporary self-signed cert and upload it to IBM Verify"
CERT_FILE=$(mktemp /tmp/ibmverify-demo-XXXX.pem)
KEY_FILE=$(mktemp /tmp/ibmverify-demo-XXXX.key)
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout "$KEY_FILE" \
  -out    "$CERT_FILE" \
  -days 1 \
  -subj "/CN=$VERIFY_LABEL/O=Demo/C=US" \
  2>/dev/null
ok "Generated temporary cert: $CERT_FILE"

$BIN cert upload \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --cert-file     "$CERT_FILE" \
  --label         "$VERIFY_LABEL"

rm -f "$CERT_FILE" "$KEY_FILE"
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "5 / 9  — cert get"
# ══════════════════════════════════════════════════════════════════════════════
step "Fetch the signer cert we just uploaded and display its metadata"
$BIN cert get \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --label         "$VERIFY_LABEL"
echo ""

# Clean up the uploaded cert so the full flow below can re-upload cleanly
$BIN cert delete \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --label         "$VERIFY_LABEL" >/dev/null 2>&1 || true

# ══════════════════════════════════════════════════════════════════════════════
header "6 / 9  — token get"
# ══════════════════════════════════════════════════════════════════════════════
step "Full cert→JWT→exchange flow — capture access token on stdout"
note "Using: --jwt-expires-in $JWT_EXPIRES_IN  --validity-days $VALIDITY_DAYS  --key-size $KEY_SIZE"
TOKEN=$($BIN token get \
  --tenant                     "$VERIFY_TENANT" \
  --sts-client-id              "$VERIFY_STS_CLIENT_ID" \
  --sts-client-secret          "$VERIFY_STS_CLIENT_SECRET" \
  --cert-manager-client-id     "$VERIFY_CERT_CLIENT_ID" \
  --cert-manager-client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --subject                    "$VERIFY_SUBJECT" \
  --issuer                     "$VERIFY_ISSUER" \
  --label                      "$VERIFY_LABEL" \
  --jwt-expires-in             "$JWT_EXPIRES_IN" \
  --validity-days              "$VALIDITY_DAYS" \
  --key-size                   "$KEY_SIZE" \
  --subject-token-type         "$SUBJECT_TOKEN_TYPE")

ok "Token captured — first 60 chars: ${TOKEN:0:60}..."
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "7 / 9  — token introspect"
# ══════════════════════════════════════════════════════════════════════════════
step "Inspect the token we just received"
$BIN token introspect \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_STS_CLIENT_ID" \
  --client-secret "$VERIFY_STS_CLIENT_SECRET" \
  --token         "$TOKEN"
echo ""

# Clean up before all-in-one run — must delete so run doesn't
# hit HTTP 400 "label already exists" from the token get step above
$BIN cert delete \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --label         "$VERIFY_LABEL" >/dev/null 2>&1 || true
sleep 1

# ══════════════════════════════════════════════════════════════════════════════
header "8 / 9  — run  (default expiration)"
# ══════════════════════════════════════════════════════════════════════════════
step "ibmverify run — generates cert, uploads, signs JWT, exchanges, introspects"
note "Progress goes to stderr — stdout is the clean token (TOKEN=\$(ibmverify run ...) works)"
FULL_TOKEN=$($BIN run \
  --tenant                     "$VERIFY_TENANT" \
  --sts-client-id              "$VERIFY_STS_CLIENT_ID" \
  --sts-client-secret          "$VERIFY_STS_CLIENT_SECRET" \
  --cert-manager-client-id     "$VERIFY_CERT_CLIENT_ID" \
  --cert-manager-client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --subject "$VERIFY_SUBJECT" \
  --issuer  "$VERIFY_ISSUER" \
  --label   "$VERIFY_LABEL")

ok "Token captured cleanly from stdout"
echo ""

# Clean up before custom-expiration run
$BIN cert delete \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --label         "$VERIFY_LABEL" >/dev/null 2>&1 || true
sleep 1

# ══════════════════════════════════════════════════════════════════════════════
header "9 / 9  — run  (custom expiration + cert options)"
# ══════════════════════════════════════════════════════════════════════════════
step "Same flow with explicit --jwt-expires-in, --validity-days, --key-size overrides"
note "--jwt-expires-in  controls how long the signed JWT lives before exchange"
note "--validity-days   controls the cert's X.509 validity window"
note "--key-size        RSA key strength: 2048 (fast) | 3072 | 4096 (default)"
note "--subject-token-type  URN sent in the token-exchange grant"
CUSTOM_TOKEN=$($BIN run \
  --tenant                     "$VERIFY_TENANT" \
  --sts-client-id              "$VERIFY_STS_CLIENT_ID" \
  --sts-client-secret          "$VERIFY_STS_CLIENT_SECRET" \
  --cert-manager-client-id     "$VERIFY_CERT_CLIENT_ID" \
  --cert-manager-client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --subject            "$VERIFY_SUBJECT" \
  --issuer             "$VERIFY_ISSUER" \
  --label              "$VERIFY_LABEL" \
  --jwt-expires-in     "$JWT_EXPIRES_IN" \
  --validity-days      "$VALIDITY_DAYS" \
  --key-size           "$KEY_SIZE" \
  --subject-token-type "$SUBJECT_TOKEN_TYPE")

ok "Custom-expiration token captured — first 60 chars: ${CUSTOM_TOKEN:0:60}..."
echo ""

# ── Final output ───────────────────────────────────────────────────────────────
echo -e "${BOLD}══════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  Full access token (use in curl / Postman):      ${RESET}"
echo -e "${BOLD}══════════════════════════════════════════════════${RESET}"
echo ""
echo "$CUSTOM_TOKEN"
echo ""
echo -e "${YELLOW}Example curl:${RESET}"
echo "  curl -s -H \"Authorization: Bearer \$CUSTOM_TOKEN\" \\"
echo "    $VERIFY_TENANT/v2.0/me"
echo ""
ok "Demo complete — all 9 sections exercised"
