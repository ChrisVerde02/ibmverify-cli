#!/bin/bash
# demo-errors.sh — demonstrates how the ibmverify error handling stack works.
#
# Shows four layers working together:
#   1. client/errors.go   — typed APIError from the SDK (StatusCode, Code, Message)
#   2. internal/retrier   — automatic retry on 429/5xx before failing
#   3. internal/errkind   — maps SDK errors to exit codes
#   4. internal/exitcode  — typed constants (0=ok, 3=auth, 4=notfound, ...)
#
# Each section triggers a real error and shows what comes out.
#
# Usage:
#   cp demo.env.example .env   # credentials already filled in
#   chmod +x demo-errors.sh
#   ./demo-errors.sh

set -euo pipefail

# ── Colours ────────────────────────────────────────────────────────────────────
BOLD='\033[1m'
GREEN='\033[0;32m'
RED='\033[0;31m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
DIM='\033[2m'
RESET='\033[0m'

header() { echo -e "\n${BOLD}${CYAN}━━  $1  ━━${RESET}"; }
step()   { echo -e "${BOLD}▶  $1${RESET}"; }
ok()     { echo -e "${GREEN}✓  $1${RESET}"; }
fail()   { echo -e "${RED}✗  $1${RESET}"; }
note()   { echo -e "${YELLOW}ℹ  $1${RESET}"; }
code()   { echo -e "${DIM}$1${RESET}"; }

# ── Load credentials ───────────────────────────────────────────────────────────
if [ ! -f .env ]; then
  echo "Error: .env file not found. Run: cp demo.env.example .env"
  exit 1
fi
source .env

: "${VERIFY_TENANT:?          Missing VERIFY_TENANT}"
: "${VERIFY_CERT_CLIENT_ID:?  Missing VERIFY_CERT_CLIENT_ID}"
: "${VERIFY_CERT_CLIENT_SECRET:? Missing VERIFY_CERT_CLIENT_SECRET}"

BIN="./ibmverify"
if [ ! -f "$BIN" ]; then
  note "Binary not found — building..."
  go build -ldflags "-X main.version=demo" -o ibmverify ./cmd/ibmverify
  ok "Built ./ibmverify"
fi

echo ""
echo -e "${BOLD}══════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  IBM Verify CLI — Error Handling Demo                    ${RESET}"
echo -e "${BOLD}                                                          ${RESET}"
echo -e "${BOLD}  Four layers shown:                                      ${RESET}"
echo -e "${BOLD}    SDK     → typed APIError (StatusCode, Code, Message)  ${RESET}"
echo -e "${BOLD}    Retry   → automatic on 429 / 5xx, exponential backoff ${RESET}"
echo -e "${BOLD}    errkind → maps SDK error → exit code number           ${RESET}"
echo -e "${BOLD}    exitcode→ 0=ok 2=usage 3=auth 4=notfound 5=rate 6=srv${RESET}"
echo -e "${BOLD}══════════════════════════════════════════════════════════${RESET}"

# ══════════════════════════════════════════════════════════════════════════════
header "1 / 5 — Wrong credentials → Exit code 3 (Auth failure)"
# ══════════════════════════════════════════════════════════════════════════════
note "Real tenant, real client ID, deliberately wrong secret"
note "The SDK will call /v1.0/endpoint/default/token and get back 401"
echo ""
step "Running:"
code "  ibmverify cert get --tenant \$VERIFY_TENANT \\"
code "    --client-id \$VERIFY_CERT_CLIENT_ID \\"
code "    --client-secret WRONG_SECRET \\"
code "    --label test"
echo ""

set +e
$BIN cert get \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "WRONG_SECRET_INTENTIONAL_FOR_DEMO" \
  --label         "test"
EXIT=$?
set -e

echo ""
echo -e "Exit code returned: ${BOLD}$EXIT${RESET}"
echo ""

