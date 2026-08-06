// Package token contains the "ibmverify token" subcommands.
package token

import (
	"github.com/spf13/cobra"
)

// TokenCmd is the "ibmverify token" parent command.
// Subcommands (get, introspect) register themselves in their own init().
var TokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage IBM Verify access tokens",
}
