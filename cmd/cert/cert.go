// Package cert contains the "ibmverify cert" subcommands.
package cert

import (
	"github.com/spf13/cobra"
)

// CertCmd is the "ibmverify cert" parent command.
// Subcommands (upload, list, delete) register themselves in their own init().
var CertCmd = &cobra.Command{
	Use:   "cert",
	Short: "Manage IBM Verify signer certificates",
}