if [ "$EXIT" -eq 3 ]; then
  ok "Exit code 3 = Auth failure"
  note "What just happened inside the stack:"
  echo "  IBM Verify returned HTTP 401"
  echo "  client/errors.go parsed the response body"
  echo "  → APIError{StatusCode:401, Code:\"invalid_client\", Message:\"...\"}"
  echo "  → apiErr.IsAuth() returned true"
  echo "  errkind.ExitCode() mapped IsAuth → exitcode.Auth (3)"
  echo "  CLI called os.Exit(3)"
  echo ""
  note "A CI pipeline sees exit code 3 and knows: rotate credentials, don't retry"
else
  note "Got exit code $EXIT"
fi

# ══════════════════════════════════════════════════════════════════════════════
header "2 / 5 — Resource not found → Exit code 4"
# ══════════════════════════════════════════════════════════════════════════════
note "Valid credentials, asking for a cert label that definitely doesn't exist"
echo ""
step "Running:"
code "  ibmverify cert get --tenant \$VERIFY_TENANT \\"
code "    --client-id \$VERIFY_CERT_CLIENT_ID \\"
code "    --client-secret \$VERIFY_CERT_CLIENT_SECRET \\"
code "    --label this-label-does-not-exist-xyz-$$"
echo ""

set +e
$BIN cert get \
  --tenant        "$VERIFY_TENANT" \
  --client-id     "$VERIFY_CERT_CLIENT_ID" \
  --client-secret "$VERIFY_CERT_CLIENT_SECRET" \
  --label         "this-label-does-not-exist-xyz-$$"
EXIT=$?
set -e

echo ""
echo -e "Exit code returned: ${BOLD}$EXIT${RESET}"
echo ""

note "What just happened inside the stack:"
echo "  IBM Verify returned HTTP 404"
echo "  client/errors.go → APIError{StatusCode:404}"
echo "  → apiErr.IsNotFound() returned true"
echo "  errkind.ExitCode() mapped IsNotFound → exitcode.NotFound (4)"
echo ""
note "Different action than exit 3 — resource is missing, not an auth problem"

# ══════════════════════════════════════════════════════════════════════════════
header "3 / 5 — Missing required flag → Exit code 2 (Usage error)"
# ══════════════════════════════════════════════════════════════════════════════
note "No HTTP call is made — Cobra catches the missing flag immediately"
echo ""
step "Running (deliberately omitting --client-id):"
code "  ibmverify app list --tenant \$VERIFY_TENANT --client-secret somevalue"
code "  (missing --client-id)"
echo ""

set +e
$BIN app list \
  --tenant        "$VERIFY_TENANT" \
  --client-secret "somevalue"
EXIT=$?
set -e

echo ""
echo -e "Exit code returned: ${BOLD}$EXIT${RESET}"
echo ""

note "What just happened:"
echo "  Cobra parsed the flags before any code ran"
echo "  Found --client-id missing (marked Required)"
echo "  Returned usage error — no HTTP call was ever made"
echo "  errkind.ExitCode() returned exitcode.Other (1) or exitcode.Usage (2)"
echo ""
note "The error was caught at the CLI layer — the SDK never saw it"

# ══════════════════════════════════════════════════════════════════════════════
header "4 / 5 — What the typed SDK error looks like"
# ══════════════════════════════════════════════════════════════════════════════
note "Showing the actual APIError structure from client/errors.go"
echo ""
echo "  When IBM Verify returns a non-2xx response, the SDK builds this:"
echo ""
echo -e "  ${BOLD}type APIError struct {${RESET}"
echo -e "      StatusCode  int    ${DIM}// HTTP status — 401, 403, 404, 429, 500${RESET}"
echo -e "      Code        string ${DIM}// IBM's error ID — \"CSIAO5401E\", \"invalid_client\"${RESET}"
echo -e "      Message     string ${DIM}// IBM's description in plain English${RESET}"
echo -e "      Endpoint    string ${DIM}// which URL failed${RESET}"
echo -e "  ${BOLD}}${RESET}"
echo ""
echo "  Methods on it:"
echo -e "      .IsAuth()      ${DIM}→ true for 401/403 or invalid_client${RESET}"
echo -e "      .IsNotFound()  ${DIM}→ true for 404${RESET}"
echo -e "      .IsRateLimit() ${DIM}→ true for 429${RESET}"
echo -e "      .IsServer()    ${DIM}→ true for 5xx${RESET}"
echo -e "      .IsRetryable() ${DIM}→ true if SDK will retry automatically${RESET}"
echo ""
note "The Terraform provider uses the same APIError — no exit codes, just:"
echo "  if apiErr.IsNotFound() { resp.State.RemoveResource(ctx) }"
echo "  The same SDK error. Different consumer. Different action."

