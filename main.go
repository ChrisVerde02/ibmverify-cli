package main

import (
	"github.com/ChrisVerde02/ibmverify-cli/cmd"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/cert"
	"github.com/ChrisVerde02/ibmverify-cli/cmd/token"
)

func main() {
	// Attach the cert and token parent commands to the root.
	// Their subcommands (get, list, delete, introspect) are registered
	// automatically inside each file's init().
	cmd.Root().AddCommand(cert.CertCmd)
	cmd.Root().AddCommand(token.TokenCmd)

	cmd.Execute()
}
