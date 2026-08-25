#!/bin/bash
# demo3.sh — showcase of apps, users, and API clients (DCR) in ibmverify CLI.
#
# This demonstrates the three new domains added in v1.6.x that were not
# covered in demo2.sh. It shows the full lifecycle for each domain:
#   create → list → get → (use) → delete
#
# Usage:
#   cp demo.env.example .env   # fill in your credentials
#   chmod +x demo3.sh
#   ./demo3.sh
#
# Additional .env vars needed beyond demo2.sh:
#   VERIFY_APP_CLIENT_ID       — API client with manageAppAccessAdmin entitlement
#   VERIFY_APP_CLIENT_SECRET   — secret for above
#   VERIFY_USER_CLIENT_ID      — API client with manageUserGroups entitlement
#   VERIFY_USER_CLIENT_SECRET  — secret for above
#   VERIFY_DCR_CLIENT_ID       — API client with manageAPIClients entitlement
#   VERIFY_DCR_CLIENT_SECRET   — secret for above
#   VERIFY_APP_TEMPLATE_ID     — template ID for app creation (get from app list)

set -e

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
: "${VERIFY_TENANT:?            Missing VERIFY_TENANT in .env}"
: "${VERIFY_APP_CLIENT_ID:?     Missing VERIFY_APP_CLIENT_ID in .env}"
: "${VERIFY_APP_CLIENT_SECRET:? Missing VERIFY_APP_CLIENT_SECRET in .env}"
: "${VERIFY_USER_CLIENT_ID:?    Missing VERIFY_USER_CLIENT_ID in .env}"
: "${VERIFY_USER_CLIENT_SECRET:? Missing VERIFY_USER_CLIENT_SECRET in .env}"
: "${VERIFY_DCR_CLIENT_ID:?     Missing VERIFY_DCR_CLIENT_ID in .env}"
: "${VERIFY_DCR_CLIENT_SECRET:? Missing VERIFY_DCR_CLIENT_SECRET in .env}"
: "${VERIFY_APP_TEMPLATE_ID:?   Missing VERIFY_APP_TEMPLATE_ID in .env}"

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
echo -e "${BOLD}  IBM Verify CLI — apps / users / API clients     ${RESET}"
echo -e "${BOLD}  $VERSION                                        ${RESET}"
echo -e "${BOLD}  Tenant: $VERIFY_TENANT                         ${RESET}"
echo -e "${BOLD}══════════════════════════════════════════════════${RESET}"

# ══════════════════════════════════════════════════════════════════════════════
header "1 / 9  — app list  (browse existing applications)"
# ══════════════════════════════════════════════════════════════════════════════
step "List all applications on the tenant — text table"
$BIN app list \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_APP_CLIENT_ID" \
  --client-secret "$VERIFY_APP_CLIENT_SECRET"