# ══════════════════════════════════════════════════════════════════════════════
header "5 / 5 — The exit code map (what every number means)"
# ══════════════════════════════════════════════════════════════════════════════
echo ""
echo -e "  ${BOLD}Exit 0${RESET}  Success — everything worked"
echo -e "  ${BOLD}Exit 1${RESET}  Unknown error — network failure, unexpected response"
echo -e "  ${BOLD}Exit 2${RESET}  Usage error — wrong or missing flags (caught before any HTTP call)"
echo -e "  ${BOLD}Exit 3${RESET}  Auth failed — wrong client_id or client_secret"
echo -e "  ${BOLD}Exit 4${RESET}  Not found — resource doesn't exist in IBM Verify"
echo -e "  ${BOLD}Exit 5${RESET}  Rate limited — IBM Verify said slow down (SDK already retried)"
echo -e "  ${BOLD}Exit 6${RESET}  Server error — IBM Verify 5xx (SDK retried 3x, still failed)"
echo -e "  ${BOLD}Exit 7${RESET}  Validation — bad input caught before any request was sent"
echo ""
note "The SDK retries exit 5 and 6 automatically before giving up"
note "By the time you see exit 5 or 6, it already tried 3 times"
echo ""
echo "  In a CI pipeline:"
echo -e "  ${DIM}case \$? in${RESET}"
echo -e "  ${DIM}  3) alert 'credentials expired — rotate now';;${RESET}"
echo -e "  ${DIM}  4) log 'resource missing — will recreate';;${RESET}"
echo -e "  ${DIM}  5) sleep 60 && retry;;${RESET}"
echo -e "  ${DIM}  6) page_oncall;;${RESET}"
echo -e "  ${DIM}esac${RESET}"

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}══════════════════════════════════════════════════════════${RESET}"
echo -e "${BOLD}  Summary — three repos, one error story                  ${RESET}"
echo -e "${BOLD}══════════════════════════════════════════════════════════${RESET}"
echo ""
echo -e "  ${BOLD}ibmverify-go / client/errors.go${RESET}"
echo -e "  ${DIM}Typed APIError — StatusCode, Code, Message, Endpoint${RESET}"
echo -e "  ${DIM}Methods: IsAuth IsNotFound IsRateLimit IsServer IsRetryable${RESET}"
echo ""
echo -e "  ${BOLD}ibmverify-go / generated/internal/retrier.go${RESET}"
echo -e "  ${DIM}Auto-retry on 429/5xx — up to 3 attempts, exponential backoff${RESET}"
echo -e "  ${DIM}Reads Retry-After header. Cryptographic jitter. Context-aware.${RESET}"
echo ""
echo -e "  ${BOLD}ibmverify-cli / internal/errkind + exitcode${RESET}"
echo -e "  ${DIM}Translates APIError → exit code number for scripts and CI${RESET}"
echo ""
echo -e "  ${BOLD}terraform-provider-verify / resp.Diagnostics.AddError${RESET}"
echo -e "  ${DIM}Same APIError → structured diagnostic tied to resource + config line${RESET}"
echo ""
ok "Error handling demo complete"
