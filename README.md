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

## IBM Verify setup

Two API clients are required in your IBM Verify tenant:

| Client | Required entitlement | Used for |
|---|---|---|
| STS client | Token exchange | Exchanging a JWT for an access token, introspection |
| Cert-manager client | `manageCerts` | Uploading signer certificates |

## Requirements

- Go 1.21+