echo ""
note "Same output as JSON — pipe into jq for scripting"
$BIN app list \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_APP_CLIENT_ID" \
  --client-secret "$VERIFY_APP_CLIENT_SECRET" \
  -o json | head -20
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "2 / 9  — app create"
# ══════════════════════════════════════════════════════════════════════════════
step "Create a demo application from template '$VERIFY_APP_TEMPLATE_ID'"
APP_ID=$($BIN app create \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_APP_CLIENT_ID" \
  --client-secret "$VERIFY_APP_CLIENT_SECRET" \
  --name          "ibmverify-demo-app" \
  --template-id   "$VERIFY_APP_TEMPLATE_ID" \
  -o json | python3 -c "
import sys, json, re
d = json.load(sys.stdin)
# IBM returns id in _links.self.href: /appaccess/v1.0/applications/<id>
href = d.get('_links',{}).get('self',{}).get('href','')
m = re.search(r'/applications/(\d+)', href)
print(m.group(1) if m else d.get('id',''))
" 2>/dev/null || echo "")

if [ -n "$APP_ID" ]; then
  ok "Application created — ID: $APP_ID"
else
  note "App created (ID not returned by this template — check app list)"
  APP_ID=""
fi
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "3 / 9  — app get"
# ══════════════════════════════════════════════════════════════════════════════
if [ -n "$APP_ID" ]; then
  step "Fetch the app we just created by ID"
  $BIN app get \
    --tenant        "$VERIFY_TENANT" \
    --client-id     "$VERIFY_APP_CLIENT_ID" \
    --client-secret "$VERIFY_APP_CLIENT_SECRET" \
    --id            "$APP_ID"
else
  note "Skipping app get — no ID available from create response"
fi
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "4 / 9  — app delete  (clean up demo app)"
# ══════════════════════════════════════════════════════════════════════════════
if [ -n "$APP_ID" ]; then
  step "Delete the demo application"
  $BIN app delete \
    --tenant        "$VERIFY_TENANT" \
    --client-id     "$VERIFY_APP_CLIENT_ID" \
    --client-secret "$VERIFY_APP_CLIENT_SECRET" \
    --id            "$APP_ID" \
    && ok "Application deleted"
else
  note "Skipping app delete — no ID available"
fi
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "5 / 9  — user list  (with SCIM filter)"
# ══════════════════════════════════════════════════════════════════════════════
step "List all users"
$BIN user list \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_USER_CLIENT_ID" \
  --client-secret "$VERIFY_USER_CLIENT_SECRET"
echo ""

step "Filter to a specific user by userName (SCIM filter)"
$BIN user list \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_USER_CLIENT_ID" \
  --client-secret "$VERIFY_USER_CLIENT_SECRET" \
  --filter        "userName eq \"${VERIFY_SUBJECT:-bretton}\""
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "6 / 9  — user create"
# ══════════════════════════════════════════════════════════════════════════════
DEMO_USERNAME="ibmverify-demo-user-$$"
step "Create a demo user: $DEMO_USERNAME"
USER_ID=$($BIN user create \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_USER_CLIENT_ID" \
  --client-secret "$VERIFY_USER_CLIENT_SECRET" \
  --username      "$DEMO_USERNAME" \
  --password      "Demo@12345!" \
  --display-name  "Demo User" \
  -o json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('id',''))" 2>/dev/null || echo "")

if [ -n "$USER_ID" ]; then
  ok "User created — ID: $USER_ID"
else
  note "User created (could not parse ID from response)"
  USER_ID=""
fi
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "7 / 9  — user get"
# ══════════════════════════════════════════════════════════════════════════════
if [ -n "$USER_ID" ]; then
  step "Fetch the user we just created"
  $BIN user get \
    --tenant        "$VERIFY_TENANT" \
    --client-id     "$VERIFY_USER_CLIENT_ID" \
    --client-secret "$VERIFY_USER_CLIENT_SECRET" \
    --id            "$USER_ID"

  step "Same user as JSON"
  $BIN user get \
    --tenant        "$VERIFY_TENANT" \
    --client-id     "$VERIFY_USER_CLIENT_ID" \
    --client-secret "$VERIFY_USER_CLIENT_SECRET" \
    --id            "$USER_ID" \
    -o json
else
  note "Skipping user get — no ID available"
fi
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "8 / 9  — user delete  (clean up demo user)"
# ══════════════════════════════════════════════════════════════════════════════
if [ -n "$USER_ID" ]; then
  step "Delete the demo user"
  $BIN user delete \
    --tenant        "$VERIFY_TENANT" \
    --client-id     "$VERIFY_USER_CLIENT_ID" \
    --client-secret "$VERIFY_USER_CLIENT_SECRET" \
    --id            "$USER_ID" \
    && ok "User deleted"
else
  note "Skipping user delete — no ID available"
fi
echo ""

# ══════════════════════════════════════════════════════════════════════════════
header "9 / 9  — apiclient create / list / get / delete  (DCR)"
# ══════════════════════════════════════════════════════════════════════════════
step "List existing API clients"
$BIN apiclient list \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_DCR_CLIENT_ID" \
  --client-secret "$VERIFY_DCR_CLIENT_SECRET"
echo ""

step "Create a new API client with manageUserGroups entitlement"
note "clientSecret is returned ONCE here — capture it immediately"
DCR_RESPONSE=$($BIN apiclient create \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_DCR_CLIENT_ID" \
  --client-secret "$VERIFY_DCR_CLIENT_SECRET" \
  --name          "ibmverify-demo-client-$$" \
  --entitlements  "manageUserGroups" \
  --enabled \
  -o json)

echo "$DCR_RESPONSE"
NEW_CLIENT_ID=$(echo "$DCR_RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('clientId',''))" 2>/dev/null || echo "")
NEW_CLIENT_SECRET=$(echo "$DCR_RESPONSE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('clientSecret',''))" 2>/dev/null || echo "")

if [ -n "$NEW_CLIENT_ID" ]; then
  ok "API client created — ID: $NEW_CLIENT_ID"
  note "clientSecret (only shown once): ${NEW_CLIENT_SECRET:0:8}..."
else
  note "Could not parse clientId from response"
fi
echo ""

if [ -n "$NEW_CLIENT_ID" ]; then
  step "Get the new API client by ID"
  $BIN apiclient get \
    --tenant        "$VERIFY_TENANT" \
    --client-id     "$VERIFY_DCR_CLIENT_ID" \
    --client-secret "$VERIFY_DCR_CLIENT_SECRET" \
    --id            "$NEW_CLIENT_ID"
  echo ""

  step "Delete the demo API client"
  $BIN apiclient delete \
    --tenant        "$VERIFY_TENANT" \
    --client-id     "$VERIFY_DCR_CLIENT_ID" \
    --client-secret "$VERIFY_DCR_CLIENT_SECRET" \
    --id            "$NEW_CLIENT_ID" \
    && ok "API client deleted"
fi
echo ""

# ── Final summary ──────────────────────────────────────────────────────────────
echo -e "${BOLD}══════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  demo3.sh complete — domains exercised:          ${RESET}"
echo -e "${BOLD}  ✓ app  list / create / get / delete             ${RESET}"
echo -e "${BOLD}  ✓ user list / create / get / delete             ${RESET}"
echo -e "${BOLD}  ✓ apiclient list / create / get / delete (DCR)  ${RESET}"
echo -e "${BOLD}══════════════════════════════════════════════════${RESET}"
echo ""
ok "All 9 sections complete"
